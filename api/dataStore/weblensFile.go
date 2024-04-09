package dataStore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
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

	NOT using this interface will yeild slow and destructive
	results when attempting to modify the real filesystem underneath.
	Using this is required to keep the database, cache, and real filesystem
	in sync.
*/

// Returns a semi-deep copy of the file descriptor. All only-locally-relevent
// feilds are copied, however the media and children are the same references
// as the original version
func (f *weblensFile) Copy() types.WeblensFile {
	// Copy values of wf struct
	c := *f

	// Create unique versions of pointers that are only relevent locally
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

	// Return pointer to copy
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

	if f.absolutePath == "" {
		return ""
	}

	f.id = types.FileId(util.GlobbyHash(8, GuaranteeRelativePath(f.absolutePath)))

	return f.id
}

// Returns the filename of the file
func (f *weblensFile) Filename() string {
	if f.filename == "" {
		f.filename = filepath.Base(f.absolutePath)
	}
	return f.filename
}

// Returns string of absolute path to file
func (f *weblensFile) GetAbsPath() string {
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
		if f.GetParent() == &mediaRoot {
			f.owner = GetUser(types.Username(f.Filename()))
		} else {
			f.owner = f.GetParent().Owner()
		}
	}
	return f.owner
}

// Return a pointer to the media represented by this file,
// or a non-nil error if the media cannot be found.
func (f *weblensFile) GetMedia() (_ types.Media, err error) {
	if f.media != nil {
		return f.media, nil
	} else {
		return nil, ErrNoMedia
	}

	// if f.IsDir() {
	// 	return nil, ErrDirNotAllowed
	// }

	// err = loadMediaByFile(f)
	// m = f.media

	// return
}

func (f *weblensFile) SetMedia(m types.Media) error {
	if f.media != nil && f.media.Id() != "" {
		if f.media != m {
			return errors.New("attempted to reasign media on file descriptor that already has media")
		}
		return nil
	}
	mediaMapLock.Lock()
	if mediaMap[m.Id()] == nil {
		mediaMapLock.Unlock()
		return errors.New("attempted to assign media to file that is not in media map")
	}
	mediaMapLock.Unlock()
	f.media = m

	return nil
}

func (f *weblensFile) ClearMedia() {
	f.media = nil
}

// Check if the file exists on the real filesystem below
//
// f.absolutePath must be set
func (f *weblensFile) Exists() bool {
	_, err := os.Stat(f.absolutePath)
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
	if f.size != 0 {
		return f.size, nil
	}

	if f.Id() == "EXTERNAL_ROOT" {
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

	return os.Open(f.absolutePath)
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
	_, err = osFile.Read(data)

	return
}

func (f *weblensFile) Write(data []byte) error {
	if f.IsDir() {
		return fmt.Errorf("attempt to write to directory")
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
		return fmt.Errorf("attempt to write to directory")
	}

	realFile, err := os.OpenFile(f.GetAbsPath(), os.O_CREATE|os.O_WRONLY, 0664)
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
		return fmt.Errorf("attempt to write to directory")
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
		childPath := filepath.Join(f.absolutePath, file.Name())
		childId := types.FileId(util.GlobbyHash(8, GuaranteeRelativePath(childPath)))

		f.childLock.Lock()
		if _, ok := f.children[childId]; ok {
			f.childLock.Unlock()
			continue
		}
		f.childLock.Unlock()

		singleChild := weblensFile{
			id:           childId,
			absolutePath: childPath,
			isDir:        boolPointer(file.IsDir()),
			filename:     file.Name(),
			owner:        f.Owner(),
			tasksLock:    &sync.Mutex{},
			childLock:    &sync.Mutex{},
			children:     map[types.FileId]types.WeblensFile{},
		}

		err := FsTreeInsert(&singleChild, f)
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

func (f *weblensFile) GetChild(childId types.FileId) (child *weblensFile) {
	if f.children == nil {
		return
	}

	f.childLock.Lock()
	child = f.children[childId].(*weblensFile)
	f.childLock.Unlock()

	return
}

func (f *weblensFile) GetChildren() []types.WeblensFile {
	if !f.IsDir() {
		return []types.WeblensFile{}
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()
	return util.MapToSlicePure(f.children)
}

func (f *weblensFile) AddChild(child types.WeblensFile) {
	if !f.IsDir() {
		util.Error.Println("Attempting to add child to non-directory")
		return
	}

	f.childLock.Lock()
	if f.children[child.Id()] != nil {
		f.childLock.Unlock()
		util.Warning.Printf("Child (%s) of %s already exists", child.Filename(), f.Filename())
		return
	}
	f.children[child.Id()] = child
	f.childLock.Unlock()
}

func (f *weblensFile) GetChildrenInfo(acc types.AccessMeta) []types.FileInfo {
	f.childLock.Lock()
	defer f.childLock.Unlock()
	childs := util.MapToSlicePure(f.children)

	childrenInfo := util.FilterMap(childs, func(file types.WeblensFile) (types.FileInfo, bool) {
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
		return fmt.Errorf("directory or file already exists")
	}
	if f.isDir == nil {
		return fmt.Errorf("cannot create self with nil self type")
	}

	var err error
	if f.IsDir() {
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
	f.Id()
	return nil
}

func (f *weblensFile) MarshalJSON() ([]byte, error) {
	acc := NewAccessMeta("").SetRequestMode(MarshalFile)
	format, err := f.FormatFileInfo(acc)
	if err != nil {
		return nil, err
	}
	return json.Marshal(format)
}

func (f *weblensFile) FormatFileInfo(access types.AccessMeta) (formattedInfo types.FileInfo, err error) {
	if f == nil {
		return formattedInfo, fmt.Errorf("cannot get file info of nil wf")
	}

	if access == nil {
		return formattedInfo, fmt.Errorf("cannot get file info without access context")
	}

	if dirIgnore[f.Filename()] {
		return formattedInfo, fmt.Errorf("filename in blocklist")
	}

	if !CanAccessFile(f, access) {
		err = ErrNoFileAccess
		return
	}

	var imported bool = true
	var m types.Media

	if !f.IsDir() {
		m, err = f.GetMedia()
		if err != nil {
			imported = false
		}
	}

	var size int64
	size, err = f.Size()
	if err != nil {
		util.ErrTrace(err, fmt.Sprintf("Failed to get file size of [ %s (Id: %s) ]", f.absolutePath, f.id))
		return
	}

	var displayable bool
	var friendlyName string
	mType, _ := f.GetMediaType()
	if mType != nil {
		displayable = mType.IsDisplayable()
		friendlyName = mType.FriendlyName()
	}

	// shares := f.GetShares()
	var parentId types.FileId
	if f.Owner() != WEBLENS_ROOT_USER && CanAccessFile(f.GetParent(), access) {
		parentId = f.GetParent().Id()
	}
	if access == nil {
		util.Warning.Println("NIL ACCESS")
	}

	shares := util.Filter(f.GetShares(), func(s types.Share) bool {
		return slices.Contains(s.GetAccessors(), access.User().GetUsername())
	})

	tmpF := types.WeblensFile(f)
	pathBits := []string{}
	for tmpF != nil && tmpF.Owner() != WEBLENS_ROOT_USER {
		if tmpF.GetParent() == &mediaRoot {
			pathBits = append(pathBits, "HOME")
			break
		} else if access.UsingShare() != nil && tmpF.Id() == types.FileId(access.UsingShare().GetContentId()) {
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

	// pathString := GuaranteeRelativePath(f.absolutePath)
	// pathString = strings.Replace(pathString, "/"+f.Owner().String()+"/"+".user_trash", "TRASH", 1)
	// pathString = strings.Replace(pathString, "/"+f.Owner().String(), "HOME", 1)

	formattedInfo = types.FileInfo{
		Id:               f.Id(),
		Imported:         imported,
		Displayable:      displayable,
		IsDir:            f.IsDir(),
		Modifiable:       f.Filename() != ".user_trash" && f.Owner() != WEBLENS_ROOT_USER && f != &externalRoot,
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
	}

	return formattedInfo, nil
}

/*
Recursively perform fn on f, first, and all children of f. This takes a "Depth first" approach as shown below.

Files are acted on in the order of their index number here, starting with the root

		f1
	   /  \
	  f2  f5
	 /  \
	f3  f4
*/
func (f *weblensFile) RecursiveMap(fn func(types.WeblensFile)) {
	fn(f)
	if !f.IsDir() {
		return
	}

	children := f.GetChildren()

	for _, c := range children {
		c.RecursiveMap(fn)
	}
}

/*
Recursively perform fn on leaves, first, and work back up the tree.
This will not call fn on the root file.
This takes a "Depth first" approach, but note this
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
func (f *weblensFile) LeafMap(fn func(types.WeblensFile)) {
	if !f.IsDir() {
		return
	}

	children := f.GetChildren()

	childrenWithChildren := util.Filter(children, func(c types.WeblensFile) bool { return len(c.GetChildren()) != 0 })
	for _, c := range childrenWithChildren {
		c.LeafMap(fn)
	}

	for _, c := range children {
		fn(c)
	}
}

/*
Perform fn on f and all parents of f, not including the media root or other static directories.

Files are acted on in the order of their index number here, starting with the caller, children are never accessed

	f3 <- Parent of f2
	|
	f2 <- Parent of f1
	|
	f1 <- Called from
*/
func (f *weblensFile) BubbleMap(fn func(types.WeblensFile)) {
	if f == nil || f.owner == WEBLENS_ROOT_USER {
		return
	}
	fn(f)

	parent := f.GetParent()
	parent.BubbleMap(fn)
}

func (f *weblensFile) AddTask(t types.Task) {
	// if f.tasksLock == nil {
	// 	f.tasksLock = &sync.Mutex{}
	// }

	f.tasksLock.Lock()
	f.tasksUsing = util.AddToSet(f.tasksUsing, []types.Task{t})
	f.tasksLock.Unlock()
}

// Returns the tasks currently using this file
func (f *weblensFile) GetTasks() []types.Task {
	return f.tasksUsing
}

func (f *weblensFile) RemoveTask(tId types.TaskId) (exists bool) {
	// if f.tasksLock == nil {
	// 	f.tasksLock = &sync.Mutex{}
	// }

	f.tasksLock.Lock()
	f.tasksUsing, _, exists = util.YoinkFunc(f.tasksUsing, func(t types.Task) bool { return t.TaskId() == tId })
	f.tasksLock.Unlock()
	return
}

func (f *weblensFile) GetShares() []types.Share {
	if f.shares == nil {
		f.shares = []*fileShareData{}
	}
	shs := util.Map(f.shares, func(sh *fileShareData) types.Share { return sh })
	return shs
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
		f.shares = []*fileShareData{}
	}
	f.shares = append(f.shares, s.(*fileShareData))
}

func (f *weblensFile) RemoveShare(sId types.ShareId) (err error) {
	if f.shares == nil {
		return ErrNoShare
	}

	var e bool
	f.shares, _, e = util.YoinkFunc(f.shares, func(fShare *fileShareData) bool { return fShare.ShareId == sId })
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
	err = fddb.updateFileShare(f.shares[index].ShareId, s.(*fileShareData))
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
	origSize := f.size
	var newSize int64 = 0

	stat, err := os.Stat(f.absolutePath)
	if err != nil {
		switch err {
		case fs.ErrNotExist:
			return ErrNoFile
		default:
			return
		}
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
		if len(c) == 0 && globalCaster != nil {
			c = append(c, globalCaster)
		}
		util.Each(c, func(c types.BroadcasterAgent) { c.PushFileUpdate(f) })
	}

	return
}

func (f *weblensFile) removeChild(childId types.FileId) {
	if f.children == nil {
		util.Debug.Println("attempt to remove child on wf where children map is nil")
		return
	}
	f.childLock.Lock()
	if f.children[childId] == nil {
		f.childLock.Unlock()
		util.Warning.Printf("Child (%s) of %s does not exist (attempting to remove)", childId, f.Filename())
		return
	}

	delete(f.children, childId)
	f.childLock.Unlock()
}
