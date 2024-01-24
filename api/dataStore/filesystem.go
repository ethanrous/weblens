package dataStore

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

var mediaRoot WeblensFile = WeblensFile{
	id:       "0",
	filename: "MEDIA_ROOT",
	owner:    "SYS",

	isDir:        boolPointer(true),
	absolutePath: util.GetMediaRoot(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFile{},
}

var tmpRoot WeblensFile = WeblensFile{
	id:       "1",
	filename: "TMP_ROOT",
	owner:    "SYS",

	isDir:        boolPointer(true),
	absolutePath: util.GetTmpDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFile{},
}

var takeoutRoot WeblensFile = WeblensFile{
	id:       "3",
	filename: "TAKEOUT_ROOT",
	owner:    "SYS",

	isDir:        boolPointer(true),
	absolutePath: util.GetTakeoutDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFile{},
}

// Initial loading of folders from the database on first read
var initFolderIds []string

// Initialize the filesystem.
func FsInit() {
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

	// Compute size for the whole tree
	mediaRoot.LeafMap(func(wf *WeblensFile) {
		wf.recompSize()
	})

	// Dump initial folder slice
	initFolderIds = nil

	// Put the file tree into safty mode, it stays like this for the rest of runtime
	safety = true

	util.Debug.Println("Filesystem initialized without error")

}

// Take a (possibly) absolutePath (string), and return a path to the same location, relative to media root (from .env)
func GuaranteeRelativePath(absolutePath string) string {
	absolutePrefix := util.GetMediaRoot()
	relativePath := filepath.Join("/", strings.TrimPrefix(absolutePath, absolutePrefix))
	return relativePath
}

// Take a possibly absolute `path` (string), and return a path to the same location, relative to the given users home directory
// Returns an error if the file is not in the users home directory, or tries to access the "SYS" home directory, which does not exist
func GuaranteeUserRelativePath(path, username string) (string, error) {
	if username == "SYS" {
		return "", fmt.Errorf("attempt to get relative path with SYS user")
	}

	absolutePrefix := filepath.Join(util.GetMediaRoot(), username)
	if isAbsolutePath(path) && !strings.HasPrefix(path, absolutePrefix) {
		return "", fmt.Errorf("attempt to get user relative path for a file not in user's home directory\n File: %s\nUser: %s", path, username)
	}

	relativePath := filepath.Join("/", strings.TrimPrefix(path, absolutePrefix))
	return relativePath, nil
}

func GuaranteeAbsolutePath(relativePath string) string {
	if isAbsolutePath(relativePath) {
		util.Warning.Printf("Relative path was already absolute path: %s", relativePath)
		return relativePath
	}

	absolutePrefix := util.GetMediaRoot()
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func isAbsolutePath(mysteryPath string) bool {
	return strings.HasPrefix(mysteryPath, util.GetMediaRoot())
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

func CreateUserHomeDir(username string) {
	homeDir, err := MkDir(GetMediaDir(), username)
	util.DisplayError(err)

	_, err = MkDir(homeDir, ".user_trash")
	util.DisplayError(err)
}

func GetUserHomeDir(username string) *WeblensFile {
	homeDirHash := util.HashOfString(8, filepath.Join(GuaranteeRelativePath(mediaRoot.absolutePath), username))
	homeDir := FsTreeGet(homeDirHash)
	return homeDir
}

func GetUserTrashDir(username string) *WeblensFile {
	trashPath := filepath.Join(GuaranteeRelativePath(mediaRoot.absolutePath), username, ".user_trash")
	trashDirHash := util.HashOfString(8, trashPath)
	trash := FsTreeGet(trashDirHash)
	return trash
}

func NewTakeoutZip(zipName string) (*WeblensFile, bool, error) {
	newZip := &WeblensFile{}

	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	newZip.absolutePath = filepath.Join(takeoutRoot.String(), zipName)
	if v := FsTreeGet(newZip.Id()); v != nil {
		return newZip, true, nil
	}

	err := FsTreeInsert(newZip, &takeoutRoot)
	if err != nil {
		return nil, false, err
	}

	return newZip, false, nil
}

func MkDir(parentFolder *WeblensFile, newDirName string) (*WeblensFile, error) {
	d := &WeblensFile{}
	d.absolutePath = filepath.Join(parentFolder.absolutePath, newDirName)
	d.isDir = boolPointer(true)

	if d.Exists() {
		existsErr := fmt.Errorf("trying create dir that already exists").(alreadyExists)
		existingFile := FsTreeGet(d.Id())

		if existingFile == nil {
			err := FsTreeInsert(d, parentFolder)
			if err != nil {
				return d, err
			}
			existingFile = d
		}

		return existingFile, existsErr
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
		return f, fmt.Errorf("trying create file that already exists").(alreadyExists)
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

func MoveFileToTrash(file *WeblensFile) error {
	if !file.Exists() {
		return fmt.Errorf("attempting to move a non-existant file to trash")
	}

	err := FsTreeMove(file, GetUserTrashDir(file.Owner()), file.Filename()+time.Now().Format(".2006-01-02T15.04.05"), true)
	if err != nil {
		return err
	}

	return nil
}

// Removes file being pointed to from the tree and deletes it from the real filesystem
func PermenantlyDeleteFile(file *WeblensFile) {
	FsTreeRemove(file)
}

func GetTmpDir() *WeblensFile {
	return &tmpRoot
}

func GetMediaDir() *WeblensFile {
	return &mediaRoot
}
