package file

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/fs"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var _ file_model.FileService = &FileServiceImpl{}

type FileServiceImpl struct {
	files map[file_system.Filepath]*file_model.WeblensFileImpl

	contentIdCache map[string]*file_model.WeblensFileImpl

	fileTaskLink map[string][]*task_model.Task

	treesLock sync.RWMutex

	contentIdLock sync.RWMutex

	fileTaskLock sync.RWMutex
}

type FolderCoverPair struct {
	FolderId  string `bson:"folderId"`
	ContentId string `bson:"coverId"`
}

func NewFileService(
	ctx context.ContextZ,
	logger *zerolog.Logger,
) (*FileServiceImpl, error) {

	fs := &FileServiceImpl{
		fileTaskLink: make(map[string][]*task_model.Task),
	}

	// for _, tree := range trees {
	// 	fs.trees[tree.GetRoot().GetPortablePath().RootName()] = tree
	// }
	//
	// if usersTree, ok := fs.trees[UsersTreeKey]; ok {
	// 	err := fs.ResizeDown(ctx, usersTree.GetRoot(), nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

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

func (fs *FileServiceImpl) GetFileById(id string) (*file_model.WeblensFileImpl, error) {
	return nil, errors.New("not implemented")
}

func (fs *FileServiceImpl) GetFileByFilepath(filepath fs.Filepath) (*file_model.WeblensFileImpl, error) {
	return nil, errors.New("not implemented")
}

func (fs *FileServiceImpl) GetFileByContentId(contentId string) (*file_model.WeblensFileImpl, error) {
	if fs.contentIdCache == nil {
		err := fs.loadContentIdCache()
		if err != nil {
			return nil, err
		}
	}

	fs.contentIdLock.RLock()
	if f, ok := fs.contentIdCache[contentId]; ok {
		fs.contentIdLock.RUnlock()
		return f, nil
	}
	fs.contentIdLock.RUnlock()

	err := fs.loadContentIdCache()
	if err != nil {
		return nil, err
	}

	fs.contentIdLock.RLock()
	if f, ok := fs.contentIdCache[contentId]; ok {
		fs.contentIdLock.RUnlock()
		return f, nil
	}
	fs.contentIdLock.RUnlock()

	return nil, errors.WithStack(file_model.ErrFileNotFound)
}

func (fs *FileServiceImpl) GetFiles(ids []string) ([]*file_model.WeblensFileImpl, []string, error) {
	// var files []*file_model.WeblensFileImpl
	// var lostFiles []string
	// for _, id := range ids {
	// 	lt := usersTree.GetJournal().Get(id)
	// 	if lt == nil {
	// 		lostFiles = append(lostFiles, id)
	// 		continue
	// 		// return nil, nil, errors.WithStack(werror.ErrNoLifetime.WithArg(id))
	// 	}
	// 	if lt.GetLatestAction().ActionType == history.FileDelete {
	// 		contentId := lt.GetContentId()
	// 		if contentId == "" {
	// 			lostFiles = append(lostFiles, id)
	// 			continue
	// 		}
	// 		f, err := fs.trees[RestoreTreeKey].GetRoot().GetChild(contentId)
	// 		if err != nil {
	// 			lostFiles = append(lostFiles, id)
	// 			continue
	// 		}
	// 		files = append(files, f)
	// 	} else {
	// 		f := usersTree.Get(id)
	// 		if f == nil {
	// 			lostFiles = append(lostFiles, id)
	// 			continue
	// 		}
	// 		files = append(files, f)
	// 	}
	// }
	// return files, lostFiles, nil
	return nil, nil, errors.New("not implemented")
}

func (fs *FileServiceImpl) GetMediaCacheByFilename(thumbFileName string) (*file_model.WeblensFileImpl, error) {
	cachePath := ThumbsDirPath.Child(thumbFileName, false)

	cacheFile, ok := fs.files[cachePath]
	if ok {
		return cacheFile, nil
	}

	return nil, errors.WithStack(file_model.ErrFileNotFound)
}

func (fs *FileServiceImpl) NewCacheFile(mediaId, quality string, pageNum int) (*file_model.WeblensFileImpl, error) {
	switch media_model.MediaQuality(quality) {
	case media_model.LowRes, media_model.HighRes:
		break
	default:
		return nil, errors.New("invalid quality")
	}

	filename := cacheFilename(mediaId, quality, pageNum)

	childPath := ThumbsDirPath.Child(filename, false)
	return touch(childPath)
}

func (fs *FileServiceImpl) DeleteCacheFile(f *file_model.WeblensFileImpl) error {
	if !isCacheFile(f.GetPortablePath()) {
		return errors.New("trying to delete non-cache file")
	}

	return remove(f.GetPortablePath())
}

func (fs *FileServiceImpl) CreateFile(parent *file_model.WeblensFileImpl, filename string, data ...[]byte) (
	*file_model.WeblensFileImpl, error,
) {
	childPath := parent.GetPortablePath().Child(filename, false)

	child, err := touch(childPath)
	if err != nil {
		return nil, err
	}

	// newF, err := fs.trees[UsersTreeKey].Touch(parent, filename, event, data...)
	// if err != nil {
	// 	return nil, err
	// }
	for _, d := range data {
		_, err = child.Write(d)
		if err != nil {
			return nil, err
		}
	}

	return child, nil
}

func (fs *FileServiceImpl) CreateFolder(parent *file_model.WeblensFileImpl, folderName string) (*file_model.WeblensFileImpl, error) {
	childPath := parent.GetPortablePath().Child(folderName, false)

	dir, err := mkdir(childPath)
	if err != nil {
		return nil, err
	}

	return dir, nil
}

func (fs *FileServiceImpl) MoveFilesToTrash(ctx context.ContextZ, files []*file_model.WeblensFileImpl, user *user_model.User, share *share_model.FileShare) error {
	if len(files) == 0 {
		return nil
	}

	owner, err := GetFileOwner(ctx, files[0])
	if err != nil {
		return err
	}

	trashPath := file_system.BuildFilePath(UsersTreeKey, owner.Username, UserTrashDirName)

	// event := journal.NewEvent()
	// oldParent := files[0].GetParent()

	for _, file := range files {
		if !file.Exists() {
			return errors.WithStack(file_model.ErrFileNotFound)
		}
		if IsFileInTrash(file) {
			return errors.Wrapf(file_model.ErrFileAlreadyExists, "Cannot move file [%s] to trash because it is already in trash", file.GetPortablePath())
		}

		childPath, err := MakeUniqueChildName(trashPath, file.GetPortablePath().Filename())
		if err != nil {
			return err
		}

		err = rename(file.GetPortablePath(), childPath)
		if err != nil {
			return err
		}
	}

	// err = fs.ResizeUp(ctx, oldParent, event)
	// if err != nil {
	// 	ctx.Log().Error().Stack().Err(err).Msg("")
	// }
	//
	// err = fs.ResizeUp(ctx, trash, event)
	// if err != nil {
	// 	ctx.Log().Error().Stack().Err(err).Msg("")
	// }
	//
	// tree.GetJournal().LogEvent(event)
	// err = event.Wait()
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (fs *FileServiceImpl) ReturnFilesFromTrash(ctx context.ContextZ, trashFiles []*file_model.WeblensFileImpl) error {
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

// DeleteFiles removes files being pointed to from the tree and moves them to the restore tree
func (fs *FileServiceImpl) DeleteFiles(ctx context.ContextZ, files []*file_model.WeblensFileImpl) error {
	deleteEvent := journal.NewEvent()

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	if local.Role == tower_model.BackupServerRole {
		for _, file := range files {
			err = remove(file.GetPortablePath())
			if err != nil {
				return err
			}
		}

		return nil
	}

	// All files *should* share the same parent: the trash folder, so pulling
	// just the first one to do the update on will work fine.
	trash := files[0].GetParent()

	var dirIds []file_system.Filepath
	var deletedFiles []*file_model.WeblensFileImpl

	for _, file := range files {
		err := file.RecursiveMap(
			func(f *file_model.WeblensFileImpl) error {

				// Freeze the file before it is deleted
				preDeleteFile := f.Freeze()
				contentId := f.GetContentId()
				m, err := media_model.GetMediaById(ctx, contentId)
				if err != nil {
					return err
				}
				// Remove the file from the media, if it exists
				if m != nil {
					err = media_model.RemoveFileFromMedia(ctx, m, f.ID())
					if err != nil {
						return err
					}
				}

				if f.IsDir() || f.Size() == 0 {
					// Save directory ids to be removed after all files have been moved
					dirIds = append(dirIds, f.GetPortablePath())
				} else {
					// Check if the restore file already exists, with the filename being the content id
					if contentId == "" {
						return errors.Errorf("trying to move file to restore tree without content id")
					}

					restorePath := file_system.BuildFilePath(RestoreTreeKey, f.GetContentId())

					// Check if the file already exists in the restore tree
					if exists(restorePath) {
						err = remove(f.GetPortablePath())
						if err != nil {
							return err
						}
						deletedFiles = append(deletedFiles, preDeleteFile)

						return nil
					}

					// Move file from users tree to the restore tree. Files later can be hard-linked back
					// from the restore tree to the users tree, but will not be moved back.
					err = rename(f.GetPortablePath(), restorePath)
					if err != nil {
						return err
					}

					ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("File [%s] moved from users tree to restore tree", f.GetPortablePath()) })

				}
				deletedFiles = append(deletedFiles, preDeleteFile)

				return nil
			},
		)
		if err != nil {
			return err
		}
	}

	// We need to make sure we delete the bottom most directories first,
	// since deleting a directory that is not empty will error. So we save
	// the directories until here, and then delete them in reverse order (bottom up).
	slices.Reverse(dirIds)
	for _, dirId := range dirIds {
		err = remove(dirId)
		if err != nil {
			return err
		}
	}

	err = fs.ResizeUp(ctx, trash, deleteEvent)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) RestoreFiles(ctx context.ContextZ, ids []string, newParent *file_model.WeblensFileImpl, restoreTime time.Time) error {
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

func (fs *FileServiceImpl) RestoreHistory(ctx context.ContextZ, actions []*history.FileAction) error {
	return errors.New("not implemented")

	// journal := fs.trees[UsersTreeKey].GetJournal()
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
	// 	if fs.trees[UsersTreeKey].Get(lt.ID()) != nil {
	// 		continue
	// 	}
	//
	// 	// parentId := latest.GetParentId()
	// 	parent, err := fs.getFileByIdAndRoot(latest.GetParentId(), UsersTreeKey)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	newF := file_model.NewWeblensFile(lt.ID(), portable.Filename(), parent, true)
	// 	err = fs.trees[UsersTreeKey].Add(newF)
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

func (fs *FileServiceImpl) NewZip(zipName string, owner *user_model.User) (*file_model.WeblensFileImpl, error) {

	return nil, errors.New("not implemented")
	// cacheTree := fs.trees[CachesTreeKey]
	// if cacheTree == nil {
	// 	return nil, errors.WithStack(file_model.ErrFileTreeNotFound)
	// }
	//
	// cacheRoot := cacheTree.GetRoot()
	//
	// takeoutDir, err := cacheRoot.GetChild("takeout")
	// if err != nil {
	// 	return nil, err
	// }
	//
	// zipFile, err := cacheTree.Touch(takeoutDir, zipName, nil)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// return zipFile, nil
}

func (fs *FileServiceImpl) GetZip(id string) (*file_model.WeblensFileImpl, error) {
	return nil, errors.New("not implemented")

	// takeoutFile := fs.trees[CachesTreeKey].Get(id)
	// if takeoutFile == nil {
	// 	return nil, file_model.ErrFileNotFound
	// }
	// if takeoutFile.GetParent().Filename() != "takeout" {
	// 	return nil, file_model.ErrFileNotFound
	// }
	//
	// return takeoutFile, nil
}

func (fs *FileServiceImpl) MoveFiles(ctx context.ContextZ, files []*file_model.WeblensFileImpl, destFolder *file_model.WeblensFileImpl, treeName string) error {
	if len(files) == 0 {
		return nil
	}

	// prevParent := files[0].GetParent()
	moveUpdates := map[string][]*file_model.WeblensFileImpl{}

	for _, file := range files {
		preFile := file.Freeze()
		newFilename, err := MakeUniqueChildName(destFolder.GetPortablePath(), file.GetPortablePath().Filename())

		err = rename(file.GetPortablePath(), newFilename)
		if err != nil {
			return err
		}

		key := preFile.GetParentId() + "->" + file.GetParentId()
		if moveUpdates[key] == nil {
			moveUpdates[key] = []*file_model.WeblensFileImpl{file}
		} else {
			moveUpdates[key] = append(moveUpdates[key], file)
		}
	}

	// TODO: Implelement caster calls
	// for key, moves := range moveUpdates {
	// 	keys := strings.Split(key, "->")
	// 	caster.PushFilesMove(keys[0], keys[1], moves)
	// }

	// err := fs.ResizeUp(ctx, destFolder, event)
	// if err != nil {
	// 	return err
	// }
	//
	// err = fs.ResizeUp(ctx, prevParent, event)
	// if err != nil {
	// 	return err
	// }
	//
	// tree.GetJournal().LogEvent(event)

	return nil
}

func (fs *FileServiceImpl) RenameFile(file *file_model.WeblensFileImpl, newName string) error {
	// preFile := file.Freeze()
	rename(file.GetPortablePath(), file.GetPortablePath().Dir().Child(newName, false))

	// TODO: Implement caster calls
	// caster.PushFileMove(preFile, file)

	return nil
}

func (fs *FileServiceImpl) NewBackupFile(ctx context.ContextZ, lt *history.FileAction) (*file_model.WeblensFileImpl, error) {
	// filename := lt.GetLatestPath().Filename()
	//
	// tree := fs.trees[lt.ServerId]
	// if tree == nil {
	// 	return nil, errors.Wrapf(file_model.ErrFileTreeNotFound, "no tree for remote with id [%s]", lt.ServerId)
	// }
	//
	// restoreTree := fs.trees[RestoreTreeKey]
	// if restoreTree == nil {
	// 	return nil, errors.Wrap(file_model.ErrFileTreeNotFound, "no restore file tree")
	// }
	//
	// if lt.GetIsDir() {
	// 	// If there is no path (i.e. the dir has been deleted), skip as there is
	// 	// no need to create a directory that no longer exists, just so long as it is
	// 	// included in the history, which is now is.
	// 	if lt.GetLatestPath().RootName() == "" {
	// 		return nil, nil
	// 	}
	//
	// 	// Find the directory's parent. This should already exist since we always create
	// 	// the backup file structure in order from parent to child.
	// 	latestAction := lt.GetLatestAction()
	// 	parent := tree.Get(latestAction.ParentId)
	// 	if parent == nil {
	// 		return nil, errors.Wrapf(file_model.ErrFileNotFound, "looking for parent of [%s]: %s", latestAction.DestinationPath, latestAction.ParentId)
	// 	}
	//
	// 	// Create the directory object and add it to the tree
	// 	newDir := file_model.NewWeblensFile(lt.ID(), filename, parent, true)
	// 	err := tree.Add(newDir)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	ctx.Log().Trace().Msgf("Creating backup dir %s", newDir.GetPortablePath())
	//
	// 	// Create the directory on disk
	// 	err = newDir.CreateSelf()
	// 	if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
	// 		return nil, err
	// 	}
	// 	return nil, nil
	// }
	//
	// if lt.GetContentId() == "" && lt.GetLatestSize() != 0 {
	// 	return nil, errors.WithStack(file_model.ErrNoContentId)
	// } else if lt.GetContentId() == "" {
	// 	return nil, nil
	// }
	//
	// var restoreFile *file_model.WeblensFileImpl
	// if restoreFile, _ = restoreTree.GetRoot().GetChild(lt.GetContentId()); restoreFile == nil {
	// 	var err error
	// 	restoreFile, err = restoreTree.Touch(restoreTree.GetRoot(), lt.GetContentId(), nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	restoreFile.SetContentId(lt.GetContentId())
	// } else {
	// 	_, err := restoreFile.LoadStat()
	// 	if err != nil {
	// 		return nil, errors.WithStack(err)
	// 	}
	// }
	//
	// if lt.GetLatestAction().ActionType != history.FileDelete {
	// 	portable, err := file_system.ParsePortable(lt.GetLatestAction().DestinationPath)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	// Translate from the portable path to expand the absolute path
	// 	// with the new backup tree
	// 	newPortable := portable.OverwriteRoot(lt.ServerId)
	//
	// 	latestMove := lt.GetLatestMove()
	// 	if latestMove == nil {
	// 		return nil, errors.WithStack(file_model.ErrFileNotFound)
	// 	}
	//
	// 	parent := tree.Get(latestMove.ParentId)
	// 	if parent == nil {
	// 		ctx.Log().Debug().Func(func(e *zerolog.Event) {
	// 			e.Msgf("Parent [%s] not found trying to get parent for [%s]", latestMove.ParentId, lt.Id)
	// 		})
	// 		return nil, errors.Wrapf(file_model.ErrFileNotFound, "trying to get %s", latestMove.ParentId)
	// 	}
	//
	// 	newF := file_model.NewWeblensFile(lt.ID(), newPortable.Filename(), parent, false)
	//
	// 	err = tree.Add(newF)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	if newF.Exists() {
	// 		ctx.Log().Warn().Msgf("File [%s] already exists, overwriting", newF.AbsPath())
	// 		err = newF.Remove()
	// 		if err != nil {
	// 			return nil, errors.WithStack(err)
	// 		}
	// 	}
	//
	// 	ctx.Log().Trace().Func(func(e *zerolog.Event) {
	// 		e.Msgf("Linking %s -> %s", restoreFile.GetPortablePath().ToPortable(), portable.ToPortable())
	// 	})
	// 	err = os.Link(restoreFile.AbsPath(), newF.AbsPath())
	// 	if err != nil {
	// 		return nil, errors.WithStack(err)
	// 	}
	// }
	//
	// return restoreFile, nil
	return nil, errors.New("not implemented")
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

func (fs *FileServiceImpl) ResizeUp(ctx context.ContextZ, f *file_model.WeblensFileImpl, event *history.FileEvent) error {
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

	return errors.New("not implemented")
}

func (fs *FileServiceImpl) ResizeDown(ctx context.ContextZ, f *file_model.WeblensFileImpl, event *history.FileEvent) error {
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
	return errors.New("not implemented")
}

func (fs *FileServiceImpl) loadContentIdCache() error {
	// fs.contentIdLock.Lock()
	// defer fs.contentIdLock.Unlock()
	// fs.contentIdCache = make(map[string]*file_model.WeblensFileImpl)
	//
	// _ = fs.trees[RestoreTreeKey].GetRoot().LeafMap(
	// 	func(f *file_model.WeblensFileImpl) error {
	// 		if f.IsDir() {
	// 			return nil
	// 		}
	// 		fs.contentIdCache[f.Filename()] = f
	// 		return nil
	// 	},
	// )
	//
	// if usersTree := fs.trees[UsersTreeKey]; usersTree != nil {
	// 	_ = usersTree.GetRoot().LeafMap(
	// 		func(f *file_model.WeblensFileImpl) error {
	// 			if f.IsDir() {
	// 				return nil
	// 			}
	// 			contentId := f.GetContentId()
	// 			if contentId != "" {
	// 				if _, ok := fs.contentIdCache[contentId]; !ok {
	// 					fs.contentIdCache[contentId] = f
	// 				}
	// 			}
	// 			return nil
	// 		},
	// 	)
	// }
	//
	// return nil
	return errors.New("not implemented")
}

func GenerateContentId(f *file_model.WeblensFileImpl) (string, error) {
	if f.IsDir() {
		return "", errors.Errorf("cannot hash directory")
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
	}

	fileSize := f.Size()

	if fileSize == 0 {
		return "", errors.WithStack(file_model.ErrEmptyFile)
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
	}

	// Read up to 1MB at a time
	bufSize := math.Min(float64(fileSize), 1000*1000)
	buf := make([]byte, int64(bufSize))
	newHash := sha256.New()
	fp, err := f.Readable()
	if err != nil {
		return "", err
	}

	if closer, ok := fp.(io.Closer); ok {
		defer func(fp io.Closer) {
			_ = fp.Close()
			// if err != nil {
			// 	ctx.Log().Error().Stack().Err(err).Msg("")
			// }
		}(closer)
	}

	_, err = io.CopyBuffer(newHash, fp, buf)
	if err != nil {
		return "", err
	}

	contentId := base64.URLEncoding.EncodeToString(newHash.Sum(nil))[:20]
	f.SetContentId(contentId)

	return contentId, nil
}

func (fs *FileServiceImpl) CreateUserHome(ctx *context.ContextZ, user *user_model.User) error {
	parent, err := fs.GetFileByFilepath(UsersRootPath)
	if err != nil {
		return err
	}
	home, err := fs.CreateFolder(parent, user.GetUsername())
	if err != nil && !errors.Is(err, file_model.ErrDirectoryAlreadyExists) {
		return err
	}
	user.HomeId = home.ID()

	trash, err := fs.CreateFolder(home, UserTrashDirName)
	if err != nil && !errors.Is(err, file_model.ErrDirectoryAlreadyExists) {
		return err
	}
	user.TrashId = trash.ID()

	return nil
}

func GetFileOwnerName(ctx context.ContextZ, file *file_model.WeblensFileImpl) (string, error) {
	portable := file.GetPortablePath()
	if portable.RootName() != UsersTreeKey {
		return "", errors.New("trying to get owner of file not in MEDIA tree")
	}

	slashIndex := strings.Index(portable.RelativePath(), "/")
	var username string
	if slashIndex == -1 {
		username = portable.RelativePath()
	} else {
		username = portable.RelativePath()[:slashIndex]
	}

	return username, nil
}

func GetFileOwner(ctx context.ContextZ, file *file_model.WeblensFileImpl) (*user_model.User, error) {
	username, err := GetFileOwnerName(ctx, file)
	if err != nil {
		return nil, err
	}
	return user_model.GetUserByUsername(ctx, username)
}

func IsFileInTrash(f *file_model.WeblensFileImpl) bool {
	return strings.Contains(f.GetPortablePath().RelativePath(), UserTrashDirName)
}

func ContentIdFromHash(newHash hash.Hash) string {
	return base64.URLEncoding.EncodeToString(newHash.Sum(nil))[:20]
}

func MakeUniqueChildName(parent file_system.Filepath, childName string) (childPath file_system.Filepath, err error) {
	dupeCount := 0

	if !exists(parent) {
		return childPath, errors.New("parent does not exist")
	}

	// Check if the child already exists
	childPath = parent.Child(childName, false)
	for exists(childPath) {
		dupeCount++
		tmpName := fmt.Sprintf("%s (%d)", childName, dupeCount)
		childPath = parent.Child(tmpName, false)

		if dupeCount > 100 {
			return childPath, errors.New("too many duplicates")
		}
	}

	return childPath, nil
}
