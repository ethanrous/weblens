package jobs

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethanrous/weblens/models/embedding"
	"github.com/ethanrous/weblens/models/featureflags"
	job_model "github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/config"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/embed"
)

// shouldExtractTextOnScan reports whether the given extension gets text extraction on scan; image types are excluded because they are embedded visually (CLIP) during scan.
func shouldExtractTextOnScan(ext string) bool {
	ext = strings.ToLower(ext)

	return media_model.EmbedEligible(ext) && !media_model.ParseExtension(ext).SupportsImgRecog()
}

// ExtractAndEmbedFile runs extraction + embedding for one file.
func ExtractAndEmbedFile(tsk *task.Task) {
	meta, ok := tsk.GetMeta().(job_model.ExtractAndEmbedMeta)
	if !ok {
		tsk.Fail(errors.New("ExtractAndEmbedFile: wrong meta type"))

		return
	}

	// Respect the embed feature flag here so every dispatch site is gated at a single choke point.
	ctx, ok := context_service.FromContext(tsk.Ctx)
	if !ok {
		tsk.Fail(errors.New("ExtractAndEmbedFile: failed to get app context"))

		return
	}

	flags, err := featureflags.GetFlags(ctx)
	if err != nil {
		tsk.Fail(err)

		return
	}

	if !flags.EnableEmbed {
		tsk.SetResult(task.Result{"skipped": "embed_disabled"})
		tsk.Success()

		return
	}

	file := meta.File

	if file == nil || file.IsDir() || file.IsPastFile() {
		tsk.SetResult(task.Result{"skipped": "ineligible"})
		tsk.Success()

		return
	}

	ext := filepath.Ext(file.GetPortablePath().Filename())
	if !media_model.EmbedEligible(ext) {
		tsk.SetResult(task.Result{"skipped": "extension"})
		tsk.Success()

		return
	}

	if file.Size() > config.GetConfig().EmbedMaxFileSize && config.GetConfig().EmbedMaxFileSize > 0 {
		tsk.SetResult(task.Result{"skipped": "too_large"})
		tsk.Success()

		return
	}

	if embed.Default().ServiceUnavailable() {
		tsk.SetResult(task.Result{"skipped": "service_unavailable"})
		tsk.Success()

		return
	}

	absPath := file.GetPortablePath().ToAbsolute()

	modelName := currentModelName()

	same, err := embedding.CountByContentHash(tsk.Ctx, embedding.KindFileChunk, file.ID(), modelName, file.GetContentID())
	if err != nil {
		tsk.Fail(err)

		return
	}

	if same > 0 {
		tsk.SetResult(task.Result{"skipped": "unchanged"})
		tsk.Success()

		return
	}

	chunks, err := embed.Default().ExtractAndEmbedFile(tsk.Ctx, absPath, "")
	if errors.Is(err, embed.ErrExtractionFailed) {
		// Extraction failed — leave existing embeddings in place instead of pruning them away.
		tsk.SetResult(task.Result{"skipped": "extraction_failed"})
		tsk.Success()

		return
	}

	if err != nil {
		tsk.Fail(err)

		return
	}

	now := time.Now().UTC()

	for _, c := range chunks {
		err := embedding.Upsert(tsk.Ctx, embedding.Embedding{
			Kind:        embedding.KindFileChunk,
			SourceID:    file.ID(),
			ChunkIndex:  c.ChunkIndex,
			Page:        c.Page,
			Snippet:     c.Snippet,
			Vector:      c.Vector,
			Model:       modelName,
			CreatedAt:   now,
			ContentHash: file.GetContentID(),
		})
		if err != nil {
			tsk.Fail(err)

			return
		}
	}

	if err := embedding.PruneTrailingChunks(tsk.Ctx, embedding.KindFileChunk, file.ID(), len(chunks)); err != nil {
		tsk.Fail(err)

		return
	}

	tsk.SetResult(task.Result{"chunks": len(chunks)})
	tsk.Success()
}

// currentModelName must stay in sync with embed/main.py MODEL_ID.
func currentModelName() string { return "jina-clip-v2" }
