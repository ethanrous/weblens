package dataStore

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

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

var ROOT_DIR_IDS = []types.FileId{"MEDIA", "TMP", "CACHE", "TAKEOUT", "EXTERNAL"}

// Directory to hold the users home directories, and a few extras.
// This is where most of the data on the filesystem will be stored
var mediaRoot weblensFile = weblensFile{
	id:       "MEDIA",
	filename: "media",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetMediaRootPath(),

	childLock: &sync.Mutex{},
	children:  []*weblensFile{},
}

// Under the media root, content root is where backup files are stored
// when they are no longer in a user directory, or links to the real files
// when they are in use
var contentRoot weblensFile = weblensFile{
	id:       "CONTENT_LINKS",
	filename: ".content",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: mediaRoot.absolutePath + "/.content/",

	childLock: &sync.Mutex{},
	children:  []*weblensFile{},
}

var tmpRoot weblensFile = weblensFile{
	id:       "TMP",
	filename: "tmp",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetTmpDir(),

	childLock: &sync.Mutex{},
	children:  []*weblensFile{},
}

var cacheRoot weblensFile = weblensFile{
	id:       "CACHE",
	filename: "cache",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetCacheDir(),

	childLock: &sync.Mutex{},
	children:  []*weblensFile{},
}

var takeoutRoot weblensFile = weblensFile{
	id:       "TAKEOUT",
	filename: "takeout",
	owner:    WEBLENS_ROOT_USER,

	isDir:        boolPointer(true),
	absolutePath: util.GetTakeoutDir(),

	childLock: &sync.Mutex{},
	children:  []*weblensFile{},
}

var externalRoot weblensFile = weblensFile{
	id:       "EXTERNAL",
	filename: "External",
	owner:    EXTERNAL_ROOT_USER,

	isDir: boolPointer(true),

	// This is a fake directory, it houses the mounted external paths, but does not exist itself
	absolutePath: "",
	readOnly:     true,

	childLock: &sync.Mutex{},
	children:  []*weblensFile{},
}

func FsInit() {
	fileTree[mediaRoot.id] = &mediaRoot
	mediaRoot.parent = &mediaRoot
	fileTree[tmpRoot.id] = &tmpRoot
	tmpRoot.parent = &tmpRoot
	fileTree[takeoutRoot.id] = &takeoutRoot
	takeoutRoot.parent = &takeoutRoot
	fileTree[externalRoot.id] = &externalRoot
	cacheRoot.parent = &cacheRoot
	fileTree[cacheRoot.id] = &cacheRoot
	contentRoot.parent = &mediaRoot
	fileTree[contentRoot.id] = &contentRoot
	fileTree[generateFileId(contentRoot.absolutePath)] = &contentRoot

	if GetServerInfo().ServerRole() == types.Core {
		externalRoot.parent = &externalRoot
	}

	var err error
	existingBackups, err = fddb.getJournaledFiles()
	if err != nil {
		panic(err)
	}

	slices.SortFunc(existingBackups, func(a, b backupFile) int { return strings.Compare(string(a.FileId), string(b.FileId)) })

	if !mediaRoot.Exists() {
		err = mediaRoot.CreateSelf()
		if err != nil {
			panic(err)
		}
	}

	go journalWorker()
	go fileWatcher()

	if !contentRoot.Exists() {
		err = contentRoot.CreateSelf()
		if err != nil {
			panic(err)
		}
	}

	users := getUsers()
	for _, user := range users {
		homeDir := newWeblensFile(&mediaRoot, user.GetUsername().String(), true)
		homeDir.owner = user

		err := homeDir.CreateSelf()
		if err != nil {
			switch err.(type) {
			case AlreadyExistsError:
			default:
				{
					panic(err)
				}
			}
		}

		err = fsTreeInsert(homeDir, &mediaRoot)
		if err != nil {
			panic(err)
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
	mediaRoot.ReadDir()
	contentRoot.ReadDir()

	for _, path := range util.GetExternalPaths() {
		if path == "" {
			continue
		}
		s, err := os.Stat(path)
		if err != nil {
			panic(fmt.Sprintf("Could not find external path: %s", path))
		}
		extF := newWeblensFile(&externalRoot, filepath.Base(path), s.IsDir())
		fsTreeInsert(extF, &externalRoot, voidCaster)
	}

	// Compute size for the whole tree
	err = mediaRoot.LeafMap(func(wf types.WeblensFile) error {
		return wf.(*weblensFile).loadStat()
	})
	if err != nil {
		panic(err)
	}

	err = cacheRoot.LeafMap(func(wf types.WeblensFile) error {
		return wf.(*weblensFile).loadStat()
	})
	if err != nil {
		panic(err)
	}

	if externalRoot.parent != &mediaRoot {
		err = externalRoot.LeafMap(func(wf types.WeblensFile) error {
			return wf.(*weblensFile).loadStat()
		})
		if err != nil {
			panic(err)
		}
	}

	// Put the file tree into safety mode, it stays like this for the rest of runtime
	safety = true

	// clean up journal init
	existingBackups = nil

	loadUsersStaticFolders()
}

// Take a (possibly) absolutePath (string), and return a path to the same location, relative to media root (from .env)
func GuaranteeRelativePath(absolutePath string) string {
	return AbsToPortable(absolutePath).postfix
}

func AbsToPortable(absPath string) (port portablePath) {
	var short string

	for _, root_id := range ROOT_DIR_IDS {
		if root_id == "EXTERNAL" {
			continue
		}

		root_dir := FsTreeGet(root_id)
		short = strings.TrimPrefix(absPath, root_dir.GetAbsPath())
		if len(short) < len(absPath) {
			port.prefix = string(root_id)
			port.postfix = short
			return
		}
	}

	externalPaths := util.GetExternalPaths()
	for _, p := range externalPaths {
		short = strings.TrimPrefix(absPath, p)
		if len(short) < len(absPath) {
			port.prefix = "EXTERNAL"
			port.postfix = short
			return
		}
	}

	return port
}

func (port portablePath) PortableString() string {
	if len(port.postfix) != 0 && port.postfix[0:1] == "/" {
		port.postfix = port.postfix[1:]
	}
	return port.prefix + ":" + port.postfix
}

func (port portablePath) Abs() string {
	return FsTreeGet(types.FileId(port.prefix)).GetAbsPath() + port.postfix
}

func portableFromString(path string) portablePath {
	parts := strings.Split(path, ":")
	if len(parts) != 2 {
		util.ErrTrace(fmt.Errorf("%s: %s", "creating portable path does not have 2 parts", path))
		return portablePath{}
	}
	return portablePath{
		prefix:  parts[0],
		postfix: parts[1],
	}
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

var fddb *WeblensDB = NewDB()

func boolPointer(b bool) *bool {
	return &b
}

func generateFileId(absPath string) types.FileId {
	fileHash := types.FileId(util.GlobbyHash(8, AbsToPortable(absPath).PortableString()))
	return fileHash
}

func CreateUserHomeDir(user types.User) (types.WeblensFile, error) {
	homeDir, err := MkDir(GetMediaDir(), strings.ToLower(string(user.GetUsername())))
	if err != nil && err != ErrDirAlreadyExists {
		return homeDir, err
	}

	_, err = MkDir(homeDir, ".user_trash")
	if err != nil && err != ErrDirAlreadyExists {
		return homeDir, err
	}

	return homeDir, nil
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

	newZip, err = Touch(&takeoutRoot, zipName, false)
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

func newWeblensFile(parent types.WeblensFile, filename string, isDir bool) *weblensFile {
	return &weblensFile{
		parent:   parent.(*weblensFile),
		filename: filename,
		isDir:    boolPointer(isDir),
		owner:    parent.Owner(),

		tasksLock: &sync.Mutex{},
		childLock: &sync.Mutex{},
		children:  []*weblensFile{},

		size: -1,
	}
}

// Create a new dir as a child of parentFolder named newDirName. If the dir already exists,
// it will be returned along with a ErrDirAlreadyExists error.
func MkDir(parentFolder types.WeblensFile, newDirName string, c ...types.BroadcasterAgent) (types.WeblensFile, error) {
	d := newWeblensFile(parentFolder, newDirName, true)

	if d.Exists() {
		existingFile := FsTreeGet(d.Id())

		if existingFile == nil {
			err := fsTreeInsert(d, parentFolder, c...)
			if err != nil {
				return d, err
			}
			existingFile = d
		}

		return existingFile, ErrDirAlreadyExists
	}

	d.size = 0

	err := fsTreeInsert(d, parentFolder, c...)
	if err != nil {
		return d, err
	}

	err = d.CreateSelf()
	if err != nil {
		return d, err
	}

	return d, nil
}

func Touch(parentFolder types.WeblensFile, newFileName string, detach bool, c ...types.BroadcasterAgent) (types.WeblensFile, error) {
	f := newWeblensFile(parentFolder, newFileName, false)
	f.detached = detach
	e := FsTreeGet(f.Id())
	if e != nil || f.Exists() {
		return e, ErrFileAlreadyExists
	}

	err := f.CreateSelf()
	if err != nil {
		return f, err
	}

	// Detach creates the file on the real filesystem,
	// but does not add it to the tree or journal its creation
	if detach {
		return f, nil
	}

	err = fsTreeInsert(f, parentFolder, c...)
	if err != nil {
		return f, err
	}

	return f, nil
}

// AttachFile takes a detached file when it is ready to be inserted to the tree, and attaches it
func AttachFile(f types.WeblensFile, c ...types.BroadcasterAgent) error {
	if FsTreeGet(f.Id()) != nil {
		return ErrFileAlreadyExists
	}

	err := fsTreeInsert(f, f.GetParent(), c...)
	if err != nil {
		return err
	}

	tmpFile, err := os.Open("/tmp/" + f.Filename())
	if err != nil {
		return err
	}

	destFile, err := os.Create(f.GetAbsPath())
	if err != nil {
		return err
	}

	_, err = io.Copy(destFile, tmpFile)
	if err != nil {
		return err
	}

	return nil
}

func createLink(linkTo, parent types.WeblensFile, filename string) (types.WeblensFile, error) {
	link := newWeblensFile(parent, filename, false)

	err := os.Symlink(linkTo.GetAbsPath(), link.GetAbsPath())
	if err != nil {
		return nil, err
	}

	fsTreeInsert(link, parent)

	return link, nil
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

	trash := file.Owner().GetTrashFolder()
	newFilename := MakeUniqueChildName(trash, file.Filename())

	buffered := util.SliceConvert[types.BufferedBroadcasterAgent](util.Filter(c, func(b types.BroadcasterAgent) bool { return b.IsBuffered() }))
	err := FsTreeMove(file, trash, newFilename, false, buffered...)
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

	buffered := util.SliceConvert[types.BufferedBroadcasterAgent](util.Filter(c, func(b types.BroadcasterAgent) bool { return b.IsBuffered() }))
	err = FsTreeMove(trashFile, oldParent, te.OrigFilename, false, buffered...)

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
func PermanentlyDeleteFile(file types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	if len(c) == 0 {
		c = append(c, globalCaster)
	}
	err = fsTreeRemove(file, c...)

	if err != nil {
		return
	}
	fddb.removeTrashEntry(file.Id())
	return
}

func RecursiveGetMedia(folderIds ...types.FileId) (ms []types.ContentId) {
	ms = []types.ContentId{}

	for _, dId := range folderIds {
		d := FsTreeGet(dId)
		if d == nil {
			util.Warning.Println("Skipping recursive media lookup for non-existent folder")
			continue
		}
		if !d.IsDir() {
			util.Warning.Println("Skipping recursive media lookup for file that is not directory")
			continue
		}
		err := d.RecursiveMap(func(f types.WeblensFile) error {
			if !f.IsDir() && f.IsDisplayable() {
				m := MediaMapGet(f.GetContentId())

				if m != nil {
					ms = append(ms, m.Id())
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

func GetFreeSpace(path string) uint64 {
	var stat unix.Statfs_t
	unix.Statfs(path, &stat)

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
	defer fp.Close()
	_, err = io.CopyBuffer(newHash, fp, buf)
	if err != nil {
		return "", err
	}

	contentId := types.ContentId(base64.URLEncoding.EncodeToString(newHash.Sum(nil)))[:20]
	f.(*weblensFile).contentId = contentId

	return contentId, nil
}

func BackupBaseFile(remoteId string, data []byte) (baseF types.WeblensFile, err error) {
	if thisServer.Role != types.Backup {
		err = ErrNotBackup
		return
	}

	baseContentId := util.GlobbyHash(8, data)

	// Get or create dir for remote core
	remoteDir, err := MkDir(&mediaRoot, remoteId)
	if err != nil && err != ErrDirAlreadyExists {
		return
	}

	dataDir, err := MkDir(remoteDir, "data")
	if err != nil && err != ErrDirAlreadyExists {
		return
	}

	baseF, err = Touch(dataDir, baseContentId+".base", false)
	if err == ErrFileAlreadyExists {
		return
	} else if err != nil {
		return
	}

	err = baseF.Write(data)
	return
}

func CacheBaseMedia(mediaId types.ContentId, data [][]byte) (newThumb, newFullres types.WeblensFile, err error) {
	newThumb, err = Touch(&cacheRoot, string(mediaId)+"-thumbnail.wlcache", false)
	if err != nil && err != ErrFileAlreadyExists {
		return nil, nil, err
	} else if err != ErrFileAlreadyExists {
		err = newThumb.Write(data[0])
		if err != nil {
			return
		}
	}

	newFullres, err = Touch(&cacheRoot, string(mediaId)+"-fullres.wlcache", false)
	if err != nil && err != ErrFileAlreadyExists {
		return nil, nil, err
	} else if err != ErrFileAlreadyExists {
		err = newFullres.Write(data[1])
		if err != nil {
			return
		}
	}

	return
}

func getChildByName(dir types.WeblensFile, childName string) (types.WeblensFile, error) {
	if !dir.IsDir() {
		return nil, ErrDirectoryRequired
	}

	children := dir.GetChildren()
	_, child, exist := util.YoinkFunc(children, func(c types.WeblensFile) bool {
		return c.Filename() == childName
	})
	if !exist {
		return nil, ErrNoFile
	}

	return child, nil
}

func isDirByPath(absPath string) (bool, error) {
	stat, err := os.Stat(absPath)
	if err != nil {
		return false, err
	}

	return stat.IsDir(), nil
}

func MakeUniqueChildName(parent types.WeblensFile, childName string) string {
	dupeCount := 0
	_, e := parent.GetChild(childName)
	for e != ErrNoFile {
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
