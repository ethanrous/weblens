package jobs_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/embedding"
	"github.com/ethanrous/weblens/models/featureflags"
	file_model "github.com/ethanrous/weblens/models/file"
	job_model "github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlog"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/embed"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

const testRootAlias = "embedtest"

type jobTestHarness struct {
	appCtx     context_service.AppContext
	workerPool *task.WorkerPool
	tempDir    string
	pool       *task.Pool
}

func newJobTestHarness(t *testing.T) (context.Context, *jobTestHarness) {
	t.Helper()

	logger := wlog.NewZeroLogger()

	// Set up a temporary directory for file operations.
	tempDir, err := os.MkdirTemp("", "embed_job_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	alias := testRootAlias + t.Name()
	err = file_system.RegisterAbsolutePrefix(alias, tempDir)
	require.NoError(t, err)

	// Set up MongoDB with a clean embeddings collection.
	dbCtx := db.SetupTestDB(t, embedding.CollectionKey)

	// Build the AppContext with the DB from dbCtx.
	basicCtx := context_service.NewBasicContext(dbCtx, logger)
	appCtx := context_service.NewAppContext(basicCtx)
	appCtx.DB = dbCtx.Value(db.DatabaseContextKey).(*mongo.Database)

	// Start the worker pool.
	workerPool := task.NewWorkerPool(2)
	jobs.RegisterJobs(workerPool)
	workerPool.Run(appCtx)
	t.Cleanup(func() {
		// Worker pool shares the appCtx; let it go out of scope to cancel.
	})

	globalPool := workerPool.GetTaskPool(task.GlobalTaskPoolID)
	require.NotNil(t, globalPool)

	// Point the embed client at a stub server that returns valid chunk responses.
	stubServer := newEmbedStubServer(t)
	embed.Default().SetBaseURLForTesting(stubServer.URL)
	embed.Default().MarkAvailable()

	// Enable the embed feature flag so the extract-and-embed handler runs.
	require.NoError(t, featureflags.SaveFlags(appCtx, featureflags.Bundle{AllowRegistrations: true, EnableEmbed: true}))

	h := &jobTestHarness{
		appCtx:     appCtx,
		workerPool: workerPool,
		tempDir:    tempDir,
		pool:       globalPool,
	}

	return appCtx, h
}

// newEmbedStubServer starts an httptest server returning one canned 1024-dim chunk.
func newEmbedStubServer(t *testing.T) *httptest.Server {
	t.Helper()

	vec := make([]float64, 1024)
	chunks := []map[string]any{
		{"chunkIndex": 0, "snippet": "test snippet", "vector": vec},
	}
	body, err := json.Marshal(chunks)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/extract-and-embed", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv
}

// CreateTextFile writes content to a temp file and returns a WeblensFileImpl for it.
func (h *jobTestHarness) CreateTextFile(t *testing.T, relPath, content string) *file_model.WeblensFileImpl {
	t.Helper()

	alias := testRootAlias + t.Name()
	fp := file_system.BuildFilePath(alias, relPath)

	f := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       fp,
		FileID:     relPath,
		GenerateID: false,
		CreateNow:  true,
	})
	require.NotNil(t, f)

	n, err := f.Write([]byte(content))
	require.NoError(t, err)
	require.Equal(t, len(content), n)

	return f
}

// DispatchExtractAndEmbed dispatches the extract-and-embed job for the given file.
func (h *jobTestHarness) DispatchExtractAndEmbed(t *testing.T, f *file_model.WeblensFileImpl) {
	t.Helper()

	meta := job_model.ExtractAndEmbedMeta{File: f}
	_, err := h.workerPool.DispatchJob(h.appCtx, job_model.ExtractAndEmbedTask, meta, h.pool)
	require.NoError(t, err)
}

// WaitForJobs blocks until all currently queued tasks complete.
func (h *jobTestHarness) WaitForJobs(t *testing.T) {
	t.Helper()

	var wg sync.WaitGroup

	for _, tsk := range h.workerPool.GetTasks() {
		done, _ := tsk.Status()
		if done {
			continue
		}

		wg.Add(1)

		go func(tsk *task.Task) {
			defer wg.Done()

			tsk.Wait()
		}(tsk)
	}

	wg.Wait()
}

// QueryEmbeddings returns all embedding rows for the given source file ID.
func (h *jobTestHarness) QueryEmbeddings(t *testing.T, fileID string) []embedding.Embedding {
	t.Helper()

	rows, err := embedding.GetForSource(h.appCtx, fileID)
	require.NoError(t, err)

	return rows
}

// StubEmbedOffline marks the embed service as unavailable.
func (h *jobTestHarness) StubEmbedOffline() {
	embed.Default().MarkUnavailable()
}

func TestExtractAndEmbedFile_PersistsChunks(t *testing.T) {
	_, h := newJobTestHarness(t)

	f := h.CreateTextFile(t, "/notes.txt",
		"The quick brown fox jumps over the lazy dog. "+
			"All work and no play makes Jack a dull boy. "+
			"The rain in Spain falls mainly on the plain.")

	h.DispatchExtractAndEmbed(t, f)
	h.WaitForJobs(t)

	rows := h.QueryEmbeddings(t, f.ID())
	assert.NotEmpty(t, rows, "should write at least one chunk")

	for _, r := range rows {
		assert.NotEmpty(t, r.Snippet)
		assert.Equal(t, 1024, len(r.Vector))
	}
}

func TestExtractAndEmbedFile_Idempotent(t *testing.T) {
	_, h := newJobTestHarness(t)

	f := h.CreateTextFile(t, "/notes.txt", "deterministic content")

	h.DispatchExtractAndEmbed(t, f)
	h.WaitForJobs(t)

	countBefore := len(h.QueryEmbeddings(t, f.ID()))

	h.DispatchExtractAndEmbed(t, f)
	h.WaitForJobs(t)

	countAfter := len(h.QueryEmbeddings(t, f.ID()))
	assert.Equal(t, countBefore, countAfter, "second run with unchanged content should be a no-op")
}

func TestExtractAndEmbedFile_ServiceUnavailableIsNoop(t *testing.T) {
	_, h := newJobTestHarness(t)

	h.StubEmbedOffline()

	f := h.CreateTextFile(t, "/notes.txt", "anything")

	h.DispatchExtractAndEmbed(t, f)
	h.WaitForJobs(t)

	rows := h.QueryEmbeddings(t, f.ID())
	assert.Empty(t, rows, "should write nothing when service is unavailable")
}

func TestExtractAndEmbedFile_DisabledFlagIsNoop(t *testing.T) {
	_, h := newJobTestHarness(t)

	require.NoError(t, featureflags.SaveFlags(h.appCtx, featureflags.Bundle{AllowRegistrations: true, EnableEmbed: false}))

	f := h.CreateTextFile(t, "/notes.txt", "content while embed is disabled")

	h.DispatchExtractAndEmbed(t, f)
	h.WaitForJobs(t)

	rows := h.QueryEmbeddings(t, f.ID())
	assert.Empty(t, rows, "should write nothing when the embed feature flag is disabled")
}
