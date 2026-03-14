// Package file provides services for managing files, folders, and file operations in the Weblens system.
package file

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	media_model "github.com/ethanrous/weblens/models/media"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/cryptography"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlog"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/rs/zerolog"
)

var _ file_model.Service = &ServiceImpl{}

// ServiceImpl implements the FileService interface for managing files and directories.
type ServiceImpl struct {
	contentIDCache map[string]*file_model.WeblensFileImpl
	contentIDLock  sync.RWMutex
	fileTaskLink   map[string][]*task_model.Task
	fileTaskLock   sync.RWMutex
	files          map[string]*file_model.WeblensFileImpl
	treeLock       sync.RWMutex
}

// FolderCoverPair represents a mapping between a folder and its cover image.
type FolderCoverPair struct {
	FolderID  string `bson:"folderID"`
	ContentID string `bson:"coverID"`
}

// NewFileService creates and initializes a new FileService instance.
func NewFileService(
	_ context.Context,
) (*ServiceImpl, error) {
	fs := &ServiceImpl{
		fileTaskLink: make(map[string][]*task_model.Task),
		files:        make(map[string]*file_model.WeblensFileImpl),
	}

	return fs, nil
}

// Size returns the total size of files in the specified tree.
func (fs *ServiceImpl) Size(_ string) int64 {
	// tree := fs.trees[treeAlias]
	// if tree == nil {
	// 	return -1
	// }
	//
	// return tree.GetRoot().Size()
	return -1
}

// AddFile adds one or more files to the file service and their parent directories.
func (fs *ServiceImpl) AddFile(ctx context.Context, files ...*file_model.WeblensFileImpl) (err error) {
	for _, f := range files {
		newID := f.ID()
		if newID == "" {
			return wlerrors.WithStack(file_model.ErrNoFileID)
		} else if !f.IsDir() && f.Size() != 0 && f.GetContentID() == "" && f.GetPortablePath().RootName() == file_model.UsersTreeKey {
			return wlerrors.Wrapf(file_model.ErrNoContentID, "failed to add [%s] to file service", f.GetPortablePath())
		}

		// Ensure the file has a parent directory
		p := f.GetParent()
		if p == nil {
			return wlerrors.Wrapf(file_model.ErrNoParent, "failed to add file [%s] to file service", f.GetPortablePath())
		}

		// Check for existing file with the same ID
		if _, exists := fs.getFileInternal(f.ID()); exists {
			wlog.FromContext(ctx).Warn().CallerSkipFrame(1).Msgf("File [%s] already exists in file service, skipping", f.GetPortablePath())

			continue
		}

		// Add the file to its parent directory if not already present
		if _, err = p.GetChild(f.GetPortablePath().Filename()); err != nil {
			err = p.AddChild(f)
			if err != nil {
				return err
			}
		}

		fs.setFileInternal(f.ID(), f)

		wlog.FromContext(ctx).Trace().Msgf("Added file [%s] to file service with id [%s] (%p)", f.GetPortablePath(), f.ID(), f)
	}

	return nil
}

// GetFileByID retrieves a file by its unique identifier.
func (fs *ServiceImpl) GetFileByID(ctx context.Context, id string) (*file_model.WeblensFileImpl, error) {
	f, ok := fs.getFileInternal(id)

	if ok {
		if f.ID() != id {
			return nil, wlerrors.Errorf("Mismatched fileID getting file by id %s != %s", f.ID(), id)
		}

		return f, nil
	}

	path, err := journal.GetLatestPathByID(ctx, id)
	if err != nil {
		return nil, wlerrors.WrapStatus(http.StatusNotFound, wlerrors.Wrap(file_model.ErrFileNotFound, err.Error()))
	}

	return fs.GetFileByFilepath(ctx, path)
}

// GetFileByFilepath retrieves a file by its portable filepath, optionally loading directories as needed.
func (fs *ServiceImpl) GetFileByFilepath(ctx context.Context, filepath file_system.Filepath, dontLoadNew ...bool) (*file_model.WeblensFileImpl, error) {
	root, err := fs.GetFileByID(ctx, filepath.RootName())
	if err != nil {
		return nil, fmt.Errorf("failed to get root file [%s]: %w", filepath.RootName(), err)
	}

	childFile := root

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, wlerrors.WithStack(context_service.ErrNoContext)
	}

	shouldLoadNew := true
	if len(dontLoadNew) != 0 && dontLoadNew[0] {
		shouldLoadNew = false
	}

	for child := range strings.SplitSeq(filepath.RelPath, "/") {
		if child == "" {
			continue
		}

		if !childFile.ChildrenLoaded() && shouldLoadNew {
			_, err = loadOneDirectory(appCtx, childFile)
			if err != nil {
				return nil, err
			}
		} else if !shouldLoadNew {
			return nil, wlerrors.ReplaceStack(wlerrors.Errorf("failed to load childFile [%s]: %w", childFile.GetPortablePath().String(), file_model.ErrFileNotFound))
		}

		childFile, err = childFile.GetChild(child)
		if err != nil {
			return nil, err
		}
	}

	return childFile, nil
}

// GetFileByContentID retrieves a file by its content identifier.
func (fs *ServiceImpl) GetFileByContentID(ctx context.Context, contentID string) (*file_model.WeblensFileImpl, error) {
	media, err := media_model.GetMediaByContentID(ctx, contentID)
	if err != nil {
		return nil, err
	}

	for _, fID := range media.FileIDs {
		f, err := fs.GetFileByID(ctx, fID)
		if err != nil {
			if wlerrors.Is(err, file_model.ErrFileNotFound) {
				continue // Skip files that are not found
			}

			return nil, err // Return other errors
		}

		if f.GetContentID() == media.ContentID {
			return f, nil
		}

		return nil, wlerrors.Errorf("file [%s] does not match media content ID [%s]", f.GetPortablePath(), media.ContentID)
	}

	return nil, wlerrors.Errorf("Failed getting file from media: %w", file_model.ErrFileNotFound)
}

// CreateFile creates a new file in the specified parent directory with optional initial data.
func (fs *ServiceImpl) CreateFile(ctx context.Context, parent *file_model.WeblensFileImpl, filename string, data ...[]byte) (
	*file_model.WeblensFileImpl, error,
) {
	if err := cryptography.ValidateFilename(filename); err != nil {
		return nil, wlerrors.Errorf("invalid filename: %w", err)
	}

	childPath := parent.GetPortablePath().Child(filename, false)

	newF, err := touch(childPath)
	if err != nil {
		return nil, err
	}

	if len(data) != 0 {
		for _, d := range data {
			_, err = newF.Write(d)
			if err != nil {
				return nil, err
			}
		}

		wlog.FromContext(ctx).Trace().Msgf("Generating content ID for new file [%s] (size now %d ## %d)", newF.GetPortablePath(), newF.Size(), len(data))

		_, err = file_model.GenerateContentID(ctx, newF)
		if err != nil {
			return nil, err
		}
	}

	err = fs.createCommon(ctx, newF, parent)
	if err != nil {
		return nil, err
	}

	return newF, nil
}

// CreateFolder creates a new folder in the specified parent directory.
func (fs *ServiceImpl) CreateFolder(ctx context.Context, parent *file_model.WeblensFileImpl, folderName string) (*file_model.WeblensFileImpl, error) {
	if err := cryptography.ValidateFilename(folderName); err != nil {
		return nil, wlerrors.Errorf("invalid folder name: %w", err)
	}

	childPath := parent.GetPortablePath().Child(folderName, true)

	dir, err := mkdir(childPath)
	if err != nil {
		return nil, err
	}

	err = fs.createCommon(ctx, dir, parent)
	if err != nil {
		return nil, err
	}

	return dir, nil
}

// ReturnFilesFromTrash moves files from the trash back to their previous locations.
func (fs *ServiceImpl) ReturnFilesFromTrash(_ context.Context, trashFiles []*file_model.WeblensFileImpl) error { //nolint:revive
	// trash := trashFiles[0].GetParent()
	// trashPath := trash.GetPortablePath().ToPortable()
	//
	// event := journal.NewEvent()
	//
	// for _, trashEntry := range trashFiles {
	// 	// preFile := trashEntry.Freeze()
	//
	// 	if !fs.IsFileInTrash(trashEntry) {
	// 		return errors.Errorf("cannot return file from trash, file is not in trash")
	// 	}
	//
	// 	acns := journal.Get(trashEntry.ID()).Actions
	// 	if len(acns) < 2 || !strings.HasPrefix(acns[len(acns)-1].DestinationPath, trashPath) {
	// 		return errors.Errorf("cannot return file from trash, journal does not have trash destination")
	// 	}
	//
	// 	justBeforeTrash := acns[len(acns)-2]
	// 	oldParent := tree.Get(justBeforeTrash.ParentID)
	// 	if oldParent == nil {
	// 		owner, err := fs.GetFileOwner(ctx, trashEntry)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		oldParent = tree.Get(owner.HomeID)
	// 	}
	//
	// 	portablePath, err := file_system.ParsePortable(justBeforeTrash.DestinationPath)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_, err = tree.Move(trashEntry, oldParent, portablePath.Filename(), false, event)
	//
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	// err := fs.ResizeUp(ctx, trash, event)
	// if err != nil {
	// 	return err
	// }
	//
	// journal.LogEvent(event)
	//
	// return nil
	return wlerrors.New("not implemented")
}

// MoveFiles moves one or more files to a destination folder.
func (fs *ServiceImpl) MoveFiles(ctx context.Context, files []*file_model.WeblensFileImpl, destFolder *file_model.WeblensFileImpl) error {
	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		return fs.moveFilesWithTransaction(ctx, files, destFolder)
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteFiles removes files being pointed to from the tree and moves them to the restore tree.
func (fs *ServiceImpl) DeleteFiles(ctx context.Context, files ...*file_model.WeblensFileImpl) error {
	for _, f := range files {
		if f.GetPortablePath().Dir().IsRoot() {
			return wlerrors.Errorf("cannot delete user home directory [%s]", f.GetPortablePath())
		} else if f.GetPortablePath().Filename() == file_model.UserTrashDirName {
			return wlerrors.Errorf("cannot delete user trash directory [%s]", f.GetPortablePath())
		}
	}

	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		return fs.deleteFilesWithTransaction(ctx, files)
	})
	if err != nil {
		return err
	}

	return nil
}

// restorePair tracks a file to be restored along with its destination parent.
type restorePair struct {
	parent *file_model.WeblensFileImpl
	fileID string
}

// RestoreFiles restores files to a previous state from their history at the specified time.
// It reconstructs past file states and hard-links content from the RESTORE tree back into the USERS tree.
func (fs *ServiceImpl) RestoreFiles(ctx context.Context, ids []string, newParent *file_model.WeblensFileImpl, restoreTime time.Time) error {
	queue := make([]restorePair, 0, len(ids))
	for _, id := range ids {
		queue = append(queue, restorePair{parent: newParent, fileID: id})
	}

	actions := make([]history.FileAction, 0)
	restoredFiles := make([]*file_model.WeblensFileImpl, 0)

	// Actions created during a transaction automatically get an event ID and timestamp, so we don't need to set those manually here
	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			pastFile, err := journal.GetPastFileByID(ctx, current.fileID, restoreTime)
			if err != nil {
				return wlerrors.Wrapf(err, "failed to get past file [%s] at time [%s]", current.fileID, restoreTime)
			}

			filename := pastFile.GetPortablePath().Filename()

			destPath, err := MakeUniqueChildName(current.parent.GetPortablePath(), filename, pastFile.IsDir())
			if err != nil {
				return err
			}

			if pastFile.IsDir() {
				restoreDir, restoreAction, err := fs.restoreDirectory(ctx, pastFile, current.parent, destPath)
				if err != nil {
					return err
				}

				actions = append(actions, restoreAction)
				restoredFiles = append(restoredFiles, restoreDir)

				// Enqueue children for restoration
				for _, child := range pastFile.GetChildren() {
					queue = append(queue, restorePair{parent: restoreDir, fileID: child.ID()})
				}
			} else {
				restoredFile, restoreAction, err := fs.restoreRegularFile(ctx, pastFile, current.parent, destPath)
				if err != nil {
					return err
				}

				actions = append(actions, restoreAction)
				restoredFiles = append(restoredFiles, restoredFile)
			}
		}

		err := history.SaveActions(ctx, actions)
		if err != nil {
			return wlerrors.Wrap(err, "failed to save restore actions")
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = fs.ResizeUp(ctx, newParent)
	if err != nil {
		return err
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.New("failed to get app context")
	}

	parentInfo, err := reshape.WeblensFileToFileInfo(ctx, newParent)
	if err != nil {
		return err
	}

	notifs := notify.NewFileNotification(ctx, parentInfo, websocket_mod.FileUpdatedEvent)
	for _, f := range restoredFiles {
		fInfo, infoErr := reshape.WeblensFileToFileInfo(ctx, f)
		if infoErr != nil {
			continue
		}

		notifs = append(notifs, notify.NewFileNotification(ctx, fInfo, websocket_mod.FileCreatedEvent)...)
	}

	appCtx.Notify(ctx, notifs...)

	return nil
}

// RestoreHistory replays a series of file actions to restore file history.
func (fs *ServiceImpl) RestoreHistory(ctx context.Context, actions []*history.FileAction) error { //nolint:revive
	return wlerrors.New("not implemented")
}

// NewZip creates a new zip file for archiving purposes.
func (fs *ServiceImpl) NewZip(ctx context.Context, zipName string, owner *user_model.User) (*file_model.WeblensFileImpl, error) { //nolint:revive
	newZipPath := file_model.ZipsDirPath.Child(zipName, false)

	zipsDir, err := fs.GetFileByFilepath(ctx, file_model.ZipsDirPath)
	if err != nil {
		return nil, err
	}

	newZip := file_model.NewWeblensFile(file_model.NewFileOptions{Path: newZipPath, CreateNow: true, GenerateID: true})

	err = newZip.SetParent(zipsDir)
	if err != nil {
		return nil, err
	}

	err = fs.AddFile(ctx, newZip)
	if err != nil {
		return nil, err
	}

	return newZip, nil
}

// GetZip retrieves a zip file by its identifier.
func (fs *ServiceImpl) GetZip(ctx context.Context, id string) (*file_model.WeblensFileImpl, error) {
	zipPath := file_model.ZipsDirPath.Child(id, false)
	f, err := fs.GetFileByFilepath(ctx, zipPath)

	return f, err
}

// DeleteZips permanently deletes zip files from the takeout cache
func (fs *ServiceImpl) DeleteZips(ctx context.Context) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.WithStack(context_service.ErrNoContext)
	}

	zipFolder, err := fs.GetFileByFilepath(ctx, file_model.ZipsDirPath)
	if err != nil {
		return err
	}

	children, err := fs.GetChildren(ctx, zipFolder)
	if err != nil {
		return err
	}

	appCtx.Log().Debug().Msgf("Deleting [%d] zip files from takeout cache", len(children))

	err = fs.DeleteFiles(ctx, children...)
	if err != nil {
		return err
	}

	return nil
}

// RenameFile changes the name of a file and updates its path in the file service.
func (fs *ServiceImpl) RenameFile(ctx context.Context, file *file_model.WeblensFileImpl, newName string) error {
	if err := cryptography.ValidateFilename(newName); err != nil {
		return wlerrors.Errorf("invalid filename: %w", err)
	}

	parent := file.GetParent()
	if _, err := parent.GetChild(newName); err == nil {
		return wlerrors.WithStack(file_model.ErrFileAlreadyExists)
	}

	oldPath := file.GetPortablePath()
	newPath := oldPath.Dir().Child(newName, file.GetPortablePath().IsDir())

	err := rename(file.GetPortablePath(), newPath)
	if err != nil {
		return err
	}

	err = file.RecursiveMap(func(wfi *file_model.WeblensFileImpl) error {
		// Update the file's path to the new path
		newFilePath, err := wfi.GetPortablePath().ReplacePrefix(oldPath, newPath)
		if err != nil {
			return err
		}

		wfi.SetPortablePath(newFilePath)

		return nil
	})
	if err != nil {
		return err
	}

	err = parent.RemoveChild(oldPath.Filename())
	if err != nil {
		return err
	}

	err = parent.AddChild(file)
	if err != nil {
		return err
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.WithStack(context_service.ErrNoContext)
	}

	appCtx.Log().Debug().Msgf("Renaming file [%s] to [%s]", file.GetPortablePath(), newPath)

	fInfo, err := reshape.WeblensFileToFileInfo(ctx, file)
	if err != nil {
		return err
	}

	notif := notify.NewFileNotification(ctx, fInfo, websocket_mod.FileUpdatedEvent)
	appCtx.Notify(ctx, notif...)

	return nil
}

// GetChildren retrieves all child files of a directory, loading them if necessary.
func (fs *ServiceImpl) GetChildren(ctx context.Context, folder *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error) {
	if !folder.IsDir() {
		return nil, wlerrors.WithStack(file_model.ErrDirectoryRequired)
	}

	if !folder.ChildrenLoaded() {
		appCtx, ok := context_service.FromContext(ctx)
		if !ok {
			return nil, wlerrors.WithStack(context_service.ErrNoContext)
		}

		appCtx = appCtx.WithValue(doFileCreationContextKey{}, true)

		_, err := loadOneDirectory(appCtx, folder)
		if err != nil {
			return nil, err
		}
	}

	return folder.GetChildren(), nil
}

// RecursiveEnsureChildrenLoaded ensures all children are loaded for a directory and all its subdirectories.
func (fs *ServiceImpl) RecursiveEnsureChildrenLoaded(ctx context.Context, folder *file_model.WeblensFileImpl) error {
	if !folder.IsDir() {
		return wlerrors.WithStack(file_model.ErrDirectoryRequired)
	}

	err := folder.RecursiveMap(func(wfi *file_model.WeblensFileImpl) error {
		if wfi.IsDir() {
			_, err := fs.GetChildren(ctx, wfi)

			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// InitBackupDirectory initializes the backup directory for a tower instance.
func (fs *ServiceImpl) InitBackupDirectory(ctx context.Context, tower tower_model.Instance) (*file_model.WeblensFileImpl, error) {
	backupRoot, err := fs.GetFileByID(ctx, file_model.BackupTreeKey)
	if err != nil {
		return nil, err
	}

	backupDir, err := backupRoot.GetChild(tower.TowerID)
	if err == nil {
		return backupDir, nil
	}

	if !exists(backupRoot.GetPortablePath().Child(tower.TowerID, true)) {
		return mkdir(backupRoot.GetPortablePath().Child(tower.TowerID, true))
	}

	return file_model.NewWeblensFile(file_model.NewFileOptions{Path: backupRoot.GetPortablePath().Child(tower.TowerID, true)}), nil
}

// AddTask associates a task with a file for tracking purposes.
func (fs *ServiceImpl) AddTask(f *file_model.WeblensFileImpl, t *task_model.Task) error {
	fs.fileTaskLock.Lock()
	defer fs.fileTaskLock.Unlock()

	tasks, ok := fs.fileTaskLink[f.ID()]
	if !ok {
		tasks = []*task_model.Task{}
	} else if slices.Contains(tasks, t) {
		return file_model.ErrFileAlreadyHasTask
	}

	fs.fileTaskLink[f.ID()] = append(tasks, t)

	return nil
}

// RemoveTask removes a task from the file's task list.
func (fs *ServiceImpl) RemoveTask(f *file_model.WeblensFileImpl, t *task_model.Task) error {
	fs.fileTaskLock.Lock()
	defer fs.fileTaskLock.Unlock()

	tasks, ok := fs.fileTaskLink[f.ID()]
	if !ok {
		return file_model.ErrFileNoTask
	}

	i := slices.Index(tasks, t)
	if i == -1 {
		return file_model.ErrFileNoTask
	}

	fs.fileTaskLink[f.ID()] = slices.Delete(tasks, i, i+1)

	return nil
}

// GetTasks returns all tasks associated with a file.
func (fs *ServiceImpl) GetTasks(f *file_model.WeblensFileImpl) []*task_model.Task {
	fs.fileTaskLock.RLock()
	defer fs.fileTaskLock.RUnlock()

	return fs.fileTaskLink[f.ID()]
}

// ResizeUp updates the size of parent directories when a file changes.
func (fs *ServiceImpl) ResizeUp(ctx context.Context, f *file_model.WeblensFileImpl) error {
	// ctx.Log().Trace().Msgf("Resizing up [%s]", f.GetPortablePath())
	// tree := fs.trees[f.GetPortablePath().RootName()]
	// if tree == nil {
	// 	return nil
	// }
	//
	// err := tree.ResizeUp(f, event, nil)
	//
	// if err != nil {
	// 	return err
	// }
	//
	// return nil
	f, err := fs.GetFileByID(ctx, file_model.UsersTreeKey)
	if err != nil {
		return err
	}

	f.Size()

	return nil
}

// CreateUserHome initializes a home directory and trash for a new user.
func (fs *ServiceImpl) CreateUserHome(ctx context.Context, user *user_model.User) error {
	if user.Username == "" {
		return wlerrors.Wrapf(user_model.ErrUsernameTooShort, "failed to create home folder for user")
	}

	parent, err := fs.GetFileByID(ctx, file_model.UsersTreeKey)
	if err != nil {
		return err
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.ReplaceStack(context_service.ErrNoContext)
	}

	appCtx = appCtx.WithValue(doFileCreationContextKey{}, true)

	home, err := fs.CreateFolder(ctx, parent, user.GetUsername())
	if wlerrors.Is(err, file_model.ErrDirectoryAlreadyExists) {
		home, err = fs.GetFileByFilepath(appCtx, file_model.UsersRootPath.Child(user.GetUsername(), true))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	user.HomeID = home.ID()

	trash, err := fs.CreateFolder(appCtx, home, file_model.UserTrashDirName)
	if wlerrors.Is(err, file_model.ErrDirectoryAlreadyExists) {
		trash, err = fs.GetFileByFilepath(appCtx, home.GetPortablePath().Child(file_model.UserTrashDirName, true))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	user.TrashID = trash.ID()

	err = fs.AddFile(appCtx, home, trash)
	if err != nil {
		return wlerrors.Wrapf(err, "failed to add home and trash for user [%s]", user.GetUsername())
	}

	return nil
}

func (fs *ServiceImpl) getFileInternal(id string) (*file_model.WeblensFileImpl, bool) {
	fs.treeLock.RLock()
	defer fs.treeLock.RUnlock()

	f, ok := fs.files[id]

	return f, ok
}

func (fs *ServiceImpl) setFileInternal(id string, f *file_model.WeblensFileImpl) {
	fs.treeLock.Lock()
	defer fs.treeLock.Unlock()

	fs.files[id] = f
}

// SkipJournalKey can be set in the context to skip journaling for file operations.
const SkipJournalKey = "skipJournal"

// createCommon contains common logic for creating files and folders.
// This includes setting the parent, adding to the parent's children, journaling the creation action, and notifying listeners.
func (fs *ServiceImpl) createCommon(ctx context.Context, newF, parent *file_model.WeblensFileImpl) error {
	err := newF.SetParent(parent)
	if err != nil {
		return err
	}

	err = parent.AddChild(newF)
	if err != nil {
		return err
	}

	action := history.NewCreateAction(ctx, newF)

	_, ok := ctx.Value(SkipJournalKey).(bool)
	if !ok {
		err = history.SaveAction(ctx, &action)
		if err != nil {
			return err
		}
	}

	newF.SetID(action.FileID)

	err = fs.AddFile(ctx, newF)
	if err != nil {
		return err
	}

	notifier, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.New("failed to get notifier from context")
	}

	fInfo, err := reshape.WeblensFileToFileInfo(ctx, newF)
	if err != nil {
		return err
	}

	notif := notify.NewFileNotification(ctx, fInfo, websocket_mod.FileCreatedEvent)
	notifier.Notify(ctx, notif...)

	context_mod.ToZ(ctx).Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Created file at [%s] with id [%s]", newF.GetPortablePath().ToAbsolute(), newF.ID())
	})

	return nil
}

func (fs *ServiceImpl) moveFilesWithTransaction(ctx context.Context, files []*file_model.WeblensFileImpl, destFolder *file_model.WeblensFileImpl) error {
	if len(files) == 0 {
		return nil
	}

	notifier, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.New("failed to get notifier from context")
	}

	oldParents := []*file_model.WeblensFileImpl{}

	for _, file := range files {
		oldPath := file.GetPortablePath()

		newPath, err := MakeUniqueChildName(destFolder.GetPortablePath(), file.GetPortablePath().Filename(), file.IsDir())
		if err != nil {
			return err
		}

		// Remove the file from the old parent
		oldParent := file.GetParent()

		err = oldParent.RemoveChild(file.GetPortablePath().Filename())
		if err != nil {
			return err
		}

		actions := []history.FileAction{}

		_, skipJournal := ctx.Value(SkipJournalKey).(bool)

		err = file.RecursiveMap(func(wfi *file_model.WeblensFileImpl) error {
			if wfi == file {
				file.SetPortablePath(newPath)

				return nil
			}

			descendantOldPath := wfi.GetPortablePath()

			descendantNewPath, err := wfi.GetPortablePath().ReplacePrefix(oldParent.GetPortablePath().Child(oldPath.Filename(), true), destFolder.GetPortablePath().Child(file.GetPortablePath().Filename(), true))
			if err != nil {
				return err
			}

			if !skipJournal {
				action := history.NewMoveAction(ctx, descendantOldPath, descendantNewPath, wfi)
				actions = append(actions, action)
			}

			wfi.SetPortablePath(descendantNewPath)

			return nil
		})
		if err != nil {
			return err
		}

		if !skipJournal {
			action := history.NewMoveAction(ctx, oldPath, newPath, file)
			actions = append(actions, action)

			err = history.SaveActions(ctx, actions)
			if err != nil {
				return err
			}
		}

		err = rename(oldPath, newPath)
		if err != nil {
			return err
		}

		// Add the file to the new parent
		err = destFolder.AddChild(file)
		if err != nil {
			return err
		}

		err = file.SetParent(destFolder)
		if err != nil {
			return err
		}

		if !slices.ContainsFunc(oldParents, func(p *file_model.WeblensFileImpl) bool {
			return p.ID() == oldParent.ID()
		}) {
			oldParents = append(oldParents, oldParent)
		}

		fInfo, err := reshape.WeblensFileToFileInfo(ctx, file)
		if err != nil {
			return err
		}

		notif := notify.NewFileNotification(ctx, fInfo, websocket_mod.FileUpdatedEvent, notify.FileNotificationOptions{PreMoveParentID: oldParent.ID()})
		notifier.Notify(ctx, notif...)
	}

	for _, oldParent := range oldParents {
		// Notify the old parent that it has been updated
		oldParentInfo, err := reshape.WeblensFileToFileInfo(ctx, oldParent)
		if err != nil {
			return err
		}

		notif := notify.NewFileNotification(ctx, oldParentInfo, websocket_mod.FileUpdatedEvent)
		notifier.Notify(ctx, notif...)
	}

	destFolder.Size()

	destFInfo, err := reshape.WeblensFileToFileInfo(ctx, destFolder)
	if err != nil {
		return err
	}

	notif := notify.NewFileNotification(ctx, destFInfo, websocket_mod.FileUpdatedEvent)
	notifier.Notify(ctx, notif...)

	return nil
}

func (fs *ServiceImpl) deleteFilesWithTransaction(ctx context.Context, files []*file_model.WeblensFileImpl) error {
	if len(files) == 0 {
		return nil
	}

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	if local.Role == tower_model.RoleBackup {
		for _, file := range files {
			err = remove(file.GetPortablePath())
			if err != nil {
				return err
			}
		}

		return nil
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.New("failed to get app context from context")
	}

	// All files *should* share the same parent: the trash folder, so pulling
	// just the first one to do the update on will work fine.
	trash := files[0].GetParent()

	actions := []history.FileAction{}
	parents := make(map[string][]string)
	notifs := []websocket_mod.WsResponseInfo{}

	for _, file := range files {
		err := file.RecursiveMap(
			func(f *file_model.WeblensFileImpl) error {
				err = rmFileMedia(ctx, f)
				if err != nil {
					return err
				}

				newAction := history.NewDeleteAction(ctx, f)
				actions = append(actions, newAction)

				if !f.IsDir() && f.Size() != 0 {
					err = linkToRestore(ctx, f)
					if err != nil {
						return err
					}
				}

				fInfo, err := reshape.WeblensFileToFileInfo(ctx, f)
				if err != nil {
					return err
				}

				notifs = append(notifs, notify.NewFileNotification(ctx, fInfo, websocket_mod.FileDeletedEvent)...)

				return nil
			},
		)
		if err != nil {
			return err
		}

		parentID := file.GetParent().ID()
		if _, ok := parents[parentID]; !ok {
			parents[parentID] = []string{}
		}

		parents[parentID] = append(parents[parentID], file.ID())

		err = file.GetParent().RemoveChild(file.GetPortablePath().Filename())
		if err != nil {
			return err
		}

		err = remove(file.GetPortablePath())
		if err != nil {
			return err
		}

		err = file.RecursiveMap(func(wfi *file_model.WeblensFileImpl) error {
			err = fs.removeFileByID(ctx, wfi.ID())
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	appCtx.Notify(ctx, notifs...)

	err = history.SaveActions(ctx, actions)
	if err != nil {
		return err
	}

	err = fs.ResizeUp(ctx, trash)
	if err != nil {
		return err
	}

	trashInfo, err := reshape.WeblensFileToFileInfo(ctx, trash)
	if err != nil {
		return err
	}

	notif := notify.NewFileNotification(ctx, trashInfo, websocket_mod.FileUpdatedEvent)
	appCtx.Notify(ctx, notif...)

	return nil
}

// restoreDirectory creates a new directory in the destination for a past directory file.
func (fs *ServiceImpl) restoreDirectory(ctx context.Context, pastFile, parent *file_model.WeblensFileImpl, destPath file_system.Filepath) (restoreDir *file_model.WeblensFileImpl, action history.FileAction, err error) {
	restoreDir, err = mkdir(destPath)
	if err != nil {
		return
	}

	err = restoreDir.SetParent(parent)
	if err != nil {
		return
	}

	err = parent.AddChild(restoreDir)
	if err != nil {
		return
	}

	action = history.NewRestoreAction(ctx, restoreDir, pastFile.ID())

	err = fs.AddFile(ctx, restoreDir)
	if err != nil {
		return
	}

	return
}

// restoreRegularFile hard-links content from the RESTORE tree (or live USERS tree) into the destination.
func (fs *ServiceImpl) restoreRegularFile(ctx context.Context, pastFile, parent *file_model.WeblensFileImpl, destPath file_system.Filepath) (restoreFile *file_model.WeblensFileImpl, action history.FileAction, err error) {
	contentID := pastFile.GetContentID()
	if contentID == "" {
		return nil, action, wlerrors.Errorf("cannot restore file [%s]: no content ID recorded in history", pastFile.ID())
	}

	sourcePath, err := fs.findRestoreSource(ctx, pastFile.ID(), contentID)
	if err != nil {
		return nil, action, wlerrors.Wrapf(err, "failed to find restore source for file [%s]", pastFile.ID())
	}

	restoreFile = file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:      destPath,
		ContentID: contentID,
		Size:      pastFile.Size(),
	})

	err = restoreFile.SetParent(parent)
	if err != nil {
		return nil, action, err
	}

	err = parent.AddChild(restoreFile)
	if err != nil {
		return nil, action, err
	}

	// Hard-link the content from the source to the new destination
	err = os.Link(sourcePath, destPath.ToAbsolute())
	if err != nil {
		return nil, action, wlerrors.Wrapf(err, "failed to hard-link restore content from [%s] to [%s]", sourcePath, destPath.ToAbsolute())
	}

	action = history.NewRestoreAction(ctx, restoreFile, pastFile.ID())

	err = fs.AddFile(ctx, restoreFile)
	if err != nil {
		return nil, action, err
	}

	return restoreFile, action, nil
}

// findRestoreSource locates the content for a file being restored,
// checking the RESTORE tree first, then falling back to the live USERS tree.
func (fs *ServiceImpl) findRestoreSource(ctx context.Context, fileID, contentID string) (string, error) {
	restorePath := file_system.BuildFilePath(file_model.RestoreTreeKey, contentID)
	if exists(restorePath) {
		return restorePath.ToAbsolute(), nil
	}

	// Fall back to the live file if it still exists
	liveFile, err := fs.GetFileByID(ctx, fileID)
	if err == nil && liveFile != nil {
		livePath := liveFile.GetPortablePath().ToAbsolute()
		if _, statErr := os.Stat(livePath); statErr == nil {
			return livePath, nil
		}
	}

	return "", wlerrors.Errorf("content [%s] not found in restore tree or live tree for file [%s]", contentID, fileID)
}
