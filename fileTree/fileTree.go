package fileTree

import (
	"context"
	"errors"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ FileTree = (*FileTreeImpl)(nil)

type FileTreeImpl struct {
	fMap           map[FileId]*WeblensFile
	fsTreeLock     sync.RWMutex
	journalService JournalService

	rootPath  string
	rootAlias string

	root *WeblensFile
}

type MoveInfo struct {
	From *WeblensFile
	To   *WeblensFile
}

func boolPointer(b bool) *bool {
	return &b
}

func NewFileTree(rootPath, rootAlias string, hasher Hasher, journal JournalService) (FileTree, error) {
	sw := internal.NewStopwatch(rootAlias + " Filetree Init")

	if _, err := os.Stat(rootPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(rootPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	sw.Lap("Find or mkdir root directory")

	root := &WeblensFile{
		id:       "ROOT",
		parent:   nil,
		filename: filepath.Base(rootPath),
		isDir:    boolPointer(true),

		childrenMap:  map[string]*WeblensFile{},
		absolutePath: rootPath,
		portablePath: WeblensFilepath{
			rootAlias: rootAlias,
			relPath:   "",
		},
	}

	root.size.Store(-1)
	sw.Lap("Init root struct")

	tree := &FileTreeImpl{
		fMap:           map[FileId]*WeblensFile{root.id: root},
		rootPath:       rootPath,
		root:           root,
		journalService: journal,
		rootAlias:      rootAlias,
	}
	sw.Lap("Init tree struct")

	journal.SetFileTree(tree)

	internal.LabelThread(
		func(_ context.Context) {
			go journal.EventWorker()
		}, "", "Journal Worker",
	)

	internal.LabelThread(
		func(_ context.Context) {
			go journal.FileWatcher()
		}, "", "File Watcher",
	)

	sw.Lap("Launch workers")

	event := tree.GetJournal().NewEvent()
	err := tree.loadFromRoot(event, hasher)
	if err != nil {
		return nil, err
	}

	sw.Lap("Load files")

	if waiter, ok := hasher.(HashWaiter); ok {
		waiter.Wait()
	}

	sw.Lap("Wait for hashes")

	journal.LogEvent(event)
	sw.Lap("Log file event")
	sw.Stop()
	sw.PrintResults(false)

	return tree, nil
}

func (ft *FileTreeImpl) GetJournal() JournalService {
	return ft.journalService
}

func (ft *FileTreeImpl) SetJournal(j JournalService) {
	ft.journalService = j
}

func (ft *FileTreeImpl) addInternal(id FileId, f *WeblensFile) {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()

	// Do not use .ID() inside critical section, as it may need to use the locks
	ft.fMap[id] = f
	if f.id == "ROOT" {
		ft.root = f
	}
}

func (ft *FileTreeImpl) deleteInternal(id FileId) {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()
	delete(ft.fMap, id)
}

func (ft *FileTreeImpl) has(id FileId) bool {
	ft.fsTreeLock.RLock()
	defer ft.fsTreeLock.RUnlock()
	_, ok := ft.fMap[id]
	return ok
}

func (ft *FileTreeImpl) Add(f *WeblensFile) error {
	if f == nil {
		return werror.WithStack(errors.New("trying to add a nil file to file tree"))
	}
	if ft.has(f.ID()) {
		return werror.Errorf(
			"key collision on attempt to insert to filesystem tree: %s. "+
				"Existing file is at %s, new file is at %s", f.ID(), ft.Get(f.ID()).GetAbsPath(), f.GetAbsPath(),
		)
	}

	if slices.Contains(IgnoreFilenames, f.Filename()) {
		return nil
	}

	ft.addInternal(f.ID(), f)

	if f.parent == nil {
		parent := ft.Get(f.parentId)
		if parent == nil {
			return werror.Errorf("could not get parent of file to add")
		}
		f.setParentInternal(parent)
	}

	err := f.GetParent().AddChild(f)
	if err != nil {
		return err
	}

	if f.IsDir() {
		err = ft.journalService.WatchFolder(f)
		if err != nil {
			return err
		}
	}

	if f.getAbsPathInternal() == "" && f.GetPortablePath().relPath == "" {
		return werror.Errorf("Cannot add file to tree without abs path or portable path")
	} else if f.GetPortablePath().relPath == "" {
		portable, err := ft.AbsToPortable(f.getAbsPathInternal())
		if err != nil {
			return err
		}
		f.setPortable(portable)
	} else if f.getAbsPathInternal() == "" {
		abs, err := ft.PortableToAbs(f.GetPortablePath())
		if err != nil {
			return err
		}
		f.setAbsPath(abs)
	}

	return nil
}

func (ft *FileTreeImpl) Del(fId FileId, deleteEvent *FileEvent) ([]*WeblensFile, error) {
	f := ft.Get(fId)

	// If the file does not already have an id, generating the id can lock the file tree,
	// so we must do that outside of the lock here to avoid deadlock
	// f.ID()

	if !ft.has(f.id) {
		log.Warning.Println("Tried to remove key not in FsTree", f.ID())
		return nil, werror.ErrNoFileId(string(f.ID()))
	}

	err := f.GetParent().removeChild(f)
	if err != nil {
		return nil, err
	}

	var localDeleteEvent bool
	if deleteEvent == nil {
		deleteEvent = ft.GetJournal().NewEvent()
		localDeleteEvent = true
	}

	var deleted []*WeblensFile

	err = f.RecursiveMap(
		func(file *WeblensFile) error {
			deleted = append(deleted, file)
			// t := file.GetTask()
			// if t != nil {
			// 	tasks = append(tasks, t)
			// 	t.Cancel()
			// }
			// if f.GetShare() != nil {
			// 	err := types.SERV.ShareService.Del(f.GetShare().GetShareId())
			// 	if err != nil {
			// 		wlog.ErrTrace(err)
			// 	}
			// }

			// if !file.IsDir() {
			// 	contentId := file.GetContentId()
			// 	m := types.SERV.MediaRepo.Get(contentId)
			// 	if m != nil {
			// 		err := m.RemoveFile(file)
			// 		if err != nil {
			// 			return err
			// 		}
			// 	}

			// possibly bug: when a single delete event is deleting multiple of the same content
			// id you get a collision in the content folder
			// backupF, _ := ft.delDirectory.GetChild(string(contentId))
			// if contentId != "" && backupF == nil {
			// 	backupF = ft.NewFile(ft.delDirectory, string(contentId), false, nil)
			// 	err = ft.Add(backupF)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	err = os.Rename(file.GetAbsPath(), backupF.GetAbsPath())
			// 	if err != nil {
			// 		return err
			// 	}
			// } else {
			// 	err := os.Remove(file.GetAbsPath())
			// 	if err != nil {
			// 		return err
			// 	}
			// }
			// }

			ft.deleteInternal(file.ID())
			deleteEvent.NewDeleteAction(file.ID())

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	// if f.IsDir() {
	// }
	// err = os.RemoveAll(f.GetAbsPath())
	// if err != nil {
	// 	return nil, err
	// }

	if localDeleteEvent {
		ft.journalService.LogEvent(deleteEvent)
	}

	return deleted, nil
}

func (ft *FileTreeImpl) Get(fileId FileId) *WeblensFile {
	ft.fsTreeLock.RLock()
	defer ft.fsTreeLock.RUnlock()
	return ft.fMap[fileId]
}

func (ft *FileTreeImpl) GetChildren(f *WeblensFile) iter.Seq[*WeblensFile] {
	if f.childIds == nil {
		f.childIds = []FileId{}
	}
	return func(yield func(file *WeblensFile) bool) {
		for _, childId := range f.childIds {
			child := ft.Get(childId)
			if !yield(child) {
				return
			}
		}
	}
}

func (ft *FileTreeImpl) Move(
	f, newParent *WeblensFile, newFilename string, overwrite bool, event *FileEvent,
) ([]MoveInfo, error) {
	if newParent == nil {
		return nil, werror.WithStack(werror.ErrFileRequired)
	} else if !newParent.IsDir() {
		return nil, werror.WithStack(werror.ErrDirectoryRequired)
	} else if newFilename == "" {
		return nil, werror.WithStack(werror.ErrFilenameRequired)
	}

	if newFilename == f.Filename() && newParent == f.GetParent() {
		return nil, werror.WithStack(werror.ErrEmptyMove)
	}

	newAbsPath := filepath.Join(newParent.GetAbsPath(), newFilename)

	if !overwrite {
		// Check if the file at the destination exists already
		if _, err := os.Stat(newAbsPath); err == nil {
			return nil, werror.WithStack(werror.ErrFileAlreadyExists)
		}
	}

	if !f.Exists() || !newParent.Exists() {
		return nil, werror.ErrNoFileId(string(f.ID()))
	}

	oldAbsPath := f.GetAbsPath()

	// Point of no return //

	var hasExternalEvent bool
	if event == nil {
		hasExternalEvent = false
		event = ft.GetJournal().NewEvent()
	}

	// Sync file tree with new move, including f and all of its children.
	var moved []MoveInfo
	err := f.RecursiveMap(
		func(w *WeblensFile) error {
			preFile := w.Freeze()

			// Shift the root of the move operation to be a child of the new parent
			if f == w {
				err := preFile.GetParent().removeChild(w)
				if err != nil {
					return err
				}
				f.filename = newFilename
				w.setParentInternal(newParent)
				err = w.GetParent().AddChild(w)
				if err != nil {
					return err
				}
			}

			newChildAbsPath := filepath.Join(w.GetParent().GetAbsPath(), w.Filename())
			if w.IsDir() {
				newChildAbsPath += "/"
			}
			w.setAbsPath(newChildAbsPath)

			portable, err := ft.AbsToPortable(w.getAbsPathInternal())
			if err != nil {
				return err
			}
			w.setPortable(portable)

			event.NewMoveAction(preFile.ID(), w)

			moved = append(
				moved, MoveInfo{
					From: preFile,
					To:   w,
				},
			)

			w.modifiedNow()

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	err = os.Rename(oldAbsPath, newAbsPath)
	if err != nil {
		return nil, err
	}

	if !hasExternalEvent {
		ft.journalService.LogEvent(event)
	}

	return moved, nil
}

// Size gets the number of files loaded into weblens.
// This does not lock the file tree, and therefore
// cannot be trusted to be microsecond accurate, but
// it's quite close
func (ft *FileTreeImpl) Size() int {
	return len(ft.fMap)
}

func (ft *FileTreeImpl) Touch(parentFolder *WeblensFile, newFileName string, detach bool) (*WeblensFile, error) {
	absPath := filepath.Join(parentFolder.GetAbsPath(), newFileName)
	portable, err := ft.AbsToPortable(absPath)
	if err != nil {
		return nil, err
	}

	f := &WeblensFile{
		id:           ft.GenerateFileId(absPath),
		absolutePath: absPath,
		portablePath: portable,
		filename:     newFileName,
		isDir:        boolPointer(false),
		modifyDate:   time.Now(),
		parentId:     parentFolder.ID(),
		parent:       parentFolder,
		childrenMap:  map[string]*WeblensFile{},
		childIds:     []FileId{},
	}

	// TODO - convert to path
	// e := ft.Get(f.ID())
	// if e != nil || f.Exists() {
	// 	return e, werror.ErrFileAlreadyExists
	// }

	err = f.CreateSelf()
	if err != nil {
		return f, err
	}

	// Detach creates the file on the real filesystem,
	// but does not add it to the tree or journalService its creation
	if detach {
		return f, nil
	}

	err = ft.Add(f)
	if err != nil {
		return f, err
	}

	return f, nil
}

// MkDir creates a new dir as a child of parentFolder named newDirName. If the dir already exists,
// it will be returned along with a ErrDirAlreadyExists error.
func (ft *FileTreeImpl) MkDir(
	parentFolder *WeblensFile, newDirName string, event *FileEvent,
) (*WeblensFile, error) {
	if existingFile, _ := parentFolder.GetChild(newDirName); existingFile != nil {
		return existingFile, werror.ErrDirAlreadyExists
	}

	absPath := filepath.Join(parentFolder.GetAbsPath(), newDirName) + "/"

	d := &WeblensFile{
		id:           ft.GenerateFileId(absPath),
		absolutePath: absPath,
		portablePath: WeblensFilepath{},
		filename:     newDirName,
		isDir:        boolPointer(true),
		modifyDate:   time.Now(),
		parentId:     parentFolder.ID(),
		parent:       parentFolder,
		childrenMap:  map[string]*WeblensFile{},
		childIds:     []FileId{},
	}

	if d.Exists() {
		existingFile := ft.Get(d.ID())

		if existingFile == nil {
			err := ft.Add(d)
			if err != nil {
				return d, err
			}
			existingFile = d
		}

		return existingFile, werror.ErrDirAlreadyExists
	}

	d.size.Store(0)

	err := ft.Add(d)
	if err != nil {
		return nil, err
	}

	err = d.CreateSelf()
	if err != nil {
		return d, err
	}

	if event != nil {
		event.NewCreateAction(d)
	}

	return d, nil
}

// ReadDir reads the filesystem for files it does not yet have, adds them to the tree,
// and returns the newly added files
func (ft *FileTreeImpl) ReadDir(dir *WeblensFile) ([]*WeblensFile, error) {
	entries, err := os.ReadDir(dir.absolutePath)
	if err != nil {
		return nil, err
	}

	children := make([]*WeblensFile, 0, len(entries))
	for _, entry := range entries {
		if slices.Contains(IgnoreFilenames, entry.Name()) {
			continue
		}

		child, err := ft.importFromDirEntry(entry, dir)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	return children, nil
}

// // AttachFile takes a detached file when it is ready to be inserted to the tree, and attaches it
// func (ft *FileTreeImpl) AttachFile(f *WeblensFile) error {
// 	if ft.Get(f.ID()) != nil {
// 		return werror.ErrFileAlreadyExists
// 	}
//
// 	tmpPath := filepath.Join("/tmp/", f.Filename())
// 	tmpFile, err := os.Open(tmpPath)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Mask the file create event. Since we cannot insert
// 	// into the tree if the file is not present already,
// 	// we must move the file first, which would create an
// 	// insert event, which we wish to do manually below.
// 	// Masking prevents the automatic insert event.
// 	// history.MaskEvent(types.FileCreate, f.GetAbsPath())
//
// 	destFile, err := os.Create(f.GetAbsPath())
// 	if err != nil {
// 		return err
// 	}
//
// 	_, err = io.Copy(destFile, tmpFile)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = ft.Add(f)
// 	if err != nil {
// 		return err
// 	}
//
// 	return os.Remove(tmpPath)
// }

func (ft *FileTreeImpl) GetRoot() *WeblensFile {
	if ft.root == nil {
		log.Error.Println("GetRoot called on fileTree with nil root")
	}
	return ft.root
}

func (ft *FileTreeImpl) GenerateFileId(absPath string) FileId {
	return FileId(primitive.NewObjectID().Hex())
	fileHash := FileId(
		internal.GlobbyHash(
			8, NewFilePath(ft.GetRoot().GetAbsPath(), ft.GetRoot().Filename(), absPath),
		),
	)
	return fileHash
}

func (ft *FileTreeImpl) PortableToAbs(portable WeblensFilepath) (string, error) {
	if portable.RootName() != ft.rootAlias {
		return "", werror.Errorf(
			"fileTree.PortableToAbs: portable path rootAlias [%s] does not match tree rootAlias [%s]",
			portable.RootName(), ft.GetRoot().Filename(),
		)
	}

	return ft.GetRoot().GetAbsPath() + portable.RelativePath(), nil
	// return "", werror.NotImplemented("PortableToAbs")
}

func (ft *FileTreeImpl) AbsToPortable(absPath string) (WeblensFilepath, error) {
	if !strings.HasPrefix(absPath, ft.GetRoot().GetAbsPath()) {
		return WeblensFilepath{}, werror.Errorf(
			"fileTree.AbsToPortable: absPath [%s] does not match tree root prefix [%s]",
			absPath, ft.GetRoot().GetAbsPath(),
		)
	}

	return NewFilePath(ft.GetRoot().GetAbsPath(), ft.rootAlias, absPath), nil
}

var IgnoreFilenames = []string{
	".DS_Store",
}

func (ft *FileTreeImpl) loadFromRoot(event *FileEvent, hasher Hasher) error {
	lifetimesByPath := map[string]*Lifetime{}
	for _, lt := range ft.journalService.GetActiveLifetimes() {
		lifetimesByPath[lt.GetLatestFilePath()] = lt
	}

	toLoad, err := ft.ReadDir(ft.root)
	if err != nil {
		return err
	}

	for len(toLoad) != 0 {
		var fileToLoad *WeblensFile

		// Pop from slice of files to load
		fileToLoad, toLoad = toLoad[0], toLoad[1:]
		if slices.Contains(IgnoreFilenames, fileToLoad.Filename()) || (fileToLoad.Filename() == ".content") {
			continue
		}

		if activeLt, ok := lifetimesByPath[fileToLoad.GetPortablePath().ToPortable()]; ok {
			fileToLoad.setIdInternal(activeLt.ID())
			if !fileToLoad.IsDir() {
				fileToLoad.SetContentId(activeLt.ContentId)
			}
		} else {
			fileToLoad.setIdInternal(ft.GenerateFileId(fileToLoad.absolutePath))
			fileSize, err := fileToLoad.Size()
			if err != nil {
				return err
			}

			if !fileToLoad.IsDir() && fileSize != 0 {
				err = hasher.Hash(fileToLoad, event)
				if err != nil {
					return err
				}
			}
			event.NewCreateAction(fileToLoad)
		}

		err = ft.Add(fileToLoad)
		if err != nil {
			return err
		}

		if fileToLoad.IsDir() {
			children, err := ft.ReadDir(fileToLoad)
			if err != nil {
				return err
			}
			toLoad = append(toLoad, children...)
		}
	}

	return nil
}

func (ft *FileTreeImpl) importFromDirEntry(entry os.DirEntry, parent *WeblensFile) (*WeblensFile, error) {
	if parent == nil {
		return nil, werror.Errorf("Trying to add dirEntry with nil parent")
	}

	absPath := filepath.Join(parent.GetAbsPath(), entry.Name())
	if entry.IsDir() {
		absPath += "/"
	}
	portable, err := ft.AbsToPortable(absPath)
	if err != nil {
		return nil, err
	}

	info, err := entry.Info()
	if err != nil {
		return nil, err
	}

	f := &WeblensFile{
		id: "",
		absolutePath: absPath,
		portablePath: portable,
		filename:     entry.Name(),
		isDir:        boolPointer(info.IsDir()),
		modifyDate:   info.ModTime(),
		childrenMap:  map[string]*WeblensFile{},
		childIds:     []FileId{},
	}

	f.setParentInternal(parent)

	if !f.IsDir() {
		f.size.Store(info.Size())
	} else {
		f.size.Store(-1)
	}

	return f, nil
}

type FileTree interface {
	Get(id FileId) *WeblensFile
	GetRoot() *WeblensFile
	ReadDir(dir *WeblensFile) ([]*WeblensFile, error)
	Size() int

	GetJournal() JournalService
	SetJournal(JournalService)

	Add(file *WeblensFile) error
	Del(id FileId, deleteEvent *FileEvent) ([]*WeblensFile, error)
	Move(f, newParent *WeblensFile, newFilename string, overwrite bool, event *FileEvent) ([]MoveInfo, error)
	Touch(parentFolder *WeblensFile, newFileName string, detach bool) (*WeblensFile, error)
	MkDir(parentFolder *WeblensFile, newDirName string, event *FileEvent) (*WeblensFile, error)

	PortableToAbs(portable WeblensFilepath) (string, error)
	GenerateFileId(absPath string) FileId
}

type Hasher interface {
	Hash(file *WeblensFile, event *FileEvent) error
}

type HashWaiter interface {
	Hasher
	Wait()
}
