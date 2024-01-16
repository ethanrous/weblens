package dataStore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

type FileInfo struct {
	Id             string    `json:"id"`
	Imported       bool      `json:"imported"` // If the item has been loaded into the database, dictates if MediaData is set or not
	IsDir          bool      `json:"isDir"`
	Modifiable     bool      `json:"modifiable"`
	Size           int64     `json:"size"`
	ModTime        time.Time `json:"modTime"`
	Filename       string    `json:"filename"`
	ParentFolderId string    `json:"parentFolderId"`
	MediaData      Media     `json:"mediaData"`
	Owner          string    `json:"owner"`
}

var mediaRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id:       "0",
	filename: "MEDIA_ROOT",
	owner:    "SYS",

	isDir:        boolPointer(true),
	absolutePath: util.GetMediaRoot(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFileDescriptor{},
}

var tmpRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id:       "1",
	filename: "TMP_ROOT",
	owner:    "SYS",

	isDir:        boolPointer(true),
	absolutePath: util.GetTmpDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFileDescriptor{},
}

var trashRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id:       "2",
	filename: "TRASH_ROOT",
	owner:    "SYS",

	isDir:        boolPointer(true),
	absolutePath: util.GetTrashDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFileDescriptor{},
}

var takeoutRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id:       "3",
	filename: "TAKEOUT_ROOT",
	owner:    "SYS",

	isDir:        boolPointer(true),
	absolutePath: util.GetTakeoutDir(),

	childLock: &sync.Mutex{},
	children:  map[string]*WeblensFileDescriptor{},
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

func ImportHomeDirectories() error {
	files, err := os.ReadDir(mediaRoot.absolutePath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			dirPath := filepath.Join(mediaRoot.absolutePath, file.Name())
			_, err := fddb.importDirectory(dirPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getUserHomeDirectories() (homeDirs []*WeblensFileDescriptor, err error) {
	files, err := os.ReadDir(mediaRoot.absolutePath)
	if err != nil {
		return
	}
	for _, homeDir := range files {
		if !homeDir.IsDir() {
			continue
		}
		tmpHomeDescriptor := WeblensFileDescriptor{}
		tmpHomeDescriptor.absolutePath = filepath.Join(mediaRoot.absolutePath, homeDir.Name())
		tmpHomeDescriptor.owner = homeDir.Name()
		tmpHomeDescriptor.filename = homeDir.Name()
		homeDirs = append(homeDirs, &tmpHomeDescriptor)
	}
	return
}

///////////////////////////////

var fddb *Weblensdb = NewDB("")

func boolPointer(b bool) *bool {
	return &b
}

func GetUserHomeDir(username string) *WeblensFileDescriptor {
	homeDirHash := util.HashOfString(8, filepath.Join(GuaranteeRelativePath(mediaRoot.absolutePath), username))
	return FsTreeGet(homeDirHash)
}

func GetUserTrashDir(username string) *WeblensFileDescriptor {
	trashDirHash := util.HashOfString(8, filepath.Join(GuaranteeRelativePath(mediaRoot.absolutePath), username, ".user_trash"))
	return FsTreeGet(trashDirHash)
}

func NewTakeoutZip(zipName string) (*WeblensFileDescriptor, bool, error) {
	newZip := &WeblensFileDescriptor{}

	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	newZip.absolutePath = filepath.Join(takeoutRoot.String(), zipName)
	if v := FsTreeGet(newZip.Id()); v != nil {
		return newZip, true, nil
	}

	err := FsTreeInsert(newZip, takeoutRoot.Id())
	if err != nil {
		return nil, false, err
	}

	return newZip, false, nil
}

func MkDir(parentFolder *WeblensFileDescriptor, newDirName string) (*WeblensFileDescriptor, error) {
	d := &WeblensFileDescriptor{}
	d.absolutePath = filepath.Join(parentFolder.absolutePath, newDirName)
	d.isDir = boolPointer(true)
	if d.Exists() {
		return d, fmt.Errorf("trying create dir that already exists")
	}
	err := d.CreateSelf()
	if err != nil {
		return d, err
	}

	err = FsTreeInsert(d, parentFolder.Id())
	if err != nil {
		return d, err
	}

	return d, nil
}

func Touch(parentFolder *WeblensFileDescriptor, newFileName string, insert bool) (*WeblensFileDescriptor, error) {
	f := &WeblensFileDescriptor{}
	f.absolutePath = filepath.Join(parentFolder.absolutePath, newFileName)
	f.isDir = boolPointer(false)
	if f.Exists() {
		return f, fmt.Errorf("trying create file that already exists")
	}

	err := f.CreateSelf()
	if err != nil {
		return f, err
	}

	if insert {
		err = FsTreeInsert(f, parentFolder.Id())
		if err != nil {
			return f, err
		}
	}

	return f, nil
}

func MoveFileToTrash(file *WeblensFileDescriptor) error {
	if !file.Exists() {
		return fmt.Errorf("attempting to move a non-existant file to trash")
	}

	err := FsTreeMove(file, GetUserTrashDir(file.Owner()), file.Filename()+time.Now().Format(".2006-01-02T15.04.05"), true)
	if err != nil {
		return err
	}

	return nil
}

func PermenantlyDeleteFile(file *WeblensFileDescriptor) {
	FsTreeRemove(file)
}

func GetTmpDir() *WeblensFileDescriptor {
	return &tmpRoot
}

func GetMediaDir() *WeblensFileDescriptor {
	return &mediaRoot
}
