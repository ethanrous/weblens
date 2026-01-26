package tower

import (
	"context"
	"io"

	tower_model "github.com/ethanrous/weblens/models/tower"
)

// DownloadFileFromCore downloads a file from a core server and writes it to the destination.
func DownloadFileFromCore(ctx context.Context, core tower_model.Instance, fileID string, dest io.Writer) (int64, error) {
	client, err := getAPIClient(ctx, core, clientOpts{})
	if err != nil {
		return -1, err
	}

	_, req, err := client.FilesAPI.DownloadFile(ctx, fileID).Execute()
	if err != nil {
		return -1, err
	}

	defer req.Body.Close() //nolint:errcheck

	bsCopied, err := io.Copy(dest, req.Body)
	if err != nil {
		return -1, err
	}

	return bsCopied, nil
}
