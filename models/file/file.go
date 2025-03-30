package file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/option"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

/*
	WeblensFile is an essential part of the weblens logic.
	Using this (and more importantly the interface defined
	in fileTree.go) will make requests to the filesystem on the
	server machine far faster both to write (as the programmer)
	and to execute.

	Using this is required to keep the database, cache, and real filesystem
	in sync.

	NOT using these interfaces will yield slow and destructive
	results when attempting to modify the real filesystem underneath.
*/

var ErrFileNotFound = errors.New("file not found")
var ErrNilFile = errors.New("file is nil")
var ErrNoContentId = errors.New("file has no content id")
var ErrFileTreeNotFound = errors.New("file tree not found")
var ErrEmptyFile = errors.New("file is empty")
var ErrFileAlreadyHasTask = errors.New("file already has task")
var ErrFileNoTask = errors.New("file has no task")

var ErrDirectoryNotAllowed = errors.New("directory not allowed")
var ErrDirectoryRequired = errors.New("directory required")
var ErrNoChildren = errors.New("directory has no children")

var ErrDirectoryAlreadyExists = errors.New("directory already exists")
var ErrFileAlreadyExists = errors.New("file already exists")

// WeblensFileImpl implements the http.File interface
var _ http.File = (*WeblensFileImpl)(nil)

type WeblensFileImpl struct {

	// The most recent time that this file was changes on the real filesystem
	ModifyDate time.Time

	// is the real file on disk a directory or regular file
	isDir option.Option[bool]

	// Pointer to the directory that this file belongs
	parent *WeblensFileImpl

	childrenMap map[string]*WeblensFileImpl

	// The portable filepath of the file. This path can be safely translated between
	// systems with trees using the same root alias
	portablePath file_system.Filepath

	contentId string

	parentId string

	// the id of the file in the past, if a new file is occupying the same path as this file
	pastId string

	childIds []string

	buffer []byte

	// size in bytes of the file on the disk
	size atomic.Int64

	writeHead int64

	// If this file is a directory, these are the files that are housed by this directory.
	childLock sync.RWMutex

	// General RW lock on file updates to prevent data races
	updateLock sync.RWMutex

	// Lock to atomize long file events
	fileLock sync.Mutex

	// If we already have added the file to the watcher
	// See fileWatch.go
	watching bool

	// Mark file as read-only internally.
	// This should be checked before any write action is to be performed.
	// This should not be changed during run-time, only set in InitMediaRoot.
	// If a directory is `readOnly`, all children are as well
	readOnly bool

	// this file represents a file possibly not on the filesystem
	// anymore, but was at some point in the past
	pastFile bool

	// memOnly if the file is meant to only be stored in memory,
	// writing to it will only write to the buffer
	memOnly bool
}

func NewWeblensFile(path file_system.Filepath, parent *WeblensFileImpl) *WeblensFileImpl {
	f := &WeblensFileImpl{
		childrenMap: make(map[string]*WeblensFileImpl),
		isDir:       option.Of(path.IsDir()),
		parent:      parent,
	}
	if parent != nil {
		f.parentId = parent.ID()
		if f.parent.memOnly {
			f.memOnly = true
		}
	} else {
		var err error
		f.portablePath, err = file_system.ParsePortable("USERS:")
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}

	return f
}

// Freeze returns a shallow copy of the file descriptor.
func (f *WeblensFileImpl) Freeze() *WeblensFileImpl {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	newFile := &WeblensFileImpl{
		ModifyDate:   f.ModifyDate,
		isDir:        f.isDir,
		parent:       f.parent,
		childrenMap:  f.childrenMap,
		portablePath: f.portablePath,
		contentId:    f.contentId,
		parentId:     f.parentId,
		pastId:       f.pastId,
		childIds:     f.childIds,
		buffer:       f.buffer,
		writeHead:    f.writeHead,
		watching:     f.watching,
		readOnly:     f.readOnly,
		pastFile:     f.pastFile,
		memOnly:      f.memOnly,
	}

	newFile.size.Store(f.size.Load())

	return newFile
}

// ID returns the unique identifier the file, which is the
// portable path of the file.
func (f *WeblensFileImpl) ID() string {
	return f.portablePath.ToPortable()
}

// Name returns the filename of the file
func (f *WeblensFileImpl) Name() string {
	return f.portablePath.Filename()
}

func (f *WeblensFileImpl) Mode() os.FileMode {
	return 0
}

func (f *WeblensFileImpl) Sys() any {
	return nil
}

func (f *WeblensFileImpl) GetPortablePath() file_system.Filepath {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.portablePath
}

func (f *WeblensFileImpl) GetPastId() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.pastId
}

// Exists check if the file exists on the real filesystem below
func (f *WeblensFileImpl) Exists() bool {
	_, err := os.Stat(f.portablePath.ToAbsolute())
	return err == nil
}

func (f *WeblensFileImpl) IsDir() bool {
	if !f.isDir.Has() {
		stat, err := os.Stat(f.portablePath.ToAbsolute())
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
			return false
		}
		isDir := stat.IsDir()
		f.isDir = option.Of(isDir)
	}
	isDir, _ := f.isDir.Get()
	return isDir
}

func (f *WeblensFileImpl) ModTime() (t time.Time) {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	if f.ModifyDate.Unix() <= 0 {
		_, err := f.LoadStat()
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}
	return f.ModifyDate
}

func (f *WeblensFileImpl) setPastFile(isPastFile bool) {
	f.pastFile = isPastFile
}

func (f *WeblensFileImpl) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *WeblensFileImpl) Size() int64 {
	return f.size.Load()
}

func (f *WeblensFileImpl) SetSize(newSize int64) {
	f.size.Store(newSize)
}

func (f *WeblensFileImpl) SetMemOnly(memOnly bool) {
	f.memOnly = memOnly
}

// WithLock is a quick way to ensure locks on files are lifted if the function
// using the file is to panic.
func (f *WeblensFileImpl) WithLock(fn func() error) error {
	f.fileLock.Lock()
	defer f.fileLock.Unlock()
	return fn()
}

func (f *WeblensFileImpl) Read(p []byte) (n int, err error) {
	if f.memOnly {
		copied, err := io.Copy(bytes.NewBuffer(f.buffer), bytes.NewReader(p))
		return int(copied), err
	}

	fp, err := os.Open(f.portablePath.ToAbsolute())
	if err != nil {
		return 0, err
	}
	return fp.Read(p)
}

func (f *WeblensFileImpl) Close() error {
	return nil
}

func (f *WeblensFileImpl) Readdir(count int) ([]fs.FileInfo, error) {
	var children []fs.FileInfo
	f.childLock.RLock()
	for _, child := range f.childrenMap {
		children = append(children, child)
		if len(children) == count {
			break
		}
	}
	f.childLock.RUnlock()

	return children, nil
}

func (f *WeblensFileImpl) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.writeHead = offset
	case io.SeekCurrent:
		f.writeHead += offset
	case io.SeekEnd:
		f.writeHead = f.size.Load() - offset
	}

	return f.writeHead, nil
}

func (f *WeblensFileImpl) Readable() (io.Reader, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	if f.memOnly {
		return bytes.NewBuffer(f.buffer), nil
	}

	path := f.portablePath.ToAbsolute()
	return os.Open(path)
}

func (f *WeblensFileImpl) Writeable() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	return os.OpenFile(f.portablePath.ToAbsolute(), os.O_CREATE|os.O_WRONLY, 0660)
}

func (f *WeblensFileImpl) ReadAll() ([]byte, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	if f.memOnly {
		return f.buffer, nil
	}

	osFile, err := os.Open(f.portablePath.ToAbsolute())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// fileSize := f.Size()

	data, err := io.ReadAll(osFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// if len(data) != int(fileSize) {
	// 	return nil, errors.WithStack(errors.ErrBadReadCount)
	// }

	return data, nil
}

func (f *WeblensFileImpl) Write(data []byte) (int, error) {
	if f.IsDir() {
		return 0, errors.WithStack(ErrDirectoryNotAllowed)
	}

	if f.memOnly {
		f.buffer = data
		return len(data), nil
	}

	err := os.WriteFile(f.portablePath.ToAbsolute(), data, 0600)
	if err == nil {
		f.size.Store(int64(len(data)))
		f.setModifyDate(time.Now())
	}
	return len(data), errors.WithStack(err)
}

func (f *WeblensFileImpl) WriteAt(data []byte, seekLoc int64) error {
	if f.IsDir() {
		return ErrDirectoryNotAllowed
	}

	if f.memOnly {
		panic(errors.New("memOnly file write at"))
	}

	path := f.portablePath.ToAbsolute()
	realFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer func(realFile *os.File) {
		err := realFile.Close()
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}(realFile)

	wroteLen, err := realFile.WriteAt(data, seekLoc)
	if err == nil {
		f.size.Add(int64(wroteLen))
		f.setModifyDate(time.Now())
	}

	return err
}

func (f *WeblensFileImpl) Append(data []byte) error {
	if f.IsDir() {
		return ErrDirectoryNotAllowed
	}

	if f.memOnly {
		f.buffer = append(f.buffer, data...)
		return nil
	}

	realFile, err := os.OpenFile(f.portablePath.ToAbsolute(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	wroteLen, err := realFile.Write(data)
	if err == nil {
		f.size.Add(int64(wroteLen))
		f.setModifyDate(time.Now())
	}
	return err
}

func (f *WeblensFileImpl) GetChild(childName string) (*WeblensFileImpl, error) {
	f.childLock.RLock()
	defer f.childLock.RUnlock()
	if len(f.childrenMap) == 0 || childName == "" {
		return nil, errors.WithStack(ErrFileNotFound)
	}

	child := f.childrenMap[strings.ToLower(childName)]
	if child == nil {
		return nil, errors.WithStack(ErrFileNotFound)
	}

	return child, nil
}

func (f *WeblensFileImpl) GetChildren() []*WeblensFileImpl {
	if !f.IsDir() {
		return []*WeblensFileImpl{}
	}

	f.childLock.RLock()
	defer f.childLock.RUnlock()

	return slices.Collect(maps.Values(f.childrenMap))
}

func (f *WeblensFileImpl) AddChild(child *WeblensFileImpl) error {
	if !f.IsDir() {
		return errors.WithStack(ErrDirectoryRequired)
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()
	if f.childrenMap == nil {
		f.childrenMap = make(map[string]*WeblensFileImpl)
	}
	f.childrenMap[strings.ToLower(child.portablePath.Filename())] = child

	return nil
}

func (f *WeblensFileImpl) GetParent() *WeblensFileImpl {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.parent
}

func (f *WeblensFileImpl) GetParentId() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	if f.parentId == "" && f.parent != nil {
		f.parentId = f.parent.ID()
	}
	return f.parentId
}

func (f *WeblensFileImpl) CreateSelf() error {
	var err error
	if f.IsDir() {
		err = os.Mkdir(f.portablePath.ToAbsolute(), 0755)
	} else {
		_, err = os.Create(f.portablePath.ToAbsolute())
	}
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errors.WithStack(ErrFileAlreadyExists)
		}
		return errors.WithStack(err)
	}

	return nil
}

func (f *WeblensFileImpl) Remove() error {
	if f.IsDir() {
		return os.RemoveAll(f.portablePath.ToAbsolute())
	}
	return os.Remove(f.portablePath.ToAbsolute())
}

func (f *WeblensFileImpl) GetContentId() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.contentId
}

func (f *WeblensFileImpl) SetContentId(newContentId string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.contentId = newContentId
}

func (f *WeblensFileImpl) getModifyDate() time.Time {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.ModifyDate
}

func (f *WeblensFileImpl) setModifyDate(newModifyDate time.Time) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.ModifyDate = newModifyDate
}

func (f *WeblensFileImpl) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return err
	}

	portable, err := file_system.ParsePortable(data["portablePath"].(string))
	if err != nil {
		return err
	}

	f.portablePath = portable
	f.size.Store(int64(data["size"].(float64)))
	isDir := data["isDir"].(bool)
	f.isDir = option.Of(isDir)
	f.ModifyDate = time.UnixMilli(int64(data["modifyTimestamp"].(float64)))
	f.contentId = data["contentId"].(string)
	if f.ModifyDate.Unix() <= 0 {
		log.Error().Msg("File has invalid mod time")
	}

	parentId := string(data["parentId"].(string))
	f.parentId = parentId

	f.childIds = slices_mod.Map(
		slices_mod.Convert[string](data["childrenIds"].([]any)), func(cId string) string {
			return string(cId)
		},
	)

	return nil
}

func (f *WeblensFileImpl) MarshalJSON() ([]byte, error) {
	var parentId string
	if f.parent != nil {
		parentId = f.parent.ID()
	}
	if !f.IsDir() && f.Size() != 0 && f.GetContentId() == "" {
		log.Warn().Msgf("File [%s] has no content Id", f.GetPortablePath())
	}

	data := map[string]any{
		"portablePath":    f.portablePath.ToPortable(),
		"size":            f.size.Load(),
		"isDir":           f.IsDir(),
		"modifyTimestamp": f.ModTime().UnixMilli(),
		"parentId":        parentId,
		"childrenIds":     slices_mod.Map(f.GetChildren(), func(c *WeblensFileImpl) string { return c.ID() }),
		"contentId":       f.GetContentId(),
		"pastFile":        f.pastFile,
	}

	if f.ModTime().UnixMilli() < 0 {
		log.Warn().Msgf("File [%s] has invalid mod time trying to marshal", f.GetPortablePath())
	}

	return json.Marshal(data)
}

// RecursiveMap applies function fn to every file recursively
func (f *WeblensFileImpl) RecursiveMap(fn func(*WeblensFileImpl) error) error {
	err := fn(f)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return nil
	}

	children := f.GetChildren()

	for _, c := range children {
		if f == c {
			panic("RecursiveMap called on self")
		}
		err := c.RecursiveMap(fn)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
LeafMap recursively perform fn on leaves, first, and work back up the tree.
This takes an inverted "Depth first" approach. Note this
behaves very differently than RecursiveMap. See below.

Files are acted on in the order of their index number here, starting with the leftmost leaf

		fx.LeafMap(fn)
		|
		f5
	   /  \
	  f3  f4
	 /  \
	f1  f2
*/
func (f *WeblensFileImpl) LeafMap(fn func(*WeblensFileImpl) error) error {
	if f == nil {
		return errors.WithStack(ErrNilFile)
	}

	if f.IsDir() {
		for _, c := range f.GetChildren() {
			err := c.LeafMap(fn)
			if err != nil {
				return err
			}
		}
	}

	return fn(f)
}

/*
BubbleMap
Performs fn on f and all parents of f, ignoring the mediaService root or other static directories.

Files are acted on in the order of their index number below, starting with the caller, children are never accessed

	f3 <- Parent of f2
	|
	f2 <- Parent of f1
	|
	f1 <- Root caller
*/
func (f *WeblensFileImpl) BubbleMap(fn func(*WeblensFileImpl) error) error {
	if f == nil {
		return nil
	}
	err := fn(f)
	if err != nil {
		return err
	}

	parent := f.GetParent()
	if parent == nil {
		return nil
	}
	return parent.BubbleMap(fn)
}

func (f *WeblensFileImpl) IsParentOf(child *WeblensFileImpl) bool {
	if f.portablePath.RootName() != child.portablePath.RootName() {
		return false
	}
	return strings.HasPrefix(child.portablePath.RelativePath(), f.portablePath.RelativePath())
}

func (f *WeblensFileImpl) SetWatching() error {
	// if f.watching {
	// 	return errors.ErrAlreadyWatching
	// }

	f.watching = true
	return nil
}

func (f *WeblensFileImpl) IsReadOnly() bool {
	return f.readOnly
}

// LoadStat will recompute the size and modify date of the file using os.Stat. If the
// size of the file changes, LoadStat will return the newSize. If the size does not change,
// LoadStat will return -1 for the newSize. To get the current size of the file, use Size() instead.
func (f *WeblensFileImpl) LoadStat() (newSize int64, err error) {
	origSize := f.size.Load()

	if f.IsDir() {
		for _, child := range f.GetChildren() {
			newSize += child.size.Load()
		}
	} else {
		if origSize > 0 {
			return -1, nil
		}
		stat, err := os.Stat(f.portablePath.ToAbsolute())
		if err != nil {
			return -1, errors.WithStack(err)
		}
		f.updateLock.Lock()
		f.ModifyDate = stat.ModTime()
		f.updateLock.Unlock()

		newSize = stat.Size()
	}

	if newSize != origSize {
		f.size.Store(newSize)
		return newSize, nil
	}
	return -1, nil
}

func (f *WeblensFileImpl) ReplaceRoot(newRoot string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.portablePath = f.portablePath.OverwriteRoot(newRoot)
}

func (f *WeblensFileImpl) setPortable(portable file_system.Filepath) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.portablePath = portable
}

func (f *WeblensFileImpl) setParentInternal(parent *WeblensFileImpl) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.parentId = parent.ID()
	f.parent = parent
}

func (f *WeblensFileImpl) removeChild(child *WeblensFileImpl) error {
	if len(f.childrenMap) == 0 {
		return errors.WithStack(ErrNoChildren)
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()

	if _, ok := f.childrenMap[strings.ToLower(child.portablePath.Filename())]; !ok {
		return errors.WithStack(ErrFileNotFound)
	}

	delete(f.childrenMap, strings.ToLower(child.portablePath.Filename()))

	f.modifiedNow()

	return nil
}

func (f *WeblensFileImpl) modifiedNow() {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.ModifyDate = time.Now()
}

type WeblensFile interface {
	ID() string
	Write(data []byte) (int, error)
	ReadAll() ([]byte, error)
}
