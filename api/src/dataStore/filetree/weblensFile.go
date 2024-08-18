package filetree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"
)

/*
	WeblensFile is an incredibly useful part of the backend logic.
	Using these (and more importantly the interface defined
	in fileTree.go) will make requests to the filesystem on the
	server machine far faster both to write (as the programmer)
	and to execute.

	NOT using this interface will yield slow and destructive
	results when attempting to modify the real filesystem underneath.
	Using this is required to keep the database, cache, and real filesystem
	in sync.
*/

type WeblensFile struct {
	// the main way to identify a file. A file id is generated via a hash of its relative filepath
	id types.FileId

	// the file tree that this file belongs to
	tree *fileTree

	// The absolute path of the real file on disk
	absolutePath string

	// Path of the content file on a backup server
	backupPath string

	// Base of the filepath, the actual name of the file.
	filename string

	// The user to whom the file belongs.
	owner types.User

	// size in bytes of the file on the disk
	size atomic.Int64

	// is the real file on disk a directory or regular file
	isDir *bool

	// The most recent time that this file was changes on the real filesystem
	modifyDate time.Time

	// mediaService types.Media
	// This is the file id of the file in the .content folder that either holds
	// or points to the real bytes on disk content that this file should read from
	contentId types.ContentId

	// Pointer to the directory that this file belongs
	parent *WeblensFile

	// If we already have added the file to the watcher
	// See fileWatch.go
	watching bool

	// If this file is a directory, these are the files that are housed by this directory.
	childLock sync.RWMutex
	childrenMap map[string]*WeblensFile
	childIds    []types.FileId

	// General RW lock on file updates to prevent data races
	updateLock sync.RWMutex

	// task operations required to be "atomic" need to lock the file to prevent
	// multiple changes being made at the same time
	taskUsing types.Task
	tasksLock sync.Mutex

	// the share that belongs to this file
	share types.Share

	// Mark file as read-only internally.
	// This should be checked before any write action is to be taken
	// this should not be changed during run-time, only set in InitMediaRoot.
	// If a directory is `readOnly`, all children are as well
	readOnly bool

	// this file represents a file possibly not on the filesystem
	// anymore, but was at some point in the past
	pastFile bool

	// If the file is a past file, and existed at the real id above, this
	// current fileId is the location of the content right now, not in the past.
	currentId types.FileId

	// this file is currently existing outside the file tree, most likely
	// in the /tmp directory
	detached bool
}

// Copy returns a semi-deep copy of the file descriptor. All only-locally-relevant
// fields are copied, however the mediaService and children are the same references
// as the original version
func (f *WeblensFile) Copy() types.WeblensFile {
	// Copy values of wf struct
	c := *f

	// Create unique versions of pointers that are only relevant locally
	if c.isDir != nil {
		boolCopy := *c.isDir
		c.isDir = &boolCopy
	}

	c.childLock = sync.RWMutex{}
	c.tasksLock = sync.Mutex{}
	c.updateLock = sync.RWMutex{}

	// WeblensFile interface requires pointer
	return &c
}

// ID returns the unique identifier the file, and will compute it on the fly
// if it is not already initialized in the struct.
//
// This function will intentionally panic if trying to get the
// ID of a nil file.
func (f *WeblensFile) ID() types.FileId {
	if f == nil {
		wlog.ShowErr(types.WeblensErrorMsg("Tried to get ID of nil file"))
		return ""
	}

	id := f.getIdInternal()
	if id != "" {
		return id
	}

	id = f.tree.GenerateFileId(f.GetAbsPath())
	f.setIdInternal(id)
	return id
}

func (f *WeblensFile) getIdInternal() types.FileId {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.id
}

func (f *WeblensFile) setIdInternal(id types.FileId) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.id = id
}

// GetTree returns a pointer to the parent tree of the file
func (f *WeblensFile) GetTree() types.FileTree {
	// if f.tree == nil {
	// 	panic("File does not have tree")
	// }
	return f.tree
}

// Filename returns the filename of the file
func (f *WeblensFile) Filename() string {
	return f.filename
}

func (f *WeblensFile) setAbsPath(absPath string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.absolutePath = absPath
}

func (f *WeblensFile) setBackupPath(backupPath string) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.backupPath = backupPath
}

func (f *WeblensFile) getAbsPathInternal() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.absolutePath
}

func (f *WeblensFile) getBackupPathInternal() string {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.backupPath
}

// GetAbsPath returns string of the absolute path to file
func (f *WeblensFile) GetAbsPath() string {
	if f == nil {
		return ""
	}
	if f.id == "EXTERNAL" {
		return ""
	}

	if backup := f.getBackupPathInternal(); backup != "" {
		return backup
	}

	if f.id == "ROOT" {
		f.setAbsPath(util.GetMediaRootPath())
		return f.getAbsPathInternal()
	}

	if types.SERV.InstanceService.GetLocal().IsCore() || f.Owner().IsSystemUser() {
		// If this is a core server, attach filename to the and of the parent directory path
		if f.getAbsPathInternal() == "" {
			f.setAbsPath(filepath.Join(f.parent.GetAbsPath(), f.filename))
		}

		// Directories must and with a "/"
		if f.IsDir() && f.getAbsPathInternal()[len(f.getAbsPathInternal())-1:] != "/" {
			f.setAbsPath(f.getAbsPathInternal() + "/")
		}
	} else {
		// If this is a backup server, we use the backup path for the "real" path
		f.setBackupPath(filepath.Join(f.tree.Get("ROOT").GetAbsPath(), string(f.GetContentId())))
		return f.getBackupPathInternal()
	}
	return f.getAbsPathInternal()
}

func (f *WeblensFile) GetPortablePath() types.WeblensFilepath {
	return FilepathFromAbs(f.GetAbsPath())
}

// Owner returns the user that owns the file
func (f *WeblensFile) Owner() types.User {
	if f == nil {
		panic("attempt to get owner on nil wf")
	}
	if f.owner == nil {
		// Media root has itself as its parent, so we use GetParent to turn *WeblensFile to types.WeblensFile
		if f.GetParent() == f.tree.Get("MEDIA") {
			f.owner = types.SERV.UserService.Get(types.Username(f.Filename()))
			if string(f.owner.GetUsername()) != f.Filename() {
				panic(types.NewWeblensError("I don't even know man... look at Owner() on WeblensFile"))
			}
		} else {
			wlog.Debug.Println("ABS PATH", f.GetAbsPath())
			f.owner = f.GetParent().Owner()
		}
	}
	return f.owner
}

func (f *WeblensFile) SetOwner(o types.User) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.owner = o
}

// Exists check if the file exists on the real filesystem below
func (f *WeblensFile) Exists() bool {
	stat, err := f.tree.db.StatFile(f)
	if err != nil {
		return false
	}

	return stat.Exists
}

func (f *WeblensFile) IsDir() bool {
	if f.isDir == nil {
		stat, err := f.tree.db.StatFile(f)
		if err != nil {
			wlog.ErrTrace(err)
			return false
		}
		f.isDir = boolPointer(stat.IsDir)
	}
	return *f.isDir
}

func (f *WeblensFile) ModTime() (t time.Time) {
	if f.modifyDate.Unix() <= 0 {
		err := f.LoadStat()
		if err != nil {
			wlog.ErrTrace(err)
		}
	}
	return f.modifyDate
}

func (f *WeblensFile) setModTime(t time.Time) {
	f.modifyDate = t
}

func (f *WeblensFile) recomputeSize() (int64, error) {
	if f.ID() == "EXTERNAL" {
		var size int64
		util.Map(f.GetChildren(), func(c types.WeblensFile) int { sz, _ := c.Size(); size += sz; return 0 })
		f.size.Store(size)
		return f.size.Load(), nil
	}

	if f.IsDir() {
		newSize := int64(0)
		for _, c := range f.GetChildren() {
			cs, err := c.Size()
			if err != nil {
				return 0, err
			}
			newSize += cs
		}
		f.size.Store(newSize)
	} else {
		err := f.LoadStat(types.SERV.Caster)
		if err != nil {
			return f.size.Load(), types.WeblensErrorFromError(err)
		}
	}

	return f.size.Load(), nil
}

func (f *WeblensFile) Size() (int64, error) {
	if f.size.Load() <= 0 {
		return f.recomputeSize()
	}

	return f.size.Load(), nil
}

func (f *WeblensFile) Readable() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	path := f.absolutePath
	if f.detached {
		path = "/tmp/" + f.filename
	}

	return os.Open(path)
}

func (f *WeblensFile) Writeable() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	path := f.GetAbsPath()
	if f.detached {
		path = "/tmp/" + f.filename
	}

	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0660)
}

func (f *WeblensFile) ReadAll() (data []byte, err error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}
	osFile, err := os.Open(f.absolutePath)
	if err != nil {
		return
	}
	fileSize, err := f.Size()
	if err != nil {
		return
	}
	data = make([]byte, fileSize)
	r, err := osFile.Read(data)
	if r != int(f.size.Load()) {
		return nil, types.ErrBadReadCount
	}

	return
}

func (f *WeblensFile) Write(data []byte) error {
	if f.IsDir() {
		return types.ErrDirNotAllowed
	}
	err := os.WriteFile(f.GetAbsPath(), data, 0660)
	if err == nil {
		f.size.Store(int64(len(data)))
		f.modifyDate = time.Now()
	}
	return err
}

func (f *WeblensFile) WriteAt(data []byte, seekLoc int64) error {
	if f.IsDir() {
		return types.ErrDirNotAllowed
	}

	path := f.GetAbsPath()

	if f.detached {
		path = "/tmp/" + f.Filename()
	}

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

func (f *WeblensFile) Append(data []byte) error {
	if f.IsDir() {
		return types.ErrDirNotAllowed
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

func (f *WeblensFile) GetContentId() types.ContentId {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.contentId
}

func (f *WeblensFile) SetContentId(cId types.ContentId) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.contentId = cId
}

func (f *WeblensFile) loadChildrenFromIds() {
	if len(f.childIds) != 0 {
		f.childLock.Lock()
		defer f.childLock.Unlock()
		if f.childrenMap == nil {
			f.childrenMap = make(map[string]*WeblensFile)
		}
		util.Each(
			f.childIds, func(cId types.FileId) {
				child := f.tree.Get(cId)
				if child == nil {
					return
				}
				f.childrenMap[child.Filename()] = child.(*WeblensFile)
			},
		)
		f.childIds = nil
	}
}

func (f *WeblensFile) ReadDir() ([]types.WeblensFile, error) {
	if !f.IsDir() {
		return nil, fmt.Errorf("cannot read dir of regular file")
	}

	f.loadChildrenFromIds()
	if len(f.childrenMap) != 0 {
		return f.GetChildren(), nil
	}

	entries, err := f.tree.db.ReadDir(f)

	if err != nil {
		return nil, err
	}
	var children []types.WeblensFile
	for _, file := range entries {
		var u types.User
		if f == f.tree.GetRoot() {
			u = types.SERV.UserService.Get(types.Username(file.Name))
		} else {
			u = f.Owner()
		}

		singleChild := f.tree.NewFile(f, file.Name, file.IsDir, u)
		if file.Size > 0 && !singleChild.IsDir() {
			singleChild.(*WeblensFile).size.Store(file.Size)
		}

		f.childLock.Lock()
		children = append(children, singleChild)
		f.childLock.Unlock()
	}

	return children, nil
}

func (f *WeblensFile) GetChild(childName string) (types.WeblensFile, error) {
	f.loadChildrenFromIds()
	f.childLock.RLock()
	defer f.childLock.RUnlock()
	if len(f.childrenMap) == 0 || childName == "" {
		return nil, types.ErrNoFileName(childName)
	}

	child := f.childrenMap[childName]
	if child == nil {
		return nil, types.ErrNoFileName(childName)
	}

	return child, nil
}

func (f *WeblensFile) GetChildren() []types.WeblensFile {
	if !f.IsDir() {
		return []types.WeblensFile{}
	}

	f.loadChildrenFromIds()

	f.childLock.RLock()
	defer f.childLock.RUnlock()

	return util.SliceConvert[types.WeblensFile](util.MapToValues(f.childrenMap))
}

func (f *WeblensFile) AddChild(child types.WeblensFile) error {
	if !f.IsDir() {
		return types.ErrDirectoryRequired
	}

	f.loadChildrenFromIds()

	f.childLock.Lock()
	defer f.childLock.Unlock()
	f.childrenMap[child.Filename()] = child.(*WeblensFile)
	sz, _ := child.Size()
	f.size.Add(sz)

	return nil
}

func (f *WeblensFile) GetChildrenInfo(acc types.AccessMeta) []types.FileInfo {
	f.loadChildrenFromIds()

	childrenInfo := util.FilterMap(
		f.GetChildren(), func(file types.WeblensFile) (types.FileInfo, bool) {
			info, err := file.FormatFileInfo(acc)
			if err != nil {
				wlog.ErrTrace(err)
				return info, false
			}
			return info, true
		},
	)

	if childrenInfo == nil {
		return []types.FileInfo{}
	}

	return childrenInfo
}

func (f *WeblensFile) GetParent() types.WeblensFile {
	f.updateLock.RLock()
	defer f.updateLock.RUnlock()
	return f.parent
}

func (f *WeblensFile) setParentInternal(parent *WeblensFile) {
	f.updateLock.Lock()
	defer f.updateLock.Unlock()
	f.parent = parent
}

func (f *WeblensFile) CreateSelf() error {
	err := f.tree.db.TouchFile(f)
	if err != nil {
		return err
	}

	f.ID()
	return nil
}

func (f *WeblensFile) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return err
	}

	f.tree = types.SERV.FileTree.(*fileTree)

	f.id = types.FileId(data["id"].(string))
	f.absolutePath = FilepathFromPortable(data["portablePath"].(string)).ToAbsPath()
	f.filename = data["filename"].(string)
	f.owner = types.SERV.UserService.Get(types.Username(data["ownerName"].(string)))
	f.size.Store(int64(data["size"].(float64)))
	f.isDir = boolPointer(data["isDir"].(bool))
	f.modifyDate = time.UnixMilli(int64(data["modifyTimestamp"].(float64)))
	if f.modifyDate.Unix() <= 0 {
		wlog.Error.Println("AHHHH")
	}

	parentId := types.FileId(data["parentId"].(string))
	if parentId != "" {
		parent := f.tree.Get(parentId)
		if parent == nil {
			return types.ErrNoFile(parentId)
		}
		f.parent = parent.(*WeblensFile)
		err = parent.AddChild(f)
		if err != nil {
			return err
		}
	}

	f.childIds = util.Map(
		util.SliceConvert[string](data["childrenIds"].([]any)), func(cId string) types.FileId {
			return types.FileId(cId)
		},
	)

	f.share = types.SERV.ShareService.Get(types.ShareId(data["shareId"].(string)))
	f.tree.addInternal(f.id, f)

	return nil
}

func (f *WeblensFile) MarshalJSON() ([]byte, error) {
	pPath := ""
	if f.id == "EXTERNAL" {
		pPath = "EXTERNAL:"
	} else {
		pPath = FilepathFromAbs(f.absolutePath).ToPortable()
	}

	var shareId types.ShareId
	if f.GetShare() != nil {
		shareId = f.GetShare().GetShareId()
	}

	var parentId types.FileId
	if f.parent != nil {
		parentId = f.parent.ID()
	}

	data := map[string]any{
		"id":              f.id,
		"portablePath":    pPath,
		"filename":        f.filename,
		"ownerName":       f.Owner().GetUsername(),
		"size":            f.size.Load(),
		"isDir":           f.IsDir(),
		"modifyTimestamp": f.ModTime().UnixMilli(),
		"parentId":        parentId,
		"childrenIds":     util.Map(f.GetChildren(), func(c types.WeblensFile) types.FileId { return c.ID() }),
		"shareId":         shareId,
	}

	return json.Marshal(data)
}

func (f *WeblensFile) FormatFileInfo(acc types.AccessMeta) (formattedInfo types.FileInfo, err error) {
	if f == nil {
		return formattedInfo, fmt.Errorf("cannot get file info of nil wf")
	}

	if acc == nil {
		return formattedInfo, fmt.Errorf("cannot get file info without access context")
	}

	if !acc.CanAccessFile(f) {
		err = types.ErrNoFileAccess
		return
	}

	m := types.SERV.MediaRepo.Get(f.GetContentId())

	var size int64
	size, err = f.Size()
	if err != nil {
		wlog.ShowErr(err, fmt.Sprintf("Failed to get file size of [ %s (ID: %s) ]", f.absolutePath, f.id))
		return
	}

	var shareId types.ShareId
	if f.GetShare() != nil {
		shareId = f.GetShare().GetShareId()
		wlog.Debug.Println("ShareId", shareId)
	}

	var parentId types.FileId
	if f.Owner() != types.SERV.UserService.Get("WEBLENS") && acc.CanAccessFile(f.GetParent()) {
		parentId = f.GetParent().ID()
	}

	tmpF := types.WeblensFile(f)
	var pathBits []string
	for tmpF != nil && tmpF.Owner() != types.SERV.UserService.Get("WEBLENS") && tmpF.Owner() != dataStore.ExternalRootUser && acc.CanAccessFile(tmpF) {
		if tmpF.GetParent() == f.tree.GetRoot() {
			pathBits = append(pathBits, "HOME")
			break
		} else if acc.UsingShare() != nil && tmpF.ID() == types.FileId(acc.UsingShare().GetItemId()) {
			pathBits = append(pathBits, "SHARE")
			break
		} else if dataStore.IsFileInTrash(tmpF) {
			pathBits = append(pathBits, "TRASH")
			break
		}
		pathBits = append(pathBits, tmpF.Filename())
		tmpF = tmpF.GetParent()
	}
	slices.Reverse(pathBits)
	pathString := strings.Join(pathBits, "/")

	formattedInfo = types.FileInfo{
		Id:          f.ID(),
		Displayable: f.IsDisplayable(),
		IsDir:       f.IsDir(),
		Modifiable: acc.GetTime().Unix() <= 0 &&
			!dataStore.IsFileInTrash(f) &&
			f.Owner() == acc.User() &&
			f.Owner() != types.SERV.UserService.Get("WEBLENS") &&
			f != f.tree.Get("EXTERNAL") &&
			types.SERV.InstanceService.GetLocal().ServerRole() != types.Backup,
		Size:           size,
		ModTime:        f.ModTime().UnixMilli(),
		Filename:       f.Filename(),
		ParentFolderId: parentId,
		Owner:          f.Owner().GetUsername(),
		PathFromHome:   pathString,
		MediaData:      m,
		ShareId:        shareId,
		Children:       util.Map(f.GetChildren(), func(wf types.WeblensFile) types.FileId { return wf.ID() }),
		PastFile:       acc.GetTime().Unix() > 0,
	}

	return formattedInfo, nil
}

// RecursiveMap applies function fn to every file recursively
func (f *WeblensFile) RecursiveMap(fn types.FileMapFunc) error {
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
func (f *WeblensFile) LeafMap(fn types.FileMapFunc) error {
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
func (f *WeblensFile) BubbleMap(fn types.FileMapFunc) error {
	if f == nil || slices.Contains(dataStore.RootDirIds, f.ID()) {
		return nil
	}
	err := fn(f)
	if err != nil {
		return err
	}

	parent := f.GetParent()
	return parent.BubbleMap(fn)
}

func (f *WeblensFile) IsParentOf(child types.WeblensFile) bool {
	return strings.HasPrefix(child.GetAbsPath(), f.GetAbsPath())
}

func (f *WeblensFile) SetWatching() error {
	if f.watching {
		return types.ErrAlreadyWatching
	}

	f.watching = true
	return nil
}

var sleeperCount = atomic.Int64{}

func (f *WeblensFile) AddTask(t types.Task) {

	sleeperCount.Add(1)
	f.tasksLock.Lock()
	sleeperCount.Add(-1)
	f.taskUsing = t
}

// GetTask Returns the task currently using this file
func (f *WeblensFile) GetTask() types.Task {
	return f.taskUsing
}

func (f *WeblensFile) RemoveTask(tId types.TaskId) error {
	if f.taskUsing == nil {
		wlog.Error.Printf("Task ID %s tried giving up file %s, but the file does not have a task", tId, f.GetAbsPath())
		panic(types.ErrBadTask)
	}
	if f.taskUsing.TaskId() != tId {
		wlog.Error.Printf(
			"Task ID %s tried giving up file %s, but the file is owned by %s does not own it", tId, f.GetAbsPath(),
			f.taskUsing.TaskId(),
		)
		panic(types.ErrBadTask)
		return types.ErrBadTask
	}

	f.taskUsing = nil
	f.tasksLock.Unlock()

	return nil
}

func (f *WeblensFile) GetShare() types.Share {
	return f.share
}

func (f *WeblensFile) SetShare(sh types.Share) error {
	f.share = sh
	return nil
}

func (f *WeblensFile) RemoveShare(sId types.ShareId) (err error) {
	// if f.share == nil {
	//	return types.ErrNoShare
	// }
	//
	// var e bool
	// f.share, _, e = util.YoinkFunc(f.share, func(share types.ShareId) bool { return share.GetShareId() == sId })
	// if !e {
	//	err = types.ErrNoShare
	// }
	return types.ErrNotImplemented("RemoveShare weblensFile")
}

// func (f *WeblensFile) UpdateShare(s types.ShareId) (err error) {
// 	index := slices.IndexFunc(f.GetShares(), func(v types.ShareId) bool { return v.GetShareId() == s.GetShareId() })
// 	if index == -1 {
// 		return types.ErrNoShare
// 	}
// 	err = f.tree.GetShareService().updateFileShare(f.shares[index].GetShareId(),
// 		s.(*dataStore.fileShareData))
// 	if err != nil {
// 		return
// 	}
// 	if f.shares[index] != s {
// 		f.shares[index] = s.(*dataStore.fileShareData)
// 		util.Warning.Println("Replacing share in full on file")
// 	}
//
// 	return
// }

func (f *WeblensFile) IsReadOnly() bool {
	return f.readOnly
}

func (f *WeblensFile) GetMediaType() (types.MediaType, error) {
	if f.IsDir() {
		return nil, types.ErrDirNotAllowed
	}
	m := types.SERV.MediaRepo.Get(f.GetContentId())
	if m != nil {
		mt := m.GetMediaType()
		if mt != nil {
			return mt, nil
		}
	}

	mType := types.SERV.MediaRepo.TypeService().ParseExtension(f.Filename()[strings.LastIndex(f.Filename(), ".")+1:])
	return mType, nil
}

func (f *WeblensFile) IsDisplayable() bool {
	mType, _ := f.GetMediaType()
	if mType == nil {
		return false
	}

	return mType.IsDisplayable()
}

func (f *WeblensFile) LoadStat(c ...types.BroadcasterAgent) (err error) {
	if f.absolutePath == "" {
		return nil
	}

	origSize := f.size.Load()
	var newSize int64 = 0

	if f.pastFile {
		stat, err := f.tree.db.StatFile(f)
		if err != nil {
			return err
		}
		f.size.Store(stat.Size)

		// Do not update modify time if the file is a past file,
		// stat of file will give new "now" modify time, which is... not in the past.
		return nil
	}

	stat, err := f.tree.db.StatFile(f)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %s", f.absolutePath, err)
	}

	if f.IsDir() {
		children := f.GetChildren()
		util.Map(children, func(w types.WeblensFile) int { s, _ := w.Size(); newSize += s; return 0 })
	} else {
		newSize = stat.Size
	}

	f.modifyDate = stat.ModTime
	if origSize != newSize {
		f.size.Store(newSize)
		util.Each(c, func(c types.BroadcasterAgent) { c.PushFileUpdate(f) })
	}

	return
}

func (f *WeblensFile) IsDetached() bool {
	return f.detached
}

// Private

func (f *WeblensFile) hasChildren() bool {
	if !f.IsDir() {
		return false
	} else {
		return len(f.childrenMap) != 0
	}
}

func (f *WeblensFile) removeChild(child types.WeblensFile) error {
	if f.childrenMap == nil {
		return types.ErrNoChildren
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()
	delete(f.childrenMap, child.Filename())

	return nil
}
