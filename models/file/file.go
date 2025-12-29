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

	"github.com/ethanrous/weblens/modules/errors"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/option"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
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

// ErrFileNotFound is returned when a requested file cannot be found.
var ErrFileNotFound = errors.New("file not found")

// ErrNoFileID is returned when a file operation requires an ID but none is present.
var ErrNoFileID = errors.New("file has no id")

// ErrNilFile is returned when a file pointer is unexpectedly nil.
var ErrNilFile = errors.New("file is nil")

// ErrNoContentID is returned when a file operation requires a content ID but none is present.
var ErrNoContentID = errors.New("file has no content id")

// ErrFileTreeNotFound is returned when a requested file tree cannot be found.
var ErrFileTreeNotFound = errors.New("file tree not found")

// ErrEmptyFile is returned when an operation cannot be performed on an empty file.
var ErrEmptyFile = errors.New("file is empty")

// ErrFileAlreadyHasTask is returned when attempting to assign a task to a file that already has one.
var ErrFileAlreadyHasTask = errors.New("file already has task")

// ErrFileNoTask is returned when a file operation requires a task but none is present.
var ErrFileNoTask = errors.New("file has no task")

// ErrDirectoryNotAllowed is returned when a directory is provided but not allowed for the operation.
var ErrDirectoryNotAllowed = errors.New("directory not allowed")

// ErrDirectoryRequired is returned when an operation requires a directory but a regular file was provided.
var ErrDirectoryRequired = errors.New("directory required")

// ErrNoChildren is returned when a directory operation requires children but the directory has none.
var ErrNoChildren = errors.New("directory has no children")

// ErrNoParent is returned when a file operation requires a parent but none is present.
var ErrNoParent = errors.New("file has no parent")

// ErrNotChild is returned when a file is not a child of the specified directory.
var ErrNotChild = errors.New("file is not a child of given directory")

// ErrDirectoryAlreadyExists is returned when attempting to create a directory that already exists.
var ErrDirectoryAlreadyExists = errors.New("directory already exists")

// ErrFileAlreadyExists is returned when attempting to create a file that already exists.
var ErrFileAlreadyExists = errors.New("file already exists")

// WeblensFileImpl implements the http.File interface.
var _ http.File = (*WeblensFileImpl)(nil)

// WeblensFileImpl represents a file in the Weblens system with metadata and operations.
type WeblensFileImpl struct {
	id string

	// The most recent time that this file was changes on the real filesystem
	modifyDate time.Time

	// is the real file on disk a directory or regular file
	isDir option.Option[bool]

	// Pointer to the directory that this file belongs
	parent *WeblensFileImpl

	childrenMap map[string]*WeblensFileImpl

	// The portable filepath of the file. This path can be safely translated between
	// systems with trees using the same root alias
	portablePath file_system.Filepath

	contentID string

	// the id of the file in the past, if a new file is occupying the same path as this file
	pastID string

	childIDs []string

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

// Freeze returns a shallow copy of the file descriptor.
func (f *WeblensFileImpl) Freeze() *WeblensFileImpl {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	newFile := &WeblensFileImpl{
		modifyDate:   f.modifyDate,
		buffer:       f.buffer,
		childIDs:     f.childIDs,
		childrenMap:  f.childrenMap,
		contentID:    f.contentID,
		id:           f.id,
		isDir:        f.isDir,
		memOnly:      f.memOnly,
		parent:       f.parent,
		pastFile:     f.pastFile,
		pastID:       f.pastID,
		portablePath: f.portablePath,
		readOnly:     f.readOnly,
		watching:     f.watching,
		writeHead:    f.writeHead,
	}

	newFile.size.Store(f.size.Load())

	return newFile
}

// ID returns the unique identifier the file, which is the
// portable path of the file.
func (f *WeblensFileImpl) ID() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	return f.id
}

// SetID sets the unique identifier for the file.
func (f *WeblensFileImpl) SetID(id string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.id = id
}

// Name returns the filename of the file.
func (f *WeblensFileImpl) Name() string {
	return f.portablePath.Filename()
}

// Mode returns the file mode bits for the file.
func (f *WeblensFileImpl) Mode() os.FileMode {
	return 0
}

// Sys returns the underlying system-specific data for the file.
func (f *WeblensFileImpl) Sys() any {
	return nil
}

// GetPortablePath returns the portable filepath that can be safely translated between systems.
func (f *WeblensFileImpl) GetPortablePath() file_system.Filepath {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	return f.portablePath
}

// SetPortablePath sets the portable filepath for the file.
func (f *WeblensFileImpl) SetPortablePath(path file_system.Filepath) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.portablePath = path
}

// GetPastID returns the ID of the file in the past if a new file is occupying the same path.
func (f *WeblensFileImpl) GetPastID() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	return f.pastID
}

// Exists check if the file exists on the real filesystem below.
func (f *WeblensFileImpl) Exists() bool {
	if f.memOnly {
		return false
	}

	_, err := os.Stat(f.portablePath.ToAbsolute())

	return err == nil
}

// IsDir returns true if the file is a directory.
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

// ModTime returns the last modification time of the file.
func (f *WeblensFileImpl) ModTime() (t time.Time) {
	if f.pastFile {
		return f.modifyDate
	}

	if f.modifyDate.Unix() <= 0 {
		_, err := f.LoadStat()
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}

	return f.modifyDate
}

// Stat returns file information for the file.
func (f *WeblensFileImpl) Stat() (fs.FileInfo, error) {
	return f, nil
}

// Size returns the size of the file in bytes.
func (f *WeblensFileImpl) Size() int64 {
	if f.pastFile {
		return f.size.Load()
	}

	if f.size.Load() == -1 {
		_, err := f.LoadStat()
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}

	if f.IsDir() {
		f.size.Store(f.dirComputeSize())
	}

	return f.size.Load()
}

// SetSize sets the size of the file in bytes.
func (f *WeblensFileImpl) SetSize(newSize int64) {
	f.size.Store(newSize)
}

// SetMemOnly sets whether the file should only be stored in memory.
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

// Close closes the file.
func (f *WeblensFileImpl) Close() error {
	return nil
}

// Readdir reads and returns up to count directory entries.
func (f *WeblensFileImpl) Readdir(count int) ([]fs.FileInfo, error) {
	children := make([]fs.FileInfo, 0, len(f.childrenMap))
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

// Seek sets the offset for the next read or write operation.
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

// Readable returns a reader for the file contents.
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

// Writer returns a writer for the file contents.
func (f *WeblensFileImpl) Writer() (io.WriteCloser, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	return os.OpenFile(f.portablePath.ToAbsolute(), os.O_CREATE|os.O_WRONLY, os.ModePerm)
}

// ReadAll reads and returns the entire contents of the file.
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

	defer osFile.Close() //nolint:errcheck

	data, err := io.ReadAll(osFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (f *WeblensFileImpl) Write(data []byte) (int, error) {
	if f.IsDir() {
		return 0, errors.WithStack(ErrDirectoryNotAllowed)
	}

	if f.memOnly {
		f.buffer = bytes.Clone(data)

		return len(data), nil
	}

	err := os.WriteFile(f.portablePath.ToAbsolute(), data, os.ModePerm)
	if err == nil {
		f.size.Store(int64(len(data)))
		f.setModifyDate(time.Now())
	}

	return len(data), errors.WithStack(err)
}

// WriteAt writes data to the file at the specified offset.
func (f *WeblensFileImpl) WriteAt(data []byte, seekLoc int64) error {
	if f.IsDir() {
		return ErrDirectoryNotAllowed
	}

	if f.memOnly {
		requiredSize := seekLoc + int64(len(data))
		if requiredSize > int64(len(f.buffer)) {
			newBuffer := make([]byte, 0, requiredSize)
			copy(newBuffer, f.buffer)
			f.buffer = newBuffer
		}

		copy(f.buffer[seekLoc:], data)
		f.size.Store(requiredSize)

		return nil
	}

	realFile, err := os.OpenFile(f.portablePath.ToAbsolute(), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer realFile.Close() //nolint:errcheck

	wroteLen, err := realFile.WriteAt(data, seekLoc)
	if err == nil {
		f.size.Add(int64(wroteLen))
		f.setModifyDate(time.Now())
	}

	return err
}

// Append appends data to the end of the file.
func (f *WeblensFileImpl) Append(data []byte) error {
	if f.IsDir() {
		return ErrDirectoryNotAllowed
	}

	if f.memOnly {
		f.buffer = append(f.buffer, data...)

		return nil
	}

	realFile, err := os.OpenFile(f.portablePath.ToAbsolute(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
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

// GetChild returns the child file with the specified name.
func (f *WeblensFileImpl) GetChild(childName string) (*WeblensFileImpl, error) {
	f.childLock.RLock()
	defer f.childLock.RUnlock()

	if len(f.childrenMap) == 0 || childName == "" {
		return nil, errors.Errorf("%w: file %s [%p] has no children", ErrFileNotFound, f.portablePath.String(), f)
	}

	child := f.childrenMap[childName]
	if child == nil {
		return nil, errors.Errorf("%w: %s is not a child of %s", ErrFileNotFound, childName, f.portablePath.String())
	}

	return child, nil
}

// ChildrenLoaded returns true if the directory's children have been loaded into memory.
func (f *WeblensFileImpl) ChildrenLoaded() bool {
	f.childLock.RLock()
	defer f.childLock.RUnlock()

	return f.childrenMap != nil
}

// GetChildren returns all child files in the directory.
func (f *WeblensFileImpl) GetChildren() []*WeblensFileImpl {
	if !f.IsDir() {
		return []*WeblensFileImpl{}
	}

	f.childLock.RLock()
	defer f.childLock.RUnlock()

	return slices.Collect(maps.Values(f.childrenMap))
}

// InitChildren initializes the children map for a directory.
func (f *WeblensFileImpl) InitChildren() {
	if f.childrenMap == nil {
		f.childrenMap = make(map[string]*WeblensFileImpl)
	}
}

// AddChild adds a child file to the directory.
func (f *WeblensFileImpl) AddChild(child *WeblensFileImpl) error {
	if !f.IsDir() {
		return errors.WithStack(ErrDirectoryRequired)
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()

	if f.childrenMap == nil {
		f.childrenMap = make(map[string]*WeblensFileImpl)
	}

	if f.childrenMap[child.portablePath.Filename()] != nil {
		return errors.Errorf("failed to add %s as child of %s: %w", child.GetPortablePath(), f.GetPortablePath().String(), ErrFileAlreadyExists)
	}

	if child.GetPortablePath().Dir() == child.GetPortablePath() {
		return errors.Errorf("Cannot add %s as a child because it is a root folder", child.GetPortablePath())
	}

	if child.GetPortablePath().Dir() != f.GetPortablePath() {
		return errors.Errorf("Cannot make %s a child of %s: %w", child.GetPortablePath().Dir(), f.GetPortablePath(), ErrNotChild)
	}

	f.childrenMap[child.portablePath.Filename()] = child

	return nil
}

// RemoveChild removes a child file from the directory by name.
func (f *WeblensFileImpl) RemoveChild(child string) error {
	if len(f.childrenMap) == 0 {
		return errors.WithStack(ErrNoChildren)
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()

	if _, ok := f.childrenMap[child]; !ok {
		return errors.WithStack(ErrFileNotFound)
	}

	delete(f.childrenMap, child)

	f.modifiedNow()

	return nil
}

// SetParent sets the parent directory for the file.
func (f *WeblensFileImpl) SetParent(p *WeblensFileImpl) error {
	if f.GetPortablePath().Dir() != p.GetPortablePath() {
		return errors.Wrapf(ErrNotChild, "%s is not the parent of %s", p.GetPortablePath(), f.GetPortablePath())
	}

	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.parent = p

	return nil
}

// GetParent returns the parent directory of the file.
func (f *WeblensFileImpl) GetParent() *WeblensFileImpl {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	return f.parent
}

// SetModifiedTime sets the last modification time of the file.
func (f *WeblensFileImpl) SetModifiedTime(ts time.Time) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.modifyDate = ts
}

// CreateSelf creates the file or directory on the filesystem.
func (f *WeblensFileImpl) CreateSelf() error {
	var err error
	if f.IsDir() {
		err = os.Mkdir(f.portablePath.ToAbsolute(), os.ModePerm)
	} else {
		var newF *os.File

		newF, err = os.Create(f.portablePath.ToAbsolute())

		if newF != nil {
			defer newF.Close() //nolint:errcheck
		}
	}

	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errors.Errorf("failed to create file %s: %w", f.portablePath.String(), ErrFileAlreadyExists)
		}

		return errors.WithStack(err)
	}

	return nil
}

// Remove deletes the file or directory from the filesystem.
func (f *WeblensFileImpl) Remove() error {
	if f.IsDir() {
		return os.RemoveAll(f.portablePath.ToAbsolute())
	}

	return os.Remove(f.portablePath.ToAbsolute())
}

// GetContentID returns the content hash of the file.
func (f *WeblensFileImpl) GetContentID() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	return f.contentID
}

// SetContentID sets the content hash of the file.
func (f *WeblensFileImpl) SetContentID(newContentID string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.contentID = newContentID
}

// UnmarshalJSON deserializes JSON data into the WeblensFileImpl structure.
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
	f.modifyDate = time.UnixMilli(int64(data["modifyTimestamp"].(float64)))
	f.contentID = data["contentID"].(string)

	if f.modifyDate.Unix() <= 0 {
		log.Error().Msg("File has invalid mod time")
	}

	f.childIDs = slices_mod.Map(
		slices_mod.Convert[string](data["childrenIDs"].([]any)), func(cID string) string {
			return string(cID)
		},
	)

	return nil
}

// MarshalJSON serializes the WeblensFileImpl into JSON format.
func (f *WeblensFileImpl) MarshalJSON() ([]byte, error) {
	var parentID string
	if f.parent != nil {
		parentID = f.parent.ID()
	}

	if !f.IsDir() && f.Size() != 0 && f.GetContentID() == "" {
		log.Warn().Msgf("File [%s] has no content ID", f.GetPortablePath())
	}

	data := map[string]any{
		"portablePath":    f.portablePath.ToPortable(),
		"size":            f.size.Load(),
		"isDir":           f.IsDir(),
		"modifyTimestamp": f.ModTime().UnixMilli(),
		"parentID":        parentID,
		"childrenIds":     slices_mod.Map(f.GetChildren(), func(c *WeblensFileImpl) string { return c.ID() }),
		"contentID":       f.GetContentID(),
		"pastFile":        f.pastFile,
	}

	if f.ModTime().UnixMilli() < 0 {
		log.Warn().Msgf("File [%s] has invalid mod time trying to marshal", f.GetPortablePath())
	}

	return json.Marshal(data)
}

// RecursiveMap applies function fn to every file recursively.
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
		children := f.GetChildren()
		slices.SortFunc(children, func(a, b *WeblensFileImpl) int {
			return strings.Compare(a.portablePath.Filename(), b.portablePath.Filename())
		})

		for _, c := range children {
			err := c.LeafMap(fn)
			if err != nil {
				return err
			}
		}
	}

	return fn(f)
}

// BubbleMap performs fn on f and all parents of f, ignoring the mediaService root or other static directories.
// Files are acted on in the order of their index number below, starting with the caller, children are never accessed.
//
//	f3 <- Parent of f2
//	|
//	f2 <- Parent of f1
//	|
//	f1 <- Root caller
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

// IsParentOf returns true if f is an ancestor of child.
func (f *WeblensFileImpl) IsParentOf(child *WeblensFileImpl) bool {
	if f.portablePath.RootName() != child.portablePath.RootName() {
		return false
	}

	return strings.HasPrefix(child.portablePath.RelativePath(), f.portablePath.RelativePath())
}

// SetWatching marks the file as being watched for changes.
func (f *WeblensFileImpl) SetWatching() error {
	f.watching = true

	return nil
}

// IsReadOnly returns true if the file is read-only.
func (f *WeblensFileImpl) IsReadOnly() bool {
	return f.readOnly
}

// LoadStat will recompute the size and modify date of the file using os.Stat. If the
// size of the file changes, LoadStat will return the newSize. If the size does not change,
// LoadStat will return -1 for the newSize. To get the current size of the file, use Size() instead.
func (f *WeblensFileImpl) LoadStat() (newSize int64, err error) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	origSize := f.size.Load()

	stat, err := os.Stat(f.portablePath.ToAbsolute())
	if err != nil {
		return -1, errors.WithStack(err)
	}

	f.modifyDate = stat.ModTime()

	if f.IsDir() {
		newSize = f.dirComputeSize()
	} else {
		if origSize > 0 {
			return -1, nil
		}

		newSize = stat.Size()
	}

	if newSize != origSize {
		f.size.Store(newSize)

		return newSize, nil
	}

	return -1, nil
}

// ReplaceRoot changes the root alias of the file's path.
func (f *WeblensFileImpl) ReplaceRoot(newRoot string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.portablePath = f.portablePath.OverwriteRoot(newRoot)
}

// SetPastFile marks the file as a past version.
func (f *WeblensFileImpl) SetPastFile(isPastFile bool) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.pastFile = isPastFile
}

// IsPastFile returns true if the file is a past version.
func (f *WeblensFileImpl) IsPastFile() bool {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	return f.pastFile
}

func (f *WeblensFileImpl) getModifyDate() time.Time {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()

	return f.modifyDate
}

func (f *WeblensFileImpl) dirComputeSize() int64 {
	size := int64(0)

	for _, child := range f.GetChildren() {
		chsz := child.Size()
		if chsz == -1 {
			continue
		}

		size += chsz
	}

	return size
}

func (f *WeblensFileImpl) setModifyDate(newModifyDate time.Time) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.modifyDate = newModifyDate
}

func (f *WeblensFileImpl) setPortable(portable file_system.Filepath) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.portablePath = portable
}

func (f *WeblensFileImpl) setParentInternal(parent *WeblensFileImpl) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.parent = parent
}

func (f *WeblensFileImpl) modifiedNow() {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()

	f.modifyDate = time.Now()
}

// WeblensFile is an interface for file operations.
type WeblensFile interface {
	ID() string
	Write(data []byte) (int, error)
	ReadAll() ([]byte, error)
}
