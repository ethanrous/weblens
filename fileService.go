package weblens

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/websocket"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileServiceImpl struct {
	mediaTree     fileTree.FileTree
	cachesTree    fileTree.FileTree
	userService   UserService
	accessService AccessService
	mediaService  MediaService

	trashCol *mongo.Collection
}

// FileInfo is a structure for safely sending file information to the client
type FileInfo struct {
	Id fileTree.FileId `json:"id"`
	/* If the content of the file can be displayed visually.
	Say the file is a jpg, mov, arw, etc. and not a zip,
	txt, doc, directory etc. */
	Displayable bool `json:"displayable"`

	IsDir          bool              `json:"isDir"`
	Modifiable     bool              `json:"modifiable"`
	Size           int64             `json:"size"`
	ModTime        int64             `json:"modTime"`
	Filename       string            `json:"filename"`
	ParentFolderId fileTree.FileId   `json:"parentFolderId"`
	MediaData      *Media            `json:"mediaData,omitempty"`
	Owner          Username          `json:"owner"`
	PathFromHome   string            `json:"pathFromHome"`
	ShareId        ShareId           `json:"shareId,omitempty"`
	Children       []fileTree.FileId `json:"children"`
	PastFile       bool              `json:"pastFile,omitempty"`
}

func NewFileService(
	mediaTree, cacheTree fileTree.FileTree, userService UserService,
	accessService AccessService, mediaService MediaService,
) *FileServiceImpl {
	return &FileServiceImpl{
		mediaTree: mediaTree, userService: userService, cachesTree: cacheTree, accessService: accessService,
		mediaService: mediaService,
	}
}

func (fs *FileServiceImpl) GetFileByIdAndRoot(id fileTree.FileId, rootAlias string) (*fileTree.WeblensFile, error) {
	switch rootAlias {
	case "MEDIA":
		return fs.mediaTree.Get(id), nil
	}

	return nil, errors.Errorf("Could not find file with id [%s] on tree [%s]", id, rootAlias)
}

func (fs *FileServiceImpl) IsFileInTrash(f *fileTree.WeblensFile) bool {
	return strings.Contains(f.GetAbsPath(), ".user_trash")
}

func (fs *FileServiceImpl) DeleteCacheFile(f *fileTree.WeblensFile) werror.WErr {

}

func (fs *FileServiceImpl) GetFileOwner(file *fileTree.WeblensFile) *User {
	if file.GetTree() != fs.mediaTree.(*fileTree.FileTreeImpl) {
		return fs.userService.Get("WEBLENS")
	}
	panic("get file owner not impl")
}

func (fs *FileServiceImpl) MoveFileToTrash(
	file *fileTree.WeblensFile, acc types.AccessMeta, event *fileTree.FileEvent, c ...websocket.BroadcasterAgent,
) error {
	if !file.Exists() {
		return errors.Errorf("Cannot with id [%s] (%s) does not exist", file.ID(), file.GetAbsPath())
	}
	if fs.IsFileInTrash(file) {
		return errors.Errorf("Cannot move file (%s) to trash because it is already in trash", file.GetAbsPath())
	}

	if !fs.accessService.CanUserAccessFile() {
		return errors.New("User cannot access file")
	}

	te := TrashEntry{
		OrigParent:   file.GetParent().ID(),
		OrigFilename: file.Filename(),
	}

	trash := fs.GetFileOwner(file).GetTrashFolder()
	newFilename := MakeUniqueChildName(trash, file.Filename())

	// TODO
	// buffered := internal.SliceConvert[websocket.BufferedBroadcasterAgent](
	// 	internal.Filter(
	// 		c, func(b websocket.BroadcasterAgent) bool { return b.IsBuffered() },
	// 	),
	// )

	err := file.GetTree().Move(file, trash, newFilename, false, event)
	if err != nil {
		return err
	}

	te.TrashFileId = file.ID()
	err = types.SERV.StoreService.NewTrashEntry(te)
	if err != nil {
		return err
	}

	if file.GetShare() != nil {
		err = file.GetShare().SetEnabled(false)
		if err != nil {
			wlog.ShowErr(err)
		}
	}

	return nil
}

func (fs *FileServiceImpl) ReturnFileFromTrash(
	trashFile *fileTree.WeblensFile, event *fileTree.FileEvent, c ...websocket.BroadcasterAgent,
) error {
	filter := bson.D{{"trashFileId", trashFile.ID()}}
	ret := fs.trashCol.FindOne(context.Background(), filter)
	if err := ret.Err(); err != nil {
		return errors.WithStack(err)
	}

	var trashEntry TrashEntry
	err := ret.Decode(&trashEntry)

	oldParent := trashFile.GetTree().Get(trashEntry.OrigParent)
	if oldParent == nil {
		oldParent = fs.GetFileOwner(trashFile).GetTrashFolder()
	}

	buffered := internal.SliceConvert[websocket.BufferedBroadcasterAgent](
		internal.Filter(
			c, func(b websocket.BroadcasterAgent) bool { return b.IsBuffered() },
		),
	)
	err = trashFile.GetTree().Move(trashFile, oldParent, trashEntry.OrigFilename, false, event, buffered...)

	if err != nil {
		return err
	}

	res, err := fs.trashCol.DeleteOne(context.Background(), bson.M{"trashFileId": trashEntry.TrashFileId})
	if res.DeletedCount == 0 {
		return errors.New("delete trash entry did not get expected delete count")
	}
	if err != nil {
		return err
	}

	if trashFile.GetShare() != nil {
		err = trashFile.GetShare().SetEnabled(false)
		if err != nil {
			wlog.ShowErr(err)
		}
	}

	return nil
}

// PermanentlyDeleteFile removes file being pointed to from the tree and deletes it from the real filesystem
func (fs *FileServiceImpl) PermanentlyDeleteFile(file *fileTree.WeblensFile, c websocket.BroadcasterAgent) (err error) {
	ownerTrash := fs.GetFileOwner(file).GetTrashFolder()
	err = file.GetTree().Del(file.ID())

	if err != nil {
		return
	}
	c.PushFileDelete(file)

	if err != nil {
		return
	}

	err = types.SERV.StoreService.DeleteTrashEntry(file.ID())
	if err != nil {
		return
	}

	err = types.SERV.FileTree.ResizeUp(ownerTrash, c)
	if err != nil {
		return
	}

	return
}

func RecursiveGetMedia(mediaRepo MediaService, folders ...*fileTree.WeblensFile) (ms []ContentId) {
	ms = []ContentId{}

	for _, f := range folders {
		if f == nil {
			wlog.Warning.Println("Skipping recursive media lookup for non-existent folder")
			continue
		}
		if !f.IsDir() {
			if f.IsDisplayable() {
				m := mediaRepo.Get(f.GetContentId())
				if m != nil {
					ms = append(ms, m.ID())
				}
			}
			continue
		}
		err := f.RecursiveMap(
			func(f *fileTree.WeblensFile) error {
				if !f.IsDir() && f.IsDisplayable() {
					m := mediaRepo.Get(f.GetContentId())
					if m != nil {
						ms = append(ms, m.ID())
					}
				}
				return nil
			},
		)
		if err != nil {
			wlog.ShowErr(err)
		}
	}

	return
}

// func GetFreeSpace(path string) uint64 {
// 	var stat unix.Statfs_t
//
// 	err := unix.Statfs(path, &stat)
// 	if err != nil {
// 		util.ErrTrace(err)
// 		return 0
// 	}
//
// 	spaceBytes := stat.Bavail * uint64(stat.Bsize)
// 	return spaceBytes
// }

func GenerateContentId(f *fileTree.WeblensFile) (ContentId, error) {
	if f.IsDir() {
		return "", nil
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
	}

	fileSize, err := f.Size()
	if err != nil {
		return "", err
	}

	// Read up to 1MB at a time
	bufSize := math.Min(float64(fileSize), 1000*1000)
	buf := make([]byte, int64(bufSize))
	newHash := sha256.New()
	fp, err := f.Readable()
	if err != nil {
		return "", err
	}

	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			wlog.ShowErr(err)
		}
	}(fp)

	_, err = io.CopyBuffer(newHash, fp, buf)
	if err != nil {
		return "", err
	}

	contentId := ContentId(base64.URLEncoding.EncodeToString(newHash.Sum(nil)))[:20]
	f.SetContentId(contentId)

	return contentId, nil
}

func BackupBaseFile(remoteId string, data []byte, ft types.FileTree) (baseF *fileTree.WeblensFile, err error) {
	// if thisServer.Role != BackupServerServer {
	// 	err = ErrNotBackup
	// 	return
	// }

	// baseContentId := util.GlobbyHash(8, data)
	//
	// mediaRoot := ft.Get("MEDIA")
	// // Get or create dir for remote core
	// remoteDir, err := ft.MkDir(mediaRoot, remoteId)
	// if err != nil && !errors.Is(err, types.ErrDirAlreadyExists) {
	// 	return
	// }
	//
	// dataDir, err := ft.MkDir(remoteDir, "data")
	// if err != nil && !errors.Is(err, types.ErrDirAlreadyExists) {
	// 	return
	// }
	//
	// baseF, err = ft.Touch(dataDir, baseContentId+".base", false, nil)
	// if errors.Is(err, types.ErrFileAlreadyExists) {
	// 	return
	// } else if err != nil {
	// 	return
	// }
	//
	// err = baseF.Write(data)
	return nil, werror.NotImplemented("BackupBaseFile")
}

// func CacheBaseMedia(mediaId ContentId, data [][]byte, ft types.FileTree) (
// 	newThumb, newFullres *fileTree.WeblensFile, err error,
// ) {
// 	cacheRoot := ft.Get("CACHE")
// 	newThumb, err = ft.Touch(cacheRoot, string(mediaId)+"-thumbnail.cache", false, nil)
// 	if err != nil && !strings.Contains(err.Error(), "file already exists") {
// 		return nil, nil, err
// 	} else if !strings.Contains(err.Error(), "file already exists") {
// 		err = newThumb.Write(data[0])
// 		if err != nil {
// 			return
// 		}
// 	}
//
// 	newFullres, err = ft.Touch(cacheRoot, string(mediaId)+"-fullres.cache", false, nil)
// 	if err != nil && !strings.Contains(err.Error(), "file already exists") {
// 		return nil, nil, err
// 	} else if !strings.Contains(err.Error(), "file already exists") {
// 		err = newFullres.Write(data[1])
// 		if err != nil {
// 			return
// 		}
// 	}
//
// 	return
// }

func MakeUniqueChildName(parent *fileTree.WeblensFile, childName string) string {
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

var IgnoreFilenames = []string{
	".DS_Store",
}

func (fs *FileServiceImpl) FormatFileInfo(f *fileTree.WeblensFile, accessor *User, share Share) (
	formattedInfo FileInfo,
	err error,
) {
	if f == nil {
		return formattedInfo, errors.WithStack(errors.New("cannot get file info of nil wf"))
	}

	if !fs.accessService.CanUserAccessFile(accessor, f) {
		err = types.ErrNoFileAccess
		return
	}

	m := fs.mediaService.Get(f.GetContentId())

	var size int64
	size, err = f.Size()
	if err != nil {
		wlog.ShowErr(err, fmt.Sprintf("Failed to get file size of [ %s (ID: %s) ]", f.GetAbsPath(), f.ID()))
		return
	}

	var shareId ShareId
	if f.GetShare() != nil {
		shareId = f.GetShare().GetShareId()
		wlog.Debug.Println("ShareId", shareId)
	}

	var parentId fileTree.FileId
	owner := fs.GetFileOwner(f)
	if owner != fs.userService.GetRootUser() && fs.accessService.CanUserAccessFile(accessor, f.GetParent()) {
		parentId = f.GetParent().ID()
	}

	tmpF := f
	var pathBits []string
	for tmpF != nil && fs.GetFileOwner(f) != fs.userService.GetRootUser() && fs.accessService.CanUserAccessFile(
		accessor, tmpF,
	) {
		if tmpF.GetParent() == f.GetTree().GetRoot() {
			pathBits = append(pathBits, "HOME")
			break
		} else if share != nil && tmpF.ID() == fileTree.FileId(share.GetItemId()) {
			pathBits = append(pathBits, "SHARE")
			break
		} else if fs.IsFileInTrash(tmpF) {
			pathBits = append(pathBits, "TRASH")
			break
		}
		pathBits = append(pathBits, tmpF.Filename())
		tmpF = tmpF.GetParent()
	}
	slices.Reverse(pathBits)
	pathString := strings.Join(pathBits, "/")

	formattedInfo = FileInfo{
		Id:          f.ID(),
		Displayable: f.IsDisplayable(),
		IsDir:       f.IsDir(),
		Modifiable: acc.GetTime().Unix() <= 0 &&
			!fs.IsFileInTrash(f) &&
			owner == accessor &&
			f.Owner() != fs.userService.GetRootUser() &&
			f != f.GetTree().Get("EXTERNAL") &&
			InstanceService.GetLocal().ServerRole() != BackupServer,
		Size:           size,
		ModTime:        f.ModTime().UnixMilli(),
		Filename:       f.Filename(),
		ParentFolderId: parentId,
		Owner:          owner.GetUsername(),
		PathFromHome:   pathString,
		MediaData:      m,
		ShareId:        shareId,
		Children: internal.Map(
			f.GetChildren(), func(wf *fileTree.WeblensFile) fileTree.FileId { return wf.ID() },
		),
		PastFile: acc.GetTime().Unix() > 0,
	}

	return formattedInfo, nil
}

func (fs *FileServiceImpl) GetChildrenInfo(f *fileTree.WeblensFile, acc types.AccessMeta) []FileInfo {
	childrenInfo := internal.FilterMap(
		f.GetChildren(), func(file *fileTree.WeblensFile) (FileInfo, bool) {
			info, err := file.FormatFileInfo(acc)
			if err != nil {
				wlog.ErrTrace(err)
				return info, false
			}
			return info, true
		},
	)

	if childrenInfo == nil {
		return []FileInfo{}
	}

	return childrenInfo
}

func (fs *FileServiceImpl) clearTempDir() (err error) {
	tmpRoot := ft.Get("TMP")
	err = os.MkdirAll(tmpRoot.GetAbsPath(), os.ModePerm)
	if err != nil {
		return
	}

	files, err := os.ReadDir(tmpRoot.GetAbsPath())
	if err != nil {
		return
	}

	for _, file := range files {
		err := os.RemoveAll(filepath.Join(internal.GetTmpDir(), file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServiceImpl) clearTakeoutDir() error {
	takeoutRoot := fs.cachesTree.Get("TAKEOUT")
	err := os.MkdirAll(takeoutRoot.GetAbsPath(), os.ModePerm)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(takeoutRoot.GetAbsPath())
	if err != nil {
		return err
	}

	for _, file := range files {
		err := os.Remove(filepath.Join(takeoutRoot.GetAbsPath(), file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

type TrashEntry struct {
	OrigParent   fileTree.FileId `bson:"originalParentId"`
	OrigFilename string          `bson:"originalFilename"`
	TrashFileId  fileTree.FileId `bson:"trashFileId"`
}
