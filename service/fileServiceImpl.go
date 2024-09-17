package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.FileService = (*FileServiceImpl)(nil)

type FileServiceImpl struct {
	usersTree   fileTree.FileTree
	cachesTree  fileTree.FileTree
	restoreTree fileTree.FileTree

	userService     models.UserService
	accessService   models.AccessService
	mediaService    models.MediaService
	instanceService models.InstanceService

	fileTaskLink map[fileTree.FileId][]*task.Task
	fileTaskLock sync.RWMutex
	trashCol     *mongo.Collection
}

func NewFileService(
	mediaTree, cacheTree, restoreFileTree fileTree.FileTree, userService models.UserService,
	accessService models.AccessService, mediaService models.MediaService, trashCol *mongo.Collection,
) (*FileServiceImpl, error) {
	fs := &FileServiceImpl{
		usersTree:     mediaTree,
		restoreTree:   restoreFileTree,
		userService:   userService,
		cachesTree:    cacheTree,
		accessService: accessService,
		mediaService:  mediaService,
		trashCol:      trashCol,
		fileTaskLink:  make(map[fileTree.FileId][]*task.Task),
	}

	sw := internal.NewStopwatch("File Service Init")

	_, err := cacheTree.MkDir(cacheTree.GetRoot(), "thumbs", nil)
	if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
		return nil, werror.WithStack(err)
	}
	sw.Lap("Make thumbs dir")

	err = fs.ResizeDown(mediaTree.GetRoot(), nil)
	if err != nil {
		return nil, err
	}

	sw.Lap("Resize tree")
	sw.Stop()
	sw.PrintResults(false)

	return fs, nil
}

func (fs *FileServiceImpl) Size() int {
	return fs.usersTree.Size()
}

func (fs *FileServiceImpl) SetMediaService(mediaService *MediaServiceImpl) {
	fs.mediaService = mediaService
}

func (fs *FileServiceImpl) GetUserFile(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
	return fs.getFileByIdAndRoot(id, "USERS")
}

func (fs *FileServiceImpl) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, error) {
	var files []*fileTree.WeblensFileImpl
	for _, id := range ids {
		f, err := fs.getFileByIdAndRoot(id, "USERS")
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (fs *FileServiceImpl) GetFileSafe(id fileTree.FileId, user *models.User, share *models.FileShare) (
	*fileTree.WeblensFileImpl,
	error,
) {
	f := fs.usersTree.Get(id)
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
	thumbsDir, err := fs.cachesTree.GetRoot().GetChild("thumbs")
	if err != nil {
		return nil, err
	}
	return thumbsDir.GetChild(thumbFileName)
}

func (fs *FileServiceImpl) IsFileInTrash(f *fileTree.WeblensFileImpl) bool {
	return strings.Contains(f.AbsPath(), ".user_trash")
}

func (fs *FileServiceImpl) ImportFile(f *fileTree.WeblensFileImpl) error {
	return fs.usersTree.Add(f)
}

func (fs *FileServiceImpl) NewCacheFile(
	media *models.Media, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFileImpl, error) {
	filename := media.FmtCacheFileName(quality, pageNum)

	thumbsDir, err := fs.cachesTree.GetRoot().GetChild("thumbs")
	if err != nil {
		return nil, err
	}

	return fs.cachesTree.Touch(thumbsDir, filename, nil)
}

func (fs *FileServiceImpl) DeleteCacheFile(f fileTree.WeblensFile) error {
	_, err := fs.cachesTree.Remove(f.ID())
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileServiceImpl) CreateFile(parent *fileTree.WeblensFileImpl, fileName string, event *fileTree.FileEvent) (
	*fileTree.WeblensFileImpl, error,
) {
	newF, err := fs.usersTree.Touch(parent, fileName, event)
	if err != nil {
		return nil, err
	}

	return newF, nil
}

func (fs *FileServiceImpl) CreateFolder(parent *fileTree.WeblensFileImpl, folderName string, caster models.FileCaster) (
	*fileTree.WeblensFileImpl,
	error,
) {

	newF, err := fs.usersTree.MkDir(parent, folderName, nil)
	if err != nil {
		return newF, err
	}

	caster.PushFileCreate(newF)

	return newF, nil
}

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
	event := fs.usersTree.GetJournal().NewEvent()

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

		_, err = fs.usersTree.Move(file, trash, newFilename, false, event)
		if err != nil {
			return err
		}

		caster.PushFileMove(preMoveFile, file)
	}

	_, err = fs.trashCol.InsertMany(context.Background(), trashEntries)
	if err != nil {
		return err
	}

	fs.usersTree.GetJournal().LogEvent(event)
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

	event := fs.usersTree.GetJournal().NewEvent()
	for i, trashEntry := range trashEntries {
		preFile := trashFiles[i].Freeze()
		oldParent := fs.usersTree.Get(trashEntry.OrigParent)
		if oldParent == nil {
			homeId := fs.GetFileOwner(trashFiles[i]).HomeId
			oldParent = fs.usersTree.Get(homeId)
		}

		_, err = fs.usersTree.Move(trashFiles[i], oldParent, trashEntry.OrigFilename, false, event)
		c.PushFileMove(preFile, trashFiles[i])

		if err != nil {
			return err
		}
	}
	fs.usersTree.GetJournal().LogEvent(event)

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

type restoreFileInfo struct {
	Path      string            `json:"path"`
	IsDir     bool              `json:"isDir"`
	Children  []fileTree.FileId `json:"children,omitempty"`
	ContentId string            `json:"contentId,omitempty"`
}

// DeleteFiles removes files being pointed to from the tree and moves them to the restore tree
func (fs *FileServiceImpl) DeleteFiles(files []*fileTree.WeblensFileImpl, caster models.FileCaster) error {
	deleteEvent := fs.usersTree.GetJournal().NewEvent()

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
				child, err := fs.restoreTree.GetRoot().GetChild(f.ID())
				if err == nil && child.Exists() {
					err = fs.usersTree.Delete(f.ID(), deleteEvent)
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
					_, err = fs.restoreTree.GetRoot().GetChild(contentId)

					if err != nil {
						// A non-nil error here means the file does not exist, so we must move it to the restore tree

						// Add the delete for this file to the event
						// We must do this before moving/deleting the file, or the action will not be able to find the file
						deleteEvent.NewDeleteAction(f.ID())

						// Move file from users tree to the restore tree. Files later can be hard-linked back
						// from the restore tree to the users tree, but will not be moved back.
						err = fileTree.MoveFileBetweenTrees(
							f, fs.restoreTree.GetRoot(), f.GetContentId(), fs.usersTree, fs.restoreTree,
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
						err = fs.usersTree.Delete(f.ID(), deleteEvent)
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
		err := fs.usersTree.Delete(dirId, deleteEvent)
		if err != nil {
			return err
		}
	}

	fs.usersTree.GetJournal().LogEvent(deleteEvent)
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
	journal := fs.usersTree.GetJournal()
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
			if liveF := fs.usersTree.Get(toRestore.fileId); liveF == nil {
				_, err = fs.restoreTree.GetRoot().GetChild(toRestore.contentId)
				if err != nil {
					return err
				}
				existingPath = filepath.Join(fs.restoreTree.GetRoot().AbsPath(), toRestore.contentId)
			} else {
				existingPath = liveF.AbsPath()
			}

			restoredF = fileTree.NewWeblensFile(
				fs.usersTree.GenerateFileId(), oldName, toRestore.newParent, pastFile.IsDir(),
			)
			restoredF.SetContentId(pastFile.GetContentId())
			restoredF.SetSize(pastFile.Size())
			err = fs.usersTree.Add(restoredF)
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
				fs.usersTree.GenerateFileId(), oldName, toRestore.newParent, true,
			)
			err = fs.usersTree.Add(restoredF)
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

func (fs *FileServiceImpl) ReadFile(f *fileTree.WeblensFileImpl) (io.ReadCloser, error) {
	panic("not implemented")
}

func (fs *FileServiceImpl) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFileImpl, error) {
	cacheRoot := fs.cachesTree.GetRoot()
	takeoutDir, err := cacheRoot.GetChild("takeout")
	if err != nil {
		return nil, err
	}

	zipFile, err := fs.cachesTree.Touch(takeoutDir, zipName, nil)
	if err != nil {
		return nil, err
	}

	return zipFile, nil
}

func (fs *FileServiceImpl) GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
	takeoutFile := fs.cachesTree.Get(id)
	if takeoutFile == nil {
		return nil, werror.ErrNoFile
	}
	if takeoutFile.GetParent().Filename() != "takeout" {
		return nil, werror.ErrNoFile
	}

	return takeoutFile, nil
}

func (fs *FileServiceImpl) MoveFiles(
	files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, caster models.FileCaster,
) error {
	if len(files) == 0 {
		return nil
	}

	event := fs.usersTree.GetJournal().NewEvent()
	prevParent := files[0].GetParent()

	for _, file := range files {
		preFile := file.Freeze()
		newFilename := MakeUniqueChildName(destFolder, file.Filename())

		_, err := fs.usersTree.Move(file, destFolder, newFilename, false, event)
		if err != nil {
			return err
		}

		caster.PushFileMove(preFile, file)
	}

	fs.usersTree.GetJournal().LogEvent(event)

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
	_, err := fs.usersTree.Move(file, file.GetParent(), newName, false, nil)
	if err != nil {
		return err
	}

	caster.PushFileMove(preFile, file)

	return nil
}

func (fs *FileServiceImpl) GetMediaRoot() *fileTree.WeblensFileImpl {
	return fs.usersTree.GetRoot()
}

func (fs *FileServiceImpl) GetUsersJournal() fileTree.Journal {
	return fs.usersTree.GetJournal()
}

func (fs *FileServiceImpl) PathToFile(searchPath string) (*fileTree.WeblensFileImpl, error) {
	// path, err := fs.usersTree.AbsToPortable(searchPath)
	// if err != nil {
	// 	return nil, err
	// }
	if strings.HasPrefix(searchPath, "USERS:") {
		searchPath = searchPath[6:]
	}

	pathParts := strings.Split(searchPath, "/")
	workingFile := fs.usersTree.GetRoot()
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
	// folderId := fs.usersTree.GenerateFileId()
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

func GenerateContentId(f *fileTree.WeblensFileImpl) (models.ContentId, error) {
	if f.IsDir() {
		return "", nil
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
	}

	fileSize := f.Size()

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

func (fs *FileServiceImpl) ResizeUp(f *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	journal := fs.usersTree.GetJournal()
	event := journal.NewEvent()
	if err := f.BubbleMap(
		func(w *fileTree.WeblensFileImpl) error {
			return handleFileResize(w, journal, event, caster)
		},
	); err != nil {
		return err
	}

	log.Trace.Printf("Resizing up event: %d", len(event.Actions))
	fs.usersTree.GetJournal().LogEvent(event)
	event.Wait()

	return nil
}

func (fs *FileServiceImpl) ResizeDown(f *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	journal := fs.usersTree.GetJournal()
	event := journal.NewEvent()

	if err := f.LeafMap(
		func(w *fileTree.WeblensFileImpl) error {
			return handleFileResize(w, journal, event, caster)
		},
	); err != nil {
		return err
	}

	log.Trace.Printf("Resizing down event: %d", len(event.Actions))

	fs.usersTree.GetJournal().LogEvent(event)
	event.Wait()

	return nil
}

func handleFileResize(
	file *fileTree.WeblensFileImpl, journal fileTree.Journal, event *fileTree.FileEvent, caster models.FileCaster,
) error {
	if file.ID() == "ROOT" {
		return nil
	}
	newSize, err := file.LoadStat()
	if err != nil {
		return err
	}
	if newSize != -1 {
		if caster != nil {
			caster.PushFileUpdate(file, nil)
		}

		lt := journal.Get(file.ID())

		if lt == nil {
			return werror.Errorf("journal does not have lifetime to resize")
		}
		latestSize := lt.GetLatestSize()
		if latestSize != newSize {
			log.Trace.Printf("Size change for [%s] detected %d -> %d", file.GetPortablePath(), latestSize, newSize)
			event.NewSizeChangeAction(file)
		}
	}
	return err
}

func (fs *FileServiceImpl) getFileByIdAndRoot(id fileTree.FileId, rootAlias string) (*fileTree.WeblensFileImpl, error) {
	var f *fileTree.WeblensFileImpl
	switch rootAlias {
	case "USERS":
		f = fs.usersTree.Get(id)
	case "CACHES":
		f = fs.cachesTree.Get(id)
	default:
		return nil, werror.Errorf("Trying to get file on non-existent tree [%s]", rootAlias)
	}

	if f == nil {
		return nil, werror.ErrNoFile
	}

	return f, nil
}

func (fs *FileServiceImpl) clearTempDir() (err error) {
	tmpRoot := fs.cachesTree.Get("TMP")
	err = os.MkdirAll(tmpRoot.AbsPath(), os.ModePerm)
	if err != nil {
		return
	}

	files, err := os.ReadDir(tmpRoot.AbsPath())
	if err != nil {
		return
	}

	for _, file := range files {
		err := os.RemoveAll(filepath.Join(env.GetTmpDir(), file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServiceImpl) clearTakeoutDir() error {
	takeoutRoot := fs.cachesTree.Get("TAKEOUT")
	err := os.MkdirAll(takeoutRoot.AbsPath(), os.ModePerm)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(takeoutRoot.AbsPath())
	if err != nil {
		return err
	}

	for _, file := range files {
		err := os.Remove(filepath.Join(takeoutRoot.AbsPath(), file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServiceImpl) clearThumbsDir() error {
	_, err := fs.cachesTree.Remove("THUMBS")
	if err != nil {
		return err
	}

	_, err = fs.cachesTree.MkDir(fs.cachesTree.GetRoot(), "thumbs", nil)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileServiceImpl) newIndexFile(indexName string) (*fileTree.WeblensFileImpl, error) {
	indexFile, err := fs.restoreTree.Touch(fs.restoreTree.GetRoot(), indexName, &fileTree.FileEvent{})
	return indexFile, err
}

func readRestoreIndexFile(indexFile *fileTree.WeblensFileImpl) (map[string]restoreFileInfo, error) {
	if !indexFile.Exists() {
		return map[string]restoreFileInfo{}, nil
	}

	indexBytes, err := indexFile.ReadAll()
	if err != nil {
		return nil, err
	}

	indexData := map[string]restoreFileInfo{}
	if len(indexBytes) != 0 {
		err = json.Unmarshal(indexBytes, &indexData)
		if err != nil {
			return nil, werror.WithStack(err)
		}
	}

	return indexData, nil
}

func writeRestoreIndexFile(indexFile *fileTree.WeblensFileImpl, data map[string]restoreFileInfo) error {
	if len(data) != 0 {
		newIndexBytes, err := json.Marshal(data)
		if err != nil {
			return werror.WithStack(err)
		}

		_, err = indexFile.Write(newIndexBytes)
		if err != nil {
			return werror.WithStack(err)
		}
	} else {
		_, err := indexFile.Write([]byte{})
		if err != nil {
			return werror.WithStack(err)
		}
	}

	return nil
}

type TrashEntry struct {
	OrigParent   fileTree.FileId `bson:"originalParentId"`
	OrigFilename string          `bson:"originalFilename"`
	FileId       fileTree.FileId `bson:"fileId"`
}
