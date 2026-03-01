package file

import (
	"context"
	"os"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	"github.com/rs/zerolog"
)

func (fs *ServiceImpl) removeFileByID(ctx context.Context, fileID string) error {
	context_mod.ToZ(ctx).Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Removing file with id [%s] removed from file tree", fileID)
	})

	fs.treeLock.Lock()
	defer fs.treeLock.Unlock()

	delete(fs.files, fileID)

	return nil
}

func linkToRestore(ctx context.Context, file *file_model.WeblensFileImpl) error {
	if file.GetContentID() == "" {
		_, err := file_model.GenerateContentID(ctx, file)
		if err != nil {
			return err
		}
	}

	// Check if the restore file already exists, with the filename being the content id
	restorePath := file_system.BuildFilePath(file_model.RestoreTreeKey, file.GetContentID())

	if exists(restorePath) {
		// If the file already exists in the restore tree, no action is needed
		return nil
	}

	// Link file from USERS tree to the RESTORE tree. Files later can be hard-linked back
	// from the restore tree to the users tree, but will not be "moved" back.
	err := os.Link(file.GetPortablePath().ToAbsolute(), restorePath.ToAbsolute())
	if err != nil {
		return err
	}

	return nil
}

func rmFileMedia(ctx context.Context, file *file_model.WeblensFileImpl) error {
	contentID := file.GetContentID()
	if contentID == "" {
		return nil
	}

	m, err := media_model.GetMediaByContentID(ctx, contentID)
	// Remove the file from the media, if it exists
	if err == nil {
		err = media_model.RemoveFileFromMedia(ctx, m, file.ID())
		if err != nil {
			return err
		}
	}

	return nil
}
