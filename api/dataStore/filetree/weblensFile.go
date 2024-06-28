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

type weblensFile struct {
	// the main way to identify a file. A file id is generated via a hash of its relative filepath
	id types.FileId

	// the file tree that this file belongs to
	tree *fileTree

	// The absolute path of the real file on disk
	absolutePath string

	// Base of the filepath, the actual name of the file.
	filename string

	// The user to whom the file belongs.
	owner types.User

	// size in bytes of the file on the disk
	size int64

	// is the real file on disk a directory or regular file
	isDir *bool

	// The most recent time that this file was changes on the real filesystem
	modifyDate time.Time

	// mediaService types.Media
	// This is the file id of the file in the .content folder that either holds
	// or points to the real bytes on disk content that this file should read from
	contentId types.ContentId

	// Pointer to the directory that this file belongs
	parent *weblensFile

	// If we already have added the file to the watcher
	// See fileWatch.go
	watching bool

	// If this file is a directory, these are the files that are housed by this directory.
	childLock *sync.Mutex
	children  []*weblensFile

	// array of tasks that currently claim are using this file.
	// TODO: allow single task-claiming of a file for file
	// operations required to be "atomic"
	taskUsing types.Task
	tasksLock *sync.Mutex

	// the share that belongs to this file
	share types.Share

	// Mark file as read-only internally.
	// This should be checked before any write action is to be taken
	// this should not be changed during run-time, only set in FsInit.
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
func (f *weblensFile) Copy() types.WeblensFile {
	// Copy values of wf struct
	c := *f

	// Create unique versions of pointers that are only relevant locally
	if c.isDir != nil {
		boolCopy := *c.isDir
		c.isDir = &boolCopy
	}

	if c.childLock != nil {
		c.childLock = &sync.Mutex{}
	}

	if c.tasksLock != nil {
		c.tasksLock = &sync.Mutex{}
	}

	// WeblensFile interface requires pointer
	return &c
}

// ID returns the unique identifier the file, and will compute it on the fly
// if it is not already initialized in the struct.
//
// This function will intentionally panic if trying to get the
// ID of a nil file.
func (f *weblensFile) ID() types.FileId {
	if f == nil {
		panic("Tried to get ID of nil file")
	}

	if f.id != "" {
		return f.id
	}

	f.id = f.tree.GenerateFileId(f.GetAbsPath())
	return f.id
}

// GetTree returns a pointer to the parent tree of the file
func (f *weblensFile) GetTree() types.FileTree {
	if f.tree == nil {
		panic("File does not have tree")
	}
	return f.tree
}

// Filename returns the filename of the file
func (f *weblensFile) Filename() string {
	return f.filename
}

// GetAbsPath returns string of the absolute path to file
func (f *weblensFile) GetAbsPath() string {
	if f.id == "EXTERNAL" {
		return ""
	}
	if f.absolutePath == "" {
		f.absolutePath = filepath.Join(f.parent.GetAbsPath(), f.filename)
	}
	if f.IsDir() && !strings.HasSuffix(f.absolutePath, "/") {
		f.absolutePath = f.absolutePath + "/"
	}
	return f.absolutePath
}

func (f *weblensFile) GetPortablePath() types.WeblensFilepath {
	return FilepathFromAbs(f.GetAbsPath())
}

// Owner returns the username of the owner of the file
func (f *weblensFile) Owner() types.User {
	if f == nil {
		panic("attempt to get owner on nil wf")
	}
	if f.owner == nil {
		// Media root has itself as its parent, so we use GetParent to turn *weblensFile to types.WeblensFile
		if f.GetParent() == f.tree.Get("MEDIA") {
			f.owner = types.SERV.UserService.Get(types.Username(f.Filename()))
			if string(f.owner.GetUsername()) != f.Filename() {
				panic(types.NewWeblensError("I don't even know man... look at Owner() on weblensFile"))
			}
		} else {
			f.owner = f.GetParent().Owner()
		}
	}
	return f.owner
}

func (f *weblensFile) SetOwner(o types.User) {
	f.owner = o
}

// Exists check if the file exists on the real filesystem below
func (f *weblensFile) Exists() bool {
	_, err := os.Stat(f.GetAbsPath())
	return err == nil
}

func (f *weblensFile) IsDir() bool {
	if f.isDir == nil {
		stat, err := os.Stat(f.absolutePath)
		if err != nil {
			util.ErrTrace(err)
			return false
		}
		f.isDir = boolPointer(stat.IsDir())
	}
	return *f.isDir
}

func (f *weblensFile) ModTime() (t time.Time) {
	if f.modifyDate.Unix() == 0 {
		err := f.LoadStat()
		if err != nil {
			util.ErrTrace(err)
		}
	}
	return f.modifyDate
}

func (f *weblensFile) Size() (int64, error) {
	if f.size != -1 {
		return f.size, nil
	}

	if f.ID() == "EXTERNAL" {
		var size int64
		util.Map(f.GetChildren(), func(c types.WeblensFile) int { sz, _ := c.Size(); size += sz; return 0 })
		f.size = size
		return f.size, nil
	}

	err := f.LoadStat()
	return f.size, err
}

func (f *weblensFile) Read() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	path := f.absolutePath
	if f.detached {
		path = "/tmp/" + f.filename
	}

	return os.Open(path)
}

func (f *weblensFile) ReadAll() (data []byte, err error) {
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
	if r != int(f.size) {
		return nil, types.ErrBadReadCount
	}

	return
}

func (f *weblensFile) Write(data []byte) error {
	if f.IsDir() {
		return types.ErrDirNotAllowed
	}
	err := os.WriteFile(f.absolutePath, data, 0660)
	if err == nil {
		f.size = int64(len(data))
		f.modifyDate = time.Now()
	}
	return err
}

func (f *weblensFile) WriteAt(data []byte, seekLoc int64) error {
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
		f.size += int64(wroteLen)
		f.modifyDate = time.Now()
	}

	return err
}

func (f *weblensFile) Append(data []byte) error {
	if f.IsDir() {
		return types.ErrDirNotAllowed
	}
	realFile, err := os.OpenFile(f.GetAbsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	wroteLen, err := realFile.Write(data)
	if err == nil {
		f.size += int64(wroteLen)
		f.modifyDate = time.Now()
	}
	return err
}

func (f *weblensFile) ReadDir() ([]types.WeblensFile, error) {
	if !f.IsDir() || !f.Exists() {
		return nil, fmt.Errorf("invalid file to read dir")
	}
	entries, err := os.ReadDir(f.absolutePath)
	if err != nil {
		return nil, err
	}

	children := make([]types.WeblensFile, len(entries))
	for i, file := range entries {
		singleChild := f.tree.NewFile(f, file.Name(), file.IsDir())
		exist := f.tree.Get(singleChild.ID())
		if exist != nil {
			continue
		}

		f.childLock.Lock()
		_, e := slices.BinarySearchFunc(f.children, singleChild.Filename(), searchWfByFilename)
		if e {
			f.childLock.Unlock()
			continue
		}
		f.childLock.Unlock()
		children[i] = singleChild

		// err := f.AddChild(singleChild)
		// if err != nil {
		// 	return err
		// }

		// err := fsTreeInsert(singleChild, f)
		// if err != nil {
		//	var alreadyExistsError AlreadyExistsError
		//	switch {
		//	case errors.As(err, &alreadyExistsError):
		//	default:
		//		return err
		//	}
		// }
	}

	children = util.Filter(
		children, func(c types.WeblensFile) bool {
			return c != nil
		},
	)

	// f.children = children

	return children, nil
}

func (f *weblensFile) GetContentId() types.ContentId {
	return f.contentId
}

func (f *weblensFile) SetContentId(cId types.ContentId) {
	f.contentId = cId
}

var searchWfByFilename = func(a *weblensFile, b string) int {
	return strings.Compare(strings.ToLower(a.filename), strings.ToLower(b))
}

var sortWfByFilename = func(a *weblensFile, b *weblensFile) int {
	return strings.Compare(strings.ToLower(a.filename), strings.ToLower(b.filename))
}

func (f *weblensFile) GetChild(childName string) (types.WeblensFile, error) {
	if f.children == nil || childName == "" {
		return nil, types.ErrNoFile
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()

	i, e := slices.BinarySearchFunc(f.children, childName, searchWfByFilename)
	if !e {
		return nil, types.ErrNoFile
	}

	return f.children[i], nil
}

func (f *weblensFile) GetChildren() []types.WeblensFile {
	if !f.IsDir() {
		return []types.WeblensFile{}
	}

	return util.SliceConvert[types.WeblensFile](f.children)
}

func (f *weblensFile) AddChild(child types.WeblensFile) error {
	if !f.IsDir() {
		return types.ErrDirectoryRequired
	}

	if slices.ContainsFunc(
		f.children, func(file *weblensFile) bool {
			return file == child
		},
	) {
		return types.ErrChildAlreadyExists
	}

	f.childLock.Lock()
	f.children = util.InsertFunc(f.children, child.(*weblensFile), sortWfByFilename)
	f.childLock.Unlock()

	return nil
}

func (f *weblensFile) GetChildrenInfo(acc types.AccessMeta) []types.FileInfo {
	childrenInfo := util.FilterMap(
		f.children, func(file *weblensFile) (types.FileInfo, bool) {
			info, err := file.FormatFileInfo(acc)
			if err != nil {
				util.ErrTrace(err)
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

func (f *weblensFile) GetParent() types.WeblensFile {
	if f.parent == nil {
		util.ErrTrace(types.NewWeblensError("Returning parent as nil from f.GetParent"))
	} else if f.parent == f {
		return nil
	}

	return f.parent
}

func (f *weblensFile) CreateSelf() error {
	if f.Exists() {
		return types.ErrFileAlreadyExists
	}
	if f.isDir == nil {
		return fmt.Errorf("cannot create self with nil self type")
	}

	var err error
	if f.IsDir() {
		err = os.Mkdir(f.GetAbsPath(), os.FileMode(0777))
		if err != nil {
			return err
		}
	} else {
		var osFile *os.File
		if f.detached {
			osFile, err = os.Create("/tmp/" + f.filename)
		} else {
			osFile, err = os.Create(f.GetAbsPath())
		}
		if err != nil {
			return err
		}
		err = osFile.Close()
		if err != nil {
			return err
		}
	}
	f.ID()
	return nil
}

func (f *weblensFile) MarshalJSON() ([]byte, error) {
	acc := dataStore.NewAccessMeta(nil, f.GetTree()).SetRequestMode(dataStore.MarshalFile)
	format, err := f.FormatFileInfo(acc)
	if err != nil {
		return nil, err
	}
	return json.Marshal(format)
}

type FileArray []types.WeblensFile

func (fa *FileArray) UnmarshalJSON(data []byte) error {
	var tmp []*weblensFile
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	*fa = util.Map(
		tmp, func(f *weblensFile) types.WeblensFile {
			return f.tree.NewFile(
				f.parent,
				f.filename, *f.isDir,
			)
		},
	)

	return nil
}

func (f *weblensFile) UnmarshalJSON(data []byte) error {
	tmp := map[string]any{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	f.id = types.FileId(tmp["id"].(string))
	f.absolutePath = FilepathFromPortable(tmp["portablePath"].(string)).ToAbsPath()
	f.filename = tmp["filename"].(string)
	util.ShowErr(types.NewWeblensError("TODO - get user in file unmarshal"))
	// f.owner = user.GetUser(types.Username(tmp["ownerName"].(string)))
	f.size = int64(tmp["size"].(float64))
	f.isDir = boolPointer(tmp["isDir"].(bool))
	t, err := types.FromSafeTimeStr(tmp["modifyDate"].(string))
	if err != nil {
		return err
	}
	f.modifyDate = t

	// if tmp["mediaService"] != nil {
	// 	m, err := util.StructFromMap[mediaService](tmp["mediaService"].(map[string]any))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	f.mediaService = &m
	// }

	parent := f.tree.Get(types.FileId(tmp["parentId"].(string)))
	if parent == nil {
		return types.ErrNoFile
	}
	f.parent = parent.(*weblensFile)

	f.children = util.FilterMap(
		tmp["childrenIds"].([]string), func(a string) (*weblensFile, bool) {
			c := f.tree.Get(types.FileId(a))
			if c == nil {
				util.Warning.Printf("Could not find child with id %s while unmarshal-ing weblens file", a)
				return nil, false
			} else {
				return c.(*weblensFile), true
			}
		},
	)

	f.share = types.SERV.ShareService.Get(types.ShareId(tmp["shareId"].(string)))

	return nil
}

func (f *weblensFile) MarshalArchive() map[string]any {
	pPath := ""
	if f.id == "EXTERNAL" {
		pPath = "EXTERNAL:"
	} else {
		pPath = FilepathFromAbs(f.absolutePath).ToPortable()
	}
	if pPath == ":" {
		panic("empty pPath")
	}

	return map[string]any{
		"id":           f.id,
		"portablePath": pPath,
		"filename":     f.filename,
		"ownerName":    f.Owner().GetUsername(),
		"size":         f.size,
		"isDir":        f.IsDir(),
		"modifyDate":   types.SafeTime(f.modifyDate),
		// "mediaService":        m,
		"parentId":    f.parent.ID(),
		"childrenIds": util.Map(f.GetChildren(), func(c types.WeblensFile) types.FileId { return c.ID() }),
		"shareId":     f.GetShare().GetShareId(),
	}
}

func (f *weblensFile) FormatFileInfo(acc types.AccessMeta) (formattedInfo types.FileInfo, err error) {
	if f == nil {
		return formattedInfo, fmt.Errorf("cannot get file info of nil wf")
	}

	if acc == nil {
		return formattedInfo, fmt.Errorf("cannot get file info without access context")
	}

	// if dataStore.dirIgnore[f.Filename()] {
	// 	return formattedInfo, fmt.Errorf("filename in block-list")
	// }

	if !acc.CanAccessFile(f) {
		err = types.ErrNoFileAccess
		return
	}

	var imported = true
	var m types.Media

	if !f.IsDir() {
		m = types.SERV.MediaRepo.Get(f.GetContentId())
		if m == nil {
			imported = false
		} else {
			err = m.AddFile(f)
			if err != nil {
				return
			}
		}
	}

	var size int64
	size, err = f.Size()
	if err != nil {
		util.ShowErr(err, fmt.Sprintf("Failed to get file size of [ %s (ID: %s) ]", f.absolutePath, f.id))
		return
	}

	var friendlyName string
	mType, _ := f.GetMediaType()
	if mType != nil {
		friendlyName = mType.FriendlyName()
	}

	var shareId types.ShareId
	if f.GetShare() != nil {
		shareId = f.GetShare().GetShareId()
	}

	var parentId types.FileId
	if f.Owner() != dataStore.WeblensRootUser && acc.CanAccessFile(f.GetParent()) {
		parentId = f.GetParent().ID()
	}

	tmpF := types.WeblensFile(f)
	var pathBits []string
	for tmpF != nil && tmpF.Owner() != dataStore.WeblensRootUser {
		if tmpF.GetParent() == f.tree.Get("MEDIA") {
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
		Imported:    imported,
		Displayable: f.IsDisplayable(),
		IsDir:       f.IsDir(),
		Modifiable: acc.GetTime().Unix() <= 0 && f.Filename() != ".user_trash" && f.Owner() != dataStore.
			WeblensRootUser && f != f.tree.Get("EXTERNAL"),
		Size:             size,
		ModTime:          f.ModTime(),
		Filename:         f.Filename(),
		ParentFolderId:   parentId,
		FileFriendlyName: friendlyName,
		Owner:            f.Owner().GetUsername(),
		PathFromHome:     pathString,
		MediaData:        m,
		Share:            shareId,
		Children:         util.Map(f.GetChildren(), func(wf types.WeblensFile) types.FileId { return wf.ID() }),
		PastFile:         acc.GetTime().Unix() > 0,
	}

	return formattedInfo, nil
}

// RecursiveMap /*
func (f *weblensFile) RecursiveMap(fn types.FileMapFunc) error {
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
func (f *weblensFile) LeafMap(fn types.FileMapFunc) error {
	if !f.IsDir() {
		return types.ErrDirectoryRequired
	}

	children := f.GetChildren()

	childrenWithChildren := util.Filter(
		children, func(c types.WeblensFile) bool { return c.(*weblensFile).hasChildren() },
	)
	for _, c := range childrenWithChildren {
		err := c.LeafMap(fn)
		if err != nil {
			return err
		}
	}

	for _, c := range children {
		err := fn(c)
		if err != nil {
			return err
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
func (f *weblensFile) BubbleMap(fn types.FileMapFunc) error {
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

func (f *weblensFile) IsParentOf(child types.WeblensFile) bool {
	return true
}

func (f *weblensFile) SetWatching() error {
	if f.watching {
		return types.ErrAlreadyWatching
	}

	f.watching = true
	return nil
}

var sleeperCount = atomic.Int64{}

func (f *weblensFile) AddTask(t types.Task) {

	sleeperCount.Add(1)
	f.tasksLock.Lock()
	sleeperCount.Add(-1)
	f.taskUsing = t
}

// GetTask Returns the task currently using this file
func (f *weblensFile) GetTask() types.Task {
	return f.taskUsing
}

func (f *weblensFile) RemoveTask(tId types.TaskId) error {
	if f.taskUsing == nil {
		util.Error.Printf("Task ID %s tried giving up file %s, but the file does not have a task", tId, f.GetAbsPath())
		panic(types.ErrBadTask)
	}
	if f.taskUsing.TaskId() != tId {
		util.Error.Printf(
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

func (f *weblensFile) GetShare() types.Share {
	return f.share
}

func (f *weblensFile) SetShare(sh types.Share) error {
	f.share = sh
	return nil
}

func (f *weblensFile) RemoveShare(sId types.ShareId) (err error) {
	// if f.share == nil {
	//	return types.ErrNoShare
	// }
	//
	// var e bool
	// f.share, _, e = util.YoinkFunc(f.share, func(share types.Share) bool { return share.GetShareId() == sId })
	// if !e {
	//	err = types.ErrNoShare
	// }
	return
}

// func (f *weblensFile) UpdateShare(s types.Share) (err error) {
// 	index := slices.IndexFunc(f.GetShares(), func(v types.Share) bool { return v.GetShareId() == s.GetShareId() })
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

func (f *weblensFile) IsReadOnly() bool {
	return f.readOnly
}

func (f *weblensFile) GetMediaType() (types.MediaType, error) {
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

func (f *weblensFile) IsDisplayable() bool {
	mType, _ := f.GetMediaType()
	if mType == nil {
		return false
	}

	return mType.IsDisplayable()
}

func (f *weblensFile) LoadStat(c ...types.BroadcasterAgent) (err error) {
	if f.absolutePath == "" {
		return nil
	}

	origSize := f.size
	var newSize int64 = 0

	if f.pastFile {
		statPath := ""
		if f.currentId != "" {
			statPath = f.tree.Get(f.currentId).GetAbsPath()
		} else {
			statPath = filepath.Join(
				f.GetTree().Get("CONTENT_LNIKS").GetAbsPath(),
				string(f.contentId),
			)
		}

		stat, err := os.Stat(statPath)
		if err != nil {
			return err
		}
		f.size = stat.Size()

		// Do not update modify time if the file is a past file,
		// stat of file will give new "now" modify time, which is... not in the past.
		return nil
	}

	stat, err := os.Stat(f.absolutePath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %s", f.absolutePath, err)
	}

	if f.IsDir() {
		children := f.GetChildren()
		util.Map(children, func(w types.WeblensFile) int { s, _ := w.Size(); newSize += s; return 0 })
	} else {
		newSize = stat.Size()
	}

	f.modifyDate = stat.ModTime()
	if origSize != newSize {
		f.size = newSize
		util.Each(c, func(c types.BroadcasterAgent) { c.PushFileUpdate(f) })
	}

	return
}

// Private

func (f *weblensFile) hasChildren() bool {
	if !f.IsDir() {
		return false
	} else {
		return len(f.children) != 0
	}
}

func (f *weblensFile) removeChild(child types.WeblensFile) error {
	if f.children == nil {
		util.Warning.Println("attempt to remove child on wf where children map is nil")
		return types.ErrNoFile
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()
	index, e := slices.BinarySearchFunc(f.children, child.Filename(), searchWfByFilename)
	if !e {
		return types.ErrNoFile
	}
	f.children = util.Banish(f.children, index)
	return nil
}
