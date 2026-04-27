// Package file provides HTTP handlers for file and folder operations in the Weblens API.
package file

import (
	"net/http"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/share"
	zip_model "github.com/ethanrous/weblens/models/zip"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/rs/zerolog"
)

func checkFileAccess(ctx context_service.RequestContext, perms ...share.Permission) (file *file_model.WeblensFileImpl, err error) {
	fileID := ctx.Path("fileID")
	if fileID == "" {
		fileID = ctx.Path("folderID")
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

			return nil, err
		}
		// For any other error
		ctx.Error(http.StatusInternalServerError, err)

		return nil, err
	}

	files := []*file_model.WeblensFileImpl{file}

	// TODO: bypass source file checks if zip owner is accessor
	if file_model.ZipsDirPath.IsParentOf(file.GetPortablePath()) {
		zip, err := zip_model.GetZip(ctx, file.ID())
		if err != nil {
			err = wlerrors.Wrapf(err, "failed to get zip object for %s (%s)", file.GetPortablePath(), file.ID())
			ctx.Error(http.StatusInternalServerError, err)

			return nil, err
		}

		files = make([]*file_model.WeblensFileImpl, 0, len(zip.SourceFileIDs))

		for _, zippedID := range zip.SourceFileIDs {
			zipSourceFile, err := ctx.FileService.GetFileByID(ctx, zippedID)
			if err != nil && wlerrors.Is(err, file_model.ErrFileNotFound) {
				// FIXME: potentially dangerous. If a zip contains a file that is shared, and another that gets deleted,
				// the user who is shared into the other file could then access the deleted file, since we skip the check here.
				continue
			} else if err != nil {
				err = wlerrors.Wrapf(err, "failed to get file object for %s", zippedID)
				ctx.Error(http.StatusInternalServerError, err)

				return nil, err
			}

			files = append(files, zipSourceFile)
		}
	}

	for _, file := range files {
		ctx.Log().Debug().Func(func(e *zerolog.Event) {
			e = e.Str("fileID", file.ID()).Str("username", ctx.Requester.Username)

			if ctx.Share != nil {
				e = e.Str("shareID", ctx.Share.ID().Hex())
			}

			e.Msgf("Checking file access for permissions %v", perms)
		})

		// Check if the user has access to the file
		if _, err = auth.CanUserAccessFile(ctx, ctx.Requester, file, ctx.Share, perms...); err != nil {
			// If the user does not have access, return forbidden
			ctx.Error(http.StatusForbidden, err)

			return nil, err
		}
	}

	return file, nil
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
