package tower

import (
	"context"
	"io"

	tower_model "github.com/ethanrous/weblens/models/tower"
)

func DownloadFileFromCore(ctx context.Context, core tower_model.Instance, fileId string, dest io.Writer) error {
	client, err := getApiClient(ctx, core)
	if err != nil {
		return err
	}

	_, req, err := client.FilesAPI.DownloadFile(ctx, fileId).Execute()
	if err != nil {
		return err
	}

	defer req.Body.Close()

	_, err = io.Copy(dest, req.Body)
	if err != nil {
		return err
	}

	return nil
}
