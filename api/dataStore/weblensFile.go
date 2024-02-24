package dataStore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

/*
	These here are all the methods on a WeblensFileDescriptor,
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
func (f *WeblensFile) Copy() *WeblensFile {
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
func (f *WeblensFile) Id() string {
	if f == nil {
		panic("Tried to get Id of nil file")
	}
	if f.absolutePath == "" {
		return ""
	}
	if f.id == "" {
		f.id = util.GlobbyHash(8, GuaranteeRelativePath(f.absolutePath))
	}
	return f.id
}

// Returns the filename of the file
func (f *WeblensFile) Filename() string {
	if f.filename == "" {
		f.filename = filepath.Base(f.absolutePath)
	}
	return f.filename
}

// Returns string of absolute path to file
func (f *WeblensFile) String() string {
	if f.IsDir() && !strings.HasSuffix(f.absolutePath, "/") {
		f.absolutePath = f.absolutePath + "/"
	}
	return f.absolutePath
}

// Returns the username of the owner of the file
func (f *WeblensFile) Owner() string {
	if f == nil {
		panic("attempt to get owner on nil wf")
	}
	if f.owner == "" {
		if f.GetParent() == &mediaRoot {
			f.owner = f.Filename()
		} else {
			f.owner = f.GetParent().Owner()
		}
	}
	return f.owner
}

// Returns the slice of usernames who have been shared into the file.
// This does not include the owner
func (f *WeblensFile) Guests() []string {
	if f.guests == nil {
		f.guests = fddb.getFileGuests(f)
	}
	return f.guests
}

// Return a pointer to the media represented by this file,
// or a non-nil error if the media cannot be found.
func (f *WeblensFile) GetMedia() (m *Media, err error) {
	if f.media != nil {
		return f.media, nil
	}

	if f.IsDir() {
		return nil, fmt.Errorf("cannot get media of directory")
	}

	err = loadMediaByFile(f)
	m = f.media

	return
}

func (f *WeblensFile) SetMedia(m *Media) error {
	if f.media != nil && f.media.MediaId != "" {
		if f.media != m {
			return errors.New("attempted to reasign media on file descriptor that already has media")
		}
	}
	mediaMapLock.Lock()
	if mediaMap[m.MediaId] == nil {
		mediaMapLock.Unlock()
		return errors.New("attempted to assign media to file that is not in media map")
	}
	mediaMapLock.Unlock()
	f.media = m

	return nil
}

func (f *WeblensFile) ClearMedia() {
	f.media = nil
}

// Check if the file exists on the real filesystem below
//
// f.absolutePath must be set
func (f *WeblensFile) Exists() bool {
	_, err := os.Stat(f.absolutePath)
	return err == nil
}

func (f *WeblensFile) IsDir() bool {
	if f.isDir == nil {
		stat, err := os.Stat(f.absolutePath)
		if err != nil {
			util.DisplayError(err)
			return false
		}
		f.isDir = boolPointer(stat.IsDir())
	}
	return *f.isDir
}

func (f *WeblensFile) ModTime() (t time.Time) {
	stat, err := os.Stat(f.absolutePath)
	if err != nil {
		util.DisplayError(err)
		return
	}
	t = stat.ModTime()
	return
}

func (f *WeblensFile) Size() (int64, error) {
	if f.size != 0 {
		return f.size, nil
	}

	// if !f.assertExists() {
	// 	return 0, fmt.Errorf("getting size of file that does not exist")
	// }

	return f.recompSize()
}

func (f *WeblensFile) Read() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	return os.Open(f.absolutePath)
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
	_, err = osFile.Read(data)

	return
}

func (f *WeblensFile) Write(data []byte) error {
	if f.IsDir() {
		return fmt.Errorf("attempt to write to directory")
	}
	err := os.WriteFile(f.absolutePath, data, 0660)
	if err == nil {
		f.size = int64(len(data))
	}
	return err
}

func (f *WeblensFile) WriteAt(data []byte, seekLoc int64) error {
	if f.IsDir() {
		return fmt.Errorf("attempt to write to directory")
	}

	realFile, err := os.OpenFile(f.String(), os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	wroteLen, err := realFile.WriteAt(data, seekLoc)
	if err == nil {
		f.size += int64(wroteLen)
	}

	return err
}

func (f *WeblensFile) Append(data []byte) error {
	if f.IsDir() {
		return fmt.Errorf("attempt to write to directory")
	}
	realFile, err := os.OpenFile(f.String(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	wroteLen, err := realFile.Write(data)
	if err == nil {
		f.size += int64(wroteLen)
	}
	return err
}

func (f *WeblensFile) ReadDir() error {
	if !f.IsDir() || !f.Exists() {
		return fmt.Errorf("invalid file to read dir")
	}
	entries, err := os.ReadDir(f.absolutePath)
	if err != nil {
		return err
	}

	f.verifyChildren()

	for _, file := range entries {
		childPath := filepath.Join(f.absolutePath, file.Name())
		childId := util.GlobbyHash(8, GuaranteeRelativePath(childPath))

		f.childLock.Lock()
		if _, ok := f.children[childId]; ok {
			f.childLock.Unlock()
			continue
		}
		f.childLock.Unlock()

		tmpChild := WeblensFile{
			id:           childId,
			absolutePath: childPath,
			isDir:        boolPointer(file.IsDir()),
			filename:     file.Name(),
			owner:        f.Owner(),
		}

		err := FsTreeInsert(&tmpChild, f)
		if err != nil {
			switch err.(type) {
			case alreadyExists:
			default:
				return err
			}
		}
	}

	return nil
}

func (f *WeblensFile) GetChild(childId string) (child *WeblensFile) {
	if f.children == nil {
		return
	}

	f.childLock.Lock()
	child = f.children[childId]
	f.childLock.Unlock()

	return
}

func (f *WeblensFile) GetChildren() []*WeblensFile {
	if !f.IsDir() {
		return []*WeblensFile{}
	}

	f.verifyChildren()

	f.childLock.Lock()
	defer f.childLock.Unlock()
	return util.MapToSlicePure(f.children)
}

func (f *WeblensFile) AddChild(child *WeblensFile) {
	if !f.IsDir() {
		util.Error.Println("Attempting to add child to non-directory")
		return
	}

	f.verifyChildren()

	f.childLock.Lock()
	if f.children[child.Id()] != nil {
		f.childLock.Unlock()
		util.Warning.Printf("Child (%s) of %s already exists", child.Filename(), f.Filename())
		return
	}
	f.children[child.Id()] = child
	f.childLock.Unlock()
}

func (f *WeblensFile) GetChildrenInfo() []FileInfo {
	f.childLock.Lock()
	defer f.childLock.Unlock()
	childs := util.MapToSlicePure(f.children)

	err := loadManyMedias(childs)
	if err != nil {
		util.DisplayError(err)
		// Don't return here, the loading is only to speed things up
		// but an error here isn't fatal
	}

	childrenInfo := util.Map(childs, func(file *WeblensFile) FileInfo {
		info, err := file.FormatFileInfo()
		if err != nil {
			info.Id = "R"
		}
		return info
	})

	return childrenInfo
}

func (f *WeblensFile) GetParent() *WeblensFile {
	if f.parent == nil {
		util.Debug.Println("Returning parent as nil from f.GetParent")
	}

	return f.parent
}

func (f *WeblensFile) CreateSelf() error {
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

func (f *WeblensFile) CanUserAccess(username string) bool {
	// Is owner or is shared into the file (guest)
	if f.Owner() == username || slices.Contains(f.Guests(), username) {
		return true
	}
	return false
}

func (f *WeblensFile) FormatFileInfo() (formattedInfo FileInfo, err error) {
	if f == nil {
		return formattedInfo, fmt.Errorf("cannot get file info of nil wf")
	}

	if dirIgnore[f.Filename()] {
		return formattedInfo, fmt.Errorf("filename in blocklist")
	}

	var imported bool = true
	var m *Media

	if !f.IsDir() {
		m, err = f.GetMedia()
		if err != nil {
			imported = false
		}

		if m != nil {
			// Copy so we don't clear the thumbnail on the real one
			tmp := *m
			m = &tmp
			imported = m.imported
		}
	}

	var size int64
	size, err = f.Size()
	if err != nil {
		util.DisplayError(err)
		return
	}

	modTime := f.ModTime()
	if modTime.IsZero() {
		return formattedInfo, fmt.Errorf("failed to get modtime")
	}

	var displayable bool
	var friendlyName string
	mType, _ := f.GetMediaType()
	if mType != nil {
		displayable = mType.IsDisplayable
		friendlyName = mType.FriendlyName
	}

	shares, err := fddb.getWormholes(f)
	if err != nil {
		return
	}

	pathString := GuaranteeRelativePath(f.absolutePath)
	pathString = strings.Replace(pathString, "/"+f.Owner()+"/"+".user_trash", "TRASH", 1)
	pathString = strings.Replace(pathString, "/"+f.Owner(), "HOME", 1)

	formattedInfo = FileInfo{
		Id:               f.Id(),
		Imported:         imported,
		Displayable:      displayable,
		IsDir:            f.IsDir(),
		Modifiable:       f.Filename() != ".user_trash",
		Size:             size,
		ModTime:          modTime,
		Filename:         f.Filename(),
		ParentFolderId:   f.GetParent().Id(),
		FileFriendlyName: friendlyName,
		Owner:            f.Owner(),
		PathFromHome:     pathString,
		MediaData:        m,
		Shares:           shares,
		Children:         util.Map(f.GetChildren(), func(wf *WeblensFile) string { return wf.Id() }),
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
func (f *WeblensFile) RecursiveMap(fn func(*WeblensFile)) {
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
func (f *WeblensFile) LeafMap(fn func(*WeblensFile)) {
	if !f.IsDir() {
		return
	}

	children := f.GetChildren()

	childrenWithChildren := util.Filter(children, func(c *WeblensFile) bool { return c.children != nil && len(c.children) != 0 })
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
func (f *WeblensFile) BubbleMap(fn func(*WeblensFile)) {
	if f == nil || f.Id() == "0" || f.Id() == "1" || f.Id() == "2" {
		return
	}
	fn(f)

	parent := f.GetParent()
	parent.BubbleMap(fn)
}

func (f *WeblensFile) MarshalJSON() ([]byte, error) {
	size, err := f.Size()
	if err != nil {
		return nil, err
	}

	m := marshalableWF{Id: f.Id(), AbsolutePath: f.absolutePath, Filename: f.Filename(), Owner: f.Owner(), ParentFolderId: f.parent.Id(), Guests: f.Guests(), Size: size, IsDir: f.IsDir()}
	return json.Marshal(m)
}

func (f *WeblensFile) AddTask(t Task) {
	if f.tasksLock == nil {
		f.tasksLock = &sync.Mutex{}
	}

	f.tasksLock.Lock()
	f.tasksUsing = util.AddToSet(f.tasksUsing, []Task{t})
	f.tasksLock.Unlock()
}

func (f *WeblensFile) GetTasks() []Task {
	return f.tasksUsing
}

func (f *WeblensFile) RemoveTask(tId string) (exists bool) {
	if f.tasksLock == nil {
		f.tasksLock = &sync.Mutex{}
	}

	f.tasksLock.Lock()
	f.tasksUsing, _, exists = util.YoinkFunc(f.tasksUsing, func(f Task) bool { return f.TaskId() == tId })
	f.tasksLock.Unlock()
	return
}

// Private

func (f *WeblensFile) recompSize(c ...BroadcasterAgent) (size int64, err error) {
	origSize := f.size
	if f.IsDir() {
		children := f.GetChildren()
		util.Map(children, func(w *WeblensFile) int { s, _ := w.Size(); size += s; return 0 })
	} else {
		stat, err := os.Stat(f.absolutePath)
		if err != nil {
			return 0, err
		}
		size = stat.Size()
	}

	// util.Debug.Println(f.String(), "is now", size)

	if origSize != size {
		f.size = size
		if len(c) == 0 {
			c = append(c, globalCaster)
		}
		util.Each(c, func(c BroadcasterAgent) { c.PushFileUpdate(f) })
	}

	return
}

func (f *WeblensFile) removeChild(childId string) {
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

func (f *WeblensFile) verifyChildren() {
	if f.children == nil {
		f.childLock = &sync.Mutex{}
		f.childLock.Lock()
		f.children = map[string]*WeblensFile{}
		f.childLock.Unlock()
	}
}
