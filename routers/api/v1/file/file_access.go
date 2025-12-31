// Package file provides HTTP handlers for file and folder operations in the Weblens API.
package file

import (
	"net/http"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/journal"
)

func checkFileAccess(ctx context_service.RequestContext, perms ...share.Permission) (file *file_model.WeblensFileImpl, err error) {
	fileID := ctx.Path("fileID")
	if fileID == "" {
		fileID = ctx.Path("folderID")
	}

	// Check if the request is a takeout request (zip file)
	isTakeout := ctx.Query("isTakeout")
	if isTakeout == "true" {
		file, err := ctx.FileService.GetFileByID(ctx, fileID)
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
		ctx.Log().Trace().Msgf("Checking past file access for [%s] at %s", fileID, ts)

		return checkPastFileAccess(ctx, fileID, ts)
	}

	return CheckFileAccessByID(ctx, fileID, perms...)
}

// CheckFileAccessByID verifies that the requester has the specified permissions to access a file by its ID.
func CheckFileAccessByID(ctx context_service.RequestContext, fileID string, perms ...share.Permission) (file *file_model.WeblensFileImpl, err error) {
	file, err = ctx.FileService.GetFileByID(ctx, fileID)
	if err != nil {
		// Handle error if file not found
		if wlerrors.Is(err, file_model.ErrFileNotFound) {
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

func checkPastFileAccess(ctx context_service.RequestContext, fileID string, timestamp time.Time) (file *file_model.WeblensFileImpl, err error) {
	file, err = journal.GetPastFileByID(ctx, fileID, timestamp)
	if err != nil {
		// Handle error if file not found, return 404.
		// Here we both send the error to the client and return it so the caller can handle it as needed.
		// The caller of this function should never send the error to the client again to avoid duplicate responses.
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
