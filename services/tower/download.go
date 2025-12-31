package tower

import (
	"context"
	"io"

	tower_model "github.com/ethanrous/weblens/models/tower"
)

// DownloadFileFromCore downloads a file from a core server and writes it to the destination.
func DownloadFileFromCore(ctx context.Context, core tower_model.Instance, fileID string, dest io.Writer) error {
	client, err := getAPIClient(ctx, core)
	if err != nil {
		return err
	}

	_, req, err := client.FilesAPI.DownloadFile(ctx, fileID).Execute()
	if err != nil {
		return err
	}

	defer req.Body.Close() //nolint:errcheck

	_, err = io.Copy(dest, req.Body)
	if err != nil {
		return err
	}

	return nil
}
