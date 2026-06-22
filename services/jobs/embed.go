package jobs

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/embedding"
	"github.com/ethanrous/weblens/models/featureflags"
	job_model "github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/embed"
	media_service "github.com/ethanrous/weblens/services/media"
)

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

	// Files without a media row (text, documents) still get text extraction below.
	m, err := media_model.GetMediaByContentID(ctx, file.GetContentID())

	switch {
	case err == nil:
		if err := writeImageEmbedding(ctx, m, meta.ForceReIndex); err != nil {
			tsk.Fail(err)

			return
		}
	case !db.IsNotFound(err):
		tsk.Fail(err)

		return
	}

	// Photos stop here: they embed visually above, and OCR on them yields junk like watermark text.
	if !media_model.TextEmbedEligible(ext) {
		tsk.SetResult(task.Result{"skipped": "text_extension"})
		tsk.Success()

		return
	}

	same, err := embedding.CountByContentID(tsk.Ctx, embedding.KindFileChunk, file.ID(), modelName, file.GetContentID())
	if err != nil {
		tsk.Fail(err)

		return
	}

	if same > 0 && !meta.ForceReIndex {
		tsk.SetResult(task.Result{"skipped": "unchanged"})
		tsk.Success()

		return
	}

	chunks, err := embed.Default().ExtractAndEmbedFile(tsk.Ctx, absPath, "")
	if errors.Is(err, embed.ErrExtractionFailed) {
		// Extraction failed - leave existing embeddings in place instead of pruning them away.
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

// writeImageEmbedding encodes a media's cached image(s) into the embeddings collection, one row per page; non-image-recognizable types are skipped.
// When force is true, existing per-page rows are re-encoded and overwritten instead of skipped.
func writeImageEmbedding(ctx context_service.AppContext, media *media_model.Media, force bool) error {
	if !media_model.ParseMime(media.MimeType).SupportsImgRecog() {
		return nil
	}

	pageCount := max(media.GetPageCount(), 1)

	multiPage := pageCount > 1
	modelName := currentModelName()

	for page := range pageCount {
		exists, err := embedding.CountForChunk(ctx, embedding.KindImage, string(media.ContentID), modelName, page)
		if err != nil {
			return err
		}

		if exists > 0 && !force {
			continue
		}

		quality := media_model.LowRes
		if multiPage {
			quality = media_model.HighRes
		}

		cacheFile, err := media_service.GetCacheFile(ctx, media, quality, page)
		if err != nil {
			return wlerrors.Errorf("get cache file (page %d, quality %s): %w", page, quality, err)
		}

		vec, err := embed.Default().EncodeImage(ctx, cacheFile.GetPortablePath().String())
		if err != nil {
			return wlerrors.Errorf("encode image (page %d): %w", page, err)
		}

		err = embedding.Upsert(ctx, embedding.Embedding{
			Kind:       embedding.KindImage,
			SourceID:   string(media.ContentID),
			ChunkIndex: page,
			Page:       page + 1,
			Vector:     vec,
			Model:      modelName,
			CreatedAt:  time.Now().UTC(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// currentModelName must stay in sync with embed/main.py MODEL_ID.
func currentModelName() string { return "jina-clip-v2" }
