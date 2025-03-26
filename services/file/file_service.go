package file

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	wl_slices "github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/task"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	UsersTreeKey   = "USERS"
	RestoreTreeKey = "RESTORE"
	CachesTreeKey  = "CACHES"

	UserTrashDirName = ".user_trash"
	ThumbsDirName    = "thumbs"
)

type FileServiceImpl struct {
	trees map[string]fileTree.FileTree

	contentIdCache map[string]*file_model.WeblensFileImpl

	folderMedia map[file_model.FileId]string

	fileTaskLink map[file_model.FileId][]*task.Task

	folderCoverCol *mongo.Collection

	log *zerolog.Logger

	treesLock sync.RWMutex

	contentIdLock sync.RWMutex

	fileTaskLock sync.RWMutex
}

type FolderCoverPair struct {
	FolderId  file_model.FileId `bson:"folderId"`
	ContentId string            `bson:"coverId"`
}

func NewFileService(
	logger *zerolog.Logger,
	folderCoverCol *mongo.Collection,
	trees ...fileTree.FileTree,
) (*FileServiceImpl, error) {

	fs := &FileServiceImpl{
		log:          logger,
		trees:        map[string]fileTree.FileTree{},
		fileTaskLink: make(map[file_model.FileId][]*task.Task),
		folderMedia:  make(map[file_model.FileId]string),
	}

	for _, tree := range trees {
		fs.trees[tree.GetRoot().GetPortablePath().RootName()] = tree
	}

	if usersTree, ok := fs.trees[UsersTreeKey]; ok {
		err := fs.ResizeDown(usersTree.GetRoot(), nil)
		if err != nil {
			return nil, err
		}
	}

	ret, err := fs.folderCoverCol.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var folderCovers []FolderCoverPair
	err = ret.All(context.Background(), &folderCovers)
	if err != nil {
		return nil, err
	}

	for _, folderCover := range folderCovers {
		fs.folderMedia[folderCover.FolderId] = folderCover.ContentId
	}

	return fs, nil
}

func (fs *FileServiceImpl) Size(treeAlias string) int64 {
	tree := fs.trees[treeAlias]
	if tree == nil {
		return -1
	}

	return tree.GetRoot().Size()
}

func (fs *FileServiceImpl) GetFileByTree(id file_model.FileId, treeAlias string) (*file_model.WeblensFileImpl, error) {
	return fs.getFileByIdAndRoot(id, treeAlias)
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

	return nil, errors.WithStack(werror.ErrNoFile)
}

func (fs *FileServiceImpl) GetFiles(ids []file_model.FileId) ([]*file_model.WeblensFileImpl, []file_model.FileId, error) {
	usersTree := fs.trees[UsersTreeKey]
	if usersTree == nil {
		return nil, nil, errors.WithStack(werror.ErrNoFileTree.WithArg(UsersTreeKey))
	}

	var files []*file_model.WeblensFileImpl
	var lostFiles []file_model.FileId
	for _, id := range ids {
		lt := usersTree.GetJournal().Get(id)
		if lt == nil {
			lostFiles = append(lostFiles, id)
			continue
			// return nil, nil, errors.WithStack(werror.ErrNoLifetime.WithArg(id))
		}
		if lt.GetLatestAction().ActionType == fileTree.FileDelete {
			contentId := lt.GetContentId()
			if contentId == "" {
				lostFiles = append(lostFiles, id)
				continue
			}
			f, err := fs.trees[RestoreTreeKey].GetRoot().GetChild(contentId)
			if err != nil {
				lostFiles = append(lostFiles, id)
				continue
			}
			files = append(files, f)
		} else {
			f := usersTree.Get(id)
			if f == nil {
				lostFiles = append(lostFiles, id)
				continue
			}
			files = append(files, f)
		}
	}
	return files, lostFiles, nil
}

// func (fs *FileServiceImpl) GetFileSafe(id file_model.FileId, user *user_model.User, share *models.FileShare) (
// 	*file_model.WeblensFileImpl,
// 	error,
// ) {
// 	tree := fs.trees[UsersTreeKey]
// 	if tree == nil {
// 		return nil, errors.WithStack(werror.ErrNoFileTree)
// 	}
//
// 	f := tree.Get(id)
// 	if f == nil {
// 		return nil, errors.WithStack(werror.ErrNoFile.WithArg(id))
// 	}
//
// 	if !fs.accessService.CanUserAccessFile(user, f, share) {
// 		fs.log.Error().Stack().Err(werror.Errorf(
// 			"User [%s] attempted to access file at %s [%s], but they do not have access",
// 			user.GetUsername(), f.GetPortablePath(), f.ID(),
// 		)).Msg("")
// 		return nil, errors.WithStack(werror.ErrNoFileAccess)
// 	}
//
// 	return f, nil
// }

func (fs *FileServiceImpl) GetFileTreeByName(treeName string) (fileTree.FileTree, error) {
	tree := fs.trees[treeName]
	if tree == nil {
		return nil, errors.WithStack(werror.ErrNoFileTree.WithArg(treeName))
	}

	return tree, nil
}

func (fs *FileServiceImpl) GetMediaCacheByFilename(thumbFileName string) (*file_model.WeblensFileImpl, error) {
	thumbsDir, err := fs.trees[CachesTreeKey].GetRoot().GetChild(ThumbsDirName)
	if err != nil {
		return nil, err
	}
	return thumbsDir.GetChild(thumbFileName)
}

func (fs *FileServiceImpl) IsFileInTrash(f *file_model.WeblensFileImpl) bool {
	return strings.Contains(f.AbsPath(), UserTrashDirName)
}

func (fs *FileServiceImpl) NewCacheFile(
	media *media_model.Media, quality media_model.MediaQuality, pageNum int,
) (*file_model.WeblensFileImpl, error) {
	filename := media.FmtCacheFileName(quality, pageNum)

	thumbsDir, err := fs.GetThumbsDir()
	if err != nil {
		return nil, err
	}

	return fs.trees[CachesTreeKey].Touch(thumbsDir, filename, nil)
}

func (fs *FileServiceImpl) DeleteCacheFile(f file_model.WeblensFileImpl) error {
	_, err := fs.trees[CachesTreeKey].Remove(f.ID())
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileServiceImpl) CreateFile(parent *file_model.WeblensFileImpl, filename string, event *fileTree.FileEvent, data ...[]byte) (
	*file_model.WeblensFileImpl, error,
) {
	newF, err := fs.trees[UsersTreeKey].Touch(parent, filename, event, data...)
	if err != nil {
		return nil, err
	}

	return newF, nil
}

func (fs *FileServiceImpl) CreateFolder(parent *file_model.WeblensFileImpl, folderName string, event *fileTree.FileEvent) (
	*file_model.WeblensFileImpl,
	error,
) {

	newF, err := fs.trees[UsersTreeKey].MkDir(parent, folderName, event)
	if err != nil {
		return newF, err
	}

	return newF, nil
}

func (fs *FileServiceImpl) CreateUserHome(user *user_model.User) error {
	usersTree := fs.trees[UsersTreeKey]
	if usersTree == nil {
		return errors.WithStack(werror.ErrNoFileTree)
	}

	home, err := usersTree.MkDir(usersTree.GetRoot(), user.GetUsername(), nil)
	if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
		return err
	}
	user.SetHomeFolder(home)

	trash, err := usersTree.MkDir(home, UserTrashDirName, nil)
	if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
		return err
	}
	user.SetTrashFolder(trash)

	return nil
}

func (fs *FileServiceImpl) GetFileOwner(file *file_model.WeblensFileImpl) (*user_model.User, error) {
	portable := file.GetPortablePath()
	if portable.RootName() != UsersTreeKey {
		return nil, werror.Errorf("trying to get owner of file not in MEDIA tree")
	}

	slashIndex := strings.Index(portable.RelativePath(), "/")
	var username string
	if slashIndex == -1 {
		username = portable.RelativePath()
	} else {
		username = portable.RelativePath()[:slashIndex]
	}

	return user_model.GetUserByUsername(context.Background(), username)
}

func (fs *FileServiceImpl) MoveFilesToTrash(files []*file_model.WeblensFileImpl, user *user_model.User, share *share_model.FileShare) error {
	if len(files) == 0 {
		return nil
	}

	owner, err := fs.GetFileOwner(files[0])
	if err != nil {
		return err
	}

	tree := fs.trees[UsersTreeKey]
	if tree == nil {
		return errors.WithStack(werror.ErrNoFileTree)
	}
	trash := tree.Get(owner.TrashId)
	if trash == nil {
		return errors.WithStack(errors.New("trash folder does not exist"))
	}

	event := tree.GetJournal().NewEvent()

	oldParent := files[0].GetParent()

	for _, file := range files {
		if !file.Exists() {
			return werror.Errorf("Cannot with id [%s] (%s) does not exist", file.ID(), file.AbsPath())
		}
		if fs.IsFileInTrash(file) {
			return werror.Errorf("Cannot move file (%s) to trash because it is already in trash", file.AbsPath())
		}
		if !fs.accessService.CanUserAccessFile(user, file, share) {
			return errors.WithStack(werror.ErrNoFileAccess)
		}

		newFilename := MakeUniqueChildName(trash, file.Filename())
		preMoveFile := file.Freeze()

		_, err := tree.Move(file, trash, newFilename, false, event)
		if err != nil {
			return err
		}
	}

	err = fs.ResizeUp(oldParent, event)
	if err != nil {
		fs.log.Error().Stack().Err(err).Msg("")
	}

	err = fs.ResizeUp(trash, event)
	if err != nil {
		fs.log.Error().Stack().Err(err).Msg("")
	}

	tree.GetJournal().LogEvent(event)
	err = event.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) ReturnFilesFromTrash(trashFiles []*file_model.WeblensFileImpl) error {
	trash := trashFiles[0].GetParent()
	trashPath := trash.GetPortablePath().ToPortable()

	tree := fs.trees[UsersTreeKey]
	if tree == nil {
		return errors.WithStack(werror.ErrNoFileTree)
	}
	journal := tree.GetJournal()
	event := journal.NewEvent()

	for _, trashEntry := range trashFiles {
		preFile := trashEntry.Freeze()

		if !fs.IsFileInTrash(trashEntry) {
			return werror.Errorf("cannot return file from trash, file is not in trash")
		}

		acns := journal.Get(trashEntry.ID()).Actions
		if len(acns) < 2 || !strings.HasPrefix(acns[len(acns)-1].DestinationPath, trashPath) {
			return werror.Errorf("cannot return file from trash, journal does not have trash destination")
		}

		justBeforeTrash := acns[len(acns)-2]
		oldParent := tree.Get(justBeforeTrash.ParentId)
		if oldParent == nil {
			owner, err := fs.GetFileOwner(trashEntry)
			if err != nil {
				return err
			}
			oldParent = tree.Get(owner.HomeId)
		}

		_, err := tree.Move(trashEntry, oldParent, fileTree.ParsePortable(justBeforeTrash.DestinationPath).Filename(), false, event)

		if err != nil {
			return err
		}
	}

	err := fs.ResizeUp(trash, event)
	if err != nil {
		return err
	}

	journal.LogEvent(event)

	return nil
}

// DeleteFiles removes files being pointed to from the tree and moves them to the restore tree
func (fs *FileServiceImpl) DeleteFiles(ctx context.Context, files []*file_model.WeblensFileImpl, treeName string) error {
	tree := fs.trees[treeName]
	if tree == nil {
		return errors.WithStack(werror.ErrNoFileTree)
	}

	restoreTree := fs.trees[RestoreTreeKey]
	if restoreTree == nil {
		return errors.WithStack(werror.ErrNoFileTree)
	}

	deleteEvent := tree.GetJournal().NewEvent()

	local, err := tower.GetLocal(ctx)
	if err != nil {
		return err
	}

	if local.Role == tower.BackupServerRole {
		for _, file := range files {
			err := tree.Delete(file.ID(), deleteEvent)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// All files *should* share the same parent: the trash folder, so pulling
	// just the first one to do the update on will work fine.
	trash := files[0].GetParent()

	var dirIds []file_model.FileId
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

				child, err := restoreTree.GetRoot().GetChild(f.GetContentId())
				if err == nil && child.Exists() {
					err = tree.Delete(f.ID(), deleteEvent)
					if err != nil {
						return err
					}
					deletedFiles = append(deletedFiles, preDeleteFile)

					// Remove the file from the media, if it exists
					if m != nil {
						err = fs.mediaService.RemoveFileFromMedia(m, f.ID())
						if err != nil {
							return err
						}
					}

					return nil
				}

				if f.IsDir() || f.Size() == 0 {
					// Save directory ids to be removed after all files have been moved
					dirIds = append(dirIds, f.ID())
				} else {
					// Check if the restore file already exists, with the filename being the content id
					if contentId == "" {
						return werror.Errorf("trying to move file to restore tree without content id")
					}

					// Remove the file from the media, if it exists
					if m != nil {
						err = fs.mediaService.RemoveFileFromMedia(m, f.ID())
						if err != nil {
							return err
						}
					}

					// Check if the file already exists in the restore tree
					_, err = restoreTree.GetRoot().GetChild(contentId)
					if err != nil {
						// A non-nil error here means the file does not exist, so we must move it to the restore tree

						// Add the delete for this file to the event
						// We must do this before moving/deleting the file, or the action will not be able to find the file
						_, err = deleteEvent.NewDeleteAction(f.ID())
						if err != nil {
							return err
						}

						moveEvent := tree.GetJournal().NewEvent()
						// Move file from users tree to the restore tree. Files later can be hard-linked back
						// from the restore tree to the users tree, but will not be moved back.
						err = fileTree.MoveFileBetweenTrees(
							f, restoreTree.GetRoot(), f.GetContentId(), tree, restoreTree,
							moveEvent,
						)
						if err != nil {
							return err
						}

						fs.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("File [%s] moved from users tree to restore tree", f.GetPortablePath()) })

					} else {
						// If the file already is in the restore tree, we can just delete it from the users tree.
						// This should be rare since we already checked if the file exists in the index, but it is possible
						// if the index is missing or otherwise out of sync.
						err = tree.Delete(f.ID(), deleteEvent)
						if err != nil {
							return err
						}

						fs.log.Trace().Func(func(e *zerolog.Event) {
							e.Msgf("File [%s] already exists in restore tree, deleting from users tree", f.GetPortablePath())
						})
					}
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
		err := tree.Delete(dirId, deleteEvent)
		if err != nil {
			return err
		}
	}

	err = fs.ResizeUp(trash, deleteEvent)
	if err != nil {
		return err
	}

	tree.GetJournal().LogEvent(deleteEvent)

	return nil
}

func (fs *FileServiceImpl) RestoreFiles(ids []file_model.FileId, newParent *file_model.WeblensFileImpl, restoreTime time.Time) error {
	usersTree := fs.trees[UsersTreeKey]
	if usersTree == nil {
		return errors.WithStack(werror.ErrNoFileTree.WithArg(UsersTreeKey))
	}

	restoreTree := fs.trees[RestoreTreeKey]
	if restoreTree == nil {
		return errors.WithStack(werror.ErrNoFileTree.WithArg(RestoreTreeKey))
	}

	journal := usersTree.GetJournal()
	event := journal.NewEvent()

	var topFiles []*file_model.WeblensFileImpl
	type restorePair struct {
		newParent *file_model.WeblensFileImpl
		fileId    file_model.FileId
		contentId string
	}

	var restorePairs []restorePair
	for _, id := range ids {
		lt := journal.Get(id)
		if lt == nil {
			return werror.Errorf("journal does not have file to restore")
		}
		restorePairs = append(
			restorePairs, restorePair{fileId: id, newParent: newParent, contentId: lt.ContentId},
		)
	}

	for len(restorePairs) != 0 {
		toRestore := restorePairs[0]
		restorePairs = restorePairs[1:]

		pastFile, err := journal.GetPastFile(toRestore.fileId, restoreTime)
		if err != nil {
			return err
		}

		var childIds []file_model.FileId
		if pastFile.IsDir() {
			children := pastFile.GetChildren()

			childIds = wl_slices.Map(
				children, func(child *file_model.WeblensFileImpl) file_model.FileId {
					return child.ID()
				},
			)
		}

		path := pastFile.GetPortablePath().ToPortable()
		if path == "" {
			return werror.Errorf("Got empty string for portable path on past file [%s]", pastFile.AbsPath())
		}
		// Paths of directory files will have an extra / on the end, so we need to remove it
		if pastFile.IsDir() {
			path = path[:len(path)-1]
		}

		oldName := filepath.Base(path)
		newName := MakeUniqueChildName(toRestore.newParent, oldName)

		var restoredF *file_model.WeblensFileImpl
		if !pastFile.IsDir() {
			var existingPath string

			// File has been deleted, get the file from the restore tree
			if liveF := usersTree.Get(toRestore.fileId); liveF == nil {
				_, err = restoreTree.GetRoot().GetChild(toRestore.contentId)
				if err != nil {
					return err
				}
				existingPath = filepath.Join(restoreTree.GetRoot().AbsPath(), toRestore.contentId)
			} else {
				existingPath = liveF.AbsPath()
			}

			restoredF = fileTree.NewWeblensFile(
				usersTree.GenerateFileId(), newName, toRestore.newParent, pastFile.IsDir(),
			)
			restoredF.SetContentId(pastFile.GetContentId())
			restoredF.SetSize(pastFile.Size())
			err = usersTree.Add(restoredF)
			if err != nil {
				return err
			}

			fs.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Restoring file [%s] to [%s]", existingPath, restoredF.AbsPath()) })
			err = os.Link(existingPath, restoredF.AbsPath())
			if err != nil {
				return errors.WithStack(err)
			}

			if toRestore.newParent == newParent {
				topFiles = append(topFiles, restoredF)
			}

		} else {
			restoredF = fileTree.NewWeblensFile(
				usersTree.GenerateFileId(), newName, toRestore.newParent, true,
			)
			err = usersTree.Add(restoredF)
			if err != nil {
				return err
			}

			err = restoredF.CreateSelf()
			if err != nil {
				return err
			}

			for _, childId := range childIds {
				childLt := journal.Get(childId)
				if childLt == nil {
					return errors.WithStack(werror.ErrNoFile)
				}
				restorePairs = append(
					restorePairs,
					restorePair{fileId: childId, newParent: restoredF, contentId: childLt.GetContentId()},
				)
			}

			if toRestore.newParent == newParent {
				topFiles = append(topFiles, restoredF)
			}
		}

		event.NewRestoreAction(restoredF)
	}

	for _, f := range topFiles {
		err := fs.ResizeDown(f, event, caster)
		if err != nil {
			return err
		}
		err = fs.ResizeUp(f, event, caster)
		if err != nil {
			return err
		}
	}

	journal.LogEvent(event)
	err := event.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) RestoreHistory(lifetimes []*fileTree.Lifetime) error {

	journal := fs.trees[UsersTreeKey].GetJournal()

	err := journal.Add(lifetimes...)
	if err != nil {
		return err
	}

	slices.SortFunc(lifetimes, fileTree.LifetimeSorter)

	for _, lt := range lifetimes {
		latest := lt.GetLatestAction()
		if latest.GetActionType() == fileTree.FileDelete {
			continue
		}
		portable := fileTree.ParsePortable(latest.GetDestinationPath())
		if !portable.IsDir() {
			continue
		}
		if fs.trees[UsersTreeKey].Get(lt.ID()) != nil {
			continue
		}

		// parentId := latest.GetParentId()
		parent, err := fs.getFileByIdAndRoot(latest.GetParentId(), UsersTreeKey)
		if err != nil {
			return err
		}

		newF := fileTree.NewWeblensFile(lt.ID(), portable.Filename(), parent, true)
		err = fs.trees[UsersTreeKey].Add(newF)
		if err != nil {
			return err
		}
		err = newF.CreateSelf()
		if err != nil && !errors.Is(err, werror.ErrFileAlreadyExists) {
			return err
		}
	}

	return nil
}

func (fs *FileServiceImpl) NewZip(zipName string, owner *user_model.User) (*file_model.WeblensFileImpl, error) {
	cacheTree := fs.trees[CachesTreeKey]
	if cacheTree == nil {
		return nil, errors.WithStack(werror.ErrNoFileTree)
	}

	cacheRoot := cacheTree.GetRoot()

	takeoutDir, err := cacheRoot.GetChild("takeout")
	if err != nil {
		return nil, err
	}

	zipFile, err := cacheTree.Touch(takeoutDir, zipName, nil)
	if err != nil {
		return nil, err
	}

	return zipFile, nil
}

func (fs *FileServiceImpl) GetZip(id file_model.FileId) (*file_model.WeblensFileImpl, error) {
	takeoutFile := fs.trees[CachesTreeKey].Get(id)
	if takeoutFile == nil {
		return nil, werror.ErrNoFile
	}
	if takeoutFile.GetParent().Filename() != "takeout" {
		return nil, werror.ErrNoFile
	}

	return takeoutFile, nil
}

func (fs *FileServiceImpl) MoveFiles(
	files []*file_model.WeblensFileImpl, destFolder *file_model.WeblensFileImpl, treeName string, caster models.FileCaster,
) error {
	if len(files) == 0 {
		return nil
	}

	tree := fs.trees[treeName]

	event := tree.GetJournal().NewEvent()
	prevParent := files[0].GetParent()

	moveUpdates := map[string][]*file_model.WeblensFileImpl{}

	for _, file := range files {
		preFile := file.Freeze()
		newFilename := MakeUniqueChildName(destFolder, file.Filename())

		_, err := tree.Move(file, destFolder, newFilename, false, event)
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

	for key, moves := range moveUpdates {
		keys := strings.Split(key, "->")
		caster.PushFilesMove(keys[0], keys[1], moves)
	}

	err := fs.ResizeUp(destFolder, event, caster)
	if err != nil {
		return err
	}

	err = fs.ResizeUp(prevParent, event, caster)
	if err != nil {
		return err
	}

	tree.GetJournal().LogEvent(event)

	return nil
}

func (fs *FileServiceImpl) RenameFile(file *file_model.WeblensFileImpl, newName string, caster models.FileCaster) error {
	preFile := file.Freeze()
	_, err := fs.trees[UsersTreeKey].Move(file, file.GetParent(), newName, false, nil)
	if err != nil {
		return err
	}

	caster.PushFileMove(preFile, file)

	return nil
}

func (fs *FileServiceImpl) AddTree(tree fileTree.FileTree) {
	fs.treesLock.Lock()
	defer fs.treesLock.Unlock()
	fs.trees[tree.GetRoot().GetPortablePath().RootName()] = tree
}

func (fs *FileServiceImpl) NewBackupFile(lt *fileTree.Lifetime) (*file_model.WeblensFileImpl, error) {
	filename := lt.GetLatestPath().Filename()

	tree := fs.trees[lt.ServerId]
	if tree == nil {
		return nil, errors.Wrapf(werror.ErrNoFileTree, "no tree for remote with id [%s]", lt.ServerId)
	}

	restoreTree := fs.trees[RestoreTreeKey]
	if restoreTree == nil {
		return nil, errors.WithStack(werror.ErrNoFileTree.WithArg(RestoreTreeKey))
	}

	if lt.GetIsDir() {
		// If there is no path (i.e. the dir has been deleted), skip as there is
		// no need to create a directory that no longer exists, just so long as it is
		// included in the history, which is now is.
		if lt.GetLatestPath().RootName() == "" {
			fs.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping dir that has no dest path") })
			return nil, nil
		}

		// Find the directory's parent. This should already exist since we always create
		// the backup file structure in order from parent to child.
		latestAction := lt.GetLatestAction()
		parent := tree.Get(latestAction.ParentId)
		if parent == nil {
			return nil, errors.WithStack(werror.ErrNoFile.WithArg(fmt.Sprintf("looking for parent of [%s]: %s", latestAction.DestinationPath, latestAction.ParentId)))
		}

		// Create the directory object and add it to the tree
		newDir := fileTree.NewWeblensFile(lt.ID(), filename, parent, true)
		err := tree.Add(newDir)
		if err != nil {
			return nil, err
		}

		fs.log.Trace().Msgf("Creating backup dir %s", newDir.GetPortablePath())

		// Create the directory on disk
		err = newDir.CreateSelf()
		if err != nil && !errors.Is(err, werror.ErrFileAlreadyExists) {
			return nil, err
		}
		return nil, nil
	}

	if lt.GetContentId() == "" && lt.GetLatestSize() != 0 {
		return nil, errors.WithStack(werror.ErrNoContentId)
	} else if lt.GetContentId() == "" {
		return nil, nil
	}

	var restoreFile *file_model.WeblensFileImpl
	if restoreFile, _ = restoreTree.GetRoot().GetChild(lt.GetContentId()); restoreFile == nil {
		var err error
		restoreFile, err = restoreTree.Touch(restoreTree.GetRoot(), lt.GetContentId(), nil)
		if err != nil {
			return nil, err
		}
		restoreFile.SetContentId(lt.GetContentId())
	} else {
		_, err := restoreFile.LoadStat()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if lt.GetLatestAction().ActionType != fileTree.FileDelete {
		portable := fileTree.ParsePortable(lt.GetLatestAction().DestinationPath)

		// Translate from the portable path to expand the absolute path
		// with the new backup tree
		newPortable := portable.OverwriteRoot(lt.ServerId)

		latestMove := lt.GetLatestMove()
		if latestMove == nil {
			return nil, errors.WithStack(werror.ErrNoFileAction.WithArg("no latest move"))
		}

		parent := tree.Get(latestMove.ParentId)
		if parent == nil {
			fs.log.Debug().Func(func(e *zerolog.Event) {
				e.Msgf("Parent [%s] not found trying to get parent for [%s]", latestMove.ParentId, lt.Id)
			})
			return nil, errors.WithStack(werror.ErrNoFile.WithArg("trying to get " + latestMove.ParentId))
		}

		newF := fileTree.NewWeblensFile(lt.ID(), newPortable.Filename(), parent, false)

		err := tree.Add(newF)
		if err != nil {
			return nil, err
		}

		if newF.Exists() {
			fs.log.Warn().Msgf("File [%s] already exists, overwriting", newF.AbsPath())
			err = os.Remove(newF.AbsPath())
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}

		fs.log.Trace().Func(func(e *zerolog.Event) {
			e.Msgf("Linking %s -> %s", restoreFile.GetPortablePath().ToPortable(), portable.ToPortable())
		})
		err = os.Link(restoreFile.AbsPath(), newF.AbsPath())
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return restoreFile, nil
}

func (fs *FileServiceImpl) GetJournalByTree(treeName string) fileTree.Journal {
	tree := fs.trees[treeName]
	if tree == nil {
		fs.log.Error().Msgf("No tree with name %s", treeName)
		return nil
	}
	return tree.GetJournal()
}

func (fs *FileServiceImpl) SetFolderCover(folderId file_model.FileId, coverId string) error {
	tree := fs.trees[UsersTreeKey]
	if tree == nil {
		return errors.WithStack(werror.ErrNoFileTree)
	}

	folder := tree.Get(folderId)
	if folder == nil {
		return errors.WithStack(werror.ErrNoFile)
	}

	if coverId == "" {
		_, err := fs.folderCoverCol.DeleteOne(context.Background(), bson.M{"folderId": folderId})
		if err != nil {
			return errors.WithStack(err)
		}

		delete(fs.folderMedia, folderId)
		folder.SetContentId("")
		return nil

	} else if fs.folderMedia[folderId] != "" {
		_, err := fs.folderCoverCol.UpdateOne(
			context.Background(), bson.M{"folderId": folderId}, bson.M{"$set": bson.M{"coverId": coverId}},
		)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		_, err := fs.folderCoverCol.InsertOne(context.Background(), bson.M{"folderId": folderId, "coverId": coverId})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	fs.folderMedia[folderId] = coverId
	folder.SetContentId(coverId)

	return nil
}

func (fs *FileServiceImpl) GetFolderCover(folder *file_model.WeblensFileImpl) (string, error) {
	if !folder.IsDir() {
		return "", werror.ErrDirectoryRequired
	}

	if cId := folder.GetContentId(); cId != "" {
		return cId, nil
	}

	coverId := fs.folderMedia[folder.ID()]
	folder.SetContentId(coverId)

	return coverId, nil
}

func (fs *FileServiceImpl) UserPathToFile(searchPath string, user *user_model.User) (*file_model.WeblensFileImpl, error) {

	if strings.HasPrefix(searchPath, "~/") {
		searchPath = string(user.GetUsername()) + "/" + searchPath[2:]
	} else if searchPath[:1] == "/" && user.IsAdmin() {
		searchPath = searchPath[1:]
	}

	return fs.PathToFile(searchPath)
}

func (fs *FileServiceImpl) PathToFile(searchPath string) (*file_model.WeblensFileImpl, error) {
	searchPath = strings.TrimPrefix(searchPath, "USERS:")

	pathParts := strings.Split(searchPath, "/")
	workingFile := fs.trees[UsersTreeKey].GetRoot()
	for _, pathPart := range pathParts {
		if pathPart == "" {
			continue
		}
		child, err := workingFile.GetChild(pathPart)
		if err != nil {
			return nil, err
		}
		if child != nil {
			workingFile = child
		}
	}

	return workingFile, nil
}

func (fs *FileServiceImpl) AddTask(f *file_model.WeblensFileImpl, t *task.Task) error {
	fs.fileTaskLock.Lock()
	defer fs.fileTaskLock.Unlock()
	tasks, ok := fs.fileTaskLink[f.ID()]
	if !ok {
		tasks = []*task.Task{}
	} else if slices.Contains(tasks, t) {
		return werror.ErrFileAlreadyHasTask
	}

	fs.fileTaskLink[f.ID()] = append(tasks, t)
	return nil
}

func (fs *FileServiceImpl) RemoveTask(f *file_model.WeblensFileImpl, t *task.Task) error {
	fs.fileTaskLock.Lock()
	defer fs.fileTaskLock.Unlock()
	tasks, ok := fs.fileTaskLink[f.ID()]
	if !ok {
		return werror.ErrFileNoTask
	}

	i := slices.Index(tasks, t)
	if i == -1 {
		return werror.ErrFileNoTask
	}

	fs.fileTaskLink[f.ID()] = slices.Delete(tasks, i, i+1)
	return nil
}

func (fs *FileServiceImpl) GetTasks(f *file_model.WeblensFileImpl) []*task.Task {
	fs.fileTaskLock.RLock()
	defer fs.fileTaskLock.RUnlock()
	return fs.fileTaskLink[f.ID()]
}

func (fs *FileServiceImpl) ResizeUp(f *file_model.WeblensFileImpl, event *fileTree.FileEvent) error {
	fs.log.Trace().Msgf("Resizing up [%s]", f.GetPortablePath())
	tree := fs.trees[f.GetPortablePath().RootName()]
	if tree == nil {
		return nil
	}

	err := tree.ResizeUp(f, event, nil)

	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) ResizeDown(f *file_model.WeblensFileImpl, event *fileTree.FileEvent) error {
	fs.log.Trace().Msgf("Resizing down [%s]", f.GetPortablePath())
	tree := fs.trees[f.GetPortablePath().RootName()]
	if tree == nil {
		return errors.WithStack(werror.ErrNoFileTree)
	}

	err := tree.ResizeDown(f, event, nil)

	if err != nil {
		return err
	}

	fs.log.Trace().Func(func(e *zerolog.Event) {
		if event == nil {
			return
		}
		e.Msgf("Resizing down event: %d", len(event.Actions))
	})

	return nil
}

func (fs *FileServiceImpl) GetThumbsDir() (*file_model.WeblensFileImpl, error) {
	cacheTree := fs.trees[CachesTreeKey]
	if cacheTree == nil {
		return nil, errors.WithStack(werror.ErrNoFileTree.WithArg(CachesTreeKey))
	}
	return cacheTree.GetRoot().GetChild(ThumbsDirName)
}

func (fs *FileServiceImpl) getFileByIdAndRoot(id file_model.FileId, rootAlias string) (*file_model.WeblensFileImpl, error) {
	tree := fs.trees[rootAlias]
	if tree == nil {
		return nil, werror.Errorf("Trying to get file on non-existent tree [%s]", rootAlias)
	}

	f := tree.Get(id)

	if f == nil {
		return nil, errors.WithStack(werror.ErrNoFile)
	}

	return f, nil
}

func (fs *FileServiceImpl) loadContentIdCache() error {
	fs.contentIdLock.Lock()
	defer fs.contentIdLock.Unlock()
	fs.contentIdCache = make(map[string]*file_model.WeblensFileImpl)

	fs.log.Trace().Msg("Loading contentId cache")

	_ = fs.trees[RestoreTreeKey].GetRoot().LeafMap(
		func(f *file_model.WeblensFileImpl) error {
			if f.IsDir() {
				return nil
			}
			fs.contentIdCache[f.Filename()] = f
			return nil
		},
	)

	if usersTree := fs.trees[UsersTreeKey]; usersTree != nil {
		_ = usersTree.GetRoot().LeafMap(
			func(f *file_model.WeblensFileImpl) error {
				if f.IsDir() {
					return nil
				}
				contentId := f.GetContentId()
				if contentId != "" {
					if _, ok := fs.contentIdCache[contentId]; !ok {
						fs.contentIdCache[contentId] = f
					}
				}
				return nil
			},
		)
	}

	return nil
}

func GenerateContentId(f *file_model.WeblensFileImpl) (string, error) {
	if f.IsDir() {
		return "", werror.Errorf("cannot hash directory")
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
	}

	fileSize := f.Size()

	if fileSize == 0 {
		return "", errors.WithStack(werror.ErrEmptyFile.WithArg(f.AbsPath()))
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
			// 	fs.log.Error().Stack().Err(err).Msg("")
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

func ContentIdFromHash(newHash hash.Hash) string {
	return base64.URLEncoding.EncodeToString(newHash.Sum(nil))[:20]
}

func MakeUniqueChildName(parent *file_model.WeblensFileImpl, childName string) string {
	dupeCount := 0
	_, e := parent.GetChild(childName)
	for e == nil {
		dupeCount++
		tmp := fmt.Sprintf("%s (%d)", childName, dupeCount)
		_, e = parent.GetChild(tmp)
	}

	newFilename := childName
	if dupeCount != 0 {
		newFilename = fmt.Sprintf("%s (%d)", newFilename, dupeCount)
	}

	return newFilename
}
