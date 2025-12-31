// Package filetree provides file tree management functionality for organizing and navigating file hierarchies.
package filetree

// import (
// 	"os"
// 	"path/filepath"
// 	"slices"
// 	"strings"
// 	"sync"
// 	"time"
//
// 	file_model "github.com/ethanrous/weblens/models/file"
// 	"github.com/ethanrous/weblens/models/history"
// 	"github.com/ethanrous/weblens/modules/fs"
// 	"github.com/ethanrous/weblens/modules/errors"
// 	"github.com/rs/zerolog"
// 	"github.com/rs/zerolog/log"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )
//
// var _ FileTree = (*FileTreeImpl)(nil)
//
// type FileTreeImpl struct {
// 	journal *history.JournalImpl
//
// 	fMap map[string]*file_model.WeblensFileImpl
//
// 	root *file_model.WeblensFileImpl
//
// 	rootPath  string
// 	rootAlias string
//
// 	fsTreeLock sync.RWMutex
// 	log        zerolog.Logger
// }
//
// type MoveInfo struct {
// 	From *file_model.WeblensFileImpl
// 	To   *file_model.WeblensFileImpl
// }
//
// func boolPointer(b bool) *bool {
// 	return &b
// }
//
// func NewFileTree(rootPath, rootAlias string, journal *history.JournalImpl, doFileDiscovery bool, log zerolog.Logger) (*FileTreeImpl, error) {
// 	if journal == nil {
// 		return nil, errors.Errorf("Got nil journal trying to create new FileTree")
// 	}
//
// 	if rootPath[len(rootPath)-1] != '/' {
// 		rootPath = rootPath + "/"
// 	}
//
// 	if !filepath.IsAbs(rootPath) {
// 		return nil, errors.Errorf("rootPath must be an absolute path: %s", rootPath)
// 	}
//
// 	if _, err := os.Stat(rootPath); errors.Is(err, os.ErrNotExist) {
// 		err = os.MkdirAll(rootPath, os.ModePerm)
// 		if err != nil {
// 			return nil, errors.WithStack(err)
// 		}
// 	} else if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	root := &file_model.WeblensFileImpl{
// 		// id:       "ROOT",
// 		// parent:   nil,
// 		// filename: filepath.Base(rootPath),
// 		// isDir:    boolPointer(true),
// 		//
// 		// childrenMap:  map[string]*file.WeblensFileImpl{},
// 		// absolutePath: rootPath,
// 		// portablePath: fs.Filepath{
// 		// 	rootAlias: rootAlias,
// 		// 	relPath:   "",
// 		// },
// 	}
//
// 	// root.size.Store(-1)
//
// 	tree := &FileTreeImpl{
// 		// fMap:      map[string]*file.WeblensFileImpl{root.id: root},
// 		rootPath:  rootPath,
// 		root:      root,
// 		journal:   journal,
// 		rootAlias: rootAlias,
// 		log:       log,
// 	}
//
// 	event := tree.GetJournal().NewEvent()
// 	// if event.journal == nil {
// 	// 	event = nil
// 	// }
// 	err := tree.loadFromRoot(event, doFileDiscovery)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	journal.LogEvent(event)
//
// 	return tree, nil
// }
//
// func (ft *FileTreeImpl) GetJournal() *history.JournalImpl {
// 	return ft.journal
// }
//
// func (ft *FileTreeImpl) SetJournal(j *history.JournalImpl) {
// 	ft.journal = j
// }
//
// func (ft *FileTreeImpl) addInternal(id string, f *file_model.WeblensFileImpl) {
// 	ft.fsTreeLock.Lock()
// 	defer ft.fsTreeLock.Unlock()
//
// 	// log.Trace().Func(func(e *zerolog.Event) {e.Msgf("Adding %s (%s) to file tree", f.filename, f.id)})
//
// 	// Do not use .ID() inside critical section, as it may need to use the locks
// 	ft.fMap[id] = f
// 	if f.id == "ROOT" {
// 		ft.root = f
// 	}
// }
//
// func (ft *FileTreeImpl) deleteInternal(id string) {
// 	ft.fsTreeLock.Lock()
// 	defer ft.fsTreeLock.Unlock()
// 	delete(ft.fMap, id)
// }
//
// func (ft *FileTreeImpl) has(id string) bool {
// 	ft.fsTreeLock.RLock()
// 	defer ft.fsTreeLock.RUnlock()
// 	_, ok := ft.fMap[id]
// 	return ok
// }
//
// func (ft *FileTreeImpl) Add(f *file_model.WeblensFileImpl) error {
// 	if f == nil {
// 		return errors.New("trying to add a nil file to file tree")
// 	}
//
// 	if ft.has(f.ID()) {
// 		return errors.Errorf(
// 			"key collision on attempt to insert to filesystem tree: %s. "+
// 				"Existing file is at %s, new file is at %s", f.ID(), ft.Get(f.ID()).AbsPath(), f.AbsPath(),
// 		)
// 	}
//
// 	if slices.Contains(IgnoreFilenames, f.Filename()) {
// 		return nil
// 	}
//
// 	ft.addInternal(f.ID(), f)
//
// 	if f.parent == nil {
// 		parent := ft.Get(f.parentID)
// 		if parent == nil {
// 			return errors.Errorf("could not get parent of file to add")
// 		}
// 		f.setParentInternal(parent)
// 	}
//
// 	err := f.GetParent().AddChild(f)
// 	if err != nil {
// 		return err
// 	}
//
// 	// if f.IsDir() {
// 	// 	err = ft.journal.WatchFolder(f)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// }
//
// 	if f.getAbsPathInternal() == "" && f.GetPortablePath().relPath == "" {
// 		return errors.Errorf("Cannot add file to tree without abs path or portable path")
// 	} else if f.GetPortablePath().relPath == "" {
// 		portable, err := ft.AbsToPortable(f.getAbsPathInternal())
// 		if err != nil {
// 			return err
// 		}
// 		f.setPortable(portable)
// 	} else if f.getAbsPathInternal() == "" {
// 		abs, err := ft.PortableToAbs(f.GetPortablePath())
// 		if err != nil {
// 			return err
// 		}
// 		f.setAbsPath(abs)
// 	}
//
// 	return nil
// }
//
// func (ft *FileTreeImpl) Remove(id string) ([]*file_model.WeblensFileImpl, error) {
// 	f := ft.Get(id)
// 	if f == nil {
// 		return nil, errors.WithStack(errors.ErrNoFile.WithArg(id))
// 	}
//
// 	if f == ft.root {
// 		return nil, errors.Errorf("cannot delete root directory")
// 	}
//
// 	// If the file does not already have an id, generating the id can lock the file tree,
// 	// so we must do that outside of the lock here to avoid deadlock
// 	// f.ID()
//
// 	if !ft.has(f.id) {
// 		ft.log.Warn().Msgf("Tried to remove key not in FsTree [%s]", f.ID())
// 		return nil, errors.WithStack(file_model.ErrFileNotFound)
// 	}
//
// 	err := f.GetParent().removeChild(f)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var deleted []*file_model.WeblensFileImpl
// 	_ = f.RecursiveMap(
// 		func(file *file_model.WeblensFileImpl) error {
// 			deleted = append(deleted, file)
// 			ft.deleteInternal(file.ID())
//
// 			return nil
// 		},
// 	)
//
// 	return deleted, nil
// }
//
// func (ft *FileTreeImpl) Delete(id string, event *history.FileEvent) error {
// 	f := ft.Get(id)
// 	if f == nil {
// 		return errors.WithStack(file_model.ErrFileNotFound)
// 	}
//
// 	if f == ft.root {
// 		return errors.WithStack(errors.ErrRootFolder)
// 	}
//
// 	if f.IsDir() && len(f.GetChildren()) != 0 {
// 		return errors.Errorf("cannot delete non-empty directory")
// 	}
//
// 	_, err := ft.Remove(id)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = os.Remove(f.getAbsPathInternal())
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	_, err = event.NewDeleteAction(f.ID())
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (ft *FileTreeImpl) Get(fileID string) *file_model.WeblensFileImpl {
// 	ft.fsTreeLock.RLock()
// 	defer ft.fsTreeLock.RUnlock()
// 	return ft.fMap[fileID]
// }
//
// func (ft *FileTreeImpl) Move(
// 	f, newParent *file_model.WeblensFileImpl, newFilename string, overwrite bool, event *history.FileEvent,
// ) ([]MoveInfo, error) {
// 	if newParent == nil {
// 		return nil, errors.WithStack(file_model.ErrNilFile)
// 	} else if !newParent.IsDir() {
// 		return nil, errors.WithStack(file_model.ErrDirectoryRequired)
// 	} else if newFilename == "" {
// 		return nil, errors.WithStack(file_model.ErrFilenameRequired)
// 	}
//
// 	if newFilename == f.Filename() && newParent == f.GetParent() {
// 		return nil, errors.WithStack(errors.ErrEmptyMove)
// 	}
//
// 	newAbsPath := filepath.Join(newParent.AbsPath(), newFilename)
//
// 	if !overwrite {
// 		// Check if the file at the destination exists already
// 		if _, err := os.Stat(newAbsPath); err == nil {
// 			return nil, errors.WithStack(file_model.ErrFileAlreadyExists)
// 		}
// 	}
//
// 	if !f.Exists() || !newParent.Exists() {
// 		return nil, file_model.ErrFileNotFound
// 	}
//
// 	oldAbsPath := f.AbsPath()
//
// 	// Point of no return //
//
// 	var hasExternalEvent bool
// 	if event == nil {
// 		event = ft.GetJournal().NewEvent()
// 	} else {
// 		hasExternalEvent = true
// 	}
//
// 	// Sync file tree with new move, including f and all of its children.
// 	var moved []MoveInfo
// 	err := f.RecursiveMap(
// 		func(w *file_model.WeblensFileImpl) error {
// 			preFile := w.Freeze()
//
// 			// Shift the root of the move operation to be a child of the new parent
// 			if f == w {
// 				err := preFile.GetParent().removeChild(w)
// 				if err != nil {
// 					return err
// 				}
// 				f.filename = newFilename
// 				w.setParentInternal(newParent)
// 				err = w.GetParent().AddChild(w)
// 				if err != nil {
// 					return err
// 				}
// 			}
//
// 			newChildAbsPath := filepath.Join(w.GetParent().AbsPath(), w.Filename())
// 			if w.IsDir() {
// 				newChildAbsPath += "/"
// 			}
// 			w.setAbsPath(newChildAbsPath)
//
// 			portable, err := ft.AbsToPortable(w.getAbsPathInternal())
// 			if err != nil {
// 				return err
// 			}
// 			w.setPortable(portable)
//
// 			_, err = event.NewMoveAction(preFile.ID(), w)
// 			if err != nil {
// 				return err
// 			}
//
// 			moved = append(
// 				moved, MoveInfo{
// 					From: preFile,
// 					To:   w,
// 				},
// 			)
//
// 			w.modifiedNow()
//
// 			return nil
// 		},
// 	)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if overwrite {
// 		err = os.Remove(newAbsPath)
// 		if err != nil && !os.IsNotExist(err) {
// 			return nil, errors.WithStack(err)
// 		}
// 	}
//
// 	err = os.Rename(oldAbsPath, newAbsPath)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if !hasExternalEvent {
// 		ft.journal.LogEvent(event)
// 	}
//
// 	return moved, nil
// }
//
// // Size gets the number of files loaded into weblens.
// // This does not lock the file tree, and therefore
// // cannot be trusted to be microsecond accurate, but
// // it's quite close
// func (ft *FileTreeImpl) Size() int {
// 	return len(ft.fMap)
// }
//
// func (ft *FileTreeImpl) Touch(parentFolder *file_model.WeblensFileImpl, newFileName string, event *history.FileEvent, data ...[]byte) (
// 	*file_model.WeblensFileImpl, error,
// ) {
// 	childPath := parentFolder.GetPortablePath().Child(newFileName, false)
// 	absPath, err := ft.PortableToAbs(childPath)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	f := &file_model.WeblensFileImpl{
// 		id:           ft.GenerateFileID(),
// 		absolutePath: absPath,
// 		portablePath: childPath,
// 		filename:     newFileName,
// 		isDir:        boolPointer(false),
// 		modifyDate:   time.Now(),
// 		parentID:     parentFolder.ID(),
// 		parent:       parentFolder,
// 		childrenMap:  map[string]*file_model.WeblensFileImpl{},
// 		childIDs:     []FileID{},
// 	}
//
// 	err = f.CreateSelf()
// 	if err != nil {
// 		return f, err
// 	}
//
// 	if len(data) > 0 {
// 		_, err = f.Write(data[0])
// 		if err != nil {
// 			return f, err
// 		}
// 	}
//
// 	err = ft.Add(f)
// 	if err != nil {
// 		return f, err
// 	}
//
// 	if event != nil {
// 		event.NewCreateAction(f)
// 	} else {
// 		// event = ft.journal.NewEvent()
// 		// event.NewCreateAction(f)
// 		// ft.journal.LogEvent(event)
// 		// err = event.Wait()
// 		// if err != nil {
// 		// 	return f, err
// 		// }
// 	}
//
// 	return f, nil
// }
//
// // MkDir creates a new dir as a child of parentFolder named newDirName. If the dir already exists,
// // it will be returned along with a ErrDirAlreadyExists error.
// func (ft *FileTreeImpl) MkDir(
// 	parentFolder *file_model.WeblensFileImpl, newDirName string, event *history.FileEvent,
// ) (*file_model.WeblensFileImpl, error) {
// 	if existingFile, _ := parentFolder.GetChild(newDirName); existingFile != nil {
// 		return existingFile, errors.WithStack(errors.ErrDirAlreadyExists.WithArg(parentFolder.AbsPath() + newDirName))
// 	}
//
// 	absPath := filepath.Join(parentFolder.AbsPath(), newDirName) + "/"
//
// 	d := &file_model.WeblensFileImpl{
// 		id:           ft.GenerateFileID(),
// 		absolutePath: absPath,
// 		portablePath: fs.Filepath{},
// 		filename:     newDirName,
// 		isDir:        boolPointer(true),
// 		modifyDate:   time.Now(),
// 		parentID:     parentFolder.ID(),
// 		parent:       parentFolder,
// 		childrenMap:  map[string]*file_model.WeblensFileImpl{},
// 		childIDs:     []FileID{},
// 	}
//
// 	if d.Exists() {
// 		existingFile := ft.Get(d.ID())
//
// 		if existingFile == nil {
// 			err := ft.Add(d)
// 			if err != nil {
// 				return d, err
// 			}
// 			existingFile = d
// 		}
//
// 		return existingFile, errors.WithStack(errors.ErrDirAlreadyExists)
// 	}
//
// 	d.size.Store(0)
//
// 	err := ft.Add(d)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	err = d.CreateSelf()
// 	if err != nil {
// 		return d, err
// 	}
//
// 	if event != nil {
// 		event.NewCreateAction(d)
// 	} else {
// 		event = ft.journal.NewEvent()
// 		event.NewCreateAction(d)
// 		ft.journal.LogEvent(event)
// 		err = event.Wait()
// 		if err != nil {
// 			return d, err
// 		}
// 	}
//
// 	return d, nil
// }
//
// func (ft *FileTreeImpl) SetRootAlias(alias string) error {
// 	if ft.Size() != 1 {
// 		return errors.Errorf("Cannot set root alias on non-empty file tree")
// 	}
//
// 	ft.rootAlias = alias
// 	ft.root.portablePath.rootAlias = alias
//
// 	return nil
// }
//
// func (ft *FileTreeImpl) ReplaceID(existingID, newID string) error {
// 	f := ft.Get(existingID)
// 	if f == nil {
// 		return errors.WithStack(errors.ErrNoFile)
// 	}
//
// 	ft.deleteInternal(existingID)
// 	f.setIDInternal(newID)
// 	ft.addInternal(newID, f)
//
// 	return nil
// }
//
// // ReadDir reads the filesystem for files it does not yet have, adds them to the tree,
// // and returns the newly added files
// func (ft *FileTreeImpl) ReadDir(dir *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error) {
// 	entries, err := os.ReadDir(dir.absolutePath)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	children := make([]*file_model.WeblensFileImpl, 0, len(entries))
// 	for _, entry := range entries {
// 		if slices.Contains(IgnoreFilenames, entry.Name()) {
// 			continue
// 		}
//
// 		child, err := ft.importFromDirEntry(entry, dir)
// 		if err != nil {
// 			return nil, err
// 		}
// 		children = append(children, child)
// 	}
//
// 	return children, nil
// }
//
// func (ft *FileTreeImpl) GetRoot() *file_model.WeblensFileImpl {
// 	if ft.root == nil {
// 		ft.log.Error().Msg("GetRoot called on fileTree with nil root")
// 	}
// 	return ft.root
// }
//
// func (ft *FileTreeImpl) GenerateFileID() string {
// 	return primitive.NewObjectID().Hex()
// }
//
// func (ft *FileTreeImpl) PortableToAbs(portable fs.Filepath) (string, error) {
// 	if portable.RootName() != ft.rootAlias {
// 		return "", errors.Errorf(
// 			"fileTree.PortableToAbs: portable root alias [%s] does not match tree [%s] for [%s]",
// 			portable.RootName(),
// 			ft.rootAlias, portable.relPath,
// 		)
// 	}
// 	return filepath.Join(ft.GetRoot().AbsPath(), portable.RelativePath()), nil
// }
//
// func (ft *FileTreeImpl) AbsToPortable(absPath string) (fs.Filepath, error) {
// 	if !strings.HasPrefix(absPath, ft.GetRoot().AbsPath()) {
// 		return fs.Filepath{}, errors.Errorf(
// 			"fileTree.AbsToPortable: absPath [%s] does not match tree root prefix [%s]",
// 			absPath, ft.GetRoot().AbsPath(),
// 		)
// 	}
//
// 	return NewFilePath(ft.GetRoot().AbsPath(), ft.rootAlias, absPath), nil
// }
//
// func (ft *FileTreeImpl) ResizeUp(anchor *file_model.WeblensFileImpl, event *history.FileEvent, updateCallback func(newFile *file_model.WeblensFileImpl)) error {
// 	if ft.journal.IgnoreLocal() {
// 		return nil
// 	}
//
// 	externalEvent := event != nil
// 	if !externalEvent {
// 		event = ft.journal.NewEvent()
// 	}
//
// 	if err := anchor.BubbleMap(
// 		func(f *file_model.WeblensFileImpl) error {
// 			return handleFileResize(f, ft.journal, event, updateCallback)
// 		},
// 	); err != nil {
// 		return err
// 	}
//
// 	if !externalEvent {
// 		ft.journal.LogEvent(event)
// 	}
//
// 	return nil
// }
//
// func (ft *FileTreeImpl) ResizeDown(anchor *file_model.WeblensFileImpl, event *history.FileEvent, updateCallback func(newFile *file_model.WeblensFileImpl)) error {
// 	if ft.journal.IgnoreLocal() {
// 		ft.log.Trace().Msg("Ignoring local resize down")
// 		return nil
// 	}
//
// 	externalEvent := event != nil
// 	if !externalEvent {
// 		event = ft.journal.NewEvent()
// 	}
//
// 	if err := anchor.LeafMap(
// 		func(f *file_model.WeblensFileImpl) error {
// 			return handleFileResize(f, ft.journal, event, updateCallback)
// 		},
// 	); err != nil {
// 		return err
// 	}
//
// 	if !externalEvent {
// 		ft.journal.LogEvent(event)
// 	}
//
// 	return nil
// }
//
// var IgnoreFilenames = []string{
// 	".DS_Store",
// 	".content",
// }
//
// func handleFileResize(file *file_model.WeblensFileImpl, journal *history.JournalImpl, event *history.FileEvent, updateCallback func(newFile *file_model.WeblensFileImpl)) error {
// 	newSize, err := file.LoadStat()
// 	if err != nil {
// 		return err
// 	}
// 	if newSize != -1 && !journal.IgnoreLocal() && file.ID() != "ROOT" {
// 		updateCallback(file)
//
// 		lt := journal.Get(file.ID())
//
// 		if lt == nil || lt.GetLatestSize() != newSize {
// 			event.NewSizeChangeAction(file)
// 		}
// 	}
//
// 	return err
// }
//
// func (ft *FileTreeImpl) loadFromRoot(event *history.FileEvent, doFileDiscovery bool) error {
// 	start := time.Now()
//
// 	lifetimesByPath := map[string]*Lifetime{}
// 	missing := map[string]struct{}{}
// 	for _, lt := range ft.journal.GetActiveLifetimes() {
// 		// If we are discovering new files, and therefore are not mimicking another
// 		// tree, we just put the path into the map as-is.
// 		if doFileDiscovery {
// 			lifetimesByPath[lt.GetLatestAction().DestinationPath] = lt
// 			missing[lt.ID] = struct{}{}
// 			continue
// 		}
//
// 		// In the case we are handling files from another tree,
// 		// overwrite the root name so that the new files discovery matches
// 		path := ParsePortable(lt.GetLatestAction().DestinationPath)
// 		if path.RootName() != ft.rootAlias {
// 			path = path.OverwriteRoot(ft.rootAlias)
// 		}
// 		lifetimesByPath[path.ToPortable()] = lt
// 	}
//
// 	toLoad, err := ft.ReadDir(ft.root)
// 	if err != nil {
// 		return err
// 	}
//
// 	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Starting loadFromRoot with %d children", len(toLoad)) })
// 	for len(toLoad) != 0 {
// 		var fileToLoad *file_model.WeblensFileImpl
//
// 		// Pop from slice of files to load
// 		fileToLoad, toLoad = toLoad[0], toLoad[1:]
// 		if slices.Contains(IgnoreFilenames, fileToLoad.Filename()) {
// 			continue
// 		}
//
// 		portablePath := fileToLoad.GetPortablePath().ToPortable()
// 		if activeLt, ok := lifetimesByPath[portablePath]; ok {
// 			// We found this lifetime, so it is not missing, remove it from the missing map
// 			delete(missing, activeLt.ID)
//
// 			if event.journal != nil && activeLt.GetIsDir() != fileToLoad.IsDir() {
// 				activeLt.IsDir = fileToLoad.IsDir()
// 				err := event.journal.UpdateLifetime(activeLt)
// 				if err != nil {
// 					return err
// 				}
// 			}
//
// 			fileToLoad.setIDInternal(activeLt.ID())
// 			if !fileToLoad.IsDir() {
// 				fileToLoad.SetContentID(activeLt.ContentID)
// 			}
// 		} else if doFileDiscovery {
// 			fileToLoad.setIDInternal(ft.GenerateFileID())
// 			log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Discovering new file %s", fileToLoad.getIDInternal()) })
// 			event.NewCreateAction(fileToLoad)
// 		} else {
// 			log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping new file and children %s", portablePath) })
// 			continue
// 		}
//
// 		err = ft.Add(fileToLoad)
// 		if err != nil {
// 			return err
// 		}
//
// 		if fileToLoad.IsDir() {
// 			children, err := ft.ReadDir(fileToLoad)
// 			if err != nil {
// 				return err
// 			}
// 			toLoad = append(toLoad, children...)
// 		}
// 	}
//
// 	if doFileDiscovery {
// 		// If we have missing files, create delete actions for them
// 		for missingID := range missing {
// 			ft.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Removing file with missing id %s", missingID) })
// 			_, err := event.NewDeleteAction(missingID)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	err = ft.ResizeDown(ft.GetRoot(), event, func(newFile *file_model.WeblensFileImpl) {})
// 	if err != nil {
// 		return err
// 	}
//
// 	log.Trace().Func(func(e *zerolog.Event) {
// 		e.Msgf("loadFromRoot of %s complete in %s", ft.GetRoot().GetPortablePath(), time.Since(start))
// 	})
//
// 	return nil
// }
//
// func (ft *FileTreeImpl) importFromDirEntry(entry os.DirEntry, parent *file_model.WeblensFileImpl) (*file_model.WeblensFileImpl, error) {
// 	if parent == nil {
// 		return nil, errors.Errorf("Trying to add dirEntry with nil parent")
// 	}
//
// 	absPath := filepath.Join(parent.AbsPath(), entry.Name())
// 	if entry.IsDir() {
// 		absPath += "/"
// 	}
// 	portable, err := ft.AbsToPortable(absPath)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	info, err := entry.Info()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	f := &file_model.WeblensFileImpl{
// 		id:           "",
// 		absolutePath: absPath,
// 		portablePath: portable,
// 		filename:     entry.Name(),
// 		isDir:        boolPointer(info.IsDir()),
// 		modifyDate:   info.ModTime(),
// 		childrenMap:  map[string]*file_model.WeblensFileImpl{},
// 		childIDs:     []FileID{},
// 	}
//
// 	f.setParentInternal(parent)
//
// 	if !f.IsDir() {
// 		f.size.Store(info.Size())
// 	} else {
// 		f.size.Store(-1)
// 	}
//
// 	return f, nil
// }
//
// func MoveFileBetweenTrees(file, newParent *file_model.WeblensFileImpl, newName string, oldTree, newTree FileTree, event *history.FileEvent) error {
// 	if file.IsDir() {
// 		return errors.Errorf("Cannot move directory between trees")
// 	}
//
// 	_ = file.RecursiveMap(
// 		func(f *file.WeblensFileImpl) error {
// 			_, err := oldTree.Remove(f.ID())
// 			if err != nil {
// 				return err
// 			}
//
// 			err = newTree.Add(f)
// 			if err != nil {
// 				return err
// 			}
//
// 			return nil
// 		},
// 	)
//
// 	_, err := newTree.Move(file, newParent, newName, false, event)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// type FileTree interface {
// 	Get(id string) *file_model.WeblensFileImpl
// 	GetRoot() *file_model.WeblensFileImpl
// 	ReadDir(dir *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error)
// 	Size() int
//
// 	GetJournal() *history.JournalImpl
// 	SetJournal(*history.JournalImpl)
//
// 	Add(file *file_model.WeblensFileImpl) error
// 	Remove(id string) ([]*file_model.WeblensFileImpl, error)
// 	Delete(id string, event *history.FileEvent) error
// 	Move(f, newParent *file_model.WeblensFileImpl, newFilename string, overwrite bool, event *history.FileEvent) ([]MoveInfo, error)
// 	Touch(parentFolder *file_model.WeblensFileImpl, newFileName string, event *history.FileEvent, data ...[]byte) (*file_model.WeblensFileImpl, error)
// 	MkDir(parentFolder *file_model.WeblensFileImpl, newDirName string, event *history.FileEvent) (*file_model.WeblensFileImpl, error)
//
// 	SetRootAlias(alias string) error
// 	ReplaceID(oldID, newID string) error
//
// 	PortableToAbs(portable fs.Filepath) (string, error)
// 	AbsToPortable(absPath string) (fs.Filepath, error)
// 	GenerateFileID() string
//
// 	ResizeUp(anchor *file_model.WeblensFileImpl, event *history.FileEvent, updateCallback func(newFile *file_model.WeblensFileImpl)) error
// 	ResizeDown(anchor *file_model.WeblensFileImpl, event *history.FileEvent, updateCallback func(newFile *file_model.WeblensFileImpl)) error
// }
//
// type Hasher interface {
// 	Hash(file *file_model.WeblensFileImpl) error
// }
//
// type HashWaiter interface {
// 	Hasher
// 	Wait()
// }
