package fileTree

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/internal/werror"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ FileTree = (*FileTreeImpl)(nil)

type FileTreeImpl struct {
	journal Journal

	fMap map[FileId]*WeblensFileImpl

	root *WeblensFileImpl

	rootPath  string
	rootAlias string

	fsTreeLock sync.RWMutex
	log        *zerolog.Logger
}

type MoveInfo struct {
	From *WeblensFileImpl
	To   *WeblensFileImpl
}

func boolPointer(b bool) *bool {
	return &b
}

func NewFileTree(rootPath, rootAlias string, journal Journal, doFileDiscovery bool, log *zerolog.Logger) (FileTree, error) {
	if journal == nil {
		return nil, werror.Errorf("Got nil journal trying to create new FileTree")
	}

	if rootPath[len(rootPath)-1] != '/' {
		rootPath = rootPath + "/"
	}

	if !filepath.IsAbs(rootPath) {
		return nil, werror.Errorf("rootPath must be an absolute path: %s", rootPath)
	}

	if _, err := os.Stat(rootPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(rootPath, os.ModePerm)
		if err != nil {
			return nil, werror.WithStack(err)
		}
	} else if err != nil {
		return nil, werror.WithStack(err)
	}

	root := &WeblensFileImpl{
		id:       "ROOT",
		parent:   nil,
		filename: filepath.Base(rootPath),
		isDir:    boolPointer(true),

		childrenMap:  map[string]*WeblensFileImpl{},
		absolutePath: rootPath,
		portablePath: WeblensFilepath{
			rootAlias: rootAlias,
			relPath:   "",
		},
	}

	root.size.Store(-1)

	tree := &FileTreeImpl{
		fMap:      map[FileId]*WeblensFileImpl{root.id: root},
		rootPath:  rootPath,
		root:      root,
		journal:   journal,
		rootAlias: rootAlias,
		log:       log,
	}

	event := tree.GetJournal().NewEvent()
	if event.journal == nil {
		event = nil
	}
	err := tree.loadFromRoot(event, doFileDiscovery)
	if err != nil {
		return nil, err
	}

	journal.LogEvent(event)

	return tree, nil
}

func (ft *FileTreeImpl) GetJournal() Journal {
	return ft.journal
}

func (ft *FileTreeImpl) SetJournal(j Journal) {
	ft.journal = j
}

func (ft *FileTreeImpl) addInternal(id FileId, f *WeblensFileImpl) {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()

	// log.Trace().Func(func(e *zerolog.Event) {e.Msgf("Adding %s (%s) to file tree", f.filename, f.id)})

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

func (ft *FileTreeImpl) Add(f *WeblensFileImpl) error {
	if f == nil {
		return werror.WithStack(errors.New("trying to add a nil file to file tree"))
	}

	if ft.has(f.ID()) {
		return werror.Errorf(
			"key collision on attempt to insert to filesystem tree: %s. "+
				"Existing file is at %s, new file is at %s", f.ID(), ft.Get(f.ID()).AbsPath(), f.AbsPath(),
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
		err = ft.journal.WatchFolder(f)
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

func (ft *FileTreeImpl) Remove(id FileId) ([]*WeblensFileImpl, error) {
	f := ft.Get(id)
	if f == nil {
		return nil, werror.WithStack(werror.ErrNoFile.WithArg(id))
	}

	if f == ft.root {
		return nil, werror.Errorf("cannot delete root directory")
	}

	// If the file does not already have an id, generating the id can lock the file tree,
	// so we must do that outside of the lock here to avoid deadlock
	// f.ID()

	if !ft.has(f.id) {
		ft.log.Warn().Msgf("Tried to remove key not in FsTree [%s]", f.ID())
		return nil, werror.WithStack(werror.ErrNoFile.WithArg(f.ID()))
	}

	err := f.GetParent().removeChild(f)
	if err != nil {
		return nil, err
	}

	var deleted []*WeblensFileImpl
	_ = f.RecursiveMap(
		func(file *WeblensFileImpl) error {
			deleted = append(deleted, file)
			ft.deleteInternal(file.ID())

			return nil
		},
	)

	return deleted, nil
}

func (ft *FileTreeImpl) Delete(id FileId, event *FileEvent) error {
	f := ft.Get(id)
	if f == nil {
		return werror.WithStack(werror.ErrNoFile)
	}

	if f == ft.root {
		return werror.WithStack(werror.ErrRootFolder)
	}

	if f.IsDir() && len(f.GetChildren()) != 0 {
		return werror.Errorf("cannot delete non-empty directory")
	}

	_, err := ft.Remove(id)
	if err != nil {
		return err
	}

	err = os.Remove(f.getAbsPathInternal())
	if err != nil {
		return werror.WithStack(err)
	}

	_, err = event.NewDeleteAction(f.ID())
	if err != nil {
		return err
	}

	return nil
}

func (ft *FileTreeImpl) Get(fileId FileId) *WeblensFileImpl {
	ft.fsTreeLock.RLock()
	defer ft.fsTreeLock.RUnlock()
	return ft.fMap[fileId]
}

func (ft *FileTreeImpl) Move(
	f, newParent *WeblensFileImpl, newFilename string, overwrite bool, event *FileEvent,
) ([]MoveInfo, error) {
	if newParent == nil {
		return nil, werror.WithStack(werror.ErrNilFile)
	} else if !newParent.IsDir() {
		return nil, werror.WithStack(werror.ErrDirectoryRequired)
	} else if newFilename == "" {
		return nil, werror.WithStack(werror.ErrFilenameRequired)
	}

	if newFilename == f.Filename() && newParent == f.GetParent() {
		return nil, werror.WithStack(werror.ErrEmptyMove)
	}

	newAbsPath := filepath.Join(newParent.AbsPath(), newFilename)

	if !overwrite {
		// Check if the file at the destination exists already
		if _, err := os.Stat(newAbsPath); err == nil {
			return nil, werror.WithStack(werror.ErrFileAlreadyExists)
		}
	}

	if !f.Exists() || !newParent.Exists() {
		return nil, werror.ErrNoFile
	}

	oldAbsPath := f.AbsPath()

	// Point of no return //

	var hasExternalEvent bool
	if event == nil {
		event = ft.GetJournal().NewEvent()
	} else {
		hasExternalEvent = true
	}

	// Sync file tree with new move, including f and all of its children.
	var moved []MoveInfo
	err := f.RecursiveMap(
		func(w *WeblensFileImpl) error {
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

			newChildAbsPath := filepath.Join(w.GetParent().AbsPath(), w.Filename())
			if w.IsDir() {
				newChildAbsPath += "/"
			}
			w.setAbsPath(newChildAbsPath)

			portable, err := ft.AbsToPortable(w.getAbsPathInternal())
			if err != nil {
				return err
			}
			w.setPortable(portable)

			_, err = event.NewMoveAction(preFile.ID(), w)
			if err != nil {
				return err
			}

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

	if overwrite {
		err = os.Remove(newAbsPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, werror.WithStack(err)
		}
	}

	err = os.Rename(oldAbsPath, newAbsPath)
	if err != nil {
		return nil, err
	}

	if !hasExternalEvent {
		ft.journal.LogEvent(event)
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

func (ft *FileTreeImpl) Touch(parentFolder *WeblensFileImpl, newFileName string, event *FileEvent, data ...[]byte) (
	*WeblensFileImpl, error,
) {
	childPath := parentFolder.GetPortablePath().Child(newFileName, false)
	absPath, err := ft.PortableToAbs(childPath)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	f := &WeblensFileImpl{
		id:           ft.GenerateFileId(),
		absolutePath: absPath,
		portablePath: childPath,
		filename:     newFileName,
		isDir:        boolPointer(false),
		modifyDate:   time.Now(),
		parentId:     parentFolder.ID(),
		parent:       parentFolder,
		childrenMap:  map[string]*WeblensFileImpl{},
		childIds:     []FileId{},
	}

	err = f.CreateSelf()
	if err != nil {
		return f, err
	}

	if len(data) > 0 {
		_, err = f.Write(data[0])
		if err != nil {
			return f, err
		}
	}

	err = ft.Add(f)
	if err != nil {
		return f, err
	}

	if event != nil {
		event.NewCreateAction(f)
	} else {
		// event = ft.journal.NewEvent()
		// event.NewCreateAction(f)
		// ft.journal.LogEvent(event)
		// err = event.Wait()
		// if err != nil {
		// 	return f, err
		// }
	}

	return f, nil
}

// MkDir creates a new dir as a child of parentFolder named newDirName. If the dir already exists,
// it will be returned along with a ErrDirAlreadyExists error.
func (ft *FileTreeImpl) MkDir(
	parentFolder *WeblensFileImpl, newDirName string, event *FileEvent,
) (*WeblensFileImpl, error) {
	if existingFile, _ := parentFolder.GetChild(newDirName); existingFile != nil {
		return existingFile, werror.WithStack(werror.ErrDirAlreadyExists.WithArg(parentFolder.AbsPath() + newDirName))
	}

	absPath := filepath.Join(parentFolder.AbsPath(), newDirName) + "/"

	d := &WeblensFileImpl{
		id:           ft.GenerateFileId(),
		absolutePath: absPath,
		portablePath: WeblensFilepath{},
		filename:     newDirName,
		isDir:        boolPointer(true),
		modifyDate:   time.Now(),
		parentId:     parentFolder.ID(),
		parent:       parentFolder,
		childrenMap:  map[string]*WeblensFileImpl{},
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

		return existingFile, werror.WithStack(werror.ErrDirAlreadyExists)
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
	} else {
		event = ft.journal.NewEvent()
		event.NewCreateAction(d)
		ft.journal.LogEvent(event)
		err = event.Wait()
		if err != nil {
			return d, err
		}
	}

	return d, nil
}

func (ft *FileTreeImpl) SetRootAlias(alias string) error {
	if ft.Size() != 1 {
		return werror.Errorf("Cannot set root alias on non-empty file tree")
	}

	ft.rootAlias = alias
	ft.root.portablePath.rootAlias = alias

	return nil
}

func (ft *FileTreeImpl) ReplaceId(existingId, newId FileId) error {
	f := ft.Get(existingId)
	if f == nil {
		return werror.WithStack(werror.ErrNoFile)
	}

	ft.deleteInternal(existingId)
	f.setIdInternal(newId)
	ft.addInternal(newId, f)

	return nil
}

// ReadDir reads the filesystem for files it does not yet have, adds them to the tree,
// and returns the newly added files
func (ft *FileTreeImpl) ReadDir(dir *WeblensFileImpl) ([]*WeblensFileImpl, error) {
	entries, err := os.ReadDir(dir.absolutePath)
	if err != nil {
		return nil, err
	}

	children := make([]*WeblensFileImpl, 0, len(entries))
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

func (ft *FileTreeImpl) GetRoot() *WeblensFileImpl {
	if ft.root == nil {
		ft.log.Error().Msg("GetRoot called on fileTree with nil root")
	}
	return ft.root
}

func (ft *FileTreeImpl) GenerateFileId() FileId {
	return primitive.NewObjectID().Hex()
}

func (ft *FileTreeImpl) PortableToAbs(portable WeblensFilepath) (string, error) {
	if portable.RootName() != ft.rootAlias {
		return "", werror.Errorf(
			"fileTree.PortableToAbs: portable root alias [%s] does not match tree [%s] for [%s]",
			portable.RootName(),
			ft.rootAlias, portable.relPath,
		)
	}
	return filepath.Join(ft.GetRoot().AbsPath(), portable.RelativePath()), nil
}

func (ft *FileTreeImpl) AbsToPortable(absPath string) (WeblensFilepath, error) {
	if !strings.HasPrefix(absPath, ft.GetRoot().AbsPath()) {
		return WeblensFilepath{}, werror.Errorf(
			"fileTree.AbsToPortable: absPath [%s] does not match tree root prefix [%s]",
			absPath, ft.GetRoot().AbsPath(),
		)
	}

	return NewFilePath(ft.GetRoot().AbsPath(), ft.rootAlias, absPath), nil
}

func (ft *FileTreeImpl) ResizeUp(anchor *WeblensFileImpl, event *FileEvent, updateCallback func(newFile *WeblensFileImpl)) error {
	if ft.journal.IgnoreLocal() {
		return nil
	}

	externalEvent := event != nil
	if !externalEvent {
		event = ft.journal.NewEvent()
	}

	if err := anchor.BubbleMap(
		func(f *WeblensFileImpl) error {
			return handleFileResize(f, ft.journal, event, updateCallback)
		},
	); err != nil {
		return err
	}

	if !externalEvent {
		ft.journal.LogEvent(event)
	}

	return nil
}

func (ft *FileTreeImpl) ResizeDown(anchor *WeblensFileImpl, event *FileEvent, updateCallback func(newFile *WeblensFileImpl)) error {
	if ft.journal.IgnoreLocal() {
		ft.log.Trace().Msg("Ignoring local resize down")
		return nil
	}

	externalEvent := event != nil
	if !externalEvent {
		event = ft.journal.NewEvent()
	}

	if err := anchor.LeafMap(
		func(f *WeblensFileImpl) error {
			return handleFileResize(f, ft.journal, event, updateCallback)
		},
	); err != nil {
		return err
	}

	if !externalEvent {
		ft.journal.LogEvent(event)
	}

	return nil
}

var IgnoreFilenames = []string{
	".DS_Store",
	".content",
}

func handleFileResize(file *WeblensFileImpl, journal Journal, event *FileEvent, updateCallback func(newFile *WeblensFileImpl)) error {
	newSize, err := file.LoadStat()
	if err != nil {
		return err
	}
	if newSize != -1 && !journal.IgnoreLocal() && file.ID() != "ROOT" {
		updateCallback(file)

		lt := journal.Get(file.ID())

		if lt == nil || lt.GetLatestSize() != newSize {
			event.NewSizeChangeAction(file)
		}
	}

	return err
}

func (ft *FileTreeImpl) loadFromRoot(event *FileEvent, doFileDiscovery bool) error {
	start := time.Now()

	lifetimesByPath := map[string]*Lifetime{}
	missing := map[string]struct{}{}
	for _, lt := range ft.journal.GetActiveLifetimes() {
		// If we are discovering new files, and therefore are not mimicking another
		// tree, we just put the path into the map as-is.
		if doFileDiscovery {
			lifetimesByPath[lt.GetLatestAction().DestinationPath] = lt
			missing[lt.Id] = struct{}{}
			continue
		}

		// In the case we are handling files from another tree,
		// overwrite the root name so that the new files discovery matches
		path := ParsePortable(lt.GetLatestAction().DestinationPath)
		if path.RootName() != ft.rootAlias {
			path = path.OverwriteRoot(ft.rootAlias)
		}
		lifetimesByPath[path.ToPortable()] = lt
	}

	toLoad, err := ft.ReadDir(ft.root)
	if err != nil {
		return err
	}

	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Starting loadFromRoot with %d children", len(toLoad)) })
	for len(toLoad) != 0 {
		var fileToLoad *WeblensFileImpl

		// Pop from slice of files to load
		fileToLoad, toLoad = toLoad[0], toLoad[1:]
		if slices.Contains(IgnoreFilenames, fileToLoad.Filename()) {
			continue
		}

		portablePath := fileToLoad.GetPortablePath().ToPortable()
		if activeLt, ok := lifetimesByPath[portablePath]; ok {
			// We found this lifetime, so it is not missing, remove it from the missing map
			delete(missing, activeLt.Id)

			if event.journal != nil && activeLt.GetIsDir() != fileToLoad.IsDir() {
				activeLt.IsDir = fileToLoad.IsDir()
				err := event.journal.UpdateLifetime(activeLt)
				if err != nil {
					return err
				}
			}

			fileToLoad.setIdInternal(activeLt.ID())
			if !fileToLoad.IsDir() {
				fileToLoad.SetContentId(activeLt.ContentId)
			}
		} else if doFileDiscovery {
			fileToLoad.setIdInternal(ft.GenerateFileId())
			log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Discovering new file %s", fileToLoad.getIdInternal()) })
			event.NewCreateAction(fileToLoad)
		} else {
			log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping new file and children %s", portablePath) })
			continue
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

	if doFileDiscovery {
		// If we have missing files, create delete actions for them
		for missingId := range missing {
			ft.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Removing file with missing id %s", missingId) })
			_, err := event.NewDeleteAction(missingId)
			if err != nil {
				return err
			}
		}
	}

	err = ft.ResizeDown(ft.GetRoot(), event, func(newFile *WeblensFileImpl) {})
	if err != nil {
		return err
	}

	log.Trace().Func(func(e *zerolog.Event) {
		e.Msgf("loadFromRoot of %s complete in %s", ft.GetRoot().GetPortablePath(), time.Since(start))
	})

	return nil
}

func (ft *FileTreeImpl) importFromDirEntry(entry os.DirEntry, parent *WeblensFileImpl) (*WeblensFileImpl, error) {
	if parent == nil {
		return nil, werror.Errorf("Trying to add dirEntry with nil parent")
	}

	absPath := filepath.Join(parent.AbsPath(), entry.Name())
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

	f := &WeblensFileImpl{
		id:           "",
		absolutePath: absPath,
		portablePath: portable,
		filename:     entry.Name(),
		isDir:        boolPointer(info.IsDir()),
		modifyDate:   info.ModTime(),
		childrenMap:  map[string]*WeblensFileImpl{},
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

func MoveFileBetweenTrees(file, newParent *WeblensFileImpl, newName string, oldTree, newTree FileTree, event *FileEvent) error {
	if file.IsDir() {
		return werror.Errorf("Cannot move directory between trees")
	}

	_ = file.RecursiveMap(
		func(f *WeblensFileImpl) error {
			_, err := oldTree.Remove(f.ID())
			if err != nil {
				return err
			}

			err = newTree.Add(f)
			if err != nil {
				return err
			}

			return nil
		},
	)

	_, err := newTree.Move(file, newParent, newName, false, event)
	if err != nil {
		return err
	}

	return nil
}

type FileTree interface {
	Get(id FileId) *WeblensFileImpl
	GetRoot() *WeblensFileImpl
	ReadDir(dir *WeblensFileImpl) ([]*WeblensFileImpl, error)
	Size() int

	GetJournal() Journal
	SetJournal(Journal)

	Add(file *WeblensFileImpl) error
	Remove(id FileId) ([]*WeblensFileImpl, error)
	Delete(id FileId, event *FileEvent) error
	Move(f, newParent *WeblensFileImpl, newFilename string, overwrite bool, event *FileEvent) ([]MoveInfo, error)
	Touch(parentFolder *WeblensFileImpl, newFileName string, event *FileEvent, data ...[]byte) (*WeblensFileImpl, error)
	MkDir(parentFolder *WeblensFileImpl, newDirName string, event *FileEvent) (*WeblensFileImpl, error)

	SetRootAlias(alias string) error
	ReplaceId(oldId, newId FileId) error

	PortableToAbs(portable WeblensFilepath) (string, error)
	AbsToPortable(absPath string) (WeblensFilepath, error)
	GenerateFileId() FileId

	ResizeUp(anchor *WeblensFileImpl, event *FileEvent, updateCallback func(newFile *WeblensFileImpl)) error
	ResizeDown(anchor *WeblensFileImpl, event *FileEvent, updateCallback func(newFile *WeblensFileImpl)) error
}

type Hasher interface {
	Hash(file *WeblensFileImpl) error
}

type HashWaiter interface {
	Hasher
	Wait()
}
