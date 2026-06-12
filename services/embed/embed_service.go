package embed

import (
	"context"

	"github.com/ethanrous/weblens/models/embedding"
	file_model "github.com/ethanrous/weblens/models/file"
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
