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

	"github.com/ethrousseau/weblens/api/dataStore/filetree"
	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"golang.org/x/sys/unix"
)

var WeblensRootUser = &user{
	Username: "WEBLENS",
}

var ExternalRootUser = &user{
	Username: "EXTERNAL",
}

var RootDirIds = []types.FileId{"MEDIA", "TMP", "CACHE", "TAKEOUT", "EXTERNAL"}

// Directory to hold the users home directories, and a few extras.
// This is where most of the data on the filesystem will be stored
// var mediaRoot = filetree.weblensFile{
// 	id:       "MEDIA",
// 	filename: "media",
// 	owner:    WeblensRootUser,
//
// 	isDir:        boolPointer(true),
// 	absolutePath: util.GetMediaRootPath(),
//
// 	childLock: &sync.Mutex{},
// 	children:  []*filetree.weblensFile{},
// }
//
// // Under the mediaRoot, contentRoot is where backup files are stored
// // when they are no longer in a user directory, or links to the real files
// // when they are in use
// var contentRoot = filetree.weblensFile{
// 	id:       "CONTENT_LINKS",
// 	filename: ".content",
// 	owner:    WeblensRootUser,
//
// 	isDir:        boolPointer(true),
// 	absolutePath: mediaRoot.absolutePath + "/.content/",
//
// 	childLock: &sync.Mutex{},
// 	children:  []*filetree.weblensFile{},
// }
//
// var tmpRoot = filetree.weblensFile{
// 	id:       "TMP",
// 	filename: "tmp",
// 	owner:    WeblensRootUser,
//
// 	isDir:        boolPointer(true),
// 	absolutePath: util.GetTmpDir(),
//
// 	childLock: &sync.Mutex{},
// 	children:  []*filetree.weblensFile{},
// }
//
// var cacheRoot = filetree.weblensFile{
// 	id:       "CACHE",
// 	filename: "cache",
// 	owner:    WeblensRootUser,
//
// 	isDir:        boolPointer(true),
// 	absolutePath: util.GetCacheDir(),
//
// 	childLock: &sync.Mutex{},
// 	children:  []*filetree.weblensFile{},
// }
//
// var takeoutRoot = filetree.weblensFile{
// 	id:       "TAKEOUT",
// 	filename: "takeout",
// 	owner:    WeblensRootUser,
//
// 	isDir:        boolPointer(true),
// 	absolutePath: util.GetTakeoutDir(),
//
// 	childLock: &sync.Mutex{},
// 	children:  []*filetree.weblensFile{},
// }
//
// var externalRoot = filetree.weblensFile{
// 	id:       "EXTERNAL",
// 	filename: "External",
// 	owner:    ExternalRootUser,
//
// 	isDir: boolPointer(true),
//
// 	// This is a fake directory, it houses the mounted external paths, but does not exist itself
// 	absolutePath: "",
// 	readOnly:     true,
//
// 	childLock: &sync.Mutex{},
// 	children:  []*filetree.weblensFile{},
// }

type dataStoreControllers struct {
	dbServer types.DatabaseService
	tasker   types.TaskPool
	caster   types.BroadcasterAgent
}

var dsc dataStoreControllers

func FsInit(tree types.FileTree, dbService types.DatabaseService, tasker types.TaskPool, caster types.BroadcasterAgent) {
	dsc = dataStoreControllers{
		dbServer: dbService,
		tasker:   tasker,
		caster:   caster,
	}

	mediaRoot, err := tree.NewRoot("MEDIA", "media", util.GetMediaRootPath(), WeblensRootUser, nil)
	if err != nil {
		panic(err)
	}
	_, err = tree.NewRoot("TMP", "tmp", util.GetTmpDir(), WeblensRootUser, nil)
	if err != nil {
		panic(err)
	}
	_, err = tree.NewRoot("TAKEOUT", "takeout", util.GetTakeoutDir(), WeblensRootUser, nil)
	if err != nil {
		panic(err)
	}
	externalRoot, err := tree.NewRoot("EXTERNAL", "External", "", ExternalRootUser, nil)
	if err != nil {
		panic(err)
	}
	cacheRoot, err := tree.NewRoot("CACHE", "Cache", util.GetCacheDir(), WeblensRootUser, nil)
	if err != nil {
		panic(err)
	}
	contentRoot, err := tree.NewRoot("CONTENT_LINKS", ".content", filepath.Join(mediaRoot.GetAbsPath(), ".content"),
		WeblensRootUser,
		mediaRoot)
	if err != nil {
		panic(err)
	}

	// if GetServerInfo().ServerRole() == types.Core {
	// 	externalRoot.parent = &externalRoot
	// }

	existingEvents, err := history.GetAllFileEvents()
	if err != nil {
		panic(err)
	}

	var existingActions []types.FileAction
	util.Each[types.FileEvent](existingEvents, func(fe types.FileEvent) {
		existingActions = append(existingActions, fe.GetActions()...)
	})

	slices.SortFunc(existingActions, func(a, b types.FileAction) int {
		return strings.Compare(string(a.GetDestinationId()), string(b.GetDestinationId()))
	})

	if !mediaRoot.Exists() {
		err = mediaRoot.CreateSelf()
		if err != nil {
			panic(err)
		}
	}

	if !contentRoot.Exists() {
		err = contentRoot.CreateSelf()
		if err != nil {
			panic(err)
		}
	}

	if !cacheRoot.Exists() {
		err = contentRoot.CreateSelf()
		if err != nil {
			panic(err)
		}
	}

	users := getUsers()
	fileEvent := history.NewFileEvent()
	for _, user := range users {
		homeDir := tree.NewFile(mediaRoot, user.GetUsername().String(), true)
		homeDir.SetOwner(user)

		err = homeDir.CreateSelf()
		if err != nil {
			var weblensError types.WeblensError
			switch {
			case errors.As(err, &weblensError):
			default:
				{
					panic(err)
				}
			}
		}

		existingActions = importFilesRecursive(homeDir, fileEvent, existingActions)

		_, err = tree.MkDir(homeDir, ".user_trash")
		if err != nil {
			var weblensError types.WeblensError
			switch {
			case errors.As(err, &weblensError):
			default:
				{
					panic(err)
				}
			}
		}
	}

	existingActions = importFilesRecursive(mediaRoot, fileEvent, existingActions)
	existingActions = importFilesRecursive(cacheRoot, fileEvent, existingActions)
	existingActions = importFilesRecursive(contentRoot, fileEvent, existingActions)

	// err = cacheRoot.ReadDir()
	// if err != nil {
	// 	panic(err)
	// }
	//
	// err = mediaRoot.ReadDir()
	// if err != nil {
	// 	panic(err)
	// }
	//
	// err = contentRoot.ReadDir()
	// if err != nil {
	// 	panic(err)
	// }

	err = tree.GetJournal().LogEvent(fileEvent)
	if err != nil {
		panic(err)
	}

	for _, path := range util.GetExternalPaths() {
		if path == "" {
			continue
		}
		s, err := os.Stat(path)
		if err != nil {
			panic(fmt.Sprintf("Could not find external path: %s", path))
		}
		extF := tree.NewFile(externalRoot, filepath.Base(path), s.IsDir())
		err = tree.Add(extF, externalRoot, nil)
		if err != nil {
			panic(err)
		}
	}

	// Compute size for the whole tree
	err = mediaRoot.LeafMap(func(wf types.WeblensFile) error {
		return wf.LoadStat()
	})
	if err != nil {
		panic(err)
	}

	err = cacheRoot.LeafMap(func(wf types.WeblensFile) error {
		return wf.LoadStat()
	})
	if err != nil {
		panic(err)
	}

	if externalRoot.GetParent() != mediaRoot {
		err = externalRoot.LeafMap(func(wf types.WeblensFile) error {
			return wf.LoadStat()
		})
		if err != nil {
			panic(err)
		}
	}

	// clean up journal init
	existingEvents = nil

	loadUsersStaticFolders(tree)
}

func importFilesRecursive(f types.WeblensFile, fileEvent types.FileEvent, existingActions []types.FileAction) []types.
	FileAction {
	var toLoad = []types.WeblensFile{f}
	for len(toLoad) != 0 {
		var fileToLoad types.WeblensFile

		// Pop from slice of files to load
		fileToLoad, toLoad = toLoad[0], toLoad[1:]

		index, e := slices.BinarySearchFunc(existingActions, fileToLoad.ID(), func(action types.FileAction,
			id types.FileId) int {
			return strings.Compare(string(action.GetDestinationId()), string(id))
		})
		if !e {
			if fileToLoad.GetContentId() == "" && !fileToLoad.IsDir() && fileToLoad.Owner() != WeblensRootUser {
				util.Debug.Println("Starting task...")
				contentId := dsc.tasker.HashFile(fileToLoad, dsc.caster).Wait().GetResult("contentId")
				if contentId != nil {
					fileToLoad.SetContentId(contentId.(types.ContentId))
				}
				fileEvent.AddAction(history.NewCreateEntry(fileToLoad.GetAbsPath(), fileToLoad.GetContentId()))
			}
		} else {
			var eAction types.FileAction
			existingActions, eAction = util.Yoink(existingActions, index)
			fileToLoad.SetContentId(eAction.GetContentId())
		}

		if !slices.Contains(RootDirIds, fileToLoad.ID()) {
			if f.GetTree().Get(fileToLoad.ID()) != nil {
				continue
				// panic(types.NewWeblensError(fmt.Sprintf("key collision on attempt to insert to filesystem tree: [%s] %s", fileToLoad.ID(), fileToLoad.GetAbsPath())))
			}
			err := f.GetTree().Add(fileToLoad, fileToLoad.GetParent())
			if err != nil {
				panic(err)
			}
		}

		if fileToLoad.IsDir() {
			err := fileToLoad.ReadDir()
			if err != nil {
				panic(err)
			}
		}

		toLoad = append(toLoad, fileToLoad.GetChildren()...)
	}

	return existingActions
}

var dirIgnore = map[string]bool{
	".DS_Store": true,
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
		err := os.Remove(filepath.Join(util.GetTmpDir(), file.Name()))
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

// /////////////////////////////

var dbServer = NewDB()

func boolPointer(b bool) *bool {
	return &b
}

func IsFileInTrash(f types.WeblensFile) bool {
	if f.Owner() == ExternalRootUser {
		return false
	}
	trashPath := f.Owner().GetTrashFolder().GetAbsPath()
	return strings.HasPrefix(f.GetAbsPath(), trashPath)
}

func NewTakeoutZip(zipName string, creatorName types.Username, ft types.FileTree) (newZip types.WeblensFile, exists bool, err error) {
	user := GetUser(creatorName)
	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	takeoutRoot := ft.Get("TAKEOUT")

	newZip, err = ft.Touch(takeoutRoot, zipName, false, user)
	if errors.Is(err, filetree.ErrFileAlreadyExists) {
		err = nil
		exists = true
	}

	return
}

func MoveFileToTrash(file types.WeblensFile, acc types.AccessMeta, c ...types.BroadcasterAgent) error {
	if !file.Exists() {
		return filetree.ErrNoFile
	}

	// if len(c) == 0 {
	// 	c = append(c, globalCaster)
	// }

	if !CanAccessFile(file, acc) {
		return ErrNoFileAccess
	}

	te := trashEntry{
		OrigParent:   file.GetParent().ID(),
		OrigFilename: file.Filename(),
	}

	trash := file.Owner().GetTrashFolder()
	newFilename := MakeUniqueChildName(trash, file.Filename())

	buffered := util.SliceConvert[types.BufferedBroadcasterAgent](util.Filter(c, func(b types.BroadcasterAgent) bool { return b.IsBuffered() }))
	err := file.GetTree().Move(file, trash, newFilename, false, buffered...)
	if err != nil {
		return err
	}

	te.TrashFileId = file.ID()
	err = dbServer.newTrashEntry(te)
	if err != nil {
		return err
	}

	for _, s := range file.GetShares() {
		s.SetEnabled(false)
		err = UpdateFileShare(s, file.GetTree())
		if err != nil {
			return err
		}
	}

	return nil
}

func ReturnFileFromTrash(trashFile types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	te, err := dbServer.getTrashEntry(trashFile.ID())
	if err != nil {
		return
	}

	oldParent := trashFile.GetTree().Get(te.OrigParent)
	trashFile.Owner()
	if oldParent == nil {
		oldParent = trashFile.Owner().GetTrashFolder()
	}

	buffered := util.SliceConvert[types.BufferedBroadcasterAgent](util.Filter(c, func(b types.BroadcasterAgent) bool { return b.IsBuffered() }))
	err = trashFile.GetTree().Move(trashFile, oldParent, te.OrigFilename, false, buffered...)

	if err != nil {
		return
	}

	err = dbServer.removeTrashEntry(te.TrashFileId)
	if err != nil {
		return
	}

	for _, s := range trashFile.GetShares() {
		s.SetEnabled(true)
		err = UpdateFileShare(s, trashFile.GetTree())
		if err != nil {
			return err
		}
	}

	return
}

// PermanentlyDeleteFile removes file being pointed to from the tree and deletes it from the real filesystem
func PermanentlyDeleteFile(file types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	// if len(c) == 0 {
	// 	c = append(c, globalCaster)
	// }

	err = file.GetTree().Del(file, c...)

	if err != nil {
		return
	}

	err = dbServer.removeTrashEntry(file.ID())
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
			if f.IsDisplayable(mediaRepo) {
				m := mediaRepo.Get(f.GetContentId())
				if m != nil {
					ms = append(ms, m.ID())
				}
			}
			continue
		}
		err := f.RecursiveMap(func(f types.WeblensFile) error {
			if !f.IsDir() && f.IsDisplayable(mediaRepo) {
				m := mediaRepo.Get(f.GetContentId())

				if m != nil {
					ms = append(ms, m.ID())
				}
			}
			return nil
		})
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
	if thisServer.Role != types.Backup {
		err = ErrNotBackup
		return
	}

	baseContentId := util.GlobbyHash(8, data)

	mediaRoot := ft.Get("MEDIA")
	// Get or create dir for remote core
	remoteDir, err := ft.MkDir(mediaRoot, remoteId)
	if err != nil && !errors.Is(err, filetree.ErrDirAlreadyExists) {
		return
	}

	dataDir, err := ft.MkDir(remoteDir, "data")
	if err != nil && !errors.Is(err, filetree.ErrDirAlreadyExists) {
		return
	}

	baseF, err = ft.Touch(dataDir, baseContentId+".base", false, nil)
	if errors.Is(err, filetree.ErrFileAlreadyExists) {
		return
	} else if err != nil {
		return
	}

	err = baseF.Write(data)
	return
}

func CacheBaseMedia(mediaId types.ContentId, data [][]byte, ft types.FileTree) (newThumb, newFullres types.WeblensFile, err error) {
	cacheRoot := ft.Get("CACHE")
	newThumb, err = ft.Touch(cacheRoot, string(mediaId)+"-thumbnail.cache", false, nil)
	if err != nil && !errors.Is(err, filetree.ErrFileAlreadyExists) {
		return nil, nil, err
	} else if !errors.Is(err, filetree.ErrFileAlreadyExists) {
		err = newThumb.Write(data[0])
		if err != nil {
			return
		}
	}

	newFullres, err = ft.Touch(cacheRoot, string(mediaId)+"-fullres.cache", false, nil)
	if err != nil && !errors.Is(err, filetree.ErrFileAlreadyExists) {
		return nil, nil, err
	} else if !errors.Is(err, filetree.ErrFileAlreadyExists) {
		err = newFullres.Write(data[1])
		if err != nil {
			return
		}
	}

	return
}

func getChildByName(dir types.WeblensFile, childName string) (types.WeblensFile, error) {
	if !dir.IsDir() {
		return nil, filetree.ErrDirectoryRequired
	}

	children := dir.GetChildren()
	_, child, exist := util.YoinkFunc(children, func(c types.WeblensFile) bool {
		return c.Filename() == childName
	})
	if !exist {
		return nil, filetree.ErrNoFile
	}

	return child, nil
}

func MakeUniqueChildName(parent types.WeblensFile, childName string) string {
	dupeCount := 0
	_, e := parent.GetChild(childName)
	for !errors.Is(e, filetree.ErrNoFile) {
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
