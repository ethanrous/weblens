package dataStore

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

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

/*
	These here are all the methods on a WeblensFile,
	an incredibly useful part of this backend logic.
	Using these (and more importantly the interface defined
	in filesystem.go and fileTree.go) will make requests to
	the filesystem on the server machine far faster both to
	write (as the programmer) and to execute.

	NOT using this interface will yield slow and destructive
	results when attempting to modify the real filesystem underneath.
	Using this is required to keep the database, cache, and real filesystem
	in sync.
*/

// Returns a semi-deep copy of the file descriptor. All only-locally-relevant
// felids are copied, however the media and children are the same references
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

// Returns the Id of the file, and will compute it on the fly
// if it is not already initialized in the struct.
//
// This function will intentionally panic if trying to get the
// Id of a nil file.
func (f *weblensFile) Id() types.FileId {
	if f == nil {
		panic("Tried to get Id of nil file")
	}

	if f.id != "" {
		return f.id
	}

	f.id = generateFileId(f.GetAbsPath())
	return f.id
}

// Returns the filename of the file
func (f *weblensFile) Filename() string {
	return f.filename
}

// Returns string of absolute path to file
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

// Returns the username of the owner of the file
func (f *weblensFile) Owner() types.User {
	if f == nil {
		panic("attempt to get owner on nil wf")
	}
	if f.owner == nil {
		// Media root has itself as its parent, so we use GetParent to turn *weblensFile to types.WeblensFile
		if f.GetParent() == mediaRoot.GetParent() {
			f.owner = GetUser(types.Username(f.Filename()))
		} else {
			f.owner = f.GetParent().Owner()
		}
	}
	return f.owner
}

// Return a pointer to the media represented by this file,
// or a non-nil error if the media cannot be found.
// func (f *weblensFile) GetMedia() (_ types.Media, err error) {
// 	if f.media != nil {
// 		return f.media, nil
// 	} else {
// 		return nil, ErrNoMedia
// 	}

// 	if f.IsDir() {
// 		return nil, ErrDirNotAllowed
// 	}

// 	err = loadMediaByFile(f)
// 	m = f.media

// 	return
// }

// func (f *weblensFile) SetMedia(m types.Media) error {
// 	if f.media != nil && f.media.Id() != "" {
// 		if f.media != m {
// 			return errors.New("attempted to reassign media on file descriptor that already has media")
// 		}
// 		return nil
// 	}
// 	mediaMapLock.Lock()
// 	if mediaMap[m.Id()] == nil {
// 		mediaMapLock.Unlock()
// 		return errors.New("attempted to assign media to file that is not in media map")
// 	}
// 	mediaMapLock.Unlock()
// 	f.media = m

// 	return nil
// }

// func (f *weblensFile) ClearMedia() {
// 	f.media = nil
// }

// Check if the file exists on the real filesystem below
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
		f.loadStat()
	}
	return f.modifyDate
}

func (f *weblensFile) Size() (int64, error) {
	if f.size != -1 {
		return f.size, nil
	}

	if f.Id() == "EXTERNAL" {
		var size int64
		util.Map(f.GetChildren(), func(c types.WeblensFile) int { sz, _ := c.Size(); size += sz; return 0 })
		f.size = size
		return f.size, nil
	}

	err := f.loadStat()
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
		return nil, ErrReadOff
	}

	return
}

func (f *weblensFile) Write(data []byte) error {
	if f.IsDir() {
		return ErrDirNotAllowed
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
		return ErrDirNotAllowed
	}

	path := f.GetAbsPath()

	if f.detached {
		path = "/tmp/" + f.Filename()
	}

	realFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer realFile.Close()

	wroteLen, err := realFile.WriteAt(data, seekLoc)
	if err == nil {
		f.size += int64(wroteLen)
		f.modifyDate = time.Now()
	}

	return err
}

func (f *weblensFile) Append(data []byte) error {
	if f.IsDir() {
		return ErrDirNotAllowed
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

func (f *weblensFile) ReadDir() error {
	if !f.IsDir() || !f.Exists() {
		return fmt.Errorf("invalid file to read dir")
	}
	entries, err := os.ReadDir(f.absolutePath)
	if err != nil {
		return err
	}

	for _, file := range entries {
		singleChild := newWeblensFile(f, file.Name(), file.IsDir())
		exist := FsTreeGet(singleChild.Id())
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

		err := fsTreeInsert(singleChild, f)
		if err != nil {
			switch err.(type) {
			case AlreadyExistsError:
			default:
				return err
			}
		}
	}

	return nil
}

func (f *weblensFile) GetContentId() types.ContentId {
	return f.contentId
}

var searchWfByFilename = func(a *weblensFile, b string) int {
	return strings.Compare(a.filename, b)
}

var sortWfByFilename = func(a *weblensFile, b *weblensFile) int {
	return strings.Compare(a.filename, b.filename)
}

func (f *weblensFile) GetChild(childName string) (types.WeblensFile, error) {
	if f.children == nil || childName == "" {
		return nil, ErrNoFile
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()

	i, e := slices.BinarySearchFunc(f.children, childName, searchWfByFilename)
	if !e {
		return nil, ErrNoFile
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
		return ErrDirectoryRequired
	}

	f.childLock.Lock()
	f.children = util.InsertFunc(f.children, child.(*weblensFile), sortWfByFilename)
	f.childLock.Unlock()

	return nil
}

func (f *weblensFile) GetChildrenInfo(acc types.AccessMeta) []types.FileInfo {
	childrenInfo := util.FilterMap(f.children, func(file *weblensFile) (types.FileInfo, bool) {
		info, err := file.FormatFileInfo(acc)
		if err != nil {
			return info, false
		}
		return info, true
	})

	return childrenInfo
}

func (f *weblensFile) GetParent() types.WeblensFile {
	if f.parent == nil {
		util.Warning.Println("Returning parent as nil from f.GetParent")
	} else if f.parent == f {
		return nil
	}

	return f.parent
}

func (f *weblensFile) CreateSelf() error {
	if f.Exists() {
		return ErrFileAlreadyExists
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
		osFile.Close()
	}
	f.Id()
	return nil
}

func (f weblensFile) MarshalJSON() ([]byte, error) {
	acc := NewAccessMeta(nil).SetRequestMode(MarshalFile)
	format, err := f.FormatFileInfo(acc)
	if err != nil {
		return nil, err
	}
	return json.Marshal(format)
}

type FileArray []types.WeblensFile

func (fa *FileArray) UnmarshalJSON(data []byte) error {
	tmp := []*weblensFile{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	*fa = FileArray(util.Map(tmp, func(f *weblensFile) types.WeblensFile { return newWeblensFile(f.parent, f.filename, *f.isDir) }))

	return nil
}

func (f *weblensFile) UnmarshalJSON(data []byte) error {
	tmp := map[string]any{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	f.id = types.FileId(tmp["id"].(string))
	f.absolutePath = portableFromString(tmp["portablePath"].(string)).Abs()
	f.filename = tmp["filename"].(string)
	f.owner = GetUser(types.Username(tmp["ownerName"].(string)))
	f.size = int64(tmp["size"].(float64))
	f.isDir = boolPointer(tmp["isDir"].(bool))
	t, err := types.FromSafeTimeStr(tmp["modifyDate"].(string))
	if err != nil {
		return err
	}
	f.modifyDate = t

	// if tmp["media"] != nil {
	// 	m, err := util.StructFromMap[media](tmp["media"].(map[string]any))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	f.media = &m
	// }

	parent := FsTreeGet(types.FileId(tmp["parentId"].(string)))
	if parent == nil {
		return ErrNoFile
	}
	f.parent = parent.(*weblensFile)

	f.children = util.FilterMap(tmp["childrenIds"].([]string), func(a string) (*weblensFile, bool) {
		c := FsTreeGet(types.FileId(a))
		if c == nil {
			util.Warning.Printf("Could not find child with id %s while unmarshal-ing weblens file", a)
			return nil, false
		} else {
			return c.(*weblensFile), true
		}
	})

	f.shares = util.Map(tmp["shareIds"].([]any), func(sId any) types.Share {
		s, _ := GetShare(types.ShareId(sId.(string)), FileShare)
		return s
	})

	return nil
}

func (f *weblensFile) MarshalArchive() map[string]any {
	// m, _ := f.GetMedia()
	// var mId types.MediaId
	// if m != nil {
	// 	mId = m.Id()
	// }

	pPath := ""
	if f.id == "EXTERNAL" {
		pPath = "EXTERNAL:"
	} else {
		pPath = AbsToPortable(f.absolutePath).PortableString()
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
		// "media":        m,
		"parentId":    f.parent.Id(),
		"childrenIds": util.Map(f.GetChildren(), func(c types.WeblensFile) types.FileId { return c.Id() }),
		"shareIds":    util.Map(f.GetShares(), func(s types.Share) types.ShareId { return s.GetShareId() }),
	}
}

func (f *weblensFile) FormatFileInfo(acc types.AccessMeta) (formattedInfo types.FileInfo, err error) {
	if f == nil {
		return formattedInfo, fmt.Errorf("cannot get file info of nil wf")
	}

	if acc == nil {
		return formattedInfo, fmt.Errorf("cannot get file info without access context")
	}

	if dirIgnore[f.Filename()] {
		return formattedInfo, fmt.Errorf("filename in block-list")
	}

	if !CanAccessFile(f, acc) {
		err = ErrNoFileAccess
		return
	}

	var imported bool = true
	var m types.Media

	if !f.IsDir() {
		m = MediaMapGet(f.GetContentId())
		if m == nil {
			imported = false
		} else {
			m.AddFile(f)
		}
	}

	var size int64
	size, err = f.Size()
	if err != nil {
		util.ShowErr(err, fmt.Sprintf("Failed to get file size of [ %s (Id: %s) ]", f.absolutePath, f.id))
		return
	}

	var friendlyName string
	mType, _ := f.GetMediaType()
	if mType != nil {
		friendlyName = mType.FriendlyName()
	}

	// shares := f.GetShares()
	var parentId types.FileId
	if f.Owner() != WEBLENS_ROOT_USER && CanAccessFile(f.GetParent(), acc) {
		parentId = f.GetParent().Id()
	}
	if acc == nil {
		util.Warning.Println("NIL ACCESS")
	}

	shares := util.Filter(f.GetShares(), func(s types.Share) bool {
		return CanAccessShare(s, acc)
	})

	tmpF := types.WeblensFile(f)
	pathBits := []string{}
	for tmpF != nil && tmpF.Owner() != WEBLENS_ROOT_USER {
		if tmpF.GetParent() == &mediaRoot {
			pathBits = append(pathBits, "HOME")
			break
		} else if acc.UsingShare() != nil && tmpF.Id() == types.FileId(acc.UsingShare().GetContentId()) {
			pathBits = append(pathBits, "SHARE")
			break
		} else if IsFileInTrash(tmpF) {
			pathBits = append(pathBits, "TRASH")
			break
		}
		pathBits = append(pathBits, tmpF.Filename())
		tmpF = tmpF.GetParent()
	}
	slices.Reverse(pathBits)
	pathString := strings.Join(pathBits, "/")

	formattedInfo = types.FileInfo{
		Id:               f.Id(),
		Imported:         imported,
		Displayable:      f.IsDisplayable(),
		IsDir:            f.IsDir(),
		Modifiable:       acc.GetTime().Unix() <= 0 && f.Filename() != ".user_trash" && f.Owner() != WEBLENS_ROOT_USER && f != &externalRoot,
		Size:             size,
		ModTime:          f.ModTime(),
		Filename:         f.Filename(),
		ParentFolderId:   parentId,
		FileFriendlyName: friendlyName,
		Owner:            f.Owner().GetUsername(),
		PathFromHome:     pathString,
		MediaData:        m,
		Shares:           shares,
		Children:         util.Map(f.GetChildren(), func(wf types.WeblensFile) types.FileId { return wf.Id() }),
		PastFile:         acc.GetTime().Unix() > 0,
	}

	return formattedInfo, nil
}

/*
Recursively perform fn on f, first, and all children of f. This takes a semi "Depth first" approach as shown below.

Files are acted on in the order of their index number shown below, starting with the root

		f1
	   /  \
	  f2  f5
	 /  \
	f3  f4
*/
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
Recursively perform fn on leaves, first, and work back up the tree.
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
		return ErrDirectoryRequired
	}

	children := f.GetChildren()

	childrenWithChildren := util.Filter(children, func(c types.WeblensFile) bool { return c.(*weblensFile).hasChildren() })
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
Perform fn on f and all parents of f, ignoring the media root or other static directories.

Files are acted on in the order of their index number below, starting with the caller, children are never accessed

	f3 <- Parent of f2
	|
	f2 <- Parent of f1
	|
	f1 <- Root caller
*/
func (f *weblensFile) BubbleMap(fn types.FileMapFunc) error {
	if f == nil || f.owner == WEBLENS_ROOT_USER {
		return nil
	}
	err := fn(f)
	if err != nil {
		return err
	}

	parent := f.GetParent()
	return parent.BubbleMap(fn)
}

var sleeperCount atomic.Int64 = atomic.Int64{}

func (f *weblensFile) AddTask(t types.Task) {
	// util.Debug.Printf("Task %s is trying to claim file %s (Sleepers: %d)", t.TaskId(), f.GetAbsPath(), sleeperCount.Load())
	sleeperCount.Add(1)
	f.tasksLock.Lock()
	sleeperCount.Add(-1)
	// util.Debug.Printf("Task %s has claimed file %s", t.TaskId(), f.GetAbsPath())
	f.taskUsing = t
}

// Returns the task currently using this file
func (f *weblensFile) GetTask() types.Task {
	return f.taskUsing
}

func (f *weblensFile) RemoveTask(tId types.TaskId) error {
	if f.taskUsing == nil {
		util.Error.Printf("Task Id %s tried giving up file %s, but the file does not have a task", tId, f.GetAbsPath())
		panic(ErrBadTask)
	}
	if f.taskUsing.TaskId() != tId {
		util.Error.Printf("Task Id %s tried giving up file %s, but the file is owned by %s does not own it", tId, f.GetAbsPath(), f.taskUsing.TaskId())
		panic(ErrBadTask)
		return ErrBadTask
	}

	f.taskUsing = nil
	f.tasksLock.Unlock()
	// util.Debug.Printf("Task %s has released file %s", tId, f.GetAbsPath())

	return nil
}

func (f *weblensFile) GetShares() []types.Share {
	if f.shares == nil {
		f.shares = []types.Share{}
	}
	return f.shares
}

func (f *weblensFile) GetShare(shareId types.ShareId) (sh types.Share, err error) {
	if f.shares == nil {
		err = ErrNoShare
		return
	}
	index := slices.IndexFunc(f.GetShares(), func(v types.Share) bool { return v.GetShareId() == shareId })
	if index == -1 {
		err = ErrNoShare
		return
	}
	sh = f.shares[index]
	return
}

func (f *weblensFile) AppendShare(s types.Share) {
	if f.shares == nil {
		f.shares = []types.Share{}
	}
	f.shares = append(f.shares, s.(*fileShareData))
}

func (f *weblensFile) RemoveShare(sId types.ShareId) (err error) {
	if f.shares == nil {
		return ErrNoShare
	}

	var e bool
	f.shares, _, e = util.YoinkFunc(f.shares, func(share types.Share) bool { return share.GetShareId() == sId })
	if !e {
		err = ErrNoShare
	}
	return
}

func (f *weblensFile) UpdateShare(s types.Share) (err error) {
	index := slices.IndexFunc(f.GetShares(), func(v types.Share) bool { return v.GetShareId() == s.GetShareId() })
	if index == -1 {
		return ErrNoShare
	}
	err = fddb.updateFileShare(f.shares[index].GetShareId(), s.(*fileShareData))
	if err != nil {
		return
	}
	if f.shares[index] != s {
		f.shares[index] = s.(*fileShareData)
		util.Warning.Println("Replacing share in full on file")
	}

	return
}

func (f *weblensFile) IsReadOnly() bool {
	return f.readOnly
}

// Private

func (f *weblensFile) loadStat(c ...types.BroadcasterAgent) (err error) {
	if f.absolutePath == "" {
		return nil
	}

	origSize := f.size
	var newSize int64 = 0

	if f.pastFile {
		statPath := ""
		if f.currentId != "" {
			statPath = FsTreeGet(f.currentId).GetAbsPath()
		} else {
			statPath = filepath.Join(contentRoot.absolutePath, string(f.contentId))
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

func (f *weblensFile) hasChildren() bool {
	if !f.IsDir() {
		return false
	} else {
		return len(f.children) != 0
	}
}

func (f *weblensFile) removeChild(child types.WeblensFile) error {
	if f.children == nil {
		util.Debug.Println("attempt to remove child on wf where children map is nil")
		return ErrNoFile
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()
	index, e := slices.BinarySearchFunc(f.children, child.Filename(), searchWfByFilename)
	if !e {
		return ErrNoFile
	}
	f.children = util.Banish(f.children, index)
	return nil
}
