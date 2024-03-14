package dataStore

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

const OWNER_SYS = "SYSTEM"

var mediaRoot WeblensFile = WeblensFile{
	id:       "0",
	filename: "MEDIA_ROOT",
	owner:    OWNER_SYS,

	isDir:        boolPointer(true),
	absolutePath: util.GetMediaRoot(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFile{},
}

var tmpRoot WeblensFile = WeblensFile{
	id:       "1",
	filename: "TMP_ROOT",
	owner:    OWNER_SYS,

	isDir:        boolPointer(true),
	absolutePath: util.GetTmpDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFile{},
}

var cacheRoot WeblensFile = WeblensFile{
	id:       "2",
	filename: "CACHE_ROOT",
	owner:    OWNER_SYS,

	isDir:        boolPointer(true),
	absolutePath: util.GetCacheDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFile{},
}

var takeoutRoot WeblensFile = WeblensFile{
	id:       "3",
	filename: "TAKEOUT_ROOT",
	owner:    OWNER_SYS,

	isDir:        boolPointer(true),
	absolutePath: util.GetTakeoutDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFile{},
}

// Initial loading of folders from the database on first read
var initFolderIds []string

func init() {
	fileTree["0"] = &mediaRoot
	fileTree["1"] = &tmpRoot
	fileTree["2"] = &takeoutRoot

	initFolders, err := fddb.getAllFolders()
	if err != nil {
		panic(err)
	}
	initFolderIds = util.Map(initFolders, func(f folderData) string { return f.FolderId })
	slices.SortFunc(initFolderIds, strings.Compare)

	users := fddb.GetUsers()

	for _, user := range users {
		homeDir, err := MkDir(&mediaRoot, user.Username)
		if err != nil {
			switch err.(type) {
			case alreadyExists:
			default:
				{
					panic(err)
				}
			}
		}

		_, err = MkDir(homeDir, ".user_trash")
		if err != nil {
			switch err.(type) {
			case alreadyExists:
			default:
				{
					panic(err)
				}
			}
		}
	}

	cacheRoot.ReadDir()

	// Compute size for the whole tree
	mediaRoot.LeafMap(func(wf *WeblensFile) {
		wf.recompSize()
	})
	cacheRoot.LeafMap(func(wf *WeblensFile) {
		wf.recompSize()
	})

	// Dump initial folder slice
	initFolderIds = nil

	// Put the file tree into safety mode, it stays like this for the rest of runtime
	safety = true
}

// Take a (possibly) absolutePath (string), and return a path to the same location, relative to media root (from .env)
func GuaranteeRelativePath(absolutePath string) string {
	absolutePrefix := util.GetMediaRoot()
	relativePath := filepath.Join("/", strings.TrimPrefix(absolutePath, absolutePrefix))
	return relativePath
}

var dirIgnore = map[string]bool{
	".DS_Store": true,
}

func ClearTempDir() (err error) {
	err = os.MkdirAll(tmpRoot.String(), os.ModePerm)
	if err != nil {
		return
	}

	files, err := os.ReadDir(tmpRoot.String())
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

func ClearTakeoutDir() error {
	os.MkdirAll(takeoutRoot.String(), os.ModePerm)
	files, err := os.ReadDir(takeoutRoot.String())
	if err != nil {
		return err
	}
	for _, file := range files {
		err := os.Remove(filepath.Join(takeoutRoot.String(), file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

///////////////////////////////

var fddb *Weblensdb = NewDB()

func boolPointer(b bool) *bool {
	return &b
}

func CreateUserHomeDir(username string) error {
	homeDir, err := MkDir(GetMediaDir(), strings.ToLower(username))
	if err != nil {
		return err
	}

	_, err = MkDir(homeDir, ".user_trash")
	if err != nil {
		return err
	}

	return nil
}

func GetUserHomeDir(username string) *WeblensFile {
	homeDirHash := util.GlobbyHash(8, filepath.Join(GuaranteeRelativePath(mediaRoot.absolutePath), username))
	homeDir := FsTreeGet(homeDirHash)
	return homeDir
}

func GetUserTrashDir(username string) *WeblensFile {
	trashPath := filepath.Join(GuaranteeRelativePath(mediaRoot.absolutePath), username, ".user_trash")
	trashDirHash := util.GlobbyHash(8, trashPath)
	trash := FsTreeGet(trashDirHash)
	return trash
}

func NewTakeoutZip(zipName string) (newZip *WeblensFile, exists bool, err error) {
	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	newZip, err = Touch(&takeoutRoot, zipName, true)
	switch err {
	case nil:
		return
	case ErrFileAlreadyExists:
		err = nil
		exists = true
	}

	return
}

func MkDir(parentFolder *WeblensFile, newDirName string) (*WeblensFile, error) {
	d := &WeblensFile{}
	d.absolutePath = filepath.Join(parentFolder.absolutePath, newDirName)
	d.isDir = boolPointer(true)

	if d.Exists() {
		existingFile := FsTreeGet(d.Id())

		if existingFile == nil {
			err := FsTreeInsert(d, parentFolder)
			if err != nil {
				return d, err
			}
			existingFile = d
		}

		return existingFile, ErrDirAlreadyExists
	}

	err := d.CreateSelf()
	if err != nil {
		return d, err
	}

	err = FsTreeInsert(d, parentFolder)
	if err != nil {
		return d, err
	}

	return d, nil
}

func Touch(parentFolder *WeblensFile, newFileName string, insert bool) (*WeblensFile, error) {
	f := &WeblensFile{}
	f.absolutePath = filepath.Join(parentFolder.absolutePath, newFileName)
	f.isDir = boolPointer(false)
	if FsTreeGet(f.Id()) != nil || f.Exists() {
		return f, ErrFileAlreadyExists
	}

	err := f.CreateSelf()
	if err != nil {
		return f, err
	}

	if insert {
		err = FsTreeInsert(f, parentFolder)
		if err != nil {
			return f, err
		}
	}

	return f, nil
}

func MoveFileToTrash(file *WeblensFile, c ...BroadcasterAgent) error {
	if !file.Exists() {
		return ErrNoFile
	}

	if len(c) == 0 {
		c = append(c, globalCaster)
	}

	err := FsTreeMove(file, GetUserTrashDir(file.Owner()), file.Filename()+time.Now().Format(".2006-01-02T15.04.05"), true, c...)
	if err != nil {
		return err
	}
	for _, s := range file.GetShares() {
		s.SetEnabled(false)
		UpdateFileShare(s)
	}

	return nil
}

// Removes file being pointed to from the tree and deletes it from the real filesystem
func PermenantlyDeleteFile(file *WeblensFile, c ...BroadcasterAgent) error {
	if len(c) != 0 {
		return FsTreeRemove(file, c...)
	} else {
		return FsTreeRemove(file, globalCaster)
	}
}

func RecursiveGetMedia(folderIds ...string) (ms []string) {
	ms = []string{}

	for _, dId := range folderIds {
		d := FsTreeGet(dId)
		if d == nil {
			util.Warning.Println("Skipping recursive media lookup for non-existant folder")
			continue
		}
		if !d.IsDir() {
			util.Warning.Println("Skipping recursive media lookup for file that is not directoy")
			continue
		}
		d.RecursiveMap(func(f *WeblensFile) {
			dis, _ := f.IsDisplayable()
			if !f.IsDir() && dis {
				m, err := f.GetMedia()
				if err != nil {
					util.DisplayError(err)
					return
				}
				if m != nil {
					ms = append(ms, m.MediaId)
				}
			}
		})
	}

	return
}

func GetCacheDir() *WeblensFile {
	return &cacheRoot
}

func GetTmpDir() *WeblensFile {
	return &tmpRoot
}

func GetMediaDir() *WeblensFile {
	return &mediaRoot
}

var ROOT_IDS []string = []string{"0", "1", "2", "3"}

func IsSystemDir(wf *WeblensFile) bool {
	return slices.Contains(ROOT_IDS, wf.Id())
}
