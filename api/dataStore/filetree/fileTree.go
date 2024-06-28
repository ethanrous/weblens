package filetree

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type fileTree struct {
	fMap           map[types.FileId]types.WeblensFile
	fsTreeLock     *sync.Mutex
	journalService types.JournalService

	db types.DatabaseService

	root         types.WeblensFile
	delDirectory types.WeblensFile
}

func boolPointer(b bool) *bool {
	return &b
}

func NewFileTree() types.FileTree {
	return &fileTree{
		fMap:       make(map[types.FileId]types.WeblensFile),
		fsTreeLock: &sync.Mutex{},
	}
}

func (ft *fileTree) Init(db types.DatabaseService) error {
	ft.db = db

	return nil
}

func (ft *fileTree) NewFile(parent types.WeblensFile, filename string, isDir bool) types.WeblensFile {
	return &weblensFile{
		tree:     parent.GetTree().(*fileTree),
		parent:   parent.(*weblensFile),
		filename: filename,
		isDir:    boolPointer(isDir),
		owner:    parent.Owner(),

		tasksLock: &sync.Mutex{},
		childLock: &sync.Mutex{},
		children:  []*weblensFile{},

		size: -1,
	}
}

func (ft *fileTree) AddRoot(r types.WeblensFile) error {
	if !r.IsDir() {
		return types.ErrDirectoryRequired
	}
	self := r.(*weblensFile)
	// Root directory must be its own parent
	self.parent = self
	ft.addInternal(r.ID(), r)

	return nil
}

func (ft *fileTree) NewRoot(
	id types.FileId, filename, absPath string, owner types.User,
	parent types.WeblensFile,
) (types.WeblensFile, error) {

	f := &weblensFile{
		id:       id,
		filename: filename,
		owner:    owner,

		isDir:        boolPointer(true),
		absolutePath: absPath,

		childLock: &sync.Mutex{},
		children:  []*weblensFile{},
	}

	if parent == nil {
		f.parent = f
	} else {
		f.parent = parent.(*weblensFile)
	}

	ft.addInternal(f.ID(), f)
	return f, nil
}

func (ft *fileTree) addInternal(id types.FileId, f types.WeblensFile) {
	realF := f.(*weblensFile)
	if realF.tree == nil {
		realF.tree = ft
	}
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()

	// Do not use .ID() inside critical section, as it may need to use the locks
	ft.fMap[id] = f
}

func (ft *fileTree) deleteInternal(id types.FileId) {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()
	delete(ft.fMap, id)
}

func (ft *fileTree) has(id types.FileId) bool {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()
	_, ok := ft.fMap[id]
	return ok
}

func (ft *fileTree) Add(f types.WeblensFile) error {
	if ft.has(f.ID()) {
		return types.NewWeblensError(
			fmt.Sprintf(
				"key collision on attempt to insert to filesystem tree: %s. "+
					"Existing file is at %s, new file is at %s", f.ID(), ft.Get(f.ID()).GetAbsPath(), f.GetAbsPath(),
			),
		)
	}

	if slices.Contains(IgnoreFilenames, f.Filename()) {
		return nil
	}

	ft.addInternal(f.ID(), f)
	err := f.GetParent().AddChild(f)
	if err != nil {
		return err
	}

	// Add system files to the map, but don't journal or push updates for them
	if f.Owner() == dataStore.WeblensRootUser {
		return nil
	}

	if f.IsDir() {
		err = ft.journalService.WatchFolder(f)
		if err != nil {
			return err
		}
	} else {
		err = ft.ResizeUp(f, types.SERV.Caster)
		if err != nil {
			util.ErrTrace(err)
		}
	}

	types.SERV.Caster.PushFileCreate(f)

	return nil
}

func (ft *fileTree) Del(fId types.FileId) (err error) {
	f := ft.Get(fId)

	// If the file does not already have an id, generating the id can lock the file tree,
	// so we must do that outside of the lock here to avoid deadlock
	// f.ID()

	realF := f.(*weblensFile)

	if !ft.has(realF.id) {
		util.Warning.Println("Tried to remove key not in FsTree", f.ID())
		return types.ErrNoFile
	}

	err = realF.parent.removeChild(f)
	if err != nil {
		return
	}

	var tasks []types.Task

	err = f.RecursiveMap(
		func(file types.WeblensFile) error {
			t := file.GetTask()
			if t != nil {
				tasks = append(tasks, t)
				t.Cancel()
			}
			if f.GetShare() != nil {
				err := types.SERV.ShareService.Del(f.GetShare().GetShareId())
				if err != nil {
					util.ErrTrace(err)
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
					backupF = ft.NewFile(ft.delDirectory, string(contentId), false)
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

			return nil
		},
	)

	if err != nil {
		util.ErrTrace(err)
		return
	}

	for _, t := range tasks {
		t.Wait()
	}

	if f.IsDir() {
		err = os.RemoveAll(f.GetAbsPath())
		if err != nil {
			util.ErrTrace(err)
			return
		}
	}

	return
}

func (ft *fileTree) Get(fileId types.FileId) types.WeblensFile {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()
	return ft.fMap[fileId]
}

func (ft *fileTree) Move(
	f, newParent types.WeblensFile, newFilename string, overwrite bool, c ...types.BufferedBroadcasterAgent,
) error {
	if f.Owner() != newParent.Owner() {
		return types.ErrIllegalFileMove
	}
	if !newParent.IsDir() {
		return types.ErrDirectoryRequired
	}

	if (newFilename == "" || newFilename == f.Filename()) && newParent == f.GetParent() {
		util.Warning.Println("Exiting early from move without updates")
		return nil
	}

	if newFilename == "" {
		newFilename = f.Filename()
	}

	newAbsPath := filepath.Join(newParent.GetAbsPath(), newFilename)

	if !overwrite {
		// Check if the file at the destination exists already
		if _, err := os.Stat(newAbsPath); err == nil {
			return types.ErrFileAlreadyExists
		}
	}

	if !f.Exists() || !newParent.Exists() {
		return types.ErrNoFile
	}

	var allTasks []types.Task
	err := f.RecursiveMap(
		func(w types.WeblensFile) error {
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
	f.(*weblensFile).filename = newFilename

	// Disable casters because we need to wait to move the files before stat-ing them for the updates
	util.Each(c, func(c types.BufferedBroadcasterAgent) { c.DisableAutoFlush() })

	// Sync file tree with new move, including f and all of its children.
	err = f.RecursiveMap(
		func(w types.WeblensFile) error {
			preFile := w.Copy()

			realW := w.(*weblensFile)
			if f == w {
				realW.parent = newParent.(*weblensFile)
			}

			err := preFile.GetParent().(*weblensFile).removeChild(w)
			if err != nil {
				return err
			}

			// The file no longer has an id, so generating the id will lock the file tree,
			// we must do that outside the lock below to avoid deadlock
			// w.ID()
			// _, err = w.Len()
			// if err != nil {
			// 	return err
			// }

			ft.deleteInternal(realW.id)

			realW.id = ""
			realW.absolutePath = filepath.Join(w.GetParent().GetAbsPath(), w.Filename())
			if realW.IsDir() {
				realW.absolutePath += "/"
			}

			ft.addInternal(realW.ID(), w)

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
						return err
					}
					err = m.RemoveFile(preFile)
					if err != nil {
						return err
					}
				}
			}

			if w.GetShare() != nil {
				w.GetShare().SetItemId(string(w.GetContentId()))
			}

			util.Each(c, func(c types.BufferedBroadcasterAgent) { c.PushFileMove(preFile, w) })
			return nil
		},
	)

	if err != nil {
		return err
	}

	err = os.Rename(oldAbsPath, newAbsPath)
	if err != nil {
		util.ErrTrace(err)
		return err
	}

	err = ft.resizeMultiple(oldParent, f.GetParent(), util.SliceConvert[types.BroadcasterAgent](c)...)
	if err != nil {
		return err
	}

	util.Each(c, func(c types.BufferedBroadcasterAgent) { c.AutoFlushEnable() })

	return nil
}

// Size gets the number of files loaded into weblens.
// This does not lock the file tree, and therefore
// cannot be trusted to be microsecond accurate, but
// it's quite close
func (ft *fileTree) Size() int {
	return len(ft.fMap)
}

func (ft *fileTree) Touch(
	parentFolder types.WeblensFile, newFileName string, detach bool, owner types.User, c ...types.BroadcasterAgent,
) (types.WeblensFile, error) {
	f := ft.NewFile(parentFolder, newFileName, false).(*weblensFile)
	f.detached = detach
	e := ft.Get(f.ID())
	if e != nil || f.Exists() {
		return e, types.ErrFileAlreadyExists
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
func (ft *fileTree) MkDir(
	parentFolder types.WeblensFile, newDirName string, c ...types.BroadcasterAgent,
) (types.WeblensFile, error) {
	d := ft.NewFile(parentFolder, newDirName, true).(*weblensFile)

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

	d.size = 0

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
	}

	return d, nil
}

// AttachFile takes a detached file when it is ready to be inserted to the tree, and attaches it
func (ft *fileTree) AttachFile(f types.WeblensFile, c ...types.BroadcasterAgent) error {
	if ft.Get(f.ID()) != nil {
		return types.ErrFileAlreadyExists
	}

	tmpPath := filepath.Join("/tmp/", f.Filename())
	tmpFile, err := os.Open(tmpPath)
	if err != nil {
		return err
	}

	// Mask the file create event. Since we cannot insert
	// into the tree if the file is not present already,
	// we must move the file first, which would create an
	// insert event, which we wish to do manually below.
	// Masking prevents the automatic insert event.
	// history.MaskEvent(types.FileCreate, f.GetAbsPath())

	destFile, err := os.Create(f.GetAbsPath())
	if err != nil {
		return err
	}

	_, err = io.Copy(destFile, tmpFile)
	if err != nil {
		return err
	}

	err = ft.Add(f)
	if err != nil {
		return err
	}

	if len(c) != 0 {
		c[0].PushFileCreate(f)
	}

	return os.Remove(tmpPath)

}

func (ft *fileTree) GenerateFileId(absPath string) types.FileId {
	fileHash := types.FileId(util.GlobbyHash(8, FilepathFromAbs(absPath)))
	return fileHash
}

func (ft *fileTree) GetRoot() types.WeblensFile {
	if ft.root == nil {
		util.Error.Println("GetRoot called on fileTree with nil root")
	}
	return ft.root
}

func (ft *fileTree) SetRoot(root types.WeblensFile) {
	ft.root = root
}

func (ft *fileTree) GetJournal() types.JournalService {
	return ft.journalService
}

func (ft *fileTree) SetJournal(j types.JournalService) {
	ft.journalService = j
}

func (ft *fileTree) SetDelDirectory(del types.WeblensFile) error {
	if del == nil {
		return types.ErrNoFile
	}

	ft.delDirectory = del
	return nil
}

// Util

func (ft *fileTree) resizeMultiple(old, new types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	// Check if either of the files are a parent of the other
	oldIsParent := strings.HasPrefix(old.GetAbsPath(), new.GetAbsPath())
	newIsParent := strings.HasPrefix(new.GetAbsPath(), old.GetAbsPath())

	if oldIsParent || !newIsParent {
		err = ft.ResizeUp(old, c...)
		if err != nil {
			return
		}
	}

	if newIsParent || !oldIsParent {
		err = ft.ResizeUp(new, c...)
		if err != nil {
			return
		}
	}

	return
}

func (ft *fileTree) ResizeUp(f types.WeblensFile, c ...types.BroadcasterAgent) error {
	return f.BubbleMap(
		func(w types.WeblensFile) error {
			return w.(*weblensFile).LoadStat(c...)
		},
	)
}

func (ft *fileTree) ResizeDown(f types.WeblensFile, c ...types.BroadcasterAgent) error {
	return f.LeafMap(
		func(w types.WeblensFile) error {
			return w.(*weblensFile).LoadStat(c...)
		},
	)
}

var IgnoreFilenames = []string{
	".DS_Store",
}
