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
	"time"

	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"golang.org/x/sys/unix"
)

var WeblensRootUser = types.User(
	&user.User{
		Username: "WEBLENS",
	},
)

var ExternalRootUser = types.User(
	&user.User{
		Username: "EXTERNAL",
	},
)

var RootDirIds = []types.FileId{"MEDIA", "TMP", "CACHE", "TAKEOUT", "EXTERNAL", "CONTENT_LINKS"}

func InitMediaRoot(tree types.FileTree) error {
	sw := util.NewStopwatch("Filesystem")

	// mediaRoot, err := tree.NewRoot("MEDIA", "media", util.GetMediaRootPath(), WeblensRootUser, nil)
	// if err != nil {
	// 	return err
	// }
	_, err := tree.NewRoot("TMP", "tmp", util.GetTmpDir(), WeblensRootUser, nil)
	if err != nil {
		return err
	}
	_, err = tree.NewRoot("TAKEOUT", "takeout", util.GetTakeoutDir(), WeblensRootUser, nil)
	if err != nil {
		return err
	}
	externalRoot, err := tree.NewRoot("EXTERNAL", "External", "", ExternalRootUser, nil)
	if err != nil {
		return err
	}
	cacheRoot, err := tree.NewRoot("CACHE", "Cache", util.GetCacheDir(), WeblensRootUser, nil)
	if err != nil {
		return err
	}
	contentRoot, err := tree.NewRoot(
		"CONTENT_LINKS", ".content", filepath.Join(tree.GetRoot().GetAbsPath(), ".content"),
		WeblensRootUser,
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

	lifetimes := tree.GetJournal().GetActiveLifetimes()
	if err != nil {
		return err
	}

	slices.SortFunc(
		lifetimes, func(a, b types.Lifetime) int {
			return strings.Compare(string(a.GetLatestFileId()), string(b.GetLatestFileId()))
		},
	)

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

	for _, u := range users {
		var homeDir types.WeblensFile
		if homeDir, err = tree.GetRoot().GetChild(u.GetUsername().String()); err != nil {
			homeDir = tree.NewFile(tree.GetRoot(), u.GetUsername().String(), true, u)
			err = tree.Add(homeDir)
			if err != nil {
				return err
			}
		}

		err = homeDir.RecursiveMap(
			func(wf types.WeblensFile) error {
				wf.SetOwner(u)
				return nil
			},
		)

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
		lifetimes, err = importFilesRecursive(homeDir, fileEvent, lifetimes)
		if err != nil {
			return err
		}

		var trashDir types.WeblensFile
		trashDir, err = homeDir.GetChild(".user_trash")
		if err != nil {
			return err
		}

		err = u.SetTrashFolder(trashDir)
		if err != nil {
			return err
		}
	}

	sw.Lap("Load users home directories")

	lifetimes, err = importFilesRecursive(tree.GetRoot(), fileEvent, lifetimes)
	if err != nil {
		return err
	}

	lifetimes, err = importFilesRecursive(contentRoot, fileEvent, lifetimes)
	if err != nil {
		return err
	}

	lifetimes, err = importFilesRecursive(cacheRoot, fileEvent, lifetimes)
	if err != nil {
		return err
	}

	sw.Lap("Load roots")

	err = tree.GetJournal().LogEvent(fileEvent)
	if err != nil {
		return err
	}

	sw.Lap("Write file event")

	for _, path := range util.GetExternalPaths() {
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

	// Compute size for the whole tree
	err = tree.GetRoot().LeafMap(
		func(wf types.WeblensFile) error {
			return wf.LoadStat()
		},
	)
	if err != nil {
		return err
	}

	err = cacheRoot.LeafMap(
		func(wf types.WeblensFile) error {
			return wf.LoadStat()
		},
	)
	if err != nil {
		return err
	}

	if externalRoot.GetParent() != tree.GetRoot() {
		err = externalRoot.LeafMap(
			func(wf types.WeblensFile) error {
				return wf.LoadStat()
			},
		)
		if err != nil {
			return err
		}
	}

	sw.Lap("Compute Sizes")

	// for _, u := range users {
	// 	homeId := tree.GenerateFileId(filepath.Join(mediaRoot.GetAbsPath(), string(u.GetUsername())) + "/")
	// 	err := u.SetHomeFolder(tree.Get(homeId))
	// 	if err != nil {
	// 		util.Error.Printf("Could not set home folder for %s to %s", u.GetUsername(), homeId)
	// 		return err
	// 	}
	//
	// 	trash, err := u.GetHomeFolder().GetChild(".user_trash")
	// 	if err != nil || trash == nil {
	// 		return err
	// 	}
	// 	err = u.SetTrashFolder(trash)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// sw.Lap("Generate missing user home directories")
	sw.Stop()
	// sw.PrintResults(false)
	// util.Debug.Println(
	// 	"Reading", timeReading, " - Searching", searching, " - Waiting", waiting, " - Creating", creating,
	// 	" - Inserting", inserting,
	// )
	return nil
}

var timeReading time.Duration
var searching time.Duration
var waiting time.Duration
var creating time.Duration
var inserting time.Duration

func importFilesRecursive(f types.WeblensFile, fileEvent types.FileEvent, lifetimes []types.Lifetime) (
	[]types.Lifetime, error,
) {
	var toLoad = []types.WeblensFile{f}
	pool := types.SERV.WorkerPool.NewTaskPool(false, nil)
	for len(toLoad) != 0 {
		var fileToLoad types.WeblensFile

		// Pop from slice of files to load
		fileToLoad, toLoad = toLoad[0], toLoad[1:]
		start := time.Now()
		if slices.Contains(IgnoreFilenames, fileToLoad.Filename()) || (fileToLoad.Filename() == "."+
			"content" && fileToLoad.ID() != "CONTENT_LINKS") {
			searching += time.Since(start)
			continue
		}
		searching += time.Since(start)

		if fileToLoad.Owner() != WeblensRootUser {
			start := time.Now()
			index, e := slices.BinarySearchFunc(
				lifetimes, fileToLoad.ID(), func(lt types.Lifetime, id types.FileId) int {
					return strings.Compare(string(lt.GetLatestFileId()), string(id))
				},
			)
			searching += time.Since(start)
			if !e {
				start := time.Now()
				if fileToLoad.GetContentId() == "" && !fileToLoad.IsDir() {
					pool.HashFile(
						fileToLoad,
						types.SERV.Caster,
					).SetPostAction(
						func(result types.TaskResult) {
							if result["contentId"] != nil {
								fileToLoad.SetContentId(result["contentId"].(types.ContentId))
							}
							fileEvent.NewCreateAction(fileToLoad)
						},
					)
				} else if fileToLoad.IsDir() {
					fileEvent.NewCreateAction(fileToLoad)
				}
				creating += time.Since(start)
			} else {
				start := time.Now()
				var life types.Lifetime
				lifetimes, life = util.Yoink(lifetimes, index)
				fileToLoad.SetContentId(life.GetContentId())
				creating += time.Since(start)
			}
		}

		if !slices.Contains(RootDirIds, fileToLoad.ID()) {
			if types.SERV.FileTree.Get(fileToLoad.ID()) == nil {
				err := types.SERV.FileTree.Add(fileToLoad)
				if err != nil {
					return nil, err
				}
			} else {
				util.Debug.Println("Skipping insert of a file already present in the tree:", fileToLoad.ID())
			}
		}

		if fileToLoad.IsDir() {
			start := time.Now()
			children, err := fileToLoad.ReadDir()
			timeReading += time.Since(start)
			if err != nil {
				return nil, err
			}
			toLoad = append(toLoad, children...)
		}
	}

	pool.SignalAllQueued()
	start := time.Now()
	pool.Wait(false)
	waiting += time.Since(start)

	return lifetimes, nil
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
	if f.Owner() == ExternalRootUser {
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
	if errors.Is(err, types.ErrFileAlreadyExists) {
		err = nil
		exists = true
	}

	return
}

func MoveFileToTrash(file types.WeblensFile, acc types.AccessMeta, c ...types.BroadcasterAgent) error {
	if !file.Exists() {
		return types.ErrNoFile
	}

	// if len(c) == 0 {
	// 	c = append(c, globalCaster)
	// }

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
	err := file.GetTree().Move(file, trash, newFilename, false, buffered...)
	if err != nil {
		return err
	}

	err = types.SERV.Database.NewTrashEntry(te)
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

func ReturnFileFromTrash(trashFile types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	te, err := types.SERV.Database.GetTrashEntry(trashFile.ID())
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
	err = trashFile.GetTree().Move(trashFile, oldParent, te.OrigFilename, false, buffered...)

	if err != nil {
		return
	}

	err = types.SERV.Database.DeleteTrashEntry(trashFile.ID())
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

	err = types.SERV.Database.DeleteTrashEntry(file.ID())
	if err != nil {
		return err
	}

	err = types.SERV.FileTree.ResizeUp(ownerTrash, c)
	if err != nil {
		return err
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

// func GetCacheDir() types.WeblensFile {
// 	return &cacheRoot
// }
//
// func GetTmpDir() types.WeblensFile {
// 	return &tmpRoot
// }
//
// func GetMediaDir() types.WeblensFile {
// 	return &mediaRoot
// }
//
// func GetExternalDir() types.WeblensFile {
// 	return &externalRoot
// }

func GetFreeSpace(path string) uint64 {
	var stat unix.Statfs_t

	err := unix.Statfs(path, &stat)
	if err != nil {
		util.ErrTrace(err)
		return 0
	}

	spaceBytes := stat.Bavail * uint64(stat.Bsize)
	return spaceBytes
}

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

	baseContentId := util.GlobbyHash(8, data)

	mediaRoot := ft.Get("MEDIA")
	// Get or create dir for remote core
	remoteDir, err := ft.MkDir(mediaRoot, remoteId)
	if err != nil && !errors.Is(err, types.ErrDirAlreadyExists) {
		return
	}

	dataDir, err := ft.MkDir(remoteDir, "data")
	if err != nil && !errors.Is(err, types.ErrDirAlreadyExists) {
		return
	}

	baseF, err = ft.Touch(dataDir, baseContentId+".base", false, nil)
	if errors.Is(err, types.ErrFileAlreadyExists) {
		return
	} else if err != nil {
		return
	}

	err = baseF.Write(data)
	return
}

func CacheBaseMedia(mediaId types.ContentId, data [][]byte, ft types.FileTree) (
	newThumb, newFullres types.WeblensFile, err error,
) {
	cacheRoot := ft.Get("CACHE")
	newThumb, err = ft.Touch(cacheRoot, string(mediaId)+"-thumbnail.cache", false, nil)
	if err != nil && !errors.Is(err, types.ErrFileAlreadyExists) {
		return nil, nil, err
	} else if !errors.Is(err, types.ErrFileAlreadyExists) {
		err = newThumb.Write(data[0])
		if err != nil {
			return
		}
	}

	newFullres, err = ft.Touch(cacheRoot, string(mediaId)+"-fullres.cache", false, nil)
	if err != nil && !errors.Is(err, types.ErrFileAlreadyExists) {
		return nil, nil, err
	} else if !errors.Is(err, types.ErrFileAlreadyExists) {
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
	for !errors.Is(e, types.ErrNoFile) {
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
