package jobs

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethanrous/weblens/models/embedding"
	job_model "github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/services/embed"
)

var eligibleExtensions = map[string]bool{
	".txt": true, ".md": true, ".csv": true, ".log": true,
	".json": true, ".yaml": true, ".yml": true,
	".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
	".vue": true, ".rs": true, ".java": true, ".c": true, ".cpp": true,
	".h": true, ".hpp": true, ".sh": true, ".rb": true, ".kt": true, ".swift": true,
	".pdf":  true,
	".docx": true, ".xlsx": true, ".pptx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".heic": true,
	".tif": true, ".tiff": true, ".bmp": true,
}

// imageExtensions are eligible for OCR on the upload path but are embedded
// visually (CLIP) during scan. We skip text extraction for them in the scan
// pipeline to avoid OCR-ing every photo on every folder scan.
var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".heic": true,
	".tif": true, ".tiff": true, ".bmp": true,
}

// shouldExtractTextOnScan reports whether a file with the given extension
// should have document-text extraction dispatched during a folder scan. It
// accepts an extension with or without a leading dot, any case. Image types
// are excluded — they are embedded visually, not via OCR, on the scan path.
func shouldExtractTextOnScan(ext string) bool {
	ext = strings.ToLower(ext)
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}

	return eligibleExtensions[ext] && !imageExtensions[ext]
}

// ExtractAndEmbedFile runs extraction + embedding for one file.
func ExtractAndEmbedFile(tsk *task.Task) {
	meta, ok := tsk.GetMeta().(job_model.ExtractAndEmbedMeta)
	if !ok {
		tsk.Fail(errors.New("ExtractAndEmbedFile: wrong meta type"))

		return
	}

	file := meta.File

	if file == nil || file.IsDir() || file.IsPastFile() {
		tsk.SetResult(task.Result{"skipped": "ineligible"})
		tsk.Success()

		return
	}

	ext := strings.ToLower(filepath.Ext(file.GetPortablePath().Filename()))
	if !eligibleExtensions[ext] {
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
