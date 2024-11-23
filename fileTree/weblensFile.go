package fileTree

import (
	"bytes"
	"encoding/json"
	"errors"
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

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
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

var _ http.File = (*WeblensFileImpl)(nil)

type FileId = string

type WeblensFileImpl struct {
	// the main way to identify a file. A file id is generated via a hash of its relative filepath
	id FileId

	// The absolute path of the real file on disk
	absolutePath string

	// The portable filepath of the file. This path can be safely translated between
	// systems with trees using the same root alias
	portablePath WeblensFilepath

	// Base of the filepath, the actual name of the file.
	filename string

	// size in bytes of the file on the disk
	size atomic.Int64

	// is the real file on disk a directory or regular file
	isDir *bool

	// The most recent time that this file was changes on the real filesystem
	modifyDate time.Time

	writeHead int64

	contentId string

	parentId FileId
	// Pointer to the directory that this file belongs
	parent *WeblensFileImpl

	// If we already have added the file to the watcher
	// See fileWatch.go
	watching bool

	// If this file is a directory, these are the files that are housed by this directory.
	childLock   sync.RWMutex
	childrenMap map[string]*WeblensFileImpl
	childIds    []FileId

	// General RW lock on file updates to prevent data races
	updateLock sync.RWMutex

	// Lock to atomize long file events
	fileLock sync.Mutex

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
	buffer  []byte
}

func NewWeblensFile(id FileId, filename string, parent *WeblensFileImpl, isDir bool) *WeblensFileImpl {
	f := &WeblensFileImpl{
		id:          id,
		childrenMap: make(map[string]*WeblensFileImpl),
		isDir:       &isDir,
		parent:      parent,
		filename:    filename,
	}
	if parent != nil {
		f.parentId = parent.ID()
		f.portablePath = parent.portablePath.Child(filename, isDir)
		if f.parent.memOnly {
			f.memOnly = true
		}
	} else {
		f.portablePath = ParsePortable("USERS:")
	}

	return f
}

// Freeze returns a "deep-enough" copy of the file descriptor. All only-locally-relevant
// fields are copied, however references, except for locks, are the same as the original version
func (f *WeblensFileImpl) Freeze() *WeblensFileImpl {
	// Copy values of wf struct
	c := *f

	// Create unique versions of pointers that are only relevant locally
	if c.isDir != nil {
		boolCopy := *c.isDir
		c.isDir = &boolCopy
	}

	c.childLock = sync.RWMutex{}
	c.updateLock = sync.RWMutex{}

	return &c
}

// ID returns the unique identifier the file, and will compute it on the fly
// if it is not already initialized in the struct.
//
// This function will intentionally panic if trying to get the
// ID of a nil file.
func (f *WeblensFileImpl) ID() FileId {
	id := f.getIdInternal()
	if id == "" {
		log.ErrTrace(werror.Errorf("Tried to ID() file with no Id and path [%s]", f.absolutePath))
		return ""
	}

	return id
}

// Filename returns the filename of the file
func (f *WeblensFileImpl) Filename() string {
	return f.filename
}

func (f *WeblensFileImpl) Name() string {
	return f.filename
}

func (f *WeblensFileImpl) Mode() os.FileMode {
	return 0
}

func (f *WeblensFileImpl) Sys() any {
	return nil
}

// AbsPath returns string of the absolute path to file
func (f *WeblensFileImpl) AbsPath() string {
	if f == nil {
		return ""
	}
	if f.id == "EXTERNAL" {
		return ""
	}

	return f.getAbsPathInternal()
}

func (f *WeblensFileImpl) GetPortablePath() WeblensFilepath {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.portablePath
}

// Exists check if the file exists on the real filesystem below
func (f *WeblensFileImpl) Exists() bool {
	_, err := os.Stat(f.absolutePath)
	return err == nil
}

func (f *WeblensFileImpl) IsDir() bool {
	if f.isDir == nil {
		stat, err := os.Stat(f.absolutePath)
		if err != nil {
			log.ErrTrace(err)
			return false
		}
		f.isDir = boolPointer(stat.IsDir())
	}
	return *f.isDir
}

func (f *WeblensFileImpl) ModTime() (t time.Time) {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	if f.modifyDate.Unix() <= 0 {
		_, err := f.LoadStat()
		if err != nil {
			log.ErrTrace(err)
		}
	}
	return f.modifyDate
}

func (f *WeblensFileImpl) setPastFile(isPastFile bool) {
	f.pastFile = isPastFile
}

func (f *WeblensFileImpl) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *WeblensFileImpl) setModTime(t time.Time) {
	f.modifyDate = t
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

	fp, err := os.Open(f.absolutePath)
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

	path := f.absolutePath
	return os.Open(path)
}

func (f *WeblensFileImpl) Writeable() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	path := f.AbsPath()
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0660)
}

func (f *WeblensFileImpl) ReadAll() ([]byte, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	if f.memOnly {
		return f.buffer, nil
	}

	osFile, err := os.Open(f.absolutePath)
	if err != nil {
		return nil, werror.WithStack(err)
	}
	// fileSize := f.Size()

	data, err := io.ReadAll(osFile)
	if err != nil {
		return nil, werror.WithStack(err)
	}
	// if len(data) != int(fileSize) {
	// 	return nil, werror.WithStack(werror.ErrBadReadCount)
	// }

	return data, nil
}

func (f *WeblensFileImpl) Write(data []byte) (int, error) {
	if f.IsDir() {
		return 0, werror.ErrDirNotAllowed
	}

	if f.memOnly {
		f.buffer = data
		return len(data), nil
	}

	err := os.WriteFile(f.AbsPath(), data, 0660)
	if err == nil {
		f.size.Store(int64(len(data)))
		f.modifyDate = time.Now()
	}
	return len(data), err
}

func (f *WeblensFileImpl) WriteAt(data []byte, seekLoc int64) error {
	if f.IsDir() {
		return werror.ErrDirNotAllowed
	}

	if f.memOnly {
		panic(werror.NotImplemented("memOnly file write at"))
	}

	path := f.AbsPath()
	realFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer func(realFile *os.File) {
		err := realFile.Close()
		if err != nil {
			log.ErrTrace(err)
		}
	}(realFile)

	wroteLen, err := realFile.WriteAt(data, seekLoc)
	if err == nil {
		f.size.Add(int64(wroteLen))
		f.modifyDate = time.Now()
	}

	return err
}

func (f *WeblensFileImpl) Append(data []byte) error {
	if f.IsDir() {
		return werror.ErrDirNotAllowed
	}

	if f.memOnly {
		f.buffer = append(f.buffer, data...)
		return nil
	}

	realFile, err := os.OpenFile(f.AbsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	wroteLen, err := realFile.Write(data)
	if err == nil {
		f.size.Add(int64(wroteLen))
		f.modifyDate = time.Now()
	}
	return err
}

func (f *WeblensFileImpl) GetChild(childName string) (*WeblensFileImpl, error) {
	f.childLock.RLock()
	defer f.childLock.RUnlock()
	if len(f.childrenMap) == 0 || childName == "" {
		return nil, werror.WithStack(werror.NewErrNoFileName(childName))
	}

	child := f.childrenMap[strings.ToLower(childName)]
	if child == nil {
		return nil, werror.WithStack(werror.NewErrNoFileName(childName))
	}

	return child, nil
}

func (f *WeblensFileImpl) GetChildren() []*WeblensFileImpl {
	if !f.IsDir() {
		return []*WeblensFileImpl{}
	}

	f.childLock.RLock()
	defer f.childLock.RUnlock()

	return internal.SliceConvert[*WeblensFileImpl](slices.Collect(maps.Values(f.childrenMap)))
}

func (f *WeblensFileImpl) AddChild(child *WeblensFileImpl) error {
	if !f.IsDir() {
		return werror.WithStack(werror.ErrDirectoryRequired)
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()
	if f.childrenMap == nil {
		f.childrenMap = make(map[string]*WeblensFileImpl)
	}
	f.childrenMap[strings.ToLower(child.Filename())] = child

	return nil
}

func (f *WeblensFileImpl) GetParent() *WeblensFileImpl {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.parent
}

func (f *WeblensFileImpl) GetParentId() FileId {
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
		err = os.Mkdir(f.AbsPath(), 0755)
	} else {
		_, err = os.Create(f.AbsPath())
	}
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return werror.WithStack(werror.ErrFileAlreadyExists)
		}
		return werror.WithStack(err)
	}

	return nil
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

func (f *WeblensFileImpl) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return err
	}

	f.id = FileId(data["id"].(string))
	f.portablePath = ParsePortable(data["portablePath"].(string))
	f.filename = data["filename"].(string)
	f.size.Store(int64(data["size"].(float64)))
	f.isDir = boolPointer(data["isDir"].(bool))
	f.modifyDate = time.UnixMilli(int64(data["modifyTimestamp"].(float64)))
	f.contentId = data["contentId"].(string)
	if f.modifyDate.Unix() <= 0 {
		log.Error.Println("AHHHH")
	}

	parentId := FileId(data["parentId"].(string))
	f.parentId = parentId

	f.childIds = internal.Map(
		internal.SliceConvert[string](data["childrenIds"].([]any)), func(cId string) FileId {
			return FileId(cId)
		},
	)

	return nil
}

func (f *WeblensFileImpl) MarshalJSON() ([]byte, error) {
	var parentId FileId
	if f.parent != nil {
		parentId = f.parent.ID()
	}
	if !f.IsDir() && f.Size() != 0 && f.GetContentId() == "" {
		log.Warning.Printf("File [%s] has no content Id", f.GetPortablePath())
	}

	data := map[string]any{
		"id":              f.id,
		"portablePath":    f.portablePath.ToPortable(),
		"filename":        f.filename,
		"size":            f.size.Load(),
		"isDir":           f.IsDir(),
		"modifyTimestamp": f.ModTime().UnixMilli(),
		"parentId":        parentId,
		"childrenIds":     internal.Map(f.GetChildren(), func(c *WeblensFileImpl) FileId { return c.ID() }),
		"contentId":       f.GetContentId(),
		"pastFile":        f.pastFile,
	}

	if f.ModTime().UnixMilli() < 0 {
		log.Warning.Printf("File [ %s ] has invalid mod time trying to marshal", f.GetPortablePath())
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
		return werror.WithStack(werror.ErrNilFile)
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
	return strings.HasPrefix(child.AbsPath(), f.AbsPath())
}

func (f *WeblensFileImpl) SetWatching() error {
	if f.watching {
		return werror.ErrAlreadyWatching
	}

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
	if f.absolutePath == "" {
		return -1, nil
	}

	origSize := f.size.Load()

	if f.IsDir() {
		for _, child := range f.GetChildren() {
			newSize += child.size.Load()
		}
	} else {
		if origSize > 0 {
			return -1, nil
		}
		stat, err := os.Stat(f.absolutePath)
		if err != nil {
			return -1, werror.WithStack(err)
		}
		f.updateLock.Lock()
		f.modifyDate = stat.ModTime()
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

func (f *WeblensFileImpl) getIdInternal() FileId {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.id
}

func (f *WeblensFileImpl) setIdInternal(id FileId) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.id = id
}

func (f *WeblensFileImpl) setAbsPath(absPath string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.absolutePath = absPath
}

func (f *WeblensFileImpl) setPortable(portable WeblensFilepath) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.portablePath = portable
}

func (f *WeblensFileImpl) getAbsPathInternal() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.absolutePath
}

func (f *WeblensFileImpl) setParentInternal(parent *WeblensFileImpl) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.parentId = parent.ID()
	f.parent = parent
}

func (f *WeblensFileImpl) removeChild(child *WeblensFileImpl) error {
	if len(f.childrenMap) == 0 {
		return werror.WithStack(werror.ErrNoChildren)
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()

	if _, ok := f.childrenMap[strings.ToLower(child.Filename())]; !ok {
		return werror.WithStack(werror.ErrNoFile)
	}

	delete(f.childrenMap, strings.ToLower(child.Filename()))

	f.modifiedNow()

	return nil
}

func (f *WeblensFileImpl) modifiedNow() {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.modifyDate = time.Now()
}

type WeblensFile interface {
	ID() FileId
	Write(data []byte) (int, error)
	ReadAll() ([]byte, error)
}
