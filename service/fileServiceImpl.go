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
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/task"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.FileService = (*FileServiceImpl)(nil)

type FileServiceImpl struct {
	trees     map[string]fileTree.FileTree
	treesLock sync.RWMutex

	contentIdCache map[models.ContentId]*fileTree.WeblensFileImpl
	contentIdLock  sync.RWMutex

	userService     models.UserService
	accessService   models.AccessService
	mediaService    models.MediaService
	instanceService models.InstanceService

	folderMedia     map[fileTree.FileId]models.ContentId
	folderMediaLock sync.RWMutex

	fileTaskLink map[fileTree.FileId][]*task.Task
	fileTaskLock sync.RWMutex

	trashCol       *mongo.Collection
	folderCoverCol *mongo.Collection
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
	instanceService models.InstanceService,
	userService models.UserService,
	accessService models.AccessService,
	mediaService models.MediaService,
	trashCol *mongo.Collection,
	folderCoverCol *mongo.Collection,
	trees ...fileTree.FileTree,
) (*FileServiceImpl, error) {
	fs := &FileServiceImpl{
		trees:           map[string]fileTree.FileTree{},
		userService:     userService,
		instanceService: instanceService,
		accessService:   accessService,
		mediaService:    mediaService,
		trashCol:        trashCol,
		folderCoverCol:  folderCoverCol,
		fileTaskLink:    make(map[fileTree.FileId][]*task.Task),
		folderMedia:     make(map[fileTree.FileId]models.ContentId),
	}

	for _, tree := range trees {
		fs.trees[tree.GetRoot().GetPortablePath().RootName()] = tree
	}

	sw := internal.NewStopwatch("File Service Init")

	if usersTree, ok := fs.trees["USERS"]; ok {
		err := fs.ResizeDown(usersTree.GetRoot(), nil)
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

	log.Trace.Println("Found folder covers", folderCovers)

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
	var files []*fileTree.WeblensFileImpl
	var lostFiles []fileTree.FileId
	for _, id := range ids {
		lt := fs.trees["USERS"].GetJournal().Get(id)
		if lt == nil {
			return nil, nil, werror.ErrNoLifetime
		}
		if lt.GetLatestAction().ActionType == fileTree.FileDelete {
			contentId := lt.GetContentId()
			if contentId == "" {
				lostFiles = append(lostFiles, id)
				continue
			}
			f, err := fs.trees["RESTORE"].GetRoot().GetChild(contentId)
			if err != nil {
				lostFiles = append(lostFiles, id)
				continue
			}
			files = append(files, f)
		} else {
			f := fs.trees["USERS"].Get(id)
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
	f := fs.trees["USERS"].Get(id)
	if f == nil {
		return nil, werror.WithStack(werror.ErrNoFile)
	}

	if !fs.accessService.CanUserAccessFile(user, f, share) {
		log.Warning.Printf(
			"Username [%s] attempted to access file at %s [%s], but they do not have access",
			user.GetUsername(), f.GetPortablePath(), f.ID(),
		)
		return nil, werror.WithStack(werror.ErrNoFileAccess)
	}

	return f, nil
}

func (fs *FileServiceImpl) GetMediaCacheByFilename(thumbFileName string) (*fileTree.WeblensFileImpl, error) {
	thumbsDir, err := fs.trees["CACHES"].GetRoot().GetChild("thumbs")
	if err != nil {
		return nil, err
	}
	return thumbsDir.GetChild(thumbFileName)
}

func (fs *FileServiceImpl) IsFileInTrash(f *fileTree.WeblensFileImpl) bool {
	return strings.Contains(f.AbsPath(), ".user_trash")
}

func (fs *FileServiceImpl) NewCacheFile(
	media *models.Media, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFileImpl, error) {
	filename := media.FmtCacheFileName(quality, pageNum)

	thumbsDir, err := fs.trees["CACHES"].GetRoot().GetChild("thumbs")
	if err != nil {
		return nil, err
	}

	return fs.trees["CACHES"].Touch(thumbsDir, filename, nil)
}

func (fs *FileServiceImpl) DeleteCacheFile(f fileTree.WeblensFile) error {
	_, err := fs.trees["CACHES"].Remove(f.ID())
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileServiceImpl) CreateFile(parent *fileTree.WeblensFileImpl, fileName string, event *fileTree.FileEvent) (
	*fileTree.WeblensFileImpl, error,
) {
	newF, err := fs.trees["USERS"].Touch(parent, fileName, event)
	if err != nil {
		return nil, err
	}

	return newF, nil
}

func (fs *FileServiceImpl) CreateFolder(parent *fileTree.WeblensFileImpl, folderName string, caster models.FileCaster) (
	*fileTree.WeblensFileImpl,
	error,
) {

	newF, err := fs.trees["USERS"].MkDir(parent, folderName, nil)
	if err != nil {
		return newF, err
	}

	caster.PushFileCreate(newF)

	return newF, nil
}

func (fs *FileServiceImpl) CreateUserHome(user *models.User) error {
	home, err := fs.trees["USERS"].MkDir(fs.trees["USERS"].GetRoot(), user.GetUsername(), nil)
	if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
		return err
	}
	user.SetHomeFolder(home)

	trash, err := fs.trees["USERS"].MkDir(home, ".user_trash", nil)
	if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
		return err
	}
	user.SetTrashFolder(trash)

	return nil
}

// func (fs *FileServiceImpl) CreateRestoreFile(lifetime *fileTree.Lifetime) (
// 	restoreFile *fileTree.WeblensFileImpl, err error,
// ) {
// 	restoreFile, err = fs.trees["RESTORE"].Touch(fs.trees["RESTORE"].GetRoot(), lifetime.GetContentId(), nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	restoreFile.SetContentId(lifetime.GetContentId())
//
// 	if lifetime.GetLatestAction().ActionType != fileTree.FileDelete {
// 		portable := fileTree.ParsePortable(lifetime.GetLatestAction().DestinationPath)
// 		newAbs, err := fs.trees["USERS"].PortableToAbs(portable)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		err = os.MkdirAll(filepath.Dir(newAbs), 0755)
// 		if err != nil {
// 			return nil, werror.WithStack(err)
// 		}
//
// 		err = os.Link(restoreFile.AbsPath(), newAbs)
// 		if err != nil {
// 			return nil, werror.WithStack(err)
// 		}
// 	}
//
// 	return restoreFile, nil
// }

func (fs *FileServiceImpl) GetFileOwner(file *fileTree.WeblensFileImpl) *models.User {
	portable := file.GetPortablePath()
	if portable.RootName() != "USERS" {
		panic(errors.New("trying to get owner of file not in MEDIA tree"))
	}
	slashIndex := strings.Index(portable.RelativePath(), "/")
	var username models.Username
	if slashIndex == -1 {
		username = portable.RelativePath()
	} else {
		username = portable.RelativePath()[:slashIndex]
	}
	u := fs.userService.Get(username)
	if u == nil {
		return fs.userService.GetRootUser()
	}
	return u
}

func (fs *FileServiceImpl) MoveFilesToTrash(
	files []*fileTree.WeblensFileImpl, user *models.User, share *models.FileShare, caster models.FileCaster,
) error {
	if len(files) == 0 {
		return nil
	}

	trashId := fs.GetFileOwner(files[0]).TrashId
	trash, err := fs.getFileByIdAndRoot(trashId, "USERS")
	event := fs.trees["USERS"].GetJournal().NewEvent()

	oldParent := files[0].GetParent()

	var trashEntries bson.A
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
		trashEntries = append(
			trashEntries,
			TrashEntry{
				OrigParent:   file.GetParentId(),
				OrigFilename: file.Filename(),
				FileId:       file.ID(),
			},
		)

		newFilename := MakeUniqueChildName(trash, file.Filename())
		preMoveFile := file.Freeze()

		_, err = fs.trees["USERS"].Move(file, trash, newFilename, false, event)
		if err != nil {
			return err
		}

		caster.PushFileMove(preMoveFile, file)
	}

	_, err = fs.trashCol.InsertMany(context.Background(), trashEntries)
	if err != nil {
		return err
	}

	fs.trees["USERS"].GetJournal().LogEvent(event)
	event.Wait()

	err = fs.ResizeUp(oldParent, caster)
	if err != nil {
		log.ErrTrace(err)
	}

	err = fs.ResizeUp(trash, caster)
	if err != nil {
		log.ErrTrace(err)
	}

	return nil
}

// DeleteFiles removes files being pointed to from the tree and moves them to the restore tree
func (fs *FileServiceImpl) ReturnFilesFromTrash(
	trashFiles []*fileTree.WeblensFileImpl, c models.FileCaster,
) error {
	fileIds := make([]fileTree.FileId, 0, len(trashFiles))
	for _, file := range trashFiles {
		fileIds = append(fileIds, file.ID())
	}
	filter := bson.D{{Key: "fileId", Value: bson.M{"$in": fileIds}}}
	ret, err := fs.trashCol.Find(context.Background(), filter)
	if err != nil {
		return werror.WithStack(err)
	}

	var trashEntries []TrashEntry
	err = ret.All(context.Background(), &trashEntries)
	if err != nil {
		return werror.WithStack(err)
	}

	if len(trashEntries) != len(trashFiles) {
		return werror.Errorf("ReturnFilesFromTrash: trashEntries count does not match trashFiles")
	}

	trash := trashFiles[0].GetParent()

	event := fs.trees["USERS"].GetJournal().NewEvent()
	for i, trashEntry := range trashEntries {
		preFile := trashFiles[i].Freeze()
		oldParent := fs.trees["USERS"].Get(trashEntry.OrigParent)
		if oldParent == nil {
			homeId := fs.GetFileOwner(trashFiles[i]).HomeId
			oldParent = fs.trees["USERS"].Get(homeId)
		}

		_, err = fs.trees["USERS"].Move(trashFiles[i], oldParent, trashEntry.OrigFilename, false, event)
		c.PushFileMove(preFile, trashFiles[i])

		if err != nil {
			return err
		}
	}
	fs.trees["USERS"].GetJournal().LogEvent(event)

	res, err := fs.trashCol.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}

	if res.DeletedCount != int64(len(trashEntries)) {
		return errors.New("delete trash entry did not get expected delete count")
	}

	err = fs.ResizeUp(trash, c)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) DeleteFiles(
	files []*fileTree.WeblensFileImpl, treeName string, caster models.FileCaster,
) error {
	tree := fs.trees[treeName]
	if tree == nil {
		return werror.WithStack(werror.ErrNoFileTree)
	}

	restoreTree := fs.trees["RESTORE"]
	if restoreTree == nil {
		return werror.WithStack(werror.ErrNoFileTree)
	}

	deleteEvent := tree.GetJournal().NewEvent()

	if fs.instanceService.GetLocal().Role == models.BackupServer {
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

	var deletedIds []fileTree.FileId
	var dirIds []fileTree.FileId

	for _, file := range files {
		if !fs.IsFileInTrash(file) {
			return werror.Errorf("cannot delete file not in trash")
		}

		deletedIds = append(deletedIds, file.ID())

		err := file.RecursiveMap(
			func(f *fileTree.WeblensFileImpl) error {

				// Freeze the file before it is deleted
				preDeleteFile := f.Freeze()

				// Check if the file is already in the index, if it does, and the file exists as well,
				// we can just delete it from the users tree
				child, err := restoreTree.GetRoot().GetChild(f.ID())
				if err == nil && child.Exists() {
					err = tree.Delete(f.ID(), deleteEvent)
					if err != nil {
						return err
					}
					caster.PushFileDelete(preDeleteFile)

					return nil
				}

				if f.IsDir() || f.Size() == 0 {
					// Save directory ids to be removed after all files have been moved
					dirIds = append(dirIds, f.ID())
				} else {
					// Check if the restore file already exists, with the filename being the content id
					contentId := f.GetContentId()
					if contentId == "" {
						return werror.Errorf("trying to move file to restore tree without content id")
					}
					_, err = restoreTree.GetRoot().GetChild(contentId)

					if err != nil {
						// A non-nil error here means the file does not exist, so we must move it to the restore tree

						// Add the delete for this file to the event
						// We must do this before moving/deleting the file, or the action will not be able to find the file
						deleteEvent.NewDeleteAction(f.ID())

						// Move file from users tree to the restore tree. Files later can be hard-linked back
						// from the restore tree to the users tree, but will not be moved back.
						err = fileTree.MoveFileBetweenTrees(
							f, restoreTree.GetRoot(), f.GetContentId(), tree, restoreTree,
							&fileTree.FileEvent{},
						)
						if err != nil {
							return err
						}

						log.Trace.Printf("File [%s] moved from users tree to restore tree", f.GetPortablePath())

					} else {
						// If the file already is in the restore tree, we can just delete it from the users tree.
						// This should be rare since we already checked if the file exists in the index, but it is possible
						// if the index is missing or otherwise out of sync.
						err = tree.Delete(f.ID(), deleteEvent)
						if err != nil {
							return err
						}

						log.Trace.Printf(
							"File [%s] already exists in restore tree, deleting from users tree",
							f.GetPortablePath(),
						)
					}
				}
				caster.PushFileDelete(preDeleteFile)

				return nil
			},
		)
		if err != nil {
			return err
		}
	}

	// We need to make sure we delete the bottom most directories first,
	// since deleting a directory that is not empty will error. So we save
	// the rest until here, and then delete them in reverse order.
	slices.Reverse(dirIds)
	for _, dirId := range dirIds {
		err := tree.Delete(dirId, deleteEvent)
		if err != nil {
			return err
		}
	}

	tree.GetJournal().LogEvent(deleteEvent)
	deleteEvent.Wait()

	_, err := fs.trashCol.DeleteMany(context.Background(), bson.M{"fileId": bson.M{"$in": deletedIds}})
	if err != nil {
		return err
	}

	err = fs.ResizeUp(trash, caster)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) RestoreFiles(
	ids []fileTree.FileId, newParent *fileTree.WeblensFileImpl, restoreTime time.Time, caster models.FileCaster,
) error {
	journal := fs.trees["USERS"].GetJournal()
	event := journal.NewEvent()

	var topFiles []*fileTree.WeblensFileImpl
	type restorePair struct {
		fileId    fileTree.FileId
		contentId models.ContentId
		newParent *fileTree.WeblensFileImpl
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
			children, err := journal.GetPastFolderChildren(pastFile, restoreTime)
			if err != nil {
				return err
			}

			childIds = internal.Map(
				children, func(child *fileTree.WeblensFileImpl) fileTree.FileId {
					return child.ID()
				},
			)
		}

		path := pastFile.GetPortablePath().ToPortable()
		// Paths of directory files will have an extra / on the end, so we need to remove it
		if pastFile.IsDir() {
			path = path[:len(path)-1]
		}

		oldName := filepath.Base(path)

		var restoredF *fileTree.WeblensFileImpl
		if !pastFile.IsDir() {
			var existingPath string

			// File has been deleted, get the file from the restore tree
			if liveF := fs.trees["USERS"].Get(toRestore.fileId); liveF == nil {
				_, err = fs.trees["RESTORE"].GetRoot().GetChild(toRestore.contentId)
				if err != nil {
					return err
				}
				existingPath = filepath.Join(fs.trees["RESTORE"].GetRoot().AbsPath(), toRestore.contentId)
			} else {
				existingPath = liveF.AbsPath()
			}

			restoredF = fileTree.NewWeblensFile(
				fs.trees["USERS"].GenerateFileId(), oldName, toRestore.newParent, pastFile.IsDir(),
			)
			restoredF.SetContentId(pastFile.GetContentId())
			restoredF.SetSize(pastFile.Size())
			err = fs.trees["USERS"].Add(restoredF)
			if err != nil {
				return err
			}

			log.Trace.Printf("Restoring file [%s] to [%s]", existingPath, restoredF.AbsPath())
			err = os.Link(existingPath, restoredF.AbsPath())
			if err != nil {
				return werror.WithStack(err)
			}

			if toRestore.newParent == newParent {
				topFiles = append(topFiles, restoredF)
			}

		} else {
			restoredF = fileTree.NewWeblensFile(
				fs.trees["USERS"].GenerateFileId(), oldName, toRestore.newParent, true,
			)
			err = fs.trees["USERS"].Add(restoredF)
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

	journal.LogEvent(event)

	event.Wait()

	for _, f := range topFiles {
		err := fs.ResizeDown(f, caster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServiceImpl) RestoreHistory(lifetimes []*fileTree.Lifetime) error {

	journal := fs.trees["USERS"].GetJournal()

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
		if fs.trees["USERS"].Get(lt.ID()) != nil {
			continue
		}

		// parentId := latest.GetParentId()
		parent, err := fs.getFileByIdAndRoot(latest.GetParentId(), "USERS")
		if err != nil {
			return err
		}

		newF := fileTree.NewWeblensFile(lt.ID(), portable.Filename(), parent, true)
		err = fs.trees["USERS"].Add(newF)
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

func (fs *FileServiceImpl) ReadFile(f *fileTree.WeblensFileImpl) (io.ReadCloser, error) {
	panic("not implemented")
}

func (fs *FileServiceImpl) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFileImpl, error) {
	cacheTree := fs.trees["CACHES"]
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
	takeoutFile := fs.trees["CACHES"].Get(id)
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

	for _, file := range files {
		preFile := file.Freeze()
		newFilename := MakeUniqueChildName(destFolder, file.Filename())

		_, err := tree.Move(file, destFolder, newFilename, false, event)
		if err != nil {
			return err
		}

		caster.PushFileMove(preFile, file)
	}

	tree.GetJournal().LogEvent(event)

	err := fs.ResizeUp(destFolder, caster)
	if err != nil {
		return err
	}

	err = fs.ResizeUp(prevParent, caster)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) RenameFile(file *fileTree.WeblensFileImpl, newName string, caster models.FileCaster) error {
	preFile := file.Freeze()
	_, err := fs.trees["USERS"].Move(file, file.GetParent(), newName, false, nil)
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
		return nil, werror.WithStack(werror.ErrNoFileTree)
	}

	if lt.GetIsDir() {
		// If there is no path (i.e. the dir has been deleted), skip as there is
		// no need to create a directory that no longer exists, just so long as it is
		// included in the history, which is now is.
		if lt.GetLatestPath().RootName() == "" {
			log.Trace.Printf("Skipping dir that has no dest path")
			return nil, nil
		}

		// Find the directory's parent. This should already exist since we always create
		// the backup file structure in order from parent to child.
		parent := tree.Get(lt.GetLatestAction().ParentId)
		if parent == nil {
			return nil, werror.WithStack(werror.ErrNoFile)
		}

		// Create the directory object and add it to the tree
		newDir := fileTree.NewWeblensFile(lt.ID(), filename, parent, true)
		err := tree.Add(newDir)
		if err != nil {
			return nil, err
		}

		log.Trace.Println("Creating backup dir", newDir.GetPortablePath())

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

	restoreFile, err := fs.trees["RESTORE"].Touch(fs.trees["RESTORE"].GetRoot(), lt.GetContentId(), nil)
	if err != nil {
		return nil, err
	}
	restoreFile.SetContentId(lt.GetContentId())

	if lt.GetLatestAction().ActionType != fileTree.FileDelete {
		portable := fileTree.ParsePortable(lt.GetLatestAction().DestinationPath)

		// Translate from the portable path to expand the absolute path
		// with the new backup tree
		newPortable := portable.OverwriteRoot(lt.ServerId)

		parent := tree.Get(lt.GetLatestMove().ParentId)
		if parent == nil {
			return nil, werror.WithStack(werror.ErrNoFile)
		}

		newF := fileTree.NewWeblensFile(lt.ID(), newPortable.Filename(), parent, false)

		err = tree.Add(newF)
		if err != nil {
			return nil, err
		}

		log.Trace.Printf("Linking %s -> %s", restoreFile.GetPortablePath().ToPortable(), portable.ToPortable())
		err = os.Link(restoreFile.AbsPath(), newF.AbsPath())
		if err != nil {
			return nil, werror.WithStack(err)
		}
	}

	return restoreFile, nil
}

func (fs *FileServiceImpl) GetUsersRoot() *fileTree.WeblensFileImpl {
	return fs.trees["USERS"].GetRoot()
}

func (fs *FileServiceImpl) GetJournalByTree(treeName string) fileTree.Journal {
	tree := fs.trees[treeName]
	if tree == nil {
		return nil
	}
	return tree.GetJournal()
}

func (fs *FileServiceImpl) SetFolderCover(folderId fileTree.FileId, coverId models.ContentId) error {
	tree := fs.trees["USERS"]
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

	coverId := fs.folderMedia[folder.ID()]

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
	// path, err := fs.trees["USERS"].AbsToPortable(searchPath)
	// if err != nil {
	// 	return nil, err
	// }
	searchPath = strings.TrimPrefix(searchPath, "USERS:")

	pathParts := strings.Split(searchPath, "/")
	workingFile := fs.trees["USERS"].GetRoot()
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

	// if strings.HasPrefix(searchPath, "~/") {
	// 	searchPath = "MEDIA:" + string(u.GetUsername()) + "/" + searchPath[2:]
	// } else if searchPath[:1] == "/" && u.IsAdmin() {
	// 	searchPath = "MEDIA:" + searchPath[1:]
	// } else {
	// 	return nil, nil, werror.Errorf("Bad search path: %s", searchPath)
	// }
	//
	// lastSlashIndex := strings.LastIndex(searchPath, "/")
	// if lastSlashIndex == -1 {
	// 	if !strings.HasSuffix(searchPath, "/") {
	// 		searchPath += "/"
	// 	}
	// 	lastSlashIndex = len(searchPath) - 1
	// }
	// folderId := fs.trees["USERS"].GenerateFileId()
	//
	// folder, err := fs.GetFileSafe(folderId, u, share)
	// if err != nil {
	// 	// ctx.JSON(http.StatusOK, gin.H{"children": []string{}, "folder": nil})
	// 	return nil, nil, err
	// }
	//
	// postFix := searchPath[lastSlashIndex+1:]
	// allChildren := folder.GetChildren()
	// childNames := internal.Map(
	// 	allChildren, func(c *fileTree.WeblensFileImpl) string {
	// 		return c.Filename()
	// 	},
	// )
	//
	// matches := fuzzy.RankFindFold(postFix, childNames)
	// slices.SortFunc(
	// 	matches, func(a, b fuzzy.Rank) int {
	// 		diff := a.Distance - b.Distance
	// 		if diff != 0 {
	// 			return diff
	// 		}
	//
	// 		return allChildren[a.OriginalIndex].ModTime().Compare(allChildren[b.OriginalIndex].ModTime())
	// 	},
	// )
	//
	// children := internal.FilterMap(
	// 	matches, func(match fuzzy.Rank) (*fileTree.WeblensFileImpl, bool) {
	// 		f := allChildren[match.OriginalIndex]
	// 		if f.ID() == u.TrashId {
	// 			return nil, false
	// 		}
	// 		return f, true
	// 	},
	// )
	//
	// return folder, children, nil
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

func (fs *FileServiceImpl) ResizeUp(f *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	tree := fs.trees["USERS"]
	if tree == nil {
		return nil
	}

	journal := tree.GetJournal()
	event := journal.NewEvent()
	if err := f.BubbleMap(
		func(w *fileTree.WeblensFileImpl) error {
			return handleFileResize(w, journal, event, caster)
		},
	); err != nil {
		return err
	}

	log.Trace.Printf("Resizing up event: %d", len(event.Actions))
	tree.GetJournal().LogEvent(event)
	event.Wait()

	return nil
}

func (fs *FileServiceImpl) ResizeDown(f *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	tree := fs.trees[f.GetPortablePath().RootName()]
	if tree == nil {
		return werror.WithStack(werror.ErrNoFileTree)
	}

	journal := tree.GetJournal()
	event := journal.NewEvent()

	if err := f.LeafMap(
		func(w *fileTree.WeblensFileImpl) error {
			return handleFileResize(w, journal, event, caster)
		},
	); err != nil {
		return err
	}

	log.Trace.Printf("Resizing down event: %d", len(event.Actions))

	journal.LogEvent(event)
	event.Wait()

	return nil
}

func (fs *FileServiceImpl) resizeMultiple(old, new *fileTree.WeblensFileImpl, caster models.FileCaster) (err error) {
	// Check if either of the files are a parent of the other
	oldIsParent := strings.HasPrefix(old.AbsPath(), new.AbsPath())
	newIsParent := strings.HasPrefix(new.AbsPath(), old.AbsPath())

	if oldIsParent || !newIsParent {
		err = fs.ResizeUp(old, caster)
		if err != nil {
			return
		}
	}

	if newIsParent || !oldIsParent {
		err = fs.ResizeUp(new, caster)
		if err != nil {
			return
		}
	}

	return
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

	log.Trace.Println("Loading contentId cache")

	_ = fs.trees["RESTORE"].GetRoot().LeafMap(
		func(f *fileTree.WeblensFileImpl) error {
			if f.IsDir() {
				return nil
			}
			fs.contentIdCache[f.Filename()] = f
			return nil
		},
	)

	if usersTree := fs.trees["USERS"]; usersTree != nil {
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

func handleFileResize(
	file *fileTree.WeblensFileImpl, journal fileTree.Journal, event *fileTree.FileEvent, caster models.FileCaster,
) error {
	// if journal.IgnoreLocal() {
	// 	return nil
	// }
	//
	// if file.ID() == "ROOT" {
	// 	return nil
	// }
	newSize, err := file.LoadStat()
	if err != nil {
		return err
	}
	if newSize != -1 && !journal.IgnoreLocal() && file.ID() != "ROOT" {
		if caster != nil {
			caster.PushFileUpdate(file, nil)
		}

		lt := journal.Get(file.ID())

		if lt == nil {
			return werror.Errorf("journal does not have lifetime [%s] to resize", file.ID())
		}
		latestSize := lt.GetLatestSize()
		if latestSize != newSize {
			log.Trace.Printf("Size change for [%s] detected %d -> %d", file.GetPortablePath(), latestSize, newSize)
			event.NewSizeChangeAction(file)
		}
	}

	return err
}

func GenerateContentId(f *fileTree.WeblensFileImpl) (models.ContentId, error) {
	if f.IsDir() {
		werror.Errorf("cannot hash directory")
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
		// t.Success("Skipping file which already has content ID", meta.File.AbsPath())
	}

	fileSize := f.Size()

	if fileSize == 0 {
		return "", nil
		// t.Success("Skipping file with no content: ", meta.File.AbsPath())
	}

	if f.IsDir() {
		return "", nil
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
			err := fp.Close()
			if err != nil {
				log.ShowErr(err)
			}
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
