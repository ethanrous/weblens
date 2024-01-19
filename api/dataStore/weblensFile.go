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

	// Return pointer to copy
	return &c
}

// Retrieve the error field set internally in
// the file descriptor. If multiple errors have
// occurred, only the most recent will be shown.
// A non-nil f.err shall never be reset to nil
func (f *WeblensFile) Err() error {
	if f == nil {
		return fmt.Errorf("error check on nil file descriptor")
	}
	return f.err
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
	if f.id == "" {
		f.id = util.HashOfString(8, GuaranteeRelativePath(f.absolutePath))
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

	m, err = fddb.GetMediaByFile(f, false)
	if err != nil {
		return
	}

	f.media = m

	return
}

func (f *WeblensFile) SetMedia(m *Media) error {
	if f.media != nil && f.media.FileHash != "" {
		if f.media != m {
			return errors.New("attempted to reasign media on file descriptor that already has media")
		}
		return nil
	}
	f.media = m

	return nil
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
			f.err = fmt.Errorf("failed to get stat of filepath checking if IsDir: %s", err)
			return false
		}
		f.isDir = boolPointer(stat.IsDir())
	}
	return *f.isDir
}

func (f *WeblensFile) ModTime() *time.Time {
	stat, err := os.Stat(f.absolutePath)
	if err != nil {
		util.DisplayError(err)
		return nil
	}
	modTime := stat.ModTime()
	return &modTime
}

func (f *WeblensFile) Size() (int64, error) {
	if f.size != 0 {
		return f.size, nil
	}

	if !f.assertExists() {
		return 0, fmt.Errorf("getting size of file that does not exist")
	}

	return f.recompSize()
}

func (f *WeblensFile) Read() (*os.File, error) {
	if f.IsDir() {
		return nil, fmt.Errorf("attempt to read from directory")
	}

	return os.Open(f.absolutePath)
}

func (f *WeblensFile) Write(data []byte) error {
	if f.IsDir() {
		return fmt.Errorf("attempt to write to directory")
	}
	err := os.WriteFile(f.absolutePath, data, 0660)
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
		tmpChild := WeblensFile{}
		tmpChild.absolutePath = filepath.Join(f.absolutePath, file.Name())
		f.childLock.Lock()
		if _, ok := f.children[tmpChild.Id()]; ok {
			f.childLock.Unlock()
			continue
		}
		f.childLock.Unlock()
		tmpChild.filename = file.Name()
		tmpChild.owner = f.Owner()
		err := FsTreeInsert(&tmpChild, f)
		if err != nil {
			switch err.(type) {
			case alreadyExists:
			default:
				return err
			}
		}
	}
	if f.isDir == nil {
		f.isDir = boolPointer(true)
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
		return nil
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
	childrenInfo := util.MapToSliceMutate(f.children, func(_ string, file *WeblensFile) FileInfo {
		info, err := file.FormatFileInfo()
		if err != nil {
			info.Id = "R"
		}
		return info
	})
	return childrenInfo
}

func (f *WeblensFile) GetParent() *WeblensFile {
	if f.id == "0" {
		util.Error.Println(fmt.Errorf("cannot get parent of media root"))
		return f
	}
	if f.parent == nil {
		util.Error.Println("Returning parent as nil from file GetParent")
	}

	return f.parent
}

func (f *WeblensFile) CreateSelf() error {
	if f.Exists() {
		return fmt.Errorf("directory already exists")
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

func (f *WeblensFile) UserCanAccess(username string) bool {
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
	if modTime == nil {
		return formattedInfo, fmt.Errorf("failed to get modtime")
	}

	displayable, _ := f.IsDisplayable()

	formattedInfo = FileInfo{Id: f.Id(), Imported: imported, Displayable: displayable, IsDir: f.IsDir(), Modifiable: f.Filename() != ".user_trash", Size: size, ModTime: *modTime, Filename: f.Filename(), ParentFolderId: f.GetParent().Id(), MediaData: m, Owner: f.Owner()}

	return formattedInfo, nil
}

// Recursively perform fn on f and all children of f. This takes a "Depth first" approach.
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

// Perform fn on f and all parents of f, not including the media root.
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

// Private

func (f *WeblensFile) recompSize() (size int64, err error) {
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

	if origSize != size {
		f.size = size
		caster.PushFileUpdate(f)
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

func (f *WeblensFile) assertExists() bool {
	if f.Exists() {
		return true
	}
	FsTreeRemove(f)
	return false
}

func (f *WeblensFile) verifyChildren() {
	if f.children == nil {
		f.childLock = &sync.Mutex{}
		f.childLock.Lock()
		f.children = map[string]*WeblensFile{}
		f.childLock.Unlock()
	}
}
