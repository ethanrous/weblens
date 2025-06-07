package file

import (
	"context"
	"fmt"
	"net/http"
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
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	file_system "github.com/ethanrous/weblens/modules/fs"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/rs/zerolog"
)

var _ file_model.FileService = &FileServiceImpl{}

type FileServiceImpl struct {
	contentIdCache map[string]*file_model.WeblensFileImpl
	contentIdLock  sync.RWMutex
	fileTaskLink   map[string][]*task_model.Task
	fileTaskLock   sync.RWMutex
	files          map[string]*file_model.WeblensFileImpl
	treeLock       sync.RWMutex
}

type FolderCoverPair struct {
	FolderId  string `bson:"folderId"`
	ContentId string `bson:"coverId"`
}

func NewFileService(
	ctx context.Context,
) (*FileServiceImpl, error) {
	fs := &FileServiceImpl{
		fileTaskLink: make(map[string][]*task_model.Task),
		files:        make(map[string]*file_model.WeblensFileImpl),
	}

	return fs, nil
}

func (fs *FileServiceImpl) Size(treeAlias string) int64 {
	// tree := fs.trees[treeAlias]
	// if tree == nil {
	// 	return -1
	// }
	//
	// return tree.GetRoot().Size()

	return -1
}

func (fs *FileServiceImpl) AddFile(c context.Context, files ...*file_model.WeblensFileImpl) (err error) {
	ctx, ok := context_service.FromContext(c)
	if !ok {
		return errors.New("failed to get context from context")
	}

	for _, f := range files {
		newId := f.ID()
		if newId == "" {
			return errors.WithStack(file_model.ErrNoFileId)
		} else if !f.IsDir() && f.Size() != 0 && f.GetContentId() == "" && f.GetPortablePath().RootName() == file_model.UsersTreeKey {
			return errors.Wrapf(file_model.ErrNoContentId, "failed to add [%s] to file service", f.GetPortablePath())
		}

		p := f.GetParent()
		if p == nil {
			return errors.WithStack(file_model.ErrNoParent)
		}

		if _, err = p.GetChild(f.GetPortablePath().Filename()); err != nil {
			err = p.AddChild(f)
			if err != nil {
				return err
			}
		}

		if _, exists := fs.getFileInternal(f.ID()); exists {
			ctx.Log().Warn().CallerSkipFrame(1).Msgf("File [%s] already exists in file service, skipping", f.GetPortablePath())

			continue
		}

		fs.setFileInternal(f.ID(), f)
	}

	return nil
}

func (fs *FileServiceImpl) GetFileById(ctx context.Context, id string) (*file_model.WeblensFileImpl, error) {
	f, ok := fs.getFileInternal(id)

	if ok {
		return f, nil
	}

	path, err := journal.GetLatestPathById(ctx, id)
	if err != nil {
		return nil, errors.WrapStatus(http.StatusNotFound, errors.Wrap(file_model.ErrFileNotFound, err.Error()))
	}

	return fs.GetFileByFilepath(ctx, path)
}

func (fs *FileServiceImpl) GetFileByFilepath(ctx context.Context, filepath file_system.Filepath, dontLoadNew ...bool) (*file_model.WeblensFileImpl, error) {
	root, err := fs.GetFileById(ctx, filepath.RootName())
	if err != nil {
		return nil, fmt.Errorf("failed to get root file [%s]: %w", filepath.RootName(), err)
	}

	childFile := root

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, errors.WithStack(context_service.ErrNoContext)
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
			appCtx.Log().Debug().Msgf("Loading children for %s [%s]", childFile.GetPortablePath(), childFile.ID())

			_, err = loadOneDirectory(appCtx, childFile, nil)
			if err != nil {
				return nil, err
			}
		} else if !shouldLoadNew {
			return nil, file_model.ErrFileNotFound
		}

		appCtx.Log().Debug().Msgf("Getting child of %s [%s]", childFile.GetPortablePath(), childFile.ID())

		childFile, err = childFile.GetChild(child)
		if err != nil {
			return nil, err
		}
	}

	return childFile, nil
}

func (fs *FileServiceImpl) GetMediaCacheByFilename(ctx context.Context, thumbFileName string) (*file_model.WeblensFileImpl, error) {
	f := file_model.NewWeblensFile(file_model.NewFileOptions{Path: file_model.ThumbsDirPath.Child(thumbFileName, false)})
	if !f.Exists() {
		return nil, errors.WithStack(file_model.ErrFileNotFound)
	}

	return f, nil
}

func (fs *FileServiceImpl) NewCacheFile(mediaId, quality string, pageNum int) (*file_model.WeblensFileImpl, error) {
	switch media_model.MediaQuality(quality) {
	case media_model.LowRes, media_model.HighRes:
		break
	default:

		return nil, errors.New("invalid quality")
	}

	filename := getCacheFilename(mediaId, quality, pageNum)

	childPath := file_model.ThumbsDirPath.Child(filename, false)

	return touch(childPath)
}

func (fs *FileServiceImpl) DeleteCacheFile(f *file_model.WeblensFileImpl) error {
	if !isCacheFile(f.GetPortablePath()) {
		return errors.New("trying to delete non-cache file")
	}

	return remove(f.GetPortablePath())
}

func (fs *FileServiceImpl) CreateFile(ctx context.Context, parent *file_model.WeblensFileImpl, filename string, data ...[]byte) (
	*file_model.WeblensFileImpl, error,
) {
	childPath := parent.GetPortablePath().Child(filename, false)

	newF, err := touch(childPath)
	if err != nil {
		return nil, err
	}

	for _, d := range data {
		_, err = newF.Write(d)
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

func (fs *FileServiceImpl) CreateFolder(ctx context.Context, parent *file_model.WeblensFileImpl, folderName string) (*file_model.WeblensFileImpl, error) {
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

func (fs *FileServiceImpl) ReturnFilesFromTrash(ctx context.Context, trashFiles []*file_model.WeblensFileImpl) error {
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
	// 	oldParent := tree.Get(justBeforeTrash.ParentId)
	// 	if oldParent == nil {
	// 		owner, err := fs.GetFileOwner(ctx, trashEntry)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		oldParent = tree.Get(owner.HomeId)
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

	return errors.New("not implemented")
}

func (fs *FileServiceImpl) MoveFiles(ctx context.Context, files []*file_model.WeblensFileImpl, destFolder *file_model.WeblensFileImpl) error {
	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		return fs.moveFilesWithTransaction(ctx, files, destFolder)
	})

	if err != nil {
		return err
	}

	return nil
}

// DeleteFiles removes files being pointed to from the tree and moves them to the restore tree.
func (fs *FileServiceImpl) DeleteFiles(ctx context.Context, files ...*file_model.WeblensFileImpl) error {
	for _, f := range files {
		if f.GetPortablePath().Dir().IsRoot() {
			return errors.Errorf("cannot delete user home directory [%s]", f.GetPortablePath())
		} else if f.GetPortablePath().Filename() == file_model.UserTrashDirName {
			return errors.Errorf("cannot delete user trash directory [%s]", f.GetPortablePath())
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

func (fs *FileServiceImpl) RestoreFiles(ctx context.Context, ids []string, newParent *file_model.WeblensFileImpl, restoreTime time.Time) error {
	// event := journal.NewEvent()
	//
	// var topFiles []*file_model.WeblensFileImpl
	// type restorePair struct {
	// 	newParent *file_model.WeblensFileImpl
	// 	fileId    string
	// 	contentId string
	// }
	//
	// var restorePairs []restorePair
	// // for _, id := range ids {
	// // 	journal
	// // 	lt := journal.Get(id)
	// // 	if lt == nil {
	// // 		return errors.Errorf("journal does not have file to restore")
	// // 	}
	// // 	restorePairs = append(
	// // 		restorePairs, restorePair{fileId: id, newParent: newParent, contentId: lt.ContentId},
	// // 	)
	// // }
	//
	// for len(restorePairs) != 0 {
	// 	toRestore := restorePairs[0]
	// 	restorePairs = restorePairs[1:]
	//
	// 	pastFile, err := journal.GetPastFile(toRestore.fileId, restoreTime)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	var childIds []string
	// 	if pastFile.IsDir() {
	// 		children := pastFile.GetChildren()
	//
	// 		childIds = wl_slices.Map(
	// 			children, func(child *file_model.WeblensFileImpl) string {
	// 				return child.ID()
	// 			},
	// 		)
	// 	}
	//
	// 	path := pastFile.GetPortablePath().ToPortable()
	// 	if path == "" {
	// 		return errors.Errorf("Got empty string for portable path on past file [%s]", pastFile.AbsPath())
	// 	}
	// 	// Paths of directory files will have an extra / on the end, so we need to remove it
	// 	if pastFile.IsDir() {
	// 		path = path[:len(path)-1]
	// 	}
	//
	// 	oldName := filepath.Base(path)
	// 	newName := MakeUniqueChildName(toRestore.newParent, oldName)
	//
	// 	var restoredF *file_model.WeblensFileImpl
	// 	if !pastFile.IsDir() {
	// 		var existingPath string
	//
	// 		// File has been deleted, get the file from the restore tree
	// 		if liveF := usersTree.Get(toRestore.fileId); liveF == nil {
	// 			_, err = restoreTree.GetRoot().GetChild(toRestore.contentId)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			existingPath = filepath.Join(restoreTree.GetRoot().AbsPath(), toRestore.contentId)
	// 		} else {
	// 			existingPath = liveF.AbsPath()
	// 		}
	//
	// 		restoredF = file_model.NewWeblensFile(
	// 			usersTree.GenerateFileId(), newName, toRestore.newParent, pastFile.IsDir(),
	// 		)
	// 		restoredF.SetContentId(pastFile.GetContentId())
	// 		restoredF.SetSize(pastFile.Size())
	// 		err = usersTree.Add(restoredF)
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Restoring file [%s] to [%s]", existingPath, restoredF.AbsPath()) })
	// 		err = os.Link(existingPath, restoredF.AbsPath())
	// 		if err != nil {
	// 			return errors.WithStack(err)
	// 		}
	//
	// 		if toRestore.newParent == newParent {
	// 			topFiles = append(topFiles, restoredF)
	// 		}
	//
	// 	} else {
	// 		restoredF = file_model.NewWeblensFile(
	// 			usersTree.GenerateFileId(), newName, toRestore.newParent, true,
	// 		)
	// 		err = usersTree.Add(restoredF)
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		err = restoredF.CreateSelf()
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		for _, childId := range childIds {
	// 			childLt := journal.Get(childId)
	// 			if childLt == nil {
	// 				return errors.Wrap(file_model.ErrFileNotFound, childId)
	// 			}
	// 			restorePairs = append(
	// 				restorePairs,
	// 				restorePair{fileId: childId, newParent: restoredF, contentId: childLt.GetContentId()},
	// 			)
	// 		}
	//
	// 		if toRestore.newParent == newParent {
	// 			topFiles = append(topFiles, restoredF)
	// 		}
	// 	}
	//
	// 	event.NewRestoreAction(restoredF)
	// }
	//
	// for _, f := range topFiles {
	// 	err := fs.ResizeDown(ctx, f, event)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = fs.ResizeUp(ctx, f, event)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	// journal.LogEvent(event)
	// err := event.Wait()
	// if err != nil {
	// 	return err
	// }
	//
	// return nil

	return errors.New("not implemented")
}

func (fs *FileServiceImpl) RestoreHistory(ctx context.Context, actions []*history.FileAction) error {
	return errors.New("not implemented")

	// journal := fs.trees[file_model.UsersTreeKey].GetJournal()
	//
	// err := journal.Add(lifetimes...)
	// if err != nil {
	// 	return err
	// }
	//
	// slices.SortFunc(lifetimes, history.LifetimeSorter)
	//
	// for _, lt := range lifetimes {
	// 	latest := lt.GetLatestAction()
	// 	if latest.GetActionType() == history.FileDelete {
	// 		continue
	// 	}
	// 	portable, err := file_system.ParsePortable(latest.GetDestinationPath())
	// 	if err != nil {
	// 		ctx.Log().Error().Stack().Err(err).Msg("Failed to parse portable path")
	// 		continue
	// 	}
	// 	if !portable.IsDir() {
	// 		continue
	// 	}
	// 	if fs.trees[file_model.UsersTreeKey].Get(lt.ID()) != nil {
	// 		continue
	// 	}
	//
	// 	// parentId := latest.GetParentId()
	// 	parent, err := fs.getFileByIdAndRoot(latest.GetParentId(), file_model.UsersTreeKey)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	newF := file_model.NewWeblensFile(file_model.NewFileOptions{Path: lt.ID(}), portable.Filename(), parent, true)
	// 	err = fs.trees[file_model.UsersTreeKey].Add(newF)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = newF.CreateSelf()
	// 	if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
	// 		return err
	// 	}
	// }
	//
	// return nil
}

func (fs *FileServiceImpl) NewZip(ctx context.Context, zipName string, owner *user_model.User) (*file_model.WeblensFileImpl, error) {
	newZipPath := file_model.ZipsDirPath.Child(zipName, false)

	zipsDir, err := fs.GetFileByFilepath(ctx, file_model.ZipsDirPath)
	if err != nil {
		return nil, err
	}

	newZip := file_model.NewWeblensFile(file_model.NewFileOptions{Path: newZipPath, CreateNow: true, GenerateId: true})

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

func (fs *FileServiceImpl) GetZip(ctx context.Context, id string) (*file_model.WeblensFileImpl, error) {
	zipPath := file_model.ZipsDirPath.Child(id, false)
	f, err := fs.GetFileByFilepath(ctx, zipPath)

	return f, err
}

func (fs *FileServiceImpl) RenameFile(file *file_model.WeblensFileImpl, newName string) error {
	// preFile := file.Freeze()
	err := rename(file.GetPortablePath(), file.GetPortablePath().Dir().Child(newName, false))
	if err != nil {
		return err
	}

	// TODO: Implement caster calls
	// caster.PushFileMove(preFile, file)

	return nil
}

func (fs *FileServiceImpl) GetChildren(ctx context.Context, folder *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error) {
	if !folder.IsDir() {
		return nil, errors.WithStack(file_model.ErrDirectoryRequired)
	}

	if !folder.ChildrenLoaded() {
		appCtx, ok := context_service.FromContext(ctx)
		if !ok {
			return nil, errors.WithStack(context_service.ErrNoContext)
		}

		appCtx.Log().Debug().Msgf("Loading children for folder [%s]", folder.GetPortablePath())

		appCtx = appCtx.WithValue(doFileCreationContextKey{}, true)
		_, err := loadOneDirectory(appCtx, folder, nil)
		if err != nil {
			return nil, err
		}
	}

	return folder.GetChildren(), nil
}

func (fs *FileServiceImpl) InitBackupDirectory(ctx context.Context, tower tower_model.Instance) (*file_model.WeblensFileImpl, error) {
	backupRoot, err := fs.GetFileById(ctx, file_model.BackupTreeKey)
	if err != nil {
		return nil, err
	}

	backupDir, err := backupRoot.GetChild(tower.TowerId)
	if err == nil {
		return backupDir, nil
	}

	if !exists(backupRoot.GetPortablePath().Child(tower.TowerId, true)) {
		return mkdir(backupRoot.GetPortablePath().Child(tower.TowerId, true))
	}

	return file_model.NewWeblensFile(file_model.NewFileOptions{Path: backupRoot.GetPortablePath().Child(tower.TowerId, true)}), nil
}

func (fs *FileServiceImpl) AddTask(f *file_model.WeblensFileImpl, t *task_model.Task) error {
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

func (fs *FileServiceImpl) RemoveTask(f *file_model.WeblensFileImpl, t *task_model.Task) error {
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

func (fs *FileServiceImpl) GetTasks(f *file_model.WeblensFileImpl) []*task_model.Task {
	fs.fileTaskLock.RLock()
	defer fs.fileTaskLock.RUnlock()

	return fs.fileTaskLink[f.ID()]
}

func (fs *FileServiceImpl) ResizeUp(ctx context.Context, f *file_model.WeblensFileImpl) error {
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

	f, err := fs.GetFileById(ctx, file_model.UsersTreeKey)
	if err != nil {
		return err
	}

	f.Size()

	return nil
}

func (fs *FileServiceImpl) ResizeDown(ctx context.Context, f *file_model.WeblensFileImpl) error {
	// ctx.Log().Trace().Msgf("Resizing down [%s]", f.GetPortablePath())
	// tree := fs.trees[f.GetPortablePath().RootName()]
	// if tree == nil {
	// 	return errors.WithStack(file_model.ErrFileTreeNotFound)
	// }
	//
	// err := tree.ResizeDown(f, event, nil)
	//
	// if err != nil {
	// 	return err
	// }
	//
	// ctx.Log().Trace().Func(func(e *zerolog.Event) {
	// 	if event == nil {
	// 		return
	// 	}
	// 	e.Msgf("Resizing down event: %d", len(event.Actions))
	// })
	//
	// return nil
	f.Size()

	return nil
}

func (fs *FileServiceImpl) CreateUserHome(ctx context.Context, user *user_model.User) error {
	parent, err := fs.GetFileById(ctx, file_model.UsersTreeKey)
	if err != nil {
		return err
	}

	home, err := fs.CreateFolder(ctx, parent, user.GetUsername())
	if errors.Is(err, file_model.ErrDirectoryAlreadyExists) {
		home, err = fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child(user.GetUsername(), true))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	user.HomeId = home.ID()

	trash, err := fs.CreateFolder(ctx, home, file_model.UserTrashDirName)
	if errors.Is(err, file_model.ErrDirectoryAlreadyExists) {
		trash, err = fs.GetFileByFilepath(ctx, home.GetPortablePath().Child(file_model.UserTrashDirName, true))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	user.TrashId = trash.ID()

	err = fs.AddFile(ctx, home, trash)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) getFileInternal(id string) (*file_model.WeblensFileImpl, bool) {
	fs.treeLock.RLock()
	defer fs.treeLock.RUnlock()

	f, ok := fs.files[id]

	return f, ok
}

func (fs *FileServiceImpl) setFileInternal(id string, f *file_model.WeblensFileImpl) {
	fs.treeLock.Lock()
	defer fs.treeLock.Unlock()
	fs.files[id] = f
}

const SkipJournalKey = "skipJournal"

func (fs *FileServiceImpl) createCommon(ctx context.Context, newF, parent *file_model.WeblensFileImpl) error {
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

	newF.SetId(action.FileId)

	err = fs.AddFile(ctx, newF)
	if err != nil {
		return err
	}

	notifier, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("failed to get notifier from context")
	}

	notif := notify.NewFileNotification(ctx, newF, websocket_mod.FileCreatedEvent)
	notifier.Notify(ctx, notif...)

	context_mod.ToZ(ctx).Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Created file at [%s] with id [%s]", newF.GetPortablePath().ToAbsolute(), newF.ID())
	})

	return nil
}

func (fs *FileServiceImpl) moveFilesWithTransaction(ctx context.Context, files []*file_model.WeblensFileImpl, destFolder *file_model.WeblensFileImpl) error {
	if len(files) == 0 {
		return nil
	}

	notifier, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("failed to get notifier from context")
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
	}

	for _, oldParent := range oldParents {
		// Notify the old parent that it has been updated
		notif := notify.NewFileNotification(ctx, oldParent, websocket_mod.FileUpdatedEvent)
		notifier.Notify(ctx, notif...)
	}

	destFolder.Size()

	notif := notify.NewFileNotification(ctx, destFolder, websocket_mod.FileUpdatedEvent)
	notifier.Notify(ctx, notif...)

	return nil
}

func (fs *FileServiceImpl) deleteFilesWithTransaction(ctx context.Context, files []*file_model.WeblensFileImpl) error {
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
		return errors.New("failed to get app context from context")
	}

	// All files *should* share the same parent: the trash folder, so pulling
	// just the first one to do the update on will work fine.
	trash := files[0].GetParent()

	actions := []history.FileAction{}
	parents := make(map[string][]string)

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
					if f.GetContentId() == "" {
						return errors.Errorf("cannot move file [%s] to restore tree without content id", f.GetPortablePath())
					} else {
						err = linkToRestore(ctx, f)
						if err != nil {
							return err
						}
					}
				}

				return nil
			},
		)
		if err != nil {
			return err
		}

		parentId := file.GetParent().ID()
		if _, ok := parents[parentId]; !ok {
			parents[parentId] = []string{}
		}

		parents[parentId] = append(parents[parentId], file.ID())

		err = file.GetParent().RemoveChild(file.GetPortablePath().Filename())
		if err != nil {
			return err
		}

		err = remove(file.GetPortablePath())
		if err != nil {
			return err
		}

		err = file.RecursiveMap(func(wfi *file_model.WeblensFileImpl) error {
			err = fs.removeFileById(ctx, wfi.ID())
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	// notif := notify.NewFileNotification(ctx, file, websocket_mod.FileDeletedEvent)
	// appCtx.Notify(ctx, notif...)

	err = history.SaveActions(ctx, actions)
	if err != nil {
		return err
	}

	err = fs.ResizeUp(ctx, trash)
	if err != nil {
		return err
	}

	notif := notify.NewFileNotification(ctx, trash, websocket_mod.FileUpdatedEvent)
	appCtx.Notify(ctx, notif...)

	return nil
}
