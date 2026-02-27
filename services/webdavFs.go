// Package services provides WebDAV filesystem implementation for Weblens.
package services

import (
	"context"
	iofs "io/fs"
	"os"
	"path"
	"strings"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"golang.org/x/net/webdav"
)

// Ensure WebdavFs implements the webdav.FileSystem interface.
var _ webdav.FileSystem = (*WebdavFs)(nil)

type webdavUserKeyType struct{}

var webdavUserKey = webdavUserKeyType{}

// WithWebDAVUser returns a context carrying the authenticated WebDAV user.
func WithWebDAVUser(ctx context.Context, user *user_model.User) context.Context {
	return context.WithValue(ctx, webdavUserKey, user)
}

func webdavUser(ctx context.Context) *user_model.User {
	u, _ := ctx.Value(webdavUserKey).(*user_model.User)

	return u
}

// WebdavFs implements webdav.FileSystem by delegating all operations
// to the Weblens file service. Each request is scoped to the
// authenticated user's home directory.
type WebdavFs struct {
	FileService file_model.Service
}

// resolvePath translates a WebDAV path (e.g. "/photos/img.jpg") to a
// portable Weblens filepath (e.g. "USERS:username/photos/img.jpg").
// The WebDAV root "/" maps to the user's home folder.
func resolvePath(webdavPath string, user *user_model.User, isDir bool) (fs.Filepath, error) {
	webdavPath = strings.TrimPrefix(webdavPath, "/webdav/")

	if strings.Contains(webdavPath, file_model.UserTrashDirName) {
		return fs.Filepath{}, os.ErrPermission
	}

	cleaned := path.Clean(webdavPath)
	if cleaned == "." || cleaned == "/" {
		// User's home directory
		return fs.BuildFilePath(file_model.UsersTreeKey, user.Username+"/"), nil
	}

	// Strip leading slash
	rel := strings.TrimPrefix(cleaned, "/")

	suffix := ""
	if isDir {
		suffix = "/"
	}

	return fs.BuildFilePath(file_model.UsersTreeKey, user.Username+"/"+rel+suffix), nil
}

// Stat returns file info for the named file, delegating to the file service.
func (w *WebdavFs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	user := webdavUser(ctx)
	if user == nil {
		return nil, os.ErrPermission
	}

	return w.lookupFile(ctx, name, user)
}

// OpenFile opens or creates a file. For existing files it returns a wrapper;
// when O_CREATE is set and the file doesn't exist, it creates a new one.
func (w *WebdavFs) OpenFile(ctx context.Context, name string, flag int, _ os.FileMode) (webdav.File, error) {
	start := time.Now()
	user := webdavUser(ctx)
	if user == nil {
		return nil, os.ErrPermission
	}

	f, lookupErr := w.lookupFile(ctx, name, user)
	if lookupErr != nil {
		return nil, lookupErr
		// // File doesn't exist
		// if flag&os.O_CREATE != 0 {
		// 	// Create new file
		// 	parentPath, err := resolvePath(path.Dir(name), user, true)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		//
		// 	parent, err := w.FileService.GetFileByFilepath(ctx, parentPath)
		// 	if err != nil {
		// 		return nil, wlerrors.Errorf("WebDav OpenFile failed to find parent directory at %s: %w", parentPath, file_model.ErrFileNotFound)
		// 	}
		//
		// 	newFile, err := w.FileService.CreateFile(ctx, parent, path.Base(name))
		// 	if err != nil {
		// 		return nil, err
		// 	}
		//
		// 	return &webdavFile{file: newFile, fs: w, ctx: ctx}, nil
		// }
		//
		// return nil, wlerrors.Errorf("WebDav OpenFile failed to lookup file at %s: %w", name, file_model.ErrFileNotFound)
	}

	log.FromContext(ctx).Debug().Msgf("WebDAV OpenFile lookup for %s took %s", name, time.Since(start))
	return &webdavFile{file: f, fs: w, ctx: ctx}, nil
}

// Mkdir creates a new directory via the file service.
func (w *WebdavFs) Mkdir(ctx context.Context, name string, _ os.FileMode) error {
	log.FromContext(ctx).Debug().Msgf("WebDAV Mkdir called for %s", name)
	user := webdavUser(ctx)
	if user == nil {
		return os.ErrPermission
	}

	parentPath, err := resolvePath(path.Dir(name), user, true)
	if err != nil {
		log.FromContext(ctx).Error().Stack().Err(err).Msgf("WebDAV Mkdir failed to resolve parent path for %s", name)

		return err
	}

	parent, err := w.FileService.GetFileByFilepath(ctx, parentPath)
	if err != nil {
		log.FromContext(ctx).Error().Stack().Err(err).Msgf("WebDAV Mkdir failed to find parent directory at %s", parentPath)

		return os.ErrNotExist
	}

	_, err = w.FileService.CreateFolder(ctx, parent, path.Base(name))
	if err != nil {
		log.FromContext(ctx).Error().Stack().Err(err).Msgf("WebDAV Mkdir failed for %s", name)

		return err
	}

	return nil
}

// RemoveAll deletes a file or directory via the file service.
func (w *WebdavFs) RemoveAll(ctx context.Context, name string) error {
	user := webdavUser(ctx)
	if user == nil {
		return os.ErrPermission
	}

	f, err := w.lookupFile(ctx, name, user)
	if err != nil {
		return err
	}

	return w.FileService.DeleteFiles(ctx, f)
}

// Rename moves and/or renames a file via the file service.
func (w *WebdavFs) Rename(ctx context.Context, oldName, newName string) error {
	user := webdavUser(ctx)
	if user == nil {
		return os.ErrPermission
	}

	f, err := w.lookupFile(ctx, oldName, user)
	if err != nil {
		return err
	}

	oldDir := path.Dir(oldName)
	newDir := path.Dir(newName)
	oldBase := path.Base(oldName)
	newBase := path.Base(newName)

	// If parents differ, move the file first
	if oldDir != newDir {
		destPath, err := resolvePath(newDir, user, true)
		if err != nil {
			return err
		}

		dest, err := w.FileService.GetFileByFilepath(ctx, destPath)
		if err != nil {
			return os.ErrNotExist
		}

		err = w.FileService.MoveFiles(ctx, []*file_model.WeblensFileImpl{f}, dest)
		if err != nil {
			return err
		}
	}

	// If names differ, rename
	if oldBase != newBase {
		return w.FileService.RenameFile(ctx, f, newBase)
	}

	return nil
}

func (w *WebdavFs) lookupFile(ctx context.Context, name string, user *user_model.User) (*file_model.WeblensFileImpl, error) {
	fp, err := resolvePath(name, user, true)
	if err != nil {
		return nil, err
	}

	f, err := w.FileService.GetFileByFilepath(ctx, fp)
	if err == nil {
		return f, nil
	}

	// Try as file (no trailing slash)
	fp, err = resolvePath(name, user, false)
	if err != nil {
		return nil, err
	}

	f, err = w.FileService.GetFileByFilepath(ctx, fp)
	if err != nil {
		return nil, wlerrors.Errorf("WebDav Could not lookup file at %s: %w", fp, file_model.ErrFileNotFound)
	}

	return f, nil
}

// webdavFile wraps a WeblensFileImpl to implement the webdav.File interface.
// It lazily opens an os.File for read/seek operations on non-directory files.
// The ctx field preserves the request context for operations like Readdir
// that need the AppContext but don't receive a context parameter.
type webdavFile struct {
	file   *file_model.WeblensFileImpl
	fs     *WebdavFs
	ctx    context.Context
	osFile *os.File
}

var _ webdav.File = (*webdavFile)(nil)

// Close closes the underlying os.File if one was opened.
func (wf *webdavFile) Close() error {
	return nil
}

// Read reads from the underlying file.
func (wf *webdavFile) Read(p []byte) (int, error) {
	if wf.file.IsDir() {
		return 0, os.ErrInvalid
	}

	return wf.file.Read(p)
}

// Readdir returns directory entries, filtering out the trash directory.
func (wf *webdavFile) Readdir(count int) ([]iofs.FileInfo, error) {
	if !wf.file.IsDir() {
		return nil, os.ErrInvalid
	}

	children, err := wf.fs.FileService.GetChildren(wf.ctx, wf.file)
	if err != nil {
		return nil, err
	}

	infos := make([]iofs.FileInfo, 0, len(children))
	for _, child := range children {
		// Filter out the trash directory
		if child.Name() == file_model.UserTrashDirName {
			continue
		}

		infos = append(infos, child)
	}

	if count > 0 && count < len(infos) {
		infos = infos[:count]
	}

	return infos, nil
}

// Seek seeks within the underlying os.File (lazily opened).
func (wf *webdavFile) Seek(offset int64, whence int) (int64, error) {
	log.GlobalLogger().Debug().Msgf("webdavFile Seek called with offset=%d, whence=%d", offset, whence)

	if wf.file.IsDir() {
		return 0, os.ErrInvalid
	}

	return wf.file.Seek(offset, whence)
	// return wf.osFile.Seek(offset, whence)
}

// Stat returns file info for the wrapped file.
func (wf *webdavFile) Stat() (iofs.FileInfo, error) {
	return wf.file, nil
}

// Write writes data to the file via the Weblens file model.
func (wf *webdavFile) Write(p []byte) (int, error) {
	return 0, wlerrors.New("WebDAV write operations are not supported")
}
