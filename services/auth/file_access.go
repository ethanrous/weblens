package auth

import (
	"net/http"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	takeout_model "github.com/ethanrous/weblens/models/takeout"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/rs/zerolog"
)

// CanUserAccessFileByID is the pure access check. It resolves the file by ID, expands zip
// source files when applicable, and verifies the requester has the requested permissions on
// each file. It NEVER writes to ctx.Error. Errors are wrapped with HTTP status codes via
// wlerrors.WrapStatus so callers can pass them straight to ctx.Error and the right code emerges.
//
// Status mapping:
//   - file not found → 404
//   - any other resolution error → 500
//   - permission denied on any file (including zip source files) → 403
func CanUserAccessFileByID(ctx context_service.RequestContext, fileID string, perms ...share_model.Permission) (*file_model.WeblensFileImpl, error) {
	file, err := ctx.FileService.GetFileByID(ctx, fileID)
	if err != nil {
		if wlerrors.Is(err, file_model.ErrFileNotFound) {
			return nil, wlerrors.WrapStatus(http.StatusNotFound, err)
		}

		return nil, wlerrors.WrapStatus(http.StatusInternalServerError, err)
	}

	files := []*file_model.WeblensFileImpl{file}

	// TODO: bypass source file checks if zip owner is accessor
	if file_model.ZipsDirPath.IsParentOf(file.GetPortablePath()) {
		zip, err := takeout_model.GetZip(ctx, file.ID())
		if err != nil {
			return nil, wlerrors.WrapStatus(http.StatusInternalServerError, wlerrors.Wrapf(err, "failed to get zip object for %s (%s)", file.GetPortablePath(), file.ID()))
		}

		files = make([]*file_model.WeblensFileImpl, 0, len(zip.SourceFileIDs))

		for _, zippedID := range zip.SourceFileIDs {
			zipSourceFile, srcErr := ctx.FileService.GetFileByID(ctx, zippedID)
			if srcErr != nil && wlerrors.Is(srcErr, file_model.ErrFileNotFound) {
				// FIXME: potentially dangerous. If a zip contains a file that is shared, and another that gets deleted,
				// the user who is shared into the other file could then access the deleted file, since we skip the check here.
				continue
			} else if srcErr != nil {
				return nil, wlerrors.WrapStatus(http.StatusInternalServerError, wlerrors.Wrapf(srcErr, "failed to get file object for %s", zippedID))
			}

			files = append(files, zipSourceFile)
		}
	}

	for _, f := range files {
		ctx.Log().Debug().Func(func(e *zerolog.Event) {
			e = e.Str("fileID", f.ID()).Str("username", ctx.Requester.Username)

			if ctx.Share != nil {
				e = e.Str("shareID", ctx.Share.ID().Hex())
			}

			e.Msgf("Checking file access for permissions %v", perms)
		})

		if _, accessErr := CanUserAccessFile(ctx, ctx.Requester, f, ctx.Share, perms...); accessErr != nil {
			return nil, wlerrors.WrapStatus(http.StatusForbidden, accessErr)
		}
	}

	return file, nil
}

// CanUserAccessPastFileByID resolves a file by ID at a past timestamp via journal and verifies
// the requester can see it. History snapshots are not share-aware: nil share is passed and no
// per-route permissions are enforced (matching the legacy checkPastFileAccess behavior).
func CanUserAccessPastFileByID(ctx context_service.RequestContext, fileID string, ts time.Time) (*file_model.WeblensFileImpl, error) {
	file, err := journal.GetPastFileByID(ctx, fileID, ts)
	if err != nil {
		return nil, wlerrors.WrapStatus(http.StatusNotFound, err)
	}

	if _, err := CanUserAccessFile(ctx, ctx.Requester, file, nil); err != nil {
		return nil, wlerrors.WrapStatus(http.StatusForbidden, err)
	}

	return file, nil
}

// RequireFileAccess verifies the requester has the given permissions on EVERY file in fileIDs
// and returns the resolved files. The returned slice has the same length as fileIDs and is
// indexed in the same order - callers can rely on `files[i]` corresponding to `fileIDs[i]`.
// On the first failure, it writes ctx.Error with the wrapped error and returns it so the caller
// can `return`. The status code (404/403/500) is extracted from the error by ctx.Error via
// wlerrors.AsStatus; the literal 500 passed below is just the fallback used when no status is
// wrapped.
func RequireFileAccess(ctx context_service.RequestContext, fileIDs []string, perms ...share_model.Permission) ([]*file_model.WeblensFileImpl, error) {
	files := make([]*file_model.WeblensFileImpl, 0, len(fileIDs))

	for _, id := range fileIDs {
		file, err := CanUserAccessFileByID(ctx, id, perms...)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

// RequireFileAccessOne is the single-ID convenience wrapper around RequireFileAccess. It
// avoids the `[]string{id}` + `files[0]` boilerplate at the (many) call sites that only
// need one access check.
func RequireFileAccessOne(ctx context_service.RequestContext, fileID string, perms ...share_model.Permission) (*file_model.WeblensFileImpl, error) {
	files, err := RequireFileAccess(ctx, []string{fileID}, perms...)
	if err != nil {
		return nil, err
	}

	return files[0], nil
}

// RequireAnyFileAccess returns the first file from fileIDs that the requester can access with
// the given permissions. Returns 403 only if EVERY id is denied. Hard errors (404/500) on any
// id surface immediately rather than falling through to the next id - only permission denials
// participate in the any-of semantics. This exists for media flows where a single piece of
// media may be backed by multiple files; access on any one grants access to the media.
func RequireAnyFileAccess(ctx context_service.RequestContext, fileIDs []string, perms ...share_model.Permission) (*file_model.WeblensFileImpl, error) {
	if len(fileIDs) == 0 {
		err := wlerrors.WrapStatus(http.StatusForbidden, wlerrors.New("no files to check"))
		ctx.Error(http.StatusInternalServerError, err)

		return nil, err
	}

	var lastErr error

	for _, id := range fileIDs {
		file, err := CanUserAccessFileByID(ctx, id, perms...)
		if err == nil {
			return file, nil
		}

		// Only permission denials (403) fall through to the next id; surface anything else
		// (404, 500) immediately to avoid masking hard failures with a permission-denied error.
		if status, _ := wlerrors.AsStatus(err, 0); status != http.StatusForbidden {
			ctx.Error(http.StatusInternalServerError, err)

			return nil, err
		}

		lastErr = err
	}

	ctx.Error(http.StatusInternalServerError, lastErr)

	return nil, lastErr
}
