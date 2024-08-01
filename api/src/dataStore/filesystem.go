package dataStore

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

var ExternalRootUser = types.User(
	&user.User{
		Username: "EXTERNAL",
	},
)

var RootDirIds = []types.FileId{"MEDIA", "TMP", "CACHE", "TAKEOUT", "EXTERNAL", "CONTENT_LINKS"}

func InitMediaRoot(tree types.FileTree, hashCaster types.BroadcasterAgent) error {
	types.SERV.InstanceService.AddLoading("filesystem")
	sw := util.NewStopwatch("Filesystem")

	// mediaRoot, err := tree.NewRoot("MEDIA", "media", util.GetMediaRootPath(), types.SERV.UserService.Get("WEBLENS"), nil)
	// if err != nil {
	// 	return err
	// }
	_, err := tree.NewRoot("TMP", "tmp", util.GetTmpDir(), types.SERV.UserService.Get("WEBLENS"), nil)
	if err != nil {
		return err
	}
	_, err = tree.NewRoot("TAKEOUT", "takeout", util.GetTakeoutDir(), types.SERV.UserService.Get("WEBLENS"), nil)
	if err != nil {
		return err
	}
	externalRoot, err := tree.NewRoot("EXTERNAL", "External", "", ExternalRootUser, nil)
	if err != nil {
		return err
	}
	cacheRoot, err := tree.NewRoot("CACHE", "Cache", util.GetCacheDir(), types.SERV.UserService.Get("WEBLENS"), nil)
	if err != nil {
		return err
	}
	contentRoot, err := tree.NewRoot(
		"CONTENT_LINKS", ".content", filepath.Join(tree.GetRoot().GetAbsPath(), ".content"),
		types.SERV.UserService.Get("WEBLENS"),
		tree.GetRoot(),
	)
	if err != nil {
		return err
	}

	err = tree.SetDelDirectory(tree.Get("CONTENT_LINKS"))
	if err != nil {
		return err
	}

	sw.Lap("Set roots")

	sw.Lap("Get + sort lifetimes")

	if !tree.GetRoot().Exists() {
		err = tree.GetRoot().CreateSelf()
		if err != nil {
			return err
		}
	}

	if !contentRoot.Exists() {
		err = contentRoot.CreateSelf()
		if err != nil {
			return err
		}
	}

	if !cacheRoot.Exists() {
		err = contentRoot.CreateSelf()
		if err != nil {
			return err
		}
	}

	sw.Lap("Check roots exist")

	users, err := types.SERV.UserService.GetAll()
	if err != nil {
		return err
	}

	fileEvent := history.NewFileEvent()
	hashTaskPool := types.SERV.WorkerPool.NewTaskPool(false, nil)

	for _, u := range users {
		var homeDir types.WeblensFile
		if homeDir, err = tree.GetRoot().GetChild(u.GetUsername().String()); err != nil {
			homeDir = tree.NewFile(tree.GetRoot(), u.GetUsername().String(), true, u)
			err = tree.Add(homeDir)
			if err != nil {
				return err
			}
		}

		err = u.SetHomeFolder(homeDir)
		if err != nil {
			return err
		}
		if !homeDir.Exists() {
			err = homeDir.CreateSelf()
			if err != nil {
				return err
			}
		}
		err = importFilesRecursive(homeDir, fileEvent, hashTaskPool, hashCaster)
		if err != nil {
			return err
		}

		var trashDir types.WeblensFile
		trashDir, err = homeDir.GetChild(".user_trash")
		if err != nil {
			trashDir = tree.NewFile(homeDir, ".user_trash", true, u)
			if tree.Get(trashDir.ID()) == nil {
				if !trashDir.Exists() {
					err = trashDir.CreateSelf()
					if err != nil {
						return err
					}
				}
				err = tree.Add(trashDir)
				if err != nil {
					return err
				}
			}
		}

		err = u.SetTrashFolder(trashDir)
		if err != nil {
			return err
		}
	}

	sw.Lap("Load users home directories")

	err = importFilesRecursive(contentRoot, fileEvent, hashTaskPool, hashCaster)
	if err != nil {
		return err
	}

	err = importFilesRecursive(cacheRoot, fileEvent, hashTaskPool, hashCaster)
	if err != nil {
		return err
	}

	hashTaskPool.SignalAllQueued()
	hashTaskPool.AddCleanup(
		func() {
			err = tree.GetJournal().LogEvent(fileEvent)
			if err != nil {
				util.Error.Println(err)
			}
			types.SERV.InstanceService.RemoveLoading("filesystem")
		},
	)

	sw.Lap("Load roots")

	for _, path := range util.GetExternalPaths() {
		continue
		if path == "" {
			continue
		}
		s, err := os.Stat(path)
		if err != nil {
			panic(fmt.Sprintf("Could not find external path: %s", path))
		}
		extF := tree.NewFile(externalRoot, filepath.Base(path), s.IsDir(), nil)
		err = tree.Add(extF)
		if err != nil {
			return err
		}
	}

	sw.Lap("Load external files")

	// Compute size for the whole tree, and ensure children are loaded while we're at it.
	err = tree.ResizeDown(tree.GetRoot())
	if err != nil {
		return err
	}

	err = cacheRoot.LeafMap(
		func(wf types.WeblensFile) error {
			_, err = wf.Size()
			return err
		},
	)
	if err != nil {
		return err
	}

	if externalRoot.GetParent() != tree.GetRoot() {
		err = externalRoot.LeafMap(
			func(wf types.WeblensFile) error {
				_, err = wf.Size()
				return err
			},
		)
		if err != nil {
			return err
		}
	}

	sw.Lap("Compute Sizes")

	for _, sh := range types.SERV.ShareService.GetAllShares() {
		switch sh.GetShareType() {
		case types.FileShare:
			sharedFile := types.SERV.FileTree.Get(types.FileId(sh.GetItemId()))
			if sharedFile != nil {
				err := sharedFile.SetShare(sh)
				if err != nil {
					return err
				}
			} else {
				util.Warning.Println("Ignoring possibly no longer existing file in share init")
			}
		}
	}

	sw.Lap("Link file shares")

	// hashTaskPool.Wait(false)
	// sw.Lap("Wait for hash pool")

	sw.Stop()
	sw.PrintResults(false)
	return nil
}

func importFilesRecursive(
	f types.WeblensFile, fileEvent types.FileEvent,
	hashTaskPool types.TaskPool, hashCaster types.BroadcasterAgent,
) error {
	if types.SERV.InstanceService.GetLocal().ServerRole() == types.Backup {
		return nil
	}
	var toLoad = []types.WeblensFile{f}
	for len(toLoad) != 0 {
		var fileToLoad types.WeblensFile

		// Pop from slice of files to load
		fileToLoad, toLoad = toLoad[0], toLoad[1:]
		if slices.Contains(IgnoreFilenames, fileToLoad.Filename()) || (fileToLoad.Filename() == "."+
			"content" && fileToLoad.ID() != "CONTENT_LINKS") {
			continue
		}

		if fileToLoad.Owner() != types.SERV.UserService.Get("WEBLENS") {
			lt := fileToLoad.GetTree().GetJournal().GetLifetimeByFileId(fileToLoad.ID())
			if lt == nil {
				fileSize, err := fileToLoad.Size()
				if err != nil {
					return types.WeblensErrorFromError(err)
				}

				if fileToLoad.GetContentId() == "" && !fileToLoad.IsDir() && fileSize != 0 {
					hashTaskPool.HashFile(
						fileToLoad,
						hashCaster,
					).SetPostAction(
						func(result types.TaskResult) {
							if result["contentId"] != nil {
								fileToLoad.SetContentId(result["contentId"].(types.ContentId))
								fileEvent.NewCreateAction(fileToLoad)
							} else {
								util.Error.Println("Failed to generate contentId for", fileToLoad.Filename())
							}

						},
					)
				} else if fileToLoad.IsDir() || fileSize == 0 {
					fileEvent.NewCreateAction(fileToLoad)
				}

			} else {
				fileToLoad.SetContentId(lt.GetContentId())
			}
		}

		if !slices.Contains(RootDirIds, fileToLoad.ID()) {
			if types.SERV.FileTree.Get(fileToLoad.ID()) == nil {
				err := types.SERV.FileTree.Add(fileToLoad)
				if err != nil {
					return err
				}
			} else {
				// util.Debug.Println("Skipping insert of a file already present in the tree:", fileToLoad.ID())
				// continue
			}
		}

		if fileToLoad.IsDir() {
			children, err := fileToLoad.ReadDir()
			if err != nil {
				return err
			}
			toLoad = append(toLoad, children...)
		}
	}

	return nil
}

func ClearTempDir(ft types.FileTree) (err error) {
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
		err := os.RemoveAll(filepath.Join(util.GetTmpDir(), file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func ClearTakeoutDir(ft types.FileTree) error {
	takeoutRoot := ft.Get("TAKEOUT")
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

/*------------------*/

func IsFileInTrash(f types.WeblensFile) bool {
	if f.Owner().IsSystemUser() {
		return false
	}
	trashPath := filepath.Join(
		types.SERV.FileTree.GetRoot().GetAbsPath(), string(f.Owner().GetUsername()), ".user_trash",
	)
	return strings.HasPrefix(f.GetAbsPath(), trashPath)
}

func NewTakeoutZip(zipName string, creatorName types.Username) (newZip types.WeblensFile, exists bool, err error) {
	u := types.SERV.UserService.Get(creatorName)
	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	takeoutRoot := types.SERV.FileTree.Get("TAKEOUT")

	newZip, err = types.SERV.FileTree.Touch(takeoutRoot, zipName, false, u)
	if errors.Is(err, types.ErrFileAlreadyExists()) {
		err = nil
		exists = true
	}

	return
}

func MoveFileToTrash(
	file types.WeblensFile, acc types.AccessMeta, event types.FileEvent, c ...types.BroadcasterAgent,
) error {
	if !file.Exists() {
		return types.ErrNoFile(file.ID())
	}

	if !acc.CanAccessFile(file) {
		return ErrNoFileAccess
	}

	te := types.TrashEntry{
		OrigParent:   file.GetParent().ID(),
		OrigFilename: file.Filename(),
	}

	trash := file.Owner().GetTrashFolder()
	newFilename := MakeUniqueChildName(trash, file.Filename())

	buffered := util.SliceConvert[types.BufferedBroadcasterAgent](
		util.Filter(
			c, func(b types.BroadcasterAgent) bool { return b.IsBuffered() },
		),
	)
	err := file.GetTree().Move(file, trash, newFilename, false, event, buffered...)
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
			util.ShowErr(err)
		}
	}

	return nil
}

func ReturnFileFromTrash(trashFile types.WeblensFile, event types.FileEvent, c ...types.BroadcasterAgent) (err error) {
	te, err := types.SERV.StoreService.GetTrashEntry(trashFile.ID())
	if err != nil {
		return
	}

	oldParent := trashFile.GetTree().Get(te.OrigParent)
	trashFile.Owner()
	if oldParent == nil {
		oldParent = trashFile.Owner().GetTrashFolder()
	}

	buffered := util.SliceConvert[types.BufferedBroadcasterAgent](
		util.Filter(
			c, func(b types.BroadcasterAgent) bool { return b.IsBuffered() },
		),
	)
	err = trashFile.GetTree().Move(trashFile, oldParent, te.OrigFilename, false, event, buffered...)

	if err != nil {
		return
	}

	err = types.SERV.StoreService.DeleteTrashEntry(te.TrashFileId)
	if err != nil {
		return
	}

	if trashFile.GetShare() != nil {
		err = trashFile.GetShare().SetEnabled(false)
		if err != nil {
			util.ShowErr(err)
		}
	}

	return
}

// PermanentlyDeleteFile removes file being pointed to from the tree and deletes it from the real filesystem
func PermanentlyDeleteFile(file types.WeblensFile, c types.BroadcasterAgent) (err error) {
	ownerTrash := file.Owner().GetTrashFolder()
	err = types.SERV.FileTree.Del(file.ID())

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

func RecursiveGetMedia(mediaRepo types.MediaRepo, folders ...types.WeblensFile) (ms []types.ContentId) {
	ms = []types.ContentId{}

	for _, f := range folders {
		if f == nil {
			util.Warning.Println("Skipping recursive media lookup for non-existent folder")
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
			func(f types.WeblensFile) error {
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
			util.ShowErr(err)
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

func GenerateContentId(f types.WeblensFile) (types.ContentId, error) {
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
	fp, err := f.Read()
	if err != nil {
		return "", err
	}

	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			util.ShowErr(err)
		}
	}(fp)

	_, err = io.CopyBuffer(newHash, fp, buf)
	if err != nil {
		return "", err
	}

	contentId := types.ContentId(base64.URLEncoding.EncodeToString(newHash.Sum(nil)))[:20]
	f.SetContentId(contentId)

	return contentId, nil
}

func BackupBaseFile(remoteId string, data []byte, ft types.FileTree) (baseF types.WeblensFile, err error) {
	// if thisServer.Role != types.Backup {
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
	return nil, types.ErrNotImplemented("BackupBaseFile")
}

func CacheBaseMedia(mediaId types.ContentId, data [][]byte, ft types.FileTree) (
	newThumb, newFullres types.WeblensFile, err error,
) {
	cacheRoot := ft.Get("CACHE")
	newThumb, err = ft.Touch(cacheRoot, string(mediaId)+"-thumbnail.cache", false, nil)
	if err != nil && !errors.Is(err, types.ErrFileAlreadyExists()) {
		return nil, nil, err
	} else if !errors.Is(err, types.ErrFileAlreadyExists()) {
		err = newThumb.Write(data[0])
		if err != nil {
			return
		}
	}

	newFullres, err = ft.Touch(cacheRoot, string(mediaId)+"-fullres.cache", false, nil)
	if err != nil && !errors.Is(err, types.ErrFileAlreadyExists()) {
		return nil, nil, err
	} else if !errors.Is(err, types.ErrFileAlreadyExists()) {
		err = newFullres.Write(data[1])
		if err != nil {
			return
		}
	}

	return
}

func MakeUniqueChildName(parent types.WeblensFile, childName string) string {
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
