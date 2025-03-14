package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
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
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/task"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.FileService = (*FileServiceImpl)(nil)

const (
	UsersTreeKey   = "USERS"
	RestoreTreeKey = "RESTORE"
	CachesTreeKey  = "CACHES"

	UserTrashDirName = ".user_trash"
	ThumbsDirName    = "thumbs"
)

type FileServiceImpl struct {
	userService     models.UserService
	accessService   models.AccessService
	mediaService    models.MediaService
	instanceService models.InstanceService

	trees map[string]fileTree.FileTree

	contentIdCache map[models.ContentId]*fileTree.WeblensFileImpl

	folderMedia map[fileTree.FileId]models.ContentId

	fileTaskLink map[fileTree.FileId][]*task.Task

	folderCoverCol *mongo.Collection

	log *zerolog.Logger

	treesLock sync.RWMutex

	contentIdLock sync.RWMutex

	fileTaskLock sync.RWMutex
}

type TrashEntry struct {
	OrigParent   fileTree.FileId `bson:"originalParentId"`
	OrigFilename string          `bson:"originalFilename"`
	FileId       fileTree.FileId `bson:"fileId"`
}

type FolderCoverPair struct {
	FolderId  fileTree.FileId  `bson:"folderId"`
	ContentId models.ContentId `bson:"coverId"`
}

func NewFileService(
	logger *zerolog.Logger,
	instanceService models.InstanceService,
	userService models.UserService,
	accessService models.AccessService,
	mediaService models.MediaService,
	folderCoverCol *mongo.Collection,
	trees ...fileTree.FileTree,
) (*FileServiceImpl, error) {

	fs := &FileServiceImpl{
		log:             logger,
		trees:           map[string]fileTree.FileTree{},
		userService:     userService,
		instanceService: instanceService,
		accessService:   accessService,
		mediaService:    mediaService,
		folderCoverCol:  folderCoverCol,
		fileTaskLink:    make(map[fileTree.FileId][]*task.Task),
		folderMedia:     make(map[fileTree.FileId]models.ContentId),
	}

	for _, tree := range trees {
		fs.trees[tree.GetRoot().GetPortablePath().RootName()] = tree
	}

	sw := internal.NewStopwatch("File Service Init", logger)

	if usersTree, ok := fs.trees[UsersTreeKey]; ok {
		err := fs.ResizeDown(usersTree.GetRoot(), nil, nil)
		if err != nil {
			return nil, err
		}
		sw.Lap("Resize tree")
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

	sw.Stop()
	sw.PrintResults(false)

	return fs, nil
}

func (fs *FileServiceImpl) Size(treeAlias string) int64 {
	tree := fs.trees[treeAlias]
	if tree == nil {
		return -1
	}

	// total := 0
	// for _, tree := range fs.trees {
	// 	total += tree.Size()
	// }
	return tree.GetRoot().Size()
}

func (fs *FileServiceImpl) SetMediaService(mediaService *MediaServiceImpl) {
	fs.mediaService = mediaService
}

func (fs *FileServiceImpl) GetFileByTree(id fileTree.FileId, treeAlias string) (*fileTree.WeblensFileImpl, error) {
	return fs.getFileByIdAndRoot(id, treeAlias)
}

func (fs *FileServiceImpl) GetFileByContentId(contentId models.ContentId) (*fileTree.WeblensFileImpl, error) {
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

	return nil, werror.WithStack(werror.ErrNoFile)
}

func (fs *FileServiceImpl) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, []fileTree.FileId, error) {
	usersTree := fs.trees[UsersTreeKey]
	if usersTree == nil {
		return nil, nil, werror.WithStack(werror.ErrNoFileTree.WithArg(UsersTreeKey))
	}

	var files []*fileTree.WeblensFileImpl
	var lostFiles []fileTree.FileId
	for _, id := range ids {
		lt := usersTree.GetJournal().Get(id)
		if lt == nil {
			lostFiles = append(lostFiles, id)
			continue
			// return nil, nil, werror.WithStack(werror.ErrNoLifetime.WithArg(id))
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

func (fs *FileServiceImpl) GetFileSafe(id fileTree.FileId, user *models.User, share *models.FileShare) (
	*fileTree.WeblensFileImpl,
	error,
) {
	tree := fs.trees[UsersTreeKey]
	if tree == nil {
		return nil, werror.WithStack(werror.ErrNoFileTree)
	}

	f := tree.Get(id)
	if f == nil {
		return nil, werror.WithStack(werror.ErrNoFile.WithArg(id))
	}

	if !fs.accessService.CanUserAccessFile(user, f, share) {
		fs.log.Error().Stack().Err(werror.Errorf(
			"User [%s] attempted to access file at %s [%s], but they do not have access",
			user.GetUsername(), f.GetPortablePath(), f.ID(),
		)).Msg("")
		return nil, werror.WithStack(werror.ErrNoFileAccess)
	}

	return f, nil
}

func (fs *FileServiceImpl) GetFileTreeByName(treeName string) fileTree.FileTree {
	return fs.trees[treeName]
}

func (fs *FileServiceImpl) GetMediaCacheByFilename(thumbFileName string) (*fileTree.WeblensFileImpl, error) {
	thumbsDir, err := fs.trees[CachesTreeKey].GetRoot().GetChild(ThumbsDirName)
	if err != nil {
		return nil, err
	}
	return thumbsDir.GetChild(thumbFileName)
}

func (fs *FileServiceImpl) IsFileInTrash(f *fileTree.WeblensFileImpl) bool {
	return strings.Contains(f.AbsPath(), UserTrashDirName)
}

func (fs *FileServiceImpl) NewCacheFile(
	media *models.Media, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFileImpl, error) {
	filename := media.FmtCacheFileName(quality, pageNum)

	thumbsDir, err := fs.GetThumbsDir()
	if err != nil {
		return nil, err
	}

	return fs.trees[CachesTreeKey].Touch(thumbsDir, filename, nil)
}

func (fs *FileServiceImpl) DeleteCacheFile(f fileTree.WeblensFile) error {
	_, err := fs.trees[CachesTreeKey].Remove(f.ID())
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileServiceImpl) CreateFile(parent *fileTree.WeblensFileImpl, filename string, event *fileTree.FileEvent, caster models.FileCaster) (
	*fileTree.WeblensFileImpl, error,
) {
	newF, err := fs.trees[UsersTreeKey].Touch(parent, filename, event)
	if err != nil {
		return nil, err
	}

	return newF, nil
}

func (fs *FileServiceImpl) CreateFolder(parent *fileTree.WeblensFileImpl, folderName string, event *fileTree.FileEvent, caster models.FileCaster) (
	*fileTree.WeblensFileImpl,
	error,
) {

	newF, err := fs.trees[UsersTreeKey].MkDir(parent, folderName, event)
	if err != nil {
		return newF, err
	}

	caster.PushFileCreate(newF)

	return newF, nil
}

func (fs *FileServiceImpl) CreateUserHome(user *models.User) error {
	home, err := fs.trees[UsersTreeKey].MkDir(fs.trees[UsersTreeKey].GetRoot(), user.GetUsername(), nil)
	if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
		return err
	}
	user.SetHomeFolder(home)

	trash, err := fs.trees[UsersTreeKey].MkDir(home, UserTrashDirName, nil)
	if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
		return err
	}
	user.SetTrashFolder(trash)

	return nil
}

func (fs *FileServiceImpl) GetFileOwner(file *fileTree.WeblensFileImpl) (*models.User, error) {
	portable := file.GetPortablePath()
	if portable.RootName() != UsersTreeKey {
		return nil, werror.Errorf("trying to get owner of file not in MEDIA tree")
	}

	slashIndex := strings.Index(portable.RelativePath(), "/")
	var username models.Username
	if slashIndex == -1 {
		username = portable.RelativePath()
	} else {
		username = portable.RelativePath()[:slashIndex]
	}
	u := fs.userService.Get(username)

	return u, nil
}

func (fs *FileServiceImpl) MoveFilesToTrash(
	files []*fileTree.WeblensFileImpl, user *models.User, share *models.FileShare, caster models.FileCaster,
) error {
	if len(files) == 0 {
		return nil
	}

	owner, err := fs.GetFileOwner(files[0])
	if err != nil {
		return err
	}

	tree := fs.trees[UsersTreeKey]
	if tree == nil {
		return werror.WithStack(werror.ErrNoFileTree)
	}
	trash := tree.Get(owner.TrashId)
	if trash == nil {
		return werror.WithStack(errors.New("trash folder does not exist"))
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
			return werror.WithStack(werror.ErrNoFileAccess)
		}

		newFilename := MakeUniqueChildName(trash, file.Filename())
		preMoveFile := file.Freeze()

		_, err := tree.Move(file, trash, newFilename, false, event)
		if err != nil {
			return err
		}

		caster.PushFileMove(preMoveFile, file)
	}

	err = fs.ResizeUp(oldParent, event, caster)
	if err != nil {
		fs.log.Error().Stack().Err(err).Msg("")
	}

	err = fs.ResizeUp(trash, event, caster)
	if err != nil {
		fs.log.Error().Stack().Err(err).Msg("")
	}

	tree.GetJournal().LogEvent(event)
	event.Wait()

	return nil
}

func (fs *FileServiceImpl) ReturnFilesFromTrash(
	trashFiles []*fileTree.WeblensFileImpl, c models.FileCaster,
) error {
	trash := trashFiles[0].GetParent()
	trashPath := trash.GetPortablePath().ToPortable()

	tree := fs.trees[UsersTreeKey]
	if tree == nil {
		return werror.WithStack(werror.ErrNoFileTree)
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

		c.PushFileMove(preFile, trashEntry)
	}

	err := fs.ResizeUp(trash, event, c)
	if err != nil {
		return err
	}

	journal.LogEvent(event)

	return nil
}

// DeleteFiles removes files being pointed to from the tree and moves them to the restore tree
func (fs *FileServiceImpl) DeleteFiles(
	files []*fileTree.WeblensFileImpl, treeName string, caster models.FileCaster,
) error {
	tree := fs.trees[treeName]
	if tree == nil {
		return werror.WithStack(werror.ErrNoFileTree)
	}

	restoreTree := fs.trees[RestoreTreeKey]
	if restoreTree == nil {
		return werror.WithStack(werror.ErrNoFileTree)
	}

	deleteEvent := tree.GetJournal().NewEvent()

	if fs.instanceService.GetLocal().Role == models.BackupServerRole {
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

	var dirIds []fileTree.FileId

	var deletedFiles []*fileTree.WeblensFileImpl
	for _, file := range files {
		err := file.RecursiveMap(
			func(f *fileTree.WeblensFileImpl) error {

				// Freeze the file before it is deleted
				preDeleteFile := f.Freeze()
				contentId := f.GetContentId()
				m := fs.mediaService.Get(contentId)

				child, err := restoreTree.GetRoot().GetChild(f.ID())
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

						// Move file from users tree to the restore tree. Files later can be hard-linked back
						// from the restore tree to the users tree, but will not be moved back.
						err = fileTree.MoveFileBetweenTrees(
							f, restoreTree.GetRoot(), f.GetContentId(), tree, restoreTree,
							&fileTree.FileEvent{},
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

	caster.PushFilesDelete(deletedFiles)

	err := fs.ResizeUp(trash, deleteEvent, caster)
	if err != nil {
		return err
	}

	tree.GetJournal().LogEvent(deleteEvent)

	return nil
}

func (fs *FileServiceImpl) RestoreFiles(
	ids []fileTree.FileId, newParent *fileTree.WeblensFileImpl, restoreTime time.Time, caster models.FileCaster,
) error {
	usersTree := fs.trees[UsersTreeKey]
	if usersTree == nil {
		return werror.WithStack(werror.ErrNoFileTree.WithArg(UsersTreeKey))
	}

	restoreTree := fs.trees[RestoreTreeKey]
	if restoreTree == nil {
		return werror.WithStack(werror.ErrNoFileTree.WithArg(RestoreTreeKey))
	}

	journal := usersTree.GetJournal()
	event := journal.NewEvent()

	var topFiles []*fileTree.WeblensFileImpl
	type restorePair struct {
		newParent *fileTree.WeblensFileImpl
		fileId    fileTree.FileId
		contentId models.ContentId
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

		var childIds []fileTree.FileId
		if pastFile.IsDir() {
			children := pastFile.GetChildren()

			childIds = internal.Map(
				children, func(child *fileTree.WeblensFileImpl) fileTree.FileId {
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

		var restoredF *fileTree.WeblensFileImpl
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
				return werror.WithStack(err)
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
					return werror.WithStack(werror.ErrNoFile)
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
	event.Wait()

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

func (fs *FileServiceImpl) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFileImpl, error) {
	cacheTree := fs.trees[CachesTreeKey]
	if cacheTree == nil {
		return nil, werror.ErrNoFileTree
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

func (fs *FileServiceImpl) GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
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
	files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, treeName string, caster models.FileCaster,
) error {
	if len(files) == 0 {
		return nil
	}

	tree := fs.trees[treeName]

	event := tree.GetJournal().NewEvent()
	prevParent := files[0].GetParent()

	moveUpdates := map[string][]*fileTree.WeblensFileImpl{}

	for _, file := range files {
		preFile := file.Freeze()
		newFilename := MakeUniqueChildName(destFolder, file.Filename())

		_, err := tree.Move(file, destFolder, newFilename, false, event)
		if err != nil {
			return err
		}

		key := preFile.GetParentId() + "->" + file.GetParentId()
		if moveUpdates[key] == nil {
			moveUpdates[key] = []*fileTree.WeblensFileImpl{file}
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

func (fs *FileServiceImpl) RenameFile(file *fileTree.WeblensFileImpl, newName string, caster models.FileCaster) error {
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

func (fs *FileServiceImpl) NewBackupFile(lt *fileTree.Lifetime) (*fileTree.WeblensFileImpl, error) {
	filename := lt.GetLatestPath().Filename()

	tree := fs.trees[lt.ServerId]
	if tree == nil {
		return nil, werror.WithStack(werror.ErrNoFileTree.WithArg(lt.ServerId))
	}

	restoreTree := fs.trees[RestoreTreeKey]
	if restoreTree == nil {
		return nil, werror.WithStack(werror.ErrNoFileTree.WithArg(RestoreTreeKey))
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
			return nil, werror.WithStack(werror.ErrNoFile.WithArg(fmt.Sprintf("looking for parent of [%s]: %s", latestAction.DestinationPath, latestAction.ParentId)))
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
		return nil, werror.WithStack(werror.ErrNoContentId)
	} else if lt.GetContentId() == "" {
		return nil, nil
	}

	var restoreFile *fileTree.WeblensFileImpl
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
			return nil, werror.WithStack(err)
		}
	}

	if lt.GetLatestAction().ActionType != fileTree.FileDelete {
		portable := fileTree.ParsePortable(lt.GetLatestAction().DestinationPath)

		// Translate from the portable path to expand the absolute path
		// with the new backup tree
		newPortable := portable.OverwriteRoot(lt.ServerId)

		latestMove := lt.GetLatestMove()
		if latestMove == nil {
			return nil, werror.WithStack(werror.ErrNoFileAction.WithArg("no latest move"))
		}

		parent := tree.Get(latestMove.ParentId)
		if parent == nil {
			fs.log.Debug().Func(func(e *zerolog.Event) {
				e.Msgf("Parent [%s] not found trying to get parent for [%s]", latestMove.ParentId, lt.Id)
			})
			return nil, werror.WithStack(werror.ErrNoFile.WithArg("trying to get " + latestMove.ParentId))
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
				return nil, werror.WithStack(err)
			}
		}

		fs.log.Trace().Func(func(e *zerolog.Event) {
			e.Msgf("Linking %s -> %s", restoreFile.GetPortablePath().ToPortable(), portable.ToPortable())
		})
		err = os.Link(restoreFile.AbsPath(), newF.AbsPath())
		if err != nil {
			return nil, werror.WithStack(err)
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

func (fs *FileServiceImpl) SetFolderCover(folderId fileTree.FileId, coverId models.ContentId) error {
	tree := fs.trees[UsersTreeKey]
	if tree == nil {
		return werror.ErrNoFileTree
	}

	folder := tree.Get(folderId)
	if folder == nil {
		return werror.ErrNoFile
	}

	if coverId == "" {
		_, err := fs.folderCoverCol.DeleteOne(context.Background(), bson.M{"folderId": folderId})
		if err != nil {
			return werror.WithStack(err)
		}

		delete(fs.folderMedia, folderId)
		folder.SetContentId("")
		return nil

	} else if fs.folderMedia[folderId] != "" {
		_, err := fs.folderCoverCol.UpdateOne(
			context.Background(), bson.M{"folderId": folderId}, bson.M{"$set": bson.M{"coverId": coverId}},
		)
		if err != nil {
			return werror.WithStack(err)
		}
	} else {
		_, err := fs.folderCoverCol.InsertOne(context.Background(), bson.M{"folderId": folderId, "coverId": coverId})
		if err != nil {
			return werror.WithStack(err)
		}
	}

	fs.folderMedia[folderId] = coverId
	folder.SetContentId(coverId)

	return nil
}

func (fs *FileServiceImpl) GetFolderCover(folder *fileTree.WeblensFileImpl) (models.ContentId, error) {
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

func (fs *FileServiceImpl) UserPathToFile(searchPath string, user *models.User) (*fileTree.WeblensFileImpl, error) {

	if strings.HasPrefix(searchPath, "~/") {
		searchPath = string(user.GetUsername()) + "/" + searchPath[2:]
	} else if searchPath[:1] == "/" && user.IsAdmin() {
		searchPath = searchPath[1:]
	}

	return fs.PathToFile(searchPath)
}

func (fs *FileServiceImpl) PathToFile(searchPath string) (*fileTree.WeblensFileImpl, error) {
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

func (fs *FileServiceImpl) AddTask(f *fileTree.WeblensFileImpl, t *task.Task) error {
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

func (fs *FileServiceImpl) RemoveTask(f *fileTree.WeblensFileImpl, t *task.Task) error {
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

	fs.fileTaskLink[f.ID()] = internal.Banish(tasks, i)
	return nil
}

func (fs *FileServiceImpl) GetTasks(f *fileTree.WeblensFileImpl) []*task.Task {
	fs.fileTaskLock.RLock()
	defer fs.fileTaskLock.RUnlock()
	return fs.fileTaskLink[f.ID()]
}

func (fs *FileServiceImpl) ResizeUp(f *fileTree.WeblensFileImpl, event *fileTree.FileEvent, caster models.FileCaster) error {
	tree := fs.trees[f.GetPortablePath().RootName()]
	if tree == nil {
		return nil
	}

	err := tree.ResizeUp(f, event, func(f *fileTree.WeblensFileImpl) {
		if caster != nil {
			caster.PushFileUpdate(f, nil)
		}
	})

	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) ResizeDown(f *fileTree.WeblensFileImpl, event *fileTree.FileEvent, caster models.FileCaster) error {
	fs.log.Debug().Msgf("Resizing down %s", f.GetPortablePath())
	tree := fs.trees[f.GetPortablePath().RootName()]
	if tree == nil {
		return werror.ErrNoFileTree
	}

	err := tree.ResizeDown(f, event, func(f *fileTree.WeblensFileImpl) {
		if caster != nil {
			caster.PushFileUpdate(f, nil)
		}
	})

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

func (fs *FileServiceImpl) GetThumbsDir() (*fileTree.WeblensFileImpl, error) {
	cacheTree := fs.trees[CachesTreeKey]
	if cacheTree == nil {
		return nil, werror.WithStack(werror.ErrNoFileTree.WithArg(CachesTreeKey))
	}
	return cacheTree.GetRoot().GetChild(ThumbsDirName)
}

func (fs *FileServiceImpl) getFileByIdAndRoot(id fileTree.FileId, rootAlias string) (*fileTree.WeblensFileImpl, error) {
	tree := fs.trees[rootAlias]
	if tree == nil {
		return nil, werror.Errorf("Trying to get file on non-existent tree [%s]", rootAlias)
	}

	f := tree.Get(id)

	if f == nil {
		return nil, werror.WithStack(werror.ErrNoFile)
	}

	return f, nil
}

func (fs *FileServiceImpl) loadContentIdCache() error {
	fs.contentIdLock.Lock()
	defer fs.contentIdLock.Unlock()
	fs.contentIdCache = make(map[models.ContentId]*fileTree.WeblensFileImpl)

	fs.log.Trace().Msg("Loading contentId cache")

	_ = fs.trees[RestoreTreeKey].GetRoot().LeafMap(
		func(f *fileTree.WeblensFileImpl) error {
			if f.IsDir() {
				return nil
			}
			fs.contentIdCache[f.Filename()] = f
			return nil
		},
	)

	if usersTree := fs.trees[UsersTreeKey]; usersTree != nil {
		_ = usersTree.GetRoot().LeafMap(
			func(f *fileTree.WeblensFileImpl) error {
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

func GenerateContentId(f *fileTree.WeblensFileImpl) (models.ContentId, error) {
	if f.IsDir() {
		return "", werror.Errorf("cannot hash directory")
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
	}

	fileSize := f.Size()

	if fileSize == 0 {
		return "", werror.WithStack(werror.ErrEmptyFile.WithArg(f.AbsPath()))
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

func ContentIdFromHash(newHash hash.Hash) models.ContentId {
	return base64.URLEncoding.EncodeToString(newHash.Sum(nil))[:20]
}

func MakeUniqueChildName(parent *fileTree.WeblensFileImpl, childName string) string {
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
