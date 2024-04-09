package dataStore

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"golang.org/x/sys/unix"
)

var WEBLENS_ROOT_USER *user = &user{
	Username: "WEBLENS",
}

var EXTERNAL_ROOT_USER *user = &user{
	Username: "EXTERNAL",
}

var mediaRoot weblensFile = weblensFile{
	id:       "MEDIA_ROOT",
	filename: "media",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetMediaRootPath(),

	childLock: &sync.Mutex{},
	children:  map[types.FileId]types.WeblensFile{},
}

var tmpRoot weblensFile = weblensFile{
	id:       "TMP_ROOT",
	filename: "tmp",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetTmpDir(),

	childLock: &sync.Mutex{},
	children:  map[types.FileId]types.WeblensFile{},
}

var cacheRoot weblensFile = weblensFile{
	id:       "CACHE_ROOT",
	filename: "cache",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetCacheDir(),

	childLock: &sync.Mutex{},
	children:  map[types.FileId]types.WeblensFile{},
}

var takeoutRoot weblensFile = weblensFile{
	id:       "TAKEOUT_ROOT",
	filename: "takeout",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetTakeoutDir(),

	childLock: &sync.Mutex{},
	children:  map[types.FileId]types.WeblensFile{},
}

var externalRoot weblensFile = weblensFile{
	id:       "EXTERNAL_ROOT",
	filename: "External",
	owner:    EXTERNAL_ROOT_USER,

	isDir: boolPointer(true),

	// This is a fake directory, it houses the mounted external paths, but does not exist itself
	absolutePath: "",
	readOnly:     true,

	childLock: &sync.Mutex{},
	children:  map[types.FileId]types.WeblensFile{},
}

func FsInit() {
	fileTree[mediaRoot.id] = &mediaRoot
	mediaRoot.parent = &mediaRoot
	fileTree[tmpRoot.id] = &tmpRoot
	tmpRoot.parent = &tmpRoot
	fileTree[takeoutRoot.id] = &takeoutRoot
	takeoutRoot.parent = &takeoutRoot
	fileTree[externalRoot.id] = &externalRoot
	externalRoot.parent = &externalRoot

	users := GetUsers()

	for _, user := range users {
		homeDir, err := MkDir(&mediaRoot, user.GetUsername().String())
		if err != nil {
			switch err.(type) {
			case AlreadyExistsError:
			default:
				{
					panic(err)
				}
			}
		}

		_, err = MkDir(homeDir, ".user_trash")
		if err != nil {
			switch err.(type) {
			case AlreadyExistsError:
			default:
				{
					panic(err)
				}
			}
		}
	}

	cacheRoot.ReadDir()

	for _, path := range util.GetExternalPaths() {
		extF := &weblensFile{absolutePath: path, parent: &externalRoot}
		FsTreeInsert(extF, &externalRoot, voidCaster)
	}

	// Compute size for the whole tree
	mediaRoot.LeafMap(func(wf types.WeblensFile) {
		wf.(*weblensFile).loadStat()
	})
	cacheRoot.LeafMap(func(wf types.WeblensFile) {
		wf.(*weblensFile).loadStat()
	})
	externalRoot.LeafMap(func(wf types.WeblensFile) {
		wf.(*weblensFile).loadStat()
	})

	// Put the file tree into safety mode, it stays like this for the rest of runtime
	safety = true

	loadUsersStaticFolders()
}

// Take a (possibly) absolutePath (string), and return a path to the same location, relative to media root (from .env)
func GuaranteeRelativePath(absolutePath string) string {
	absolutePrefix := util.GetMediaRootPath()
	relativePath := filepath.Join("/", strings.TrimPrefix(absolutePath, absolutePrefix))
	return relativePath
}

var dirIgnore = map[string]bool{
	".DS_Store": true,
}

func ClearTempDir() (err error) {
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

func ClearTakeoutDir() error {
	os.MkdirAll(takeoutRoot.GetAbsPath(), os.ModePerm)
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

///////////////////////////////

var fddb *Weblensdb = NewDB()

func boolPointer(b bool) *bool {
	return &b
}

func CreateUserHomeDir(username types.Username) error {
	homeDir, err := MkDir(GetMediaDir(), strings.ToLower(username.String()))
	if err != nil {
		return err
	}

	_, err = MkDir(homeDir, ".user_trash")
	if err != nil {
		return err
	}

	return nil
}

func generateFileId(path string) types.FileId {
	fileHash := types.FileId(util.GlobbyHash(8, filepath.Join(GuaranteeRelativePath(path))))
	return fileHash
}

func GetUserTrashDir(username types.Username) types.WeblensFile {
	trashPath := filepath.Join(GuaranteeRelativePath(mediaRoot.absolutePath), username.String(), ".user_trash")
	trashDirHash := types.FileId(util.GlobbyHash(8, trashPath))
	trash := FsTreeGet(trashDirHash)
	return trash
}

func IsFileInTrash(f types.WeblensFile) bool {
	if f.Owner() == EXTERNAL_ROOT_USER {
		return false
	}
	trashPath := f.Owner().GetTrashFolder().GetAbsPath()
	return strings.HasPrefix(f.GetAbsPath(), trashPath)
}

func NewTakeoutZip(zipName string, creatorName types.Username) (newZip types.WeblensFile, exists bool, err error) {
	user := GetUser(creatorName)
	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	newZip, err = Touch(&takeoutRoot, zipName, true)
	switch err {
	case nil:
		newZip.(*weblensFile).owner = user
		return
	case ErrFileAlreadyExists:
		err = nil
		exists = true
	}

	return
}

func MkDir(parentFolder types.WeblensFile, newDirName string) (types.WeblensFile, error) {
	d := &weblensFile{
		absolutePath: filepath.Join(parentFolder.GetAbsPath(), newDirName),
		isDir:        boolPointer(true),
		// tasksLock:    &sync.Mutex{},
		// childLock:    &sync.Mutex{},
		// children:     map[types.FileId]types.WeblensFile{},
	}

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

func Touch(parentFolder types.WeblensFile, newFileName string, insert bool, c ...types.BroadcasterAgent) (types.WeblensFile, error) {
	f := &weblensFile{
		absolutePath: filepath.Join(parentFolder.GetAbsPath(), newFileName),
		isDir:        boolPointer(false),
		// tasksLock:    &sync.Mutex{},
		// childLock:    &sync.Mutex{},
		// children:     map[types.FileId]types.WeblensFile{},
	}

	if FsTreeGet(f.Id()) != nil || f.Exists() {
		return f, ErrFileAlreadyExists
	}

	err := f.CreateSelf()
	if err != nil {
		return f, err
	}

	if insert {
		err = FsTreeInsert(f, parentFolder, c...)
		if err != nil {
			return f, err
		}
	}

	return f, nil
}

func MoveFileToTrash(file types.WeblensFile, c ...types.BroadcasterAgent) error {
	if !file.Exists() {
		return ErrNoFile
	}

	if len(c) == 0 {
		c = append(c, globalCaster)
	}

	te := trashEntry{
		OrigParent:   file.GetParent().Id(),
		OrigFilename: file.Filename(),
	}

	newFilename := file.Filename() + time.Now().Format(".2006-01-02T15.04.05")
	err := FsTreeMove(file, file.Owner().GetTrashFolder(), newFilename, true, c...)
	if err != nil {
		return err
	}

	te.TrashFileId = file.Id()
	err = fddb.newTrashEntry(te)
	if err != nil {
		return err
	}

	for _, s := range file.GetShares() {
		s.SetEnabled(false)
		UpdateFileShare(s)
	}

	return nil
}

func ReturnFileFromTrash(trashFile types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	te, err := fddb.getTrashEntry(trashFile.Id())
	if err != nil {
		return
	}

	oldParent := FsTreeGet(te.OrigParent)
	trashFile.Owner()
	if oldParent == nil {
		oldParent = trashFile.Owner().GetTrashFolder()
	}
	err = FsTreeMove(trashFile, oldParent, te.OrigFilename, false, c...)

	if err != nil {
		return
	}

	err = fddb.removeTrashEntry(te.TrashFileId)
	if err != nil {
		return
	}

	for _, s := range trashFile.GetShares() {
		s.SetEnabled(true)
		UpdateFileShare(s)
	}

	return
}

// Removes file being pointed to from the tree and deletes it from the real filesystem
func PermenantlyDeleteFile(file types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	if len(c) == 0 {
		c = append(c, globalCaster)
	}
	err = FsTreeRemove(file, c...)

	if err != nil {
		return
	}
	fddb.removeTrashEntry(file.Id())
	return
}

func RecursiveGetMedia(folderIds ...types.FileId) (ms []types.MediaId) {
	ms = []types.MediaId{}

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
		d.RecursiveMap(func(f types.WeblensFile) {
			dis, _ := f.IsDisplayable()
			if !f.IsDir() && dis {
				m, err := f.GetMedia()
				if err != nil && err != ErrNoMedia {
					util.ErrTrace(err)
					return
				} else if m != nil {
					ms = append(ms, m.Id())
				}
			}
		})
	}

	return
}

func GetCacheDir() types.WeblensFile {
	return &cacheRoot
}

func GetTmpDir() types.WeblensFile {
	return &tmpRoot
}

func GetMediaDir() types.WeblensFile {
	return &mediaRoot
}

func GetExternalDir() types.WeblensFile {
	return &externalRoot
}

var ROOT_IDS []string = []string{"0", "1", "2", "3"}

func IsSystemDir(wf types.WeblensFile) bool {
	return slices.Contains(ROOT_IDS, wf.Id().String())
}

func GetFreeSpace(path string) uint64 {
	var stat unix.Statfs_t
	unix.Statfs(path, &stat)

	spaceBytes := stat.Bavail * uint64(stat.Bsize)
	return spaceBytes
}
