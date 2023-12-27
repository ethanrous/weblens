package dataStore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/util"
	"github.com/google/uuid"
)

type FileInfo struct{
	Id string `json:"id"`
	Imported bool `json:"imported"` // If the item has been loaded into the database, dictates if MediaData is set or not
	IsDir bool `json:"isDir"`
	Size int `json:"size"`
	ModTime time.Time `json:"modTime"`
	Filename string `json:"filename"`
	ParentFolderId string `json:"parentFolderId"`
	MediaData Media `json:"mediaData"`
	Owner string `json:"owner"`
}


// Take a (possibly) absolutePath (string), and return a path to the same location, relative to media root (from .env)
func GuaranteeRelativePath(absolutePath string) (string) {
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

func GuaranteeAbsolutePath(relativePath string) (string) {
	if isAbsolutePath(relativePath) {
		util.Warning.Printf("Relative path was already absolute path: %s", relativePath)
		return relativePath
	}

	absolutePrefix := util.GetMediaRoot()
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func GuaranteeUserAbsolutePath(relativePath, username string) (string) {
	if username == "SYS" {
		util.Error.Panicln("Attempt to get absolute path with SYS user")
	}

	if isAbsolutePath(relativePath) {
		util.Warning.Printf("Relative path was already absolute path: %s", relativePath)
		return relativePath
	}

	relativePath = strings.TrimPrefix(relativePath, "/" + username)

	absolutePrefix := filepath.Join(util.GetMediaRoot(), username)
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func isAbsolutePath(mysteryPath string) (bool) {
	return strings.HasPrefix(mysteryPath, util.GetMediaRoot())
}

func GetOwnerFromFilepath(path string) string {
	relativePath := GuaranteeRelativePath(path)
	if string(relativePath[0]) == "/" {
		relativePath = relativePath[1:]
	}
	index := strings.Index(relativePath, "/")
	if index == -1 {
		return relativePath
	} else {
		return relativePath[:index]
	}
}

var dirIgnore = map[string]bool {
	".DS_Store": true,
}

func (f *WeblensFileDescriptor) FormatFileInfo() (FileInfo, error) {
	var formattedInfo FileInfo
	if !f.Exists() {
		return formattedInfo, fmt.Errorf("file does not exist")
	}
	if f.isDir == nil {
		return formattedInfo, fmt.Errorf("directory status not initialized")
	}

	if !dirIgnore[f.Filename] {
		var imported bool = true

		var m Media
		var mt mediaType = f.getMediaType()

		if mt.IsDisplayable {
			var err error
			m, err = f.GetMedia()
			if err != nil {
				imported = false
			}
		} else {
			m.MediaType = mt
		}

		m.Thumbnail64 = ""

		formattedInfo = FileInfo{Id: f.Id(), Imported: imported, IsDir: *f.isDir, Size: int(f.Size()), ModTime: f.ModTime(), Filename: f.Filename, ParentFolderId: f.ParentFolderId, MediaData: m, Owner: f.owner}
	} else {
		return formattedInfo, fmt.Errorf("filename in blocklist")
	}
	return formattedInfo, nil
}

func ClearTempDir() error {
	files, err := os.ReadDir(util.GetTmpDir())
	if err != nil {
		return err
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
	files, err := os.ReadDir(util.GetTakeoutDir())
	if err != nil {
		return err
	}
	for _, file := range files {
		err := os.Remove(filepath.Join(util.GetTakeoutDir(), file.Name()))
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

///////////////////////////////

var fddb *Weblensdb = NewDB("")

type WFDCreateOptions struct {
	IgnoreNonexistance *bool
	TypeHint *fileTypeHint // "dir" or "file"
}

type WFDMoveOptions struct {
	SkipMediaMove *bool
	SkipIdRecompute *bool
}

func CreateOpts() *WFDCreateOptions {
	return &WFDCreateOptions{}
}
func MoveOpts() *WFDMoveOptions {
	return &WFDMoveOptions{}
}

func (o *WFDCreateOptions) SetIgnoreNonexistance(i bool) *WFDCreateOptions {
	o.IgnoreNonexistance = &i
	return o
}

type fileTypeHint string
const (
	dir = "dir"
	file = "file"
)

func (o *WFDCreateOptions) SetTypeHint(t fileTypeHint) *WFDCreateOptions {
	o.TypeHint = &t
	return o
}

func (o *WFDMoveOptions) SetSkipIdRecompute(b bool) *WFDMoveOptions {
	o.SkipIdRecompute = &b
	return o
}

func (o *WFDMoveOptions) SetSkipMediaMove(b bool) *WFDMoveOptions {
	o.SkipMediaMove = &b
	return o
}

func mergeCreateOpts(opts ...*WFDCreateOptions) *WFDCreateOptions {
	var finalOpt WFDCreateOptions
	for _, o := range opts {
		if o.IgnoreNonexistance != nil {
			finalOpt.IgnoreNonexistance = o.IgnoreNonexistance
		}
		if o.TypeHint != nil {
			finalOpt.TypeHint = o.TypeHint
		}
	}

	// Set defaults
	if finalOpt.IgnoreNonexistance == nil {
		finalOpt.IgnoreNonexistance = boolPointer(false)
	}
	return &finalOpt
}

func mergeMoveOpts(opts ...*WFDMoveOptions) *WFDMoveOptions {
	var finalOpt WFDMoveOptions
	for _, o := range opts {
		if o.SkipMediaMove != nil {
			finalOpt.SkipMediaMove = o.SkipMediaMove
		}
		if o.SkipIdRecompute != nil {
			finalOpt.SkipIdRecompute = o.SkipIdRecompute
		}
	}

	// Set defaults
	if finalOpt.SkipMediaMove == nil {
		finalOpt.SkipMediaMove = boolPointer(false)
	}
	if finalOpt.SkipIdRecompute == nil {
		finalOpt.SkipIdRecompute = boolPointer(false)
	}
	return &finalOpt
}

type WeblensFileDescriptor struct {
	id string
	ParentFolderId string
	Filename string
	owner string

	isDir *bool
	absolutePath string
	err error
}

func boolPointer(b bool) *bool {
	return &b
}

var mediaRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id: "0",
	ParentFolderId: "0",
	Filename: "MEDIA_ROOT",
	owner: "SYS",

	isDir: boolPointer(true),
	absolutePath: util.GetMediaRoot(),
}

var tmpRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id: "1",
	ParentFolderId: "1",
	Filename: "TMP_ROOT",
	owner: "SYS",

	isDir: boolPointer(true),
	absolutePath: util.GetTmpDir(),
}

var trashRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id: "2",
	ParentFolderId: "2",
	Filename: "TRASH_ROOT",
	owner: "SYS",

	isDir: boolPointer(true),
	absolutePath: util.GetTrashDir(),
}

var takeoutRoot WeblensFileDescriptor = WeblensFileDescriptor{
	id: "3",
	ParentFolderId: "3",
	Filename: "TAKEOUT_ROOT",
	owner: "SYS",

	isDir: boolPointer(true),
	absolutePath: util.GetTakeoutDir(),
}

func WFDByFolderId(folderId string) *WeblensFileDescriptor {
	var f *WeblensFileDescriptor = &WeblensFileDescriptor{}
	if folderId == "" {
		f.err = fmt.Errorf("unable to get folder with empty id")
		return f
	}
	if folderId == "0" {
		f := mediaRoot // Dont return pointer to mediaRoot, copy first
		return &f
	}

	f.isDir = boolPointer(true)
	folderData := fddb.getFolderById(folderId)
	f.absolutePath = GuaranteeAbsolutePath(folderData.DirPath)
	f.id = folderData.FolderId
	f.ParentFolderId = folderData.ParentId

	f.init(&WFDCreateOptions{})

	return f
}

func GetWFD(parentFolderId, filename string, opts ...*WFDCreateOptions) *WeblensFileDescriptor {
	folderData := fddb.getFolderById(parentFolderId)
	var f *WeblensFileDescriptor = &WeblensFileDescriptor{}

	f.absolutePath = filepath.Join(GuaranteeAbsolutePath(folderData.DirPath), filename)
	f.ParentFolderId = parentFolderId
	f.Filename = filename

	f.init(opts...)

	return f
}

func WFDByPath(path string) *WeblensFileDescriptor {
	var f *WeblensFileDescriptor = &WeblensFileDescriptor{}
	f.absolutePath = path

	f.init()

	return f
}

// Create a WFD as a child to the instance this method is called on, named childFilename.
// Useful for pointing to a file location that does not exist, and calling CreateSelf().
// Note that the underlying file could still exist, if it need not exist, call this and check elsewhere with .Exists().
// Also, this is perhaps the greatest function name ever
func (f *WeblensFileDescriptor) CreateGhostChild(childFilename string) *WeblensFileDescriptor {
	var ghostChild *WeblensFileDescriptor = &WeblensFileDescriptor{}

	ghostChild.absolutePath = filepath.Join(f.absolutePath, childFilename)
	ghostChild.ParentFolderId = f.Id()
	ghostChild.owner = f.owner
	ghostChild.Filename = childFilename

	return ghostChild
}

func (f *WeblensFileDescriptor) init(opts... *WFDCreateOptions) {
	if f.absolutePath == "" {
		f.err = fmt.Errorf("cannot init WFD without an absolute path")
		return
	}

	if f.Filename == "" {
		f.Filename = filepath.Base(f.absolutePath)
	}
	if f.owner == "" {
		f.owner = GetOwnerFromFilepath(f.absolutePath)
	}

	opt := mergeCreateOpts(opts...)

	if opt.TypeHint != nil {
		f.isDir = boolPointer(*opt.TypeHint == dir)
	}

	if !f.Exists() {
		if !*opt.IgnoreNonexistance {
			f.err = fmt.Errorf("not continuing WFD init on non-existent file: %s", f.absolutePath)
		}
		return
	}

	f.IsDir()
	if f.Err() != nil {
		return
	}

	if f.id == "" && f.ParentFolderId == "" {
		if *f.isDir {
			folderData, err := fddb.importDirectory(f.absolutePath)
			if err != nil {
				f.err = fmt.Errorf("failure to get wl file descriptor by path (directory): %s", err)
				return
			}

			f.id = folderData.FolderId
			f.ParentFolderId = folderData.ParentId
		} else {
			parentPath := filepath.Dir(f.absolutePath)
			parent, err := fddb.getFolderByPath(parentPath)
			if err != nil {
				_, err := fddb.importDirectory(parentPath)
				util.FailOnError(err, "Failure to get wl file descriptor by path")
			}
			parentId := parent.FolderId
			f.ParentFolderId = parentId

			m, _ := f.GetMedia()
			// util.Warn(err)
			f.id = m.FileHash
		}
	}
}

func (f *WeblensFileDescriptor) Copy() *WeblensFileDescriptor {
	c := *f
	return &c
}

// Retrieve the error field set internally in
// the file descriptor. If multiple errors have
// occurred, only the most recent will be shown.
// A non-nil err will only ever be set to nil
// if .ClrErr() is called
func (f *WeblensFileDescriptor) Err() error {
	return f.err
}

// Sets the error field internally to nil,
// and returns the error value that was set.
// Use this call sparingly, chances are there
// is a better way to do what you want than ignoring
// errors.
func (f *WeblensFileDescriptor) ClrErr() error {
	err := f.err
	f.err = nil
	return err
}

// This function does not return an error, and instead
// will set the err feild internally. To check this error,
// call .Err() afterwards.
func (f *WeblensFileDescriptor) Id() string {
	if f.id != "" {
		return f.id
	}
	if f.isDir == nil {
		f.err = fmt.Errorf("cannot get Id without knowing self type (isDir = nil)")
	}
	if *f.isDir {
		// Check if folder is already in the database
		folderData, err := fddb.importDirectory(f.absolutePath)
		if err != nil {
			f.err = fmt.Errorf("failed to import directory: %s", err)
		}
		f.id = folderData.FolderId
	} else if f.IsDisplayable() {
		m, err := f.GetMedia()
		if err != nil {
			f.err = fmt.Errorf("failed to get media: %s", err)
		}
		f.id = m.FileHash
	} else {
		f.id = util.HashOfString(8, f.absolutePath)
	}

	return f.id
}

// Returns string of absolute path to file
func (f *WeblensFileDescriptor) String() string {
	if f.IsDir() && !strings.HasSuffix(f.absolutePath, "/") {
		f.absolutePath = f.absolutePath + "/"
	}
	return f.absolutePath
}

func (f *WeblensFileDescriptor) Owner() string {
	if f.owner == "" {
		f.owner = GetOwnerFromFilepath(f.absolutePath)
	}
	return f.owner
}

func (f *WeblensFileDescriptor) GetMedia() (Media, error) {
	if f.isDir != nil && *f.isDir {
		return Media{}, fmt.Errorf("cannot get media of directory")
	}
	m, err := fddb.getMediaByPath(f.ParentFolderId, f.Filename)
	if err != nil {
		return Media{}, fmt.Errorf("failed to get media from WFD: %s", err)
	}

	if f.isDir != nil {
		f.isDir = boolPointer(false)
	}

	return m, nil
}

func (f *WeblensFileDescriptor) Exists() bool {
	_, err := os.Stat(f.absolutePath)
	return err == nil
}

func (f *WeblensFileDescriptor) IsDir() bool {
	if f.isDir == nil {
		stat, err := os.Stat(f.absolutePath)
		if err != nil {
			f.err = fmt.Errorf("failed to get stat of filepath checking if IsDir: %s", err)
			return false
		}
		f.isDir = boolPointer(stat.IsDir())
	}
	return *f.isDir
}

func (f *WeblensFileDescriptor) ModTime() time.Time {
	stat, _ := os.Stat(f.absolutePath)
	return stat.ModTime()
}

func (f *WeblensFileDescriptor) Size() int64 {
	var size int64
	if f.isDir == nil {
		err := fmt.Errorf("directory status not initialized")
		util.FailOnError(err, "")
	}
	if *f.isDir {
		var err error
		size, err = util.DirSize(f.absolutePath)
		util.FailOnError(err, "Failed to get dir size")
	} else {
		stat, err := os.Stat(f.absolutePath)
		util.FailOnError(err, "Failed to get file stats")

		size = stat.Size()
	}

	return size
}

func (f *WeblensFileDescriptor) JoinStr(extentions ...string) *WeblensFileDescriptor {
	extentions = append([]string{f.absolutePath}, extentions...)
	newFile := WFDByPath(filepath.Join(extentions...))
	return newFile
}

func (f *WeblensFileDescriptor) Read() (*os.File, error) {
	if *f.isDir {
		return nil, fmt.Errorf("attempt to read directory as file")
	}
	osFile, err := os.Open(f.absolutePath)
	return osFile, err
}

func (f *WeblensFileDescriptor) Write(data []byte) error {
	if *f.isDir {
		return fmt.Errorf("attempt to write to directory")
	}
	err := os.WriteFile(f.absolutePath, data, 0660)
	return err
}

func (f *WeblensFileDescriptor) ReadDir() ([]*WeblensFileDescriptor, error) {
	if (f.isDir != nil && !*f.isDir) || !f.Exists() {
		return nil, fmt.Errorf("invalid file to read dir")
	}
	entries, _ := os.ReadDir(f.absolutePath)

	var ret []*WeblensFileDescriptor
	for _, file := range entries {
		ff := f.JoinStr(file.Name())
		if ff.Err() != nil {
			return nil, fmt.Errorf("failed to get subfile info for %s: %s", file.Name(), ff.Err())
		}
		ret = append(ret, ff)
	}
	if f.isDir == nil {
		f.isDir = boolPointer(true)
	}
	return ret, nil
}

func (f *WeblensFileDescriptor) GetParent() *WeblensFileDescriptor {
	parent := &WeblensFileDescriptor{}

	if f.id == "0" {
		parent.err = fmt.Errorf("cannot get parent of media root")
		return parent
	}
	if f.ParentFolderId == "" {
		parent.err = fmt.Errorf("file descriptor has no parent id")
		return parent
	}
	parent = WFDByFolderId(f.ParentFolderId)
	return parent
}

func (f *WeblensFileDescriptor) CreateSelf() error {
	if f.Exists() {
		return fmt.Errorf("directory already exists")
	}
	if f.isDir == nil {
		return fmt.Errorf("cannot create self with nil self type")
	}

	var err error
	if *f.isDir {
		err = os.Mkdir(f.absolutePath, os.FileMode(0777))
		if err != nil {
			return err
		}
	} else {
		var osFile *os.File
		osFile, err = os.Create(f.absolutePath)
		if err != nil {
			return err
		}
		osFile.Close()
	}
	return nil
}

func (f *WeblensFileDescriptor) MoveTo(destination *WeblensFileDescriptor, tasker func(string, map[string]any), opts... *WFDMoveOptions) error {
	if !f.Exists() {
		return fmt.Errorf("file does not exist")
	}
	if destination.Exists() {
		return fmt.Errorf("destination already exists")
	}

	opt := mergeMoveOpts(opts...)

	err := os.Rename(f.absolutePath, destination.absolutePath)
	if err != nil {
		return err
	}

	if destination.IsDir() {
		// Remove old directory from database
		err := fddb.deleteDirectory(f.Id())
		if err != nil {
			return err
		}

	} else if f.IsDisplayable() && destination.IsDisplayable() && !*opt.SkipMediaMove {
		err := fddb.HandleMediaMove(f, destination)
		if err != nil {
			return err
		}
	} else if f.IsDisplayable() && !destination.IsDisplayable() {
		fddb.RemoveMediaByFilepath(f.ParentFolderId, f.Filename)
	} else if !f.IsDisplayable() && destination.IsDisplayable() {
		tasker("scan_file", map[string]any{"file": destination, "username": destination.Owner()})
	}

	f.ParentFolderId = destination.ParentFolderId

	if !*opt.SkipIdRecompute {
		f.Id()
	}

	*f = *destination

	return f.Err()
}

func (f *WeblensFileDescriptor) MoveUnder(newParent *WeblensFileDescriptor, overwrite bool) error {
	destination := newParent.CreateGhostChild(f.Filename)

	if !f.Exists() || !newParent.Exists() {
		return fmt.Errorf("file does not exist")
	}

	if !*newParent.isDir {
		return fmt.Errorf("new parent is not a directory")
	}

	if destination.Exists() && !overwrite {
		return fmt.Errorf("destination file already exists")
	}

	err := os.Rename(f.String(), destination.String())
	if err != nil {
		return err
	}

	destination.init(&WFDCreateOptions{})

	*f = *destination
	return nil
}

func (f *WeblensFileDescriptor) MoveToTrash() error {
	if f.IsDir() {
		fddb.deleteDirectory(f.Id())
		fddb.deleteMediaByFolder(f.Id())
	} else if f.IsDisplayable() {
		err := fddb.RemoveMediaByFilepath(f.ParentFolderId, f.Filename)
		if err != nil {
			return err
		}
	}
	err := f.MoveUnder(&trashRoot, false)
	if err != nil {
		return err
	}
	err = f.Rename(f.Filename + time.Now().Format(".2006-01-02T15.04.05"))
	return err
}

func (f *WeblensFileDescriptor) Rename(newName string) error {
	if !f.Exists() {
		return fmt.Errorf("cannot rename file that does not exist")
	}

	oldAbsPath := f.absolutePath
	f.Filename = newName
	f.absolutePath = filepath.Join(filepath.Dir(oldAbsPath), f.Filename)
	err := os.Rename(oldAbsPath, f.absolutePath)
	if err != nil {
		return err
	}

	return nil
}

func (f *WeblensFileDescriptor) UserCanAccess(username string) bool {
	if (f.owner == username) {
		return true
	}
	return fddb.CanUserAccess(f, username)
}

////

func GetUserHomeDir(username string) *WeblensFileDescriptor {
	file := WFDByPath(filepath.Join(util.GetMediaRoot(), username))
	util.FailOnError(file.Err(), "Failed to get user home directory")
	return file
}

func NewTempFile() *WeblensFileDescriptor {
	tmpFileName := uuid.New().String()
	tmpFile := tmpRoot.CreateGhostChild(tmpFileName)
	return tmpFile
}

// GetTmpFile returns the WFD to the file in the tmp directory
// with the filename as the given UUID4, or an error if it does
// not exist
func GetTmpFile(tmpFileId string) (*WeblensFileDescriptor, error) {
	tmpFile := tmpRoot.JoinStr(tmpFileId)
	if tmpFile.Err() != nil {
		return nil, tmpFile.Err()
	}
	return tmpFile, nil
}

// GetTakeoutFile returns the WFD to the file in the takeout directory
// with the filename as passed as `takeoutId`.zip, or an error if it does
// not exist
func GetTakeoutFile(takeoutId string) *WeblensFileDescriptor {
	tmpFile := takeoutRoot.JoinStr(takeoutId + ".zip")
	return tmpFile
}