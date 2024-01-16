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

	"github.com/ethrousseau/weblens/api/util"
)

/*	These here are all the methods on a WeblensFileDescriptor,
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

func (f *WeblensFileDescriptor) Copy() *WeblensFileDescriptor {
	// Copy values of wfd struct
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
func (f *WeblensFileDescriptor) Err() error {
	if f == nil {
		return fmt.Errorf("error check on nil file descriptor")
	}
	return f.err
}

// This function does not return an error, and instead
// will set the err feild internally. To check this error,
// call .Err() afterwards.
func (f *WeblensFileDescriptor) Id() string {
	if f == nil {
		panic("Tried to get Id of nil file")
	}
	if f.id == "" {
		f.id = util.HashOfString(8, GuaranteeRelativePath(f.absolutePath))
	}
	return f.id
}

func (f *WeblensFileDescriptor) Filename() string {
	if f.filename == "" {
		f.filename = filepath.Base(f.absolutePath)
	}
	return f.filename
}

// Returns string of absolute path to file
func (f *WeblensFileDescriptor) String() string {
	if f.IsDir() && !strings.HasSuffix(f.absolutePath, "/") {
		f.absolutePath = f.absolutePath + "/"
	}
	return f.absolutePath
}

func (f *WeblensFileDescriptor) Owner() string {
	if f == nil {
		panic("attempt to get owner on nil wfd")
	}
	if f.owner == "" {
		f.owner = f.parent.Owner()
	}
	return f.owner
}

func (f *WeblensFileDescriptor) Guests() []string {
	if f.guests == nil {
		f.guests = fddb.getFileGuests(f)
	}
	return f.guests
}

func (f *WeblensFileDescriptor) GetMedia() (*Media, error) {
	if f.media != nil {
		return f.media, nil
	}
	if f.isDir != nil && *f.isDir {
		return nil, fmt.Errorf("cannot get media of directory")
	}

	m, err := fddb.GetMediaByFile(f, false)
	if err != nil || m.FileHash == "" {
		return nil, fmt.Errorf("failed to get media from WFD: %s", err)
	}

	f.media = &m

	return f.media, nil
}

func (f *WeblensFileDescriptor) SetMedia(m *Media) error {
	if f.media != nil && f.media.FileHash != "" {
		if f.media != m {
			return errors.New("attempted to reasign media on file descriptor that already has media")
		}
		return nil
	}
	f.media = m

	return nil
}

func (f *WeblensFileDescriptor) Exists() bool {
	_, err := os.Stat(f.absolutePath)
	return err == nil
}

func (f *WeblensFileDescriptor) IsDir() bool {
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

func (f *WeblensFileDescriptor) ModTime() time.Time {
	stat, _ := os.Stat(f.absolutePath)
	return stat.ModTime()
}

func (f *WeblensFileDescriptor) Size() (int64, error) {
	if f.size != 0 {
		return f.size, nil
	}

	if !f.IsDir() {
		if !f.assertExists() {
			return 0, fmt.Errorf("getting size of file that does not exist")
		}
		return f.recompSize()
	}

	return f.size, nil
}

func (f *WeblensFileDescriptor) Read() (*os.File, error) {
	if *f.isDir {
		return nil, fmt.Errorf("attempt to read directory as file")
	}
	osFile, err := os.Open(f.absolutePath)
	return osFile, err
}

func (f *WeblensFileDescriptor) Write(data []byte) error {
	if *f.isDir {
		return fmt.Errorf("attempt to write to directory")
	}
	err := os.WriteFile(f.absolutePath, data, 0660)
	return err
}

func (f *WeblensFileDescriptor) ReadDir() error {
	if !f.IsDir() || !f.Exists() {
		return fmt.Errorf("invalid file to read dir")
	}
	entries, err := os.ReadDir(f.absolutePath)
	if err != nil {
		return err
	}

	for _, file := range entries {
		if file.Name() == ".DS_Store" {
			continue
		}
		tmpChild := WeblensFileDescriptor{}
		tmpChild.absolutePath = filepath.Join(f.absolutePath, file.Name())
		f.childLock.Lock()
		if _, ok := f.children[tmpChild.Id()]; ok {
			f.childLock.Unlock()
			continue
		}
		f.childLock.Unlock()
		tmpChild.filename = file.Name()
		tmpChild.owner = f.Owner()
		FsTreeInsert(&tmpChild, f.id)
	}
	if f.isDir == nil {
		f.isDir = boolPointer(true)
	}

	return nil
}

func (f *WeblensFileDescriptor) GetChildren() []*WeblensFileDescriptor {
	if f.children == nil {
		f.childLock = &sync.Mutex{}
		f.childLock.Lock()
		f.children = map[string]*WeblensFileDescriptor{}
		f.childLock.Unlock()
	}

	if len(f.children) == 0 {
		err := f.ReadDir()
		if err != nil {
			util.DisplayError(err)
			return []*WeblensFileDescriptor{}
		}
	}

	f.childLock.Lock()
	defer f.childLock.Unlock()
	return util.MapToSlicePure(f.children)
}

func (f *WeblensFileDescriptor) AddChild(child *WeblensFileDescriptor) {
	if f.children == nil {
		f.childLock = &sync.Mutex{}
		f.childLock.Lock()
		f.children = map[string]*WeblensFileDescriptor{}
		f.childLock.Unlock()
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

func (f *WeblensFileDescriptor) GetChildrenInfo() []FileInfo {
	childrenInfo := util.MapToSliceMutate(f.children, func(_ string, file *WeblensFileDescriptor) FileInfo {
		info, err := file.FormatFileInfo()
		if err != nil {
			info.Id = "R"
		}
		return info
	})
	return childrenInfo
}

func (f *WeblensFileDescriptor) GetParent() *WeblensFileDescriptor {
	if f.id == "0" {
		util.Error.Println(fmt.Errorf("cannot get parent of media root"))
		return f
	}
	if f.parent == nil {
		util.Error.Println("Returning parent as nil from file GetParent")
	}

	return f.parent
}

func (f *WeblensFileDescriptor) CreateSelf() error {
	if f.Exists() {
		return fmt.Errorf("directory already exists")
	}
	if f.isDir == nil {
		return fmt.Errorf("cannot create self with nil self type")
	}

	var err error
	if *f.isDir {
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

func (f *WeblensFileDescriptor) UserCanAccess(username string) bool {
	if f.Owner() == username || slices.Contains(f.Guests(), username) {
		return true
	}
	return false
}

func (f *WeblensFileDescriptor) FormatFileInfo() (formattedInfo FileInfo, err error) {
	if f == nil {
		return formattedInfo, fmt.Errorf("cannot get file info of nil wfd")
	}

	if !dirIgnore[f.Filename()] {
		var imported bool = true

		var m Media

		if !f.IsDir() {
			mptr, err := f.GetMedia()
			if err != nil {
				imported = false
			}

			if mptr != nil {
				m = *mptr
			}
			m.Thumbnail64 = ""
			m.MediaType = f.getMediaType()
		}

		if f.Err() != nil {
			util.DisplayError(f.err)
			return formattedInfo, f.err
		}
		var size int64
		if size, err = f.Size(); err != nil {
			util.DisplayError(err)
			return formattedInfo, err
		}

		formattedInfo = FileInfo{Id: f.Id(), Imported: imported, IsDir: f.IsDir(), Modifiable: f.Filename() != ".user_trash", Size: size, ModTime: f.ModTime(), Filename: f.Filename(), ParentFolderId: f.GetParent().Id(), MediaData: m, Owner: f.Owner()}
	} else {
		return formattedInfo, fmt.Errorf("filename in blocklist")
	}
	return formattedInfo, nil
}

func (f *WeblensFileDescriptor) RecursiveMap(fn func(*WeblensFileDescriptor)) {
	fn(f)
	if !f.IsDir() {
		return
	}

	children := f.GetChildren()

	for _, c := range children {
		c.RecursiveMap(fn)
	}
}

func (f *WeblensFileDescriptor) BubbleMap(fn func(*WeblensFileDescriptor)) {
	if f == nil || f.Id() == "0" {
		return
	}
	fn(f)

	parent := f.GetParent()
	parent.BubbleMap(fn)
}

func (f *WeblensFileDescriptor) MarshalJSON() ([]byte, error) {
	size, err := f.Size()
	if err != nil {
		return nil, err
	}

	m := marshalableWFD{Id: f.Id(), AbsolutePath: f.absolutePath, Filename: f.Filename(), Owner: f.Owner(), ParentFolderId: f.parent.Id(), Guests: f.Guests(), Size: size, IsDir: f.IsDir()}
	return json.Marshal(m)
}

// Private

func (f *WeblensFileDescriptor) recompSize() (size int64, err error) {
	if f.IsDir() {
		// return size, fmt.Errorf("attempt to recompute size of directory")
		children := f.GetChildren()

		util.Map(children, func(w *WeblensFileDescriptor) int { size += w.size; return 0 })
	} else {
		var stat fs.FileInfo
		stat, err = os.Stat(f.absolutePath)
		if err != nil {
			return
		}
		size = stat.Size()
	}

	f.size = size
	return
}

func (f *WeblensFileDescriptor) removeChild(childId string) {
	if f.children == nil {
		util.Debug.Println("attempt to remove child on wfd where children map is nil")
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

func (f *WeblensFileDescriptor) assertExists() bool {
	if f.Exists() {
		return true
	}
	FsTreeRemove(f)
	return false
}
