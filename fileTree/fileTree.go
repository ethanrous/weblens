package fileTree

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/websocket"

	"github.com/ethrousseau/weblens/api/types"
)

type FileTreeImpl struct {
	fMap           map[FileId]*WeblensFile
	fsTreeLock     sync.RWMutex
	journalService JournalService

	rootPath string

	root         *WeblensFile
	delDirectory *WeblensFile
}

func boolPointer(b bool) *bool {
	return &b
}

var RootDirIds = []FileId{"MEDIA", "TMP", "CACHE", "TAKEOUT", "EXTERNAL", "CONTENT_LINKS"}

func NewFileTree(rootPath, rootAlias string) FileTree {

	if _, err := os.Stat(rootPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(rootPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	root := &WeblensFile{
		id:       "ROOT",
		parent:   nil,
		filename: rootAlias,
		isDir:    boolPointer(true),

		childrenMap:  map[string]*WeblensFile{},
		absolutePath: rootPath,
		portablePath: WeblensFilepath{
			base: rootAlias,
			ext:  "",
		},
	}

	root.size.Store(-1)

	tree := &FileTreeImpl{
		fMap:     map[FileId]*WeblensFile{root.id: root},
		rootPath: rootPath,
		root:     root,
	}

	root.tree = tree

	return tree
}

func (ft *FileTreeImpl) InitMediaRoot(hashCaster websocket.BroadcasterAgent) error {
	InstanceService.AddLoading("filesystem")
	sw := internal.NewStopwatch("Filesystem")

	_, err := ft.NewRoot("TMP", "tmp", internal.GetTmpDir(), ft.userService.GetRootUser(), nil)
	if err != nil {
		return err
	}
	_, err = ft.NewRoot("TAKEOUT", "takeout", internal.GetTakeoutDir(), ft.userService.GetRootUser(), nil)
	if err != nil {
		return err
	}
	// externalRoot, err := ft.NewRoot("EXTERNAL", "External", "", "EXTERNAL", nil)
	// if err != nil {
	// 	return err
	// }
	cacheRoot, err := ft.NewRoot("CACHE", "Cache", internal.GetCacheDir(), ft.userService.GetRootUser(), nil)
	if err != nil {
		return err
	}
	contentRoot, err := ft.NewRoot(
		"CONTENT_LINKS", ".content", filepath.Join(ft.GetRoot().GetAbsPath(), ".content"),
		ft.userService.GetRootUser(),
		ft.GetRoot(),
	)
	if err != nil {
		return err
	}

	err = ft.SetDelDirectory(ft.Get("CONTENT_LINKS"))
	if err != nil {
		return err
	}

	sw.Lap("Set roots")

	sw.Lap("Get + sort lifetimes")

	if !ft.GetRoot().Exists() {
		err = ft.GetRoot().CreateSelf()
		if err != nil {
			return err
		}
	}

	if !contentRoot.Exists() {
		err = contentRoot.CreateSelf()
		if err != nil {
			return err
		}
	}

	if !cacheRoot.Exists() {
		err = contentRoot.CreateSelf()
		if err != nil {
			return err
		}
	}

	sw.Lap("Check roots exist")

	users, err := ft.userService.GetAll()
	if err != nil {
		return err
	}

	fileEvent := NewFileEvent()
	hashTaskPool := types.SERV.WorkerPool.NewTaskPool(false, nil)

	for _, u := range users {
		var homeDir *WeblensFile
		if homeDir, err = ft.GetRoot().GetChild(u.GetUsername().String()); err != nil {
			homeDir = ft.NewFile(ft.GetRoot(), u.GetUsername().String(), true, u)
			err = ft.Add(homeDir)
			if err != nil {
				return err
			}
		}

		err = u.SetHomeFolder(homeDir)
		if err != nil {
			return err
		}
		if !homeDir.Exists() {
			err = homeDir.CreateSelf()
			if err != nil {
				return err
			}
		}
		err = importFilesRecursive(homeDir, fileEvent, hashTaskPool, hashCaster)
		if err != nil {
			return err
		}

		var trashDir *WeblensFile
		trashDir, err = homeDir.GetChild(".user_trash")
		if err != nil {
			trashDir = ft.NewFile(homeDir, ".user_trash", true, u)
			if ft.Get(trashDir.ID()) == nil {
				if !trashDir.Exists() {
					err = trashDir.CreateSelf()
					if err != nil {
						return err
					}
				}
				err = ft.Add(trashDir)
				if err != nil {
					return err
				}
			}
		}

		err = u.SetTrashFolder(trashDir)
		if err != nil {
			return err
		}
	}

	sw.Lap("Load users home directories")

	err = importFilesRecursive(contentRoot, fileEvent, hashTaskPool, hashCaster)
	if err != nil {
		return err
	}

	err = importFilesRecursive(cacheRoot, fileEvent, hashTaskPool, hashCaster)
	if err != nil {
		return err
	}

	hashTaskPool.SignalAllQueued()
	hashTaskPool.AddCleanup(
		func() {
			err := ft.GetJournal().LogEvent(fileEvent)
			if err != nil {
				wlog.Error.Println(err)
			}
			InstanceService.RemoveLoading("filesystem")
		},
	)

	sw.Lap("Load roots")

	// for _, path := range util.GetExternalPaths() {
	// 	continue
	// 	if path == "" {
	// 		continue
	// 	}
	// 	s, err := os.Stat(path)
	// 	if err != nil {
	// 		panic(fmt.Sprintf("Could not find external path: %s", path))
	// 	}
	// 	extF := ft.NewFile(externalRoot, filepath.Base(path), s.IsDir(), nil)
	// 	err = ft.Add(extF)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	sw.Lap("Load external files")

	if InstanceService.GetLocal().IsCore() {
		// Compute size for the whole tree, and ensure children are loaded while we're at it.
		err = ft.ResizeDown(ft.GetRoot())
		if err != nil {
			return err
		}

		err = cacheRoot.LeafMap(
			func(wf *WeblensFile) error {
				_, err = wf.Size()
				return err
			},
		)
		if err != nil {
			return err
		}

		// if externalRoot.GetParent() != ft.GetRoot() {
		// 	err = externalRoot.LeafMap(
		// 		func(wf *WeblensFile) error {
		// 			_, err = wf.Size()
		// 			return err
		// 		},
		// 	)
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		sw.Lap("Compute Sizes")
	}

	for _, sh := range types.SERV.ShareService.GetAllShares() {
		switch sh.GetShareType() {
		case weblens.FileShare:
			sharedFile := types.SERV.FileTree.Get(FileId(sh.GetItemId()))
			if sharedFile != nil {
				err := sharedFile.SetShare(sh)
				if err != nil {
					return err
				}
			} else {
				wlog.Warning.Println("Ignoring possibly no longer existing file in share init")
			}
		}
	}

	sw.Lap("Link file shares")

	// err = ClearTempDir(ft)
	// if err != nil {
	// 	panic(err)
	// }
	// sw.Lap("Clear tmp dir")
	//
	// err = ClearTakeoutDir(ft)
	// if err != nil {
	// 	panic(err)
	// }
	// sw.Lap("Clear takeout dir")

	files, err := ft.GetAllFiles()
	if err != nil {
		return err
	}

	lifetimes := ft.GetJournal().GetAllLifetimes()
	ltMap := map[FileId]types.Lifetime{}
	for _, lt := range lifetimes {
		if lt.GetLatestFileId() == "" {
			if lt.GetLatestAction().GetActionType() != types.FileDelete {
				wlog.Error.Println("Skipping lifetime with no latest id, that has not been marked as delted")
			}
			continue
		}
		if _, ok := ltMap[lt.GetLatestFileId()]; ok {
			wlog.Warning.Println("Already have fileid in lifetimes", lt.GetLatestFileId())
			continue
		}
		ltMap[lt.GetLatestFileId()] = lt
	}
	for _, file := range files {
		delete(ltMap, file.ID())
	}
	if len(ltMap) != 0 {
		wlog.Error.Println("Leftover lifetimes: ", ltMap)
	}
	sw.Lap("Check for dangling lifetimes")
	sw.Stop()
	sw.PrintResults(false)

	return nil
}

func (ft *FileTreeImpl) NewFile(
	parent *WeblensFile, filename string, isDir bool,
	owner types.User,
) *WeblensFile {
	if owner == nil {
		owner = parent.Owner()
	}
	newFile := &WeblensFile{
		tree:     ft,
		parent: parent,
		filename: filename,
		isDir:    boolPointer(isDir),
		owner:    owner,
		childrenMap: map[string]*WeblensFile{},

		size: atomic.Int64{},
	}

	newFile.size.Store(-1)
	return newFile
}

func (ft *FileTreeImpl) NewRoot(
	id FileId, filename, absPath string, owner types.User,
	parent *WeblensFile,
) (*WeblensFile, error) {

	f := &WeblensFile{
		id:       id,
		filename: filename,
		owner:    owner,

		isDir:        boolPointer(true),
		absolutePath: absPath,

		childrenMap: map[string]*WeblensFile{},
	}

	if parent == nil {
		f.parent = f
	} else {
		f.parent = parent
	}

	ft.addInternal(f.ID(), f)
	return f, nil
}

func (ft *FileTreeImpl) addInternal(id FileId, f *WeblensFile) {
	realF := f
	if realF.tree == nil {
		realF.tree = ft
	}
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()

	// Do not use .ID() inside critical section, as it may need to use the locks
	ft.fMap[id] = f
	if realF.id == "ROOT" {
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
		return werror.NewWeblensError("trying to add a nil file to file tree")
	}
	if ft.has(f.ID()) {
		err := werror.NewWeblensError(
			fmt.Sprintf(
				"key collision on attempt to insert to filesystem tree: %s. "+
					"Existing file is at %s, new file is at %s", f.ID(), ft.Get(f.ID()).GetAbsPath(), f.GetAbsPath(),
			),
		)
		return err
	}

	// if !f.IsDir() && f.GetContentId() == "" && f.size.Load() != 0 {
	// 	wlog.Warning.Println("Adding file to tree with no contentId")
	// }

	if slices.Contains(IgnoreFilenames, f.Filename()) {
		return nil
	}

	ft.addInternal(f.ID(), f)
	err := f.GetParent().AddChild(f)
	if err != nil {
		return err
	}

	// Add system files to the map, but don't journal or push updates for them
	if f.Owner() == ft.userService.GetRootUser() {
		return nil
	}

	if f.IsDir() {
		err = ft.journalService.WatchFolder(f)
		if err != nil {
			return err
		}
	} else {
		// err = ft.ResizeUp(f, types.SERV.Caster)
		// if err != nil {
		// 	util.ErrTrace(err)
		// }
	}

	return nil
}

func (ft *FileTreeImpl) Del(fId FileId) (err error) {
	f := ft.Get(fId)

	// If the file does not already have an id, generating the id can lock the file tree,
	// so we must do that outside of the lock here to avoid deadlock
	// f.ID()

	realF := f

	if !ft.has(realF.id) {
		wlog.Warning.Println("Tried to remove key not in FsTree", f.ID())
		return types.ErrNoFile(realF.id)
	}

	err = realF.parent.removeChild(f)
	if err != nil {
		return
	}

	var tasks []types.Task

	deleteEvent := NewFileEvent()

	err = f.RecursiveMap(
		func(file *WeblensFile) error {
			t := file.GetTask()
			if t != nil {
				tasks = append(tasks, t)
				t.Cancel()
			}
			if f.GetShare() != nil {
				err := types.SERV.ShareService.Del(f.GetShare().GetShareId())
				if err != nil {
					wlog.ErrTrace(err)
				}
			}

			if !file.IsDir() {
				contentId := file.GetContentId()
				m := types.SERV.MediaRepo.Get(contentId)
				if m != nil {
					err := m.RemoveFile(file)
					if err != nil {
						return err
					}
				}

				// possibly bug: when a single delete event is deleting multiple of the same content
				// id you get a collision in the content folder

				backupF, _ := ft.delDirectory.GetChild(string(contentId))
				if contentId != "" && backupF == nil {
					backupF = ft.NewFile(ft.delDirectory, string(contentId), false, nil)
					err = ft.Add(backupF)
					if err != nil {
						return err
					}
					err = os.Rename(file.GetAbsPath(), backupF.GetAbsPath())
					if err != nil {
						return err
					}
				} else {
					err := os.Remove(file.GetAbsPath())
					if err != nil {
						return err
					}
				}
			}

			ft.deleteInternal(file.ID())
			deleteEvent.NewDeleteAction(file.ID())

			return nil
		},
	)

	if err != nil {
		return err
	}

	for _, t := range tasks {
		t.Wait()
	}

	if f.IsDir() {
		err = os.RemoveAll(f.GetAbsPath())
		if err != nil {
			return werror.Wrap(err)
		}
	}

	err = ft.journalService.LogEvent(deleteEvent)
	if err != nil {
		return werror.Wrap(err)
	}

	return
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
	c ...websocket.BufferedBroadcasterAgent,
) error {
	if f.Owner() != newParent.Owner() {
		return types.ErrIllegalFileMove
	}
	if !newParent.IsDir() {
		return types.ErrDirectoryRequired
	}

	if (newFilename == "" || newFilename == f.Filename()) && newParent == f.GetParent() {
		wlog.Warning.Println("Exiting early from move without updates")
		return nil
	}

	if newFilename == "" {
		newFilename = f.Filename()
	}

	newAbsPath := filepath.Join(newParent.GetAbsPath(), newFilename)

	if !overwrite {
		// Check if the file at the destination exists already
		if _, err := os.Stat(newAbsPath); err == nil {
			return types.ErrFileAlreadyExists(newAbsPath)
		}
	}

	if !f.Exists() || !newParent.Exists() {
		return types.ErrNoFile(f.ID())
	}

	var allTasks []types.Task
	err := f.RecursiveMap(
		func(w *WeblensFile) error {
			t := w.GetTask()
			if t != nil {
				allTasks = append(allTasks, t)
				t.Cancel()
			}

			return nil
		},
	)
	if err != nil {

		return err
	}

	for _, t := range allTasks {
		t.Wait()
	}

	oldAbsPath := f.GetAbsPath()
	oldParent := f.GetParent()

	// Point of no return //

	// Overwrite filename
	f.filename = newFilename

	// Disable casters because we need to wait to move the files before stat-ing them for the updates
	for cast := range slices.Values(c) {
		cast.DisableAutoFlush()
	}

	// Sync file tree with new move, including f and all of its children.
	err = f.RecursiveMap(
		func(w *WeblensFile) error {
			preFile := w.Copy()

			if f == w {
				w.setParentInternal(newParent)
			}

			err := preFile.GetParent().removeChild(w)
			if err != nil {
				return werror.Wrap(err)
			}
			w.setModTime(time.Now())

			ft.deleteInternal(w.id)

			w.setIdInternal("")
			w.absolutePath = filepath.Join(w.GetParent().GetAbsPath(), w.Filename())
			if w.IsDir() {
				w.absolutePath += "/"
			}

			ft.addInternal(w.ID(), w)

			err = w.GetParent().AddChild(w)
			if err != nil {
				return err
			}

			if w.IsDisplayable() {
				m := types.SERV.MediaRepo.Get(preFile.GetContentId())
				if m != nil {
					// Add new file first so the mediaService doesn't trash the media if if there is only 1 file
					err = m.AddFile(w)
					if err != nil {
						return werror.Wrap(err)
					}
					err = m.RemoveFile(preFile)
					if err != nil {
						return werror.Wrap(err)
					}
				}
			}

			if w.GetShare() != nil {
				w.GetShare().SetItemId(string(w.GetContentId()))
			}

			for cast := range slices.Values(c) {
				cast.PushFileMove(preFile, w)
			}
			event.NewMoveAction(preFile.ID(), w)

			return nil
		},
	)

	if err != nil {
		return werror.Wrap(err)
	}

	err = os.Rename(oldAbsPath, newAbsPath)
	if err != nil {
		return werror.Wrap(err)
	}

	err = ft.resizeMultiple(oldParent, f.GetParent(), internal.SliceConvert[websocket.BroadcasterAgent](c)...)
	if err != nil {
		return werror.Wrap(err)
	}

	for cast := range slices.Values(c) {
		cast.AutoFlushEnable()
	}

	return nil
}

// Size gets the number of files loaded into weblens.
// This does not lock the file tree, and therefore
// cannot be trusted to be microsecond accurate, but
// it's quite close
func (ft *FileTreeImpl) Size() int {
	return len(ft.fMap)
}

func (ft *FileTreeImpl) Touch(
	parentFolder *WeblensFile, newFileName string, detach bool, owner types.User,
	c ...websocket.BroadcasterAgent,
) (*WeblensFile, error) {
	f := ft.NewFile(parentFolder, newFileName, false, nil)
	f.detached = detach
	e := ft.Get(f.ID())
	if e != nil || f.Exists() {
		return e, types.ErrFileAlreadyExists(f.GetAbsPath())
	}

	err := f.CreateSelf()
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

	if owner != nil {
		f.owner = owner
	}

	if len(c) != 0 {
		c[0].PushFileCreate(f)
	}

	return f, nil
}

// MkDir creates a new dir as a child of parentFolder named newDirName. If the dir already exists,
// it will be returned along with a ErrDirAlreadyExists error.
func (ft *FileTreeImpl) MkDir(
	parentFolder *WeblensFile, newDirName string, event *FileEvent, c ...websocket.BroadcasterAgent,
) (*WeblensFile, error) {
	d := ft.NewFile(parentFolder, newDirName, true, nil)

	if d.Exists() {
		existingFile := ft.Get(d.ID())

		if existingFile == nil {
			err := ft.Add(d)
			if err != nil {
				return d, err
			}
			existingFile = d
		}

		return existingFile, types.ErrDirAlreadyExists
	}

	d.size.Store(0)

	err := ft.Add(d)
	if err != nil {
		return d, err
	}

	err = d.CreateSelf()
	if err != nil {
		return d, err
	}

	if len(c) != 0 {
		c[0].PushFileCreate(d)
	} else {
		wlog.Error.Println("MkDir: No caster")
	}

	event.NewCreateAction(d)

	return d, nil
}

// AttachFile takes a detached file when it is ready to be inserted to the tree, and attaches it
func (ft *FileTreeImpl) AttachFile(f *WeblensFile, c ...websocket.BroadcasterAgent) error {
	if ft.Get(f.ID()) != nil {
		return types.ErrFileAlreadyExists(f.GetAbsPath())
	}

	tmpPath := filepath.Join("/tmp/", f.Filename())
	tmpFile, err := os.Open(tmpPath)
	if err != nil {
		return werror.Wrap(err)
	}

	// Mask the file create event. Since we cannot insert
	// into the tree if the file is not present already,
	// we must move the file first, which would create an
	// insert event, which we wish to do manually below.
	// Masking prevents the automatic insert event.
	// history.MaskEvent(types.FileCreate, f.GetAbsPath())

	destFile, err := os.Create(f.GetAbsPath())
	if err != nil {
		return werror.Wrap(err)
	}

	_, err = io.Copy(destFile, tmpFile)
	if err != nil {
		return werror.Wrap(err)
	}

	err = ft.Add(f)
	if err != nil {
		return werror.Wrap(err)
	}

	if len(c) != 0 {
		c[0].PushFileCreate(f)
	}

	return os.Remove(tmpPath)
}

func (ft *FileTreeImpl) CreateHomeFolder(u types.User) (*WeblensFile, error) {
	mediaRoot := ft.GetRoot()
	event := NewFileEvent()
	homeDir, err := ft.MkDir(mediaRoot, strings.ToLower(string(u.GetUsername())), event)
	if err != nil && errors.Is(err, types.ErrDirAlreadyExists) {

	} else if err != nil {
		return nil, err
	}

	homeDir.SetOwner(u)

	_, err = ft.MkDir(homeDir, ".user_trash", event)
	if err != nil {
		return homeDir, err
	}

	err = ft.GetJournal().LogEvent(event)
	if err != nil {
		return homeDir, werror.Wrap(err)
	}

	return homeDir, nil
}

func (ft *FileTreeImpl) GenerateFileId(absPath string) FileId {
	fileHash := FileId(internal.GlobbyHash(8, FilepathFromAbs(absPath)))
	return fileHash
}

func (ft *FileTreeImpl) GetRoot() *WeblensFile {
	if ft.root == nil {
		wlog.Error.Println("GetRoot called on fileTree with nil root")
	}
	return ft.root
}

func (ft *FileTreeImpl) SetRoot(root *WeblensFile) {
	ft.root = root
}

func (ft *FileTreeImpl) GetJournal() JournalService {
	return ft.journalService
}

func (ft *FileTreeImpl) SetJournal(j JournalService) {
	ft.journalService = j
}

func (ft *FileTreeImpl) SetDelDirectory(del *WeblensFile) error {
	// if del == nil {
	// 	return types.ErrNoFile(del.ID())
	// }

	ft.delDirectory = del
	return nil
}

// Util

func (ft *FileTreeImpl) resizeMultiple(old, new *WeblensFile) (err error) {
	// Check if either of the files are a parent of the other
	oldIsParent := strings.HasPrefix(old.GetAbsPath(), new.GetAbsPath())
	newIsParent := strings.HasPrefix(new.GetAbsPath(), old.GetAbsPath())

	if oldIsParent || !newIsParent {
		err = ft.ResizeUp(old)
		if err != nil {
			return
		}
	}

	if newIsParent || !oldIsParent {
		err = ft.ResizeUp(new)
		if err != nil {
			return
		}
	}

	return
}

func (ft *FileTreeImpl) ResizeUp(f *WeblensFile) error {
	return f.BubbleMap(
		func(w *WeblensFile) error {
			return w.LoadStat()
		},
	)
}

func (ft *FileTreeImpl) ResizeDown(f *WeblensFile) error {
	return f.LeafMap(
		func(w *WeblensFile) error {
			_, err := w.recomputeSize()
			return err
		},
	)
}

func (ft *FileTreeImpl) GetAllFiles() ([]*WeblensFile, error) {
	ft.fsTreeLock.RLock()
	defer ft.fsTreeLock.RUnlock()
	return slices.Collect(maps.Values(ft.fMap)), nil
}

var IgnoreFilenames = []string{
	".DS_Store",
}

func importFilesRecursive(
	f *WeblensFile, fileEvent *FileEvent,
	hashTaskPool types.TaskPool, hashCaster websocket.BroadcasterAgent,
) error {
	var toLoad = []*WeblensFile{f}
	for len(toLoad) != 0 {
		var fileToLoad *WeblensFile

		// Pop from slice of files to load
		fileToLoad, toLoad = toLoad[0], toLoad[1:]
		if slices.Contains(IgnoreFilenames, fileToLoad.Filename()) || (fileToLoad.Filename() == "."+
			"content" && fileToLoad.ID() != "CONTENT_LINKS") {
			continue
		}

		if InstanceService.GetLocal().ServerRole() == BackupServer {
			if fileToLoad.IsDir() {
				toLoad = append(toLoad, fileToLoad.GetChildren()...)
			} else if fileToLoad.GetContentId() == "" {
				fileToLoad.SetContentId(types.SERV.FileTree.GetJournal().GetLifetimeByFileId(fileToLoad.ID()).GetContentId())
			}
			continue
		}

		if fileToLoad.Owner() != f.tree.userService.GetRootUser() {
			lt := fileToLoad.GetTree().GetJournal().GetLifetimeByFileId(fileToLoad.ID())
			if lt == nil {
				fileSize, err := fileToLoad.Size()
				if err != nil {
					return werror.Wrap(err)
				}

				if fileToLoad.GetContentId() == "" && !fileToLoad.IsDir() && fileSize != 0 {
					hashTaskPool.HashFile(
						fileToLoad,
						hashCaster,
					).SetPostAction(
						func(result types.TaskResult) {
							if result["contentId"] != nil {
								fileToLoad.SetContentId(result["contentId"].(weblens.ContentId))
								fileEvent.NewCreateAction(fileToLoad)
							} else {
								wlog.Error.Println("Failed to generate contentId for", fileToLoad.Filename())
							}

						},
					)
				} else if fileToLoad.IsDir() || fileSize == 0 {
					fileEvent.NewCreateAction(fileToLoad)
				}

			} else {
				fileToLoad.SetContentId(lt.GetContentId())
			}
		}

		if !slices.Contains(RootDirIds, fileToLoad.ID()) {
			if types.SERV.FileTree.Get(fileToLoad.ID()) == nil {
				err := types.SERV.FileTree.Add(fileToLoad)
				if err != nil {
					return err
				}
			} else {
				// util.Debug.Println("Skipping insert of a file already present in the tree:", fileToLoad.ID())
				// continue
			}
		}

		if fileToLoad.IsDir() {
			children, err := fileToLoad.ReadDir()
			if err != nil {
				return err
			}
			toLoad = append(toLoad, children...)
		}
	}

	return nil
}

type FileTree interface {
	Size() int

	Get(id FileId) *WeblensFile
	Add(file *WeblensFile) error
	Del(id FileId) error

	Move(f, newParent *WeblensFile, newFilename string, overwrite bool, event *FileEvent) error

	Touch(parentFolder *WeblensFile, newFileName string, detach bool) (WeblensFile, error)
	MkDir(parentFolder *WeblensFile, newDirName string, event FileEvent) (WeblensFile, error)

	AttachFile(f *WeblensFile) error

	GetRoot() *WeblensFile
	SetRoot(WeblensFile)
	GetJournal() JournalService
	SetJournal(JournalService)
	GenerateFileId(absPath string) FileId
	NewFile(parent *WeblensFile, filename string, isDir bool) *WeblensFile
	GetAllFiles() ([]WeblensFile, error)

	NewRoot(id FileId, filename, absPath string, parent *WeblensFile,) (WeblensFile, error)
	SetDelDirectory(WeblensFile) error

	InitMediaRoot() error

	ResizeUp(WeblensFile) error
	ResizeDown(WeblensFile) error
}
