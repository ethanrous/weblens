package task

import (
	"context"
	"sync"
	"time"

	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/rs/zerolog"
	"github.com/viccon/sturdyc"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewTestTask creates a Task for testing purposes.
// This allows external packages to create Task instances with specific state.
// Only available when building with -tags=test.
func NewTestTask(id, jobName string, workerID int64, exitStatus ExitStatus, result Result, startTime time.Time) *Task {
	logger := wlog.NewZeroLogger()
	ctx := wlog.WithContext(context.Background(), logger)

	return &Task{
		taskID:     id,
		jobName:    jobName,
		WorkerID:   workerID,
		exitStatus: exitStatus,
		queueState: Exited,
		result:     result,
		StartTime:  startTime,
		Ctx:        ctx,
	}
}

// testContext is a simple context that implements context_mod.Z for testing.
type testContext struct {
	context.Context

	logger *zerolog.Logger
	wg     *sync.WaitGroup
}

func (tc testContext) Database() *mongo.Database {
	return nil
}

func (tc testContext) GetCache(_ string) *sturdyc.Client[any] {
	return nil
}

func (tc testContext) Log() *zerolog.Logger {
	return tc.logger
}

func (tc testContext) WithLogger(_ zerolog.Logger) {}

func (tc testContext) WithContext(ctx context.Context) context.Context {
	return testContext{
		Context: ctx,
		logger:  tc.logger,
		wg:      tc.wg,
	}
}

func (tc testContext) Value(key any) any {
	if key == context_mod.WgKey {
		return tc.wg
	}

	return tc.Context.Value(key)
}

// NewTestContext creates a context implementing context_mod.Z for testing.
func NewTestContext() context_mod.Z {
	logger := wlog.NewZeroLogger()
	wg := &sync.WaitGroup{}

	return testContext{
		Context: context.Background(),
		logger:  logger,
		wg:      wg,
	}
}

// NewTestContextWithCancel creates a cancelable context implementing context_mod.Z for testing.
func NewTestContextWithCancel() (context_mod.Z, context.CancelFunc) {
	logger := wlog.NewZeroLogger()
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	return testContext{
		Context: ctx,
		logger:  logger,
		wg:      wg,
	}, cancel
}

// NewTestWorkerPool creates a WorkerPool for testing purposes.
func NewTestWorkerPool(numWorkers int) *WorkerPool {
	return NewWorkerPool(numWorkers)
}

// testMetadata is a simple metadata implementation for testing.
type testMetadata struct {
	jobName            string
	data               map[string]any
	metaStringOverride string
}

func (m *testMetadata) JobName() string {
	return m.jobName
}

func (m *testMetadata) MetaString() string {
	if m.metaStringOverride != "" {
		return m.metaStringOverride
	}

	return m.jobName
}

func (m *testMetadata) FormatToResult() Result {
	return Result{}
}

func (m *testMetadata) Verify() error {
	return nil
}

// NewTestMetadata creates a Metadata instance for testing.
func NewTestMetadata(jobName string) Metadata {
	return &testMetadata{
		jobName: jobName,
		data:    make(map[string]any),
	}
}

// NewTestQueueableTask creates a Task that can be queued for testing.
// Unlike NewTestTask, this creates a task in the Created state that can be queued.
func NewTestQueueableTask(id, jobName string) *Task {
	logger := wlog.NewZeroLogger()
	ctx := wlog.WithContext(context.Background(), logger)
	ctx, cancel := context.WithCancelCause(ctx)

	return &Task{
		taskID:     id,
		jobName:    jobName,
		queueState: Created,
		exitStatus: TaskNoStatus,
		work:       job{handler: func(_ *Task) {}, opts: Options{}},
		waitChan:   make(chan struct{}),
		Ctx:        ctx,
		cancelFunc: cancel,
	}
}
