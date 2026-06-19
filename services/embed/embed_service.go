package embed

import (
	"context"

	"github.com/ethanrous/weblens/models/embedding"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// FilesEmbedded reports, keyed by file ID, which of the given files already have an embedding of
// either kind (a text-chunk row keyed by file ID, or an image row keyed by content ID). It batches
// the lookup into two distinct queries rather than two counts per file to avoid an N+1 on folder listings.
func FilesEmbedded(ctx context.Context, files []*file_model.WeblensFileImpl) (map[string]bool, error) {
	fileIDs := make([]string, 0, len(files))
	contentIDs := make([]string, 0, len(files))

	for _, f := range files {
		fileIDs = append(fileIDs, f.ID())

		if cid := f.GetContentID(); cid != "" {
			contentIDs = append(contentIDs, cid)
		}
	}

	textPresent, err := embedding.SourceIDsWithEmbeddings(ctx, embedding.KindFileChunk, fileIDs)
	if err != nil {
		return nil, err
	}

	imgPresent, err := embedding.SourceIDsWithEmbeddings(ctx, embedding.KindImage, contentIDs)
	if err != nil {
		return nil, err
	}

	embedded := make(map[string]bool, len(files))

	for _, f := range files {
		_, hasText := textPresent[f.ID()]
		_, hasImg := imgPresent[f.GetContentID()]
		embedded[f.ID()] = hasText || hasImg
	}

	return embedded, nil
}

// IsEmbeddingInProgress checks if there are any pending embedding jobs for the user, which is used to determine whether to show the "embedding in progress" state in the UI.
func IsEmbeddingInProgress(ctx context.Context, user *user_model.User) (bool, error) {
	// Check if there are any pending embedding jobs for the user
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return false, wlerrors.New("failed to retrieve application context")
	}

	tasks := appCtx.TaskService.GetTasks()
	for _, t := range tasks {
		if t.GetMeta().JobName() != job.ExtractAndEmbedTask {
			continue
		}

		meta, ok := t.GetMeta().(job.ExtractAndEmbedMeta)
		if !ok {
			continue
		}

		ownerName, err := file_model.GetFileOwnerName(ctx, meta.File)
		if err != nil {
			return false, err
		}

		if ownerName == user.Username {
			return true, nil
		}
	}

	return false, nil
}
