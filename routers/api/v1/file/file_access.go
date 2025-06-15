package file

import (
	"net/http"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/journal"
)

func checkFileAccess(ctx context_service.RequestContext, perms ...share.Permission) (file *file_model.WeblensFileImpl, err error) {
	fileId := ctx.Path("fileId")
	if fileId == "" {
		fileId = ctx.Path("folderId")
	}

	// Check if the request is a takeout request (zip file)
	isTakeout := ctx.Query("isTakeout")
	if isTakeout == "true" {
		file, err := ctx.FileService.GetFileById(ctx, fileId)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return nil, err
		}

		return file, nil
	}

	ts, ok, err := context_service.TimestampFromCtx(ctx)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return nil, err
	}

	if ok {
		ctx.Log().Trace().Msgf("Checking past file access for [%s] at %s", fileId, ts)

		return checkPastFileAccess(ctx, fileId, ts)
	}

	return checkFileAccessById(ctx, fileId, perms...)
}

func checkFileAccessById(ctx context_service.RequestContext, fileId string, perms ...share.Permission) (file *file_model.WeblensFileImpl, err error) {
	file, err = ctx.FileService.GetFileById(ctx, fileId)
	if err != nil {
		// Handle error if file not found
		if errors.Is(err, file_model.ErrFileNotFound) {
			ctx.Error(http.StatusNotFound, err)

			return
		}
		// For any other error
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Check if the user has access to the file
	if _, err = auth.CanUserAccessFile(ctx, ctx.Requester, file, ctx.Share, perms...); err != nil {
		// If the user does not have access, return forbidden
		ctx.Error(http.StatusForbidden, err)

		return
	}

	return
}

func checkPastFileAccess(ctx context_service.RequestContext, fileId string, timestamp time.Time) (file *file_model.WeblensFileImpl, err error) {
	file, err = journal.GetPastFileById(ctx, fileId, timestamp)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return nil, err
	}

	// Check if the user has access to the file
	if _, err = auth.CanUserAccessFile(ctx, ctx.Requester, file, nil); err != nil {
		// If the user does not have access, return forbidden
		ctx.Error(http.StatusForbidden, err)

		return nil, err
	}

	return file, nil
}
