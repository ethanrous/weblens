package file

import (
	"net/http"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/pkg/errors"
)

func checkFileAccess(ctx context_service.RequestContext) (file *file_model.WeblensFileImpl, err error) {
	fileId := ctx.Path("fileId")
	if fileId == "" {
		fileId = ctx.Path("folderId")
	}

	// Check if the request is a takeout request (zip file)
	isTakeout := ctx.Query("isTakeout")
	if isTakeout == "true" {
		file, err = ctx.FileService.GetZip(fileId)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return nil, err
		}
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

	return checkFileAccessById(ctx, fileId)
}

func checkFileAccessById(ctx context_service.RequestContext, fileId string) (file *file_model.WeblensFileImpl, err error) {
	file, err = ctx.FileService.GetFileById(fileId)
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
	if !auth.CanUserAccessFile(ctx, ctx.Requester, file, ctx.Share) {
		// If the user does not have access, return Unauthorized
		err = errors.New("access denied to file")
		ctx.Error(http.StatusUnauthorized, err)

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
	if !auth.CanUserAccessFile(ctx, ctx.Requester, file, nil) {
		// If the user does not have access, return Unauthorized
		err = errors.New("access denied to file")
		ctx.Error(http.StatusUnauthorized, err)

		return nil, err
	}

	return file, nil
}
