package fileTree

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
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

type FileId string

type WeblensFileImpl struct {
	// the main way to identify a file. A file id is generated via a hash of its relative filepath
	id FileId

	// The absolute path of the real file on disk
	absolutePath string

	// The portable filepath of the file. This path can be safely translated between
	// systems with trees using the same root alias
	portablePath WeblensFilepath

	// Path of the content file on a backup server
	backupPath string

	// Base of the filepath, the actual name of the file.
	filename string

	// size in bytes of the file on the disk
	size atomic.Int64

	// is the real file on disk a directory or regular file
	isDir *bool

	// The most recent time that this file was changes on the real filesystem
	modifyDate time.Time

	contentId string

	parentId FileId
	// Pointer to the directory that this file belongs
	parent *WeblensFileImpl

	// If we already have added the file to the watcher
	// See fileWatch.go
	watching bool

	// If this file is a directory, these are the files that are housed by this directory.
	childLock sync.RWMutex
	childrenMap map[string]*WeblensFileImpl
	childIds []FileId

	// General RW lock on file updates to prevent data races
	updateLock sync.RWMutex

	// Mark file as read-only internally.
	// This should be checked before any write action is to be taken
	// this should not be changed during run-time, only set in InitMediaRoot.
	// If a directory is `readOnly`, all children are as well
	readOnly bool

	// this file represents a file possibly not on the filesystem
	// anymore, but was at some point in the past
	pastFile bool

	// If the file is in the trash, or a past file, this current fileId
	// is the location of the content right now, not in the past.
	currentId FileId
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

// GetAbsPath returns string of the absolute path to file
func (f *WeblensFileImpl) GetAbsPath() string {
	if f == nil {
		return ""
	}
	if f.id == "EXTERNAL" {
		return ""
	}

	return f.getAbsPathInternal()

	// if backup := f.getBackupPathInternal(); backup != "" {
	// 	return backup
	// }
	//
	// if f.id == "ROOT" {
	// 	f.setAbsPath(internal.GetMediaRootPath())
	// 	return f.getAbsPathInternal()
	// }
	//
	// if InstanceService.GetLocal().IsCore() || f.Owner().IsSystemUser() {
	// 	// If this is a core server, attach filename to the and of the parent directory path
	// 	if f.getAbsPathInternal() == "" {
	// 		f.setAbsPath(filepath.Join(f.parent.GetAbsPath(), f.filename))
	// 	}
	//
	// 	// Directories must and with a "/"
	// 	if f.IsDir() && f.getAbsPathInternal()[len(f.getAbsPathInternal())-1:] != "/" {
	// 		f.setAbsPath(f.getAbsPathInternal() + "/")
	// 	}
	// } else {
	// 	// If this is a backup server, we use the backup path for the "real" path
	// 	f.setBackupPath(filepath.Join(f.tree.Get("ROOT").GetAbsPath(), string(f.GetContentId())))
	// 	return f.getBackupPathInternal()
	// }
	// return f.getAbsPathInternal()
}

func (f *WeblensFileImpl) GetPortablePath() WeblensFilepath {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.portablePath
}

// Owner returns the user that owns the file
// func (f *WeblensFileImpl) Owner() *weblens.User {
// 	if f == nil {
// 		panic("attempt to get owner on nil wf")
// 	}
// 	if f.owner == nil {
// 		// Media root has itself as its parent, so we use GetParent to turn *WeblensFileImpl to *WeblensFileImpl
// 		if f.GetParent() == f.tree.GetRoot() {
// 			f.owner = f.tree.userService.Get(weblens.Username(f.Filename()))
// 			if string(f.owner.GetUsername()) != f.Filename() {
// 				panic(errors.New("I don't even know man... look at Owner() on WeblensFileImpl"))
// 			}
// 		} else {
// 			wlog.Debug.Println("ABS PATH", f.GetAbsPath())
// 			f.owner = f.GetParent().Owner()
// 		}
// 	}
// 	return f.owner
// }

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
	if f.modifyDate.Unix() <= 0 {
		_, err := f.LoadStat()
		if err != nil {
			log.ErrTrace(err)
		}
	}
	return f.modifyDate
}

func (f *WeblensFileImpl) setModTime(t time.Time) {
	f.modifyDate = t
}

// func (f *WeblensFileImpl) recomputeSize() (int64, error) {
// 	if f.ID() == "EXTERNAL" {
// 		var size int64
// 		internal.Map(f.GetChildren(), func(c *WeblensFileImpl) int { sz, _ := c.Size(); size += sz; return 0 })
// 		f.size.Store(size)
// 		return f.size.Load(), nil
// 	}
//
// 	if f.IsDir() {
// 		newSize := int64(0)
// 		for _, c := range f.GetChildren() {
// 			cs, err := c.Size()
// 			if err != nil {
// 				return 0, err
// 			}
// 			newSize += cs
// 		}
// 		f.size.Store(newSize)
// 	} else {
// 		_, err := f.LoadStat()
// 		if err != nil {
// 			return f.size.Load(), werror.WithStack(err)
// 		}
// 	}
//
// 	return int64(f.size.Load()), nil
// }

func (f *WeblensFileImpl) Size() (int64, error) {
	// if f.size.Load() <= 0 {
	// 	return f.recomputeSize()
	// }

	return int64(f.size.Load()), nil
}

func (f *WeblensFileImpl) Readable() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	path := f.absolutePath
	return os.Open(path)
}

func (f *WeblensFileImpl) Writeable() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	path := f.GetAbsPath()
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0660)
}

func (f *WeblensFileImpl) ReadAll() ([]byte, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}
	osFile, err := os.Open(f.absolutePath)
	if err != nil {
		return nil, werror.WithStack(err)
	}
	fileSize, err := f.Size()
	if err != nil {
		return nil, werror.WithStack(err)
	}
	data, err := internal.OracleReader(osFile, fileSize)
	if len(data) != int(fileSize) {
		return nil, werror.WithStack(werror.ErrBadReadCount)
	}

	return data, nil
}

func (f *WeblensFileImpl) Write(data []byte) error {
	if f.IsDir() {
		return werror.ErrDirNotAllowed
	}
	err := os.WriteFile(f.GetAbsPath(), data, 0660)
	if err == nil {
		f.size.Store(int64(len(data)))
		f.modifyDate = time.Now()
	}
	return err
}

func (f *WeblensFileImpl) WriteAt(data []byte, seekLoc int64) error {
	if f.IsDir() {
		return werror.ErrDirNotAllowed
	}

	path := f.GetAbsPath()
	realFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer func(realFile *os.File) {
		err := realFile.Close()
		if err != nil {

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
	realFile, err := os.OpenFile(f.GetAbsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
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

// func (f *WeblensFileImpl) ReadDir() ([]*WeblensFileImpl, error) {
// 	if !f.IsDir() {
// 		return nil, fmt.Errorf("cannot read dir of regular file")
// 	}
//
// 	if len(f.childrenMap) != 0 {
// 		return f.GetChildren(), nil
// 	}
//
// 	entries, err := f.tree.db.ReadDir(f)
//
// 	if err != nil {
// 		return nil, err
// 	}
// 	var children []*WeblensFileImpl
// 	for _, file := range entries {
//
// 		if f == f.tree.GetRoot() {
// 			// u = weblens.Username(file.Name)
// 			wlog.Error.Println("UH OH")
// 		} else {
// 			u = f.Owner()
// 		}
//
// 		singleChild := f.tree.NewFile(f, file.Name, file.IsDir, u)
// 		if file.Size > 0 && !singleChild.IsDir() {
// 			singleChild.(*WeblensFileImpl).size.Store(file.Size)
// 		}
//
// 		f.childLock.Lock()
// 		children = append(children, singleChild)
// 		f.childLock.Unlock()
// 	}
//
// 	return children, nil
// }

func (f *WeblensFileImpl) GetChild(childName string) (*WeblensFileImpl, error) {
	f.childLock.RLock()
	defer f.childLock.RUnlock()
	if len(f.childrenMap) == 0 || childName == "" {
		return nil, werror.WithStack(werror.ErrNoFileName(childName))
	}

	child := f.childrenMap[childName]
	if child == nil {
		return nil, werror.WithStack(werror.ErrNoFileName(childName))
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
	f.childrenMap[child.Filename()] = child

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
	return f.parentId
}

func (f *WeblensFileImpl) CreateSelf() error {
	var err error
	if f.IsDir() {
		err = os.Mkdir(f.GetAbsPath(), 0755)
	} else {
		_, err = os.Create(f.GetAbsPath())
	}
	if err != nil {
		if os.IsExist(err) {
			return werror.ErrFileAlreadyExists
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

	data := map[string]any{
		"id":              f.id,
		"portablePath": f.portablePath.ToPortable(),
		"filename":        f.filename,
		"size":            f.size.Load(),
		"isDir":           f.IsDir(),
		"modifyTimestamp": f.ModTime().UnixMilli(),
		"parentId":        parentId,
		"childrenIds": internal.Map(f.GetChildren(), func(c *WeblensFileImpl) FileId { return c.ID() }),
	}

	if f.ModTime().UnixMilli() < 0 {
		log.Debug.Println("AH")
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
		err := c.RecursiveMap(fn)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
LeafMap recursively perform fn on leaves, first, and work back up the tree.
This will not call fn on the root file.
This takes an inverted "Depth first" approach. Note this
behaves very differently than RecursiveMap. See below.

Files are acted on in the order of their index number here, starting with the leftmost leaf

		fx.LeafMap(fn) <- fn not called on root caller
		|
		f5
	   /  \
	  f3  f4
	 /  \
	f1  f2
*/
func (f *WeblensFileImpl) LeafMap(fn func(*WeblensFileImpl) error) error {
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
	return strings.HasPrefix(child.GetAbsPath(), f.GetAbsPath())
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

// func (f *WeblensFileImpl) GetMediaType() (types.MediaType, error) {
// 	if f.IsDir() {
// 		return nil, types.ErrDirNotAllowed
// 	}
// 	m := types.SERV.MediaRepo.Get(f.GetContentId())
// 	if m != nil {
// 		mt := m.GetMediaType()
// 		if mt != nil {
// 			return mt, nil
// 		}
// 	}
//
// 	mType := types.SERV.MediaRepo.TypeService().ParseExtension(f.Filename()[strings.LastIndex(f.Filename(), ".")+1:])
// 	return mType, nil
// }

// func (f *WeblensFileImpl) IsDisplayable() bool {
// 	mType, _ := f.GetMediaType()
// 	if mType == nil {
// 		return false
// 	}
//
// 	return mType.IsDisplayable()
// }

// LoadStat will recompute the size and modify date of the file using os.Stat. If the
// size of the file changes, LoadStat will return the newSize. If the size does not change,
// LoadStat will return -1 for the newSize.
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

	// if f.pastFile {
	// 	stat, err := f.tree.db.StatFile(f)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	f.size.Store(stat.Size())
	//
	// 	// Do not update modify time if the file is a past file,
	// 	// stat of file will give new "now" modify time, which is... not in the past.
	// 	return nil
	// }
	//
	// if err != nil {
	// 	return fmt.Errorf("failed to stat file %s: %s", f.absolutePath, err)
	// }
	//
	// if f.IsDir() {
	// 	children := f.GetChildren()
	// 	internal.Map(children, func(w *WeblensFileImpl) int { s, _ := w.Size(); newSize += s; return 0 })
	// } else {
	// 	newSize = stat.Size()
	// }
	//
	// f.modifyDate = stat.ModTime()

	// if origSize != newSize {
	// 	f.size.Store(newSize)
	// 	for cast := range slices.Values(c) {
	// 		cast.PushFileUpdate(f)
	// 	}
	// }

	// return
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

func (f *WeblensFileImpl) setBackupPath(backupPath string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.backupPath = backupPath
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

func (f *WeblensFileImpl) getBackupPathInternal() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.backupPath
}

func (f *WeblensFileImpl) hasChildren() bool {
	if !f.IsDir() {
		return false
	} else {
		return len(f.childrenMap) != 0
	}
}

func (f *WeblensFileImpl) removeChild(child *WeblensFileImpl) error {
	if len(f.childrenMap) == 0 {
		return werror.WithStack(werror.ErrNoChildren)
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()

	if _, ok := f.childrenMap[child.Filename()]; !ok {
		return werror.WithStack(werror.ErrNoFile)
	}

	delete(f.childrenMap, child.Filename())

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
	Write(data []byte) error
	ReadAll() ([]byte, error)
}