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

// IsFileEmbedded checks whether any embeddings of either kind exist for a file, which is used to determine whether the file is embedded or not.
func IsFileEmbedded(ctx context.Context, file *file_model.WeblensFileImpl) (bool, error) {
	textEmbeddings, err := embedding.CountForSource(ctx, embedding.KindFileChunk, file.ID())
	if err != nil {
		return false, err
	}

	imgEmbeddings, err := embedding.CountForSource(ctx, embedding.KindImage, file.GetContentID())
	if err != nil {
		return false, err
	}

	return textEmbeddings > 0 || imgEmbeddings > 0, nil
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
