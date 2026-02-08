package file

import (
	"context"
	"path/filepath"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	job_model "github.com/ethanrous/weblens/models/job"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/startup"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// ErrChildrenAlreadyLoaded indicates that a directory's children have already been loaded.
var ErrChildrenAlreadyLoaded = wlerrors.Errorf("children already loaded")

type doFileCreationContextKey struct{}

func init() {
	startup.RegisterHook(loadFs)
}

// LoadFilesRecursively loads a directory and all its subdirectories into the file service.
func LoadFilesRecursively(ctx context_service.AppContext, root *file_model.WeblensFileImpl) error {
	searchFiles := []*file_model.WeblensFileImpl{root}

	start := time.Now()

	for len(searchFiles) != 0 {
		var file *file_model.WeblensFileImpl

		file, searchFiles = searchFiles[0], searchFiles[1:]

		if !file.IsDir() {
			continue
		}

		newChildren, err := loadOneDirectory(ctx, file)
		if err != nil {
			return wlerrors.Errorf("Failed to load directory [%s]: %w", file.GetPortablePath(), err)
		}

		searchFiles = append(searchFiles, newChildren...)
	}

	ctx.Log().Trace().Msgf("Loaded file %s [%s] in %s", root.GetPortablePath().String(), root.ID(), time.Since(start))

	return nil
}

func (fs *ServiceImpl) makeRoot(ctx context.Context, rootPath file_system.Filepath) error {
	var f *file_model.WeblensFileImpl

	f = file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       rootPath,
		CreateNow:  true,
		GenerateID: true,
	})

	if f == nil {
		return wlerrors.New("failed to create root file")
	}

	if rootPath.RelPath == "" {
		fs.setFileInternal(rootPath.RootName(), f)
	} else {
		parent, err := fs.GetFileByID(ctx, rootPath.RootName())
		if err != nil {
			return err
		}

		err = f.SetParent(parent)
		if err != nil {
			return err
		}

		err = fs.AddFile(ctx, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadFs(ctx context.Context, cnf config.Provider) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.WithStack(context_service.ErrNoContext)
	}

	if !filepath.IsAbs(cnf.DataPath) {
		return wlerrors.Errorf("config DataPath must be an absolute path, got: [%s]", cnf.DataPath)
	}

	err := file_system.RegisterAbsolutePrefix(file_model.RestoreTreeKey, filepath.Join(cnf.DataPath, ".restore"))
	if err != nil {
		return err
	}

	err = file_system.RegisterAbsolutePrefix(file_model.CachesTreeKey, cnf.CachePath)
	if err != nil {
		return err
	}

	fs := appCtx.FileService.(*ServiceImpl)

	start := time.Now()

	for _, root := range []file_system.Filepath{file_model.CacheRootPath, file_model.ThumbsDirPath, file_model.RestoreDirPath, file_model.ZipsDirPath} {
		if err := fs.makeRoot(ctx, root); err != nil {
			return err
		}
	}

	if tower_model.Role(cnf.InitRole) == tower_model.RoleCore {
		appCtx = appCtx.WithValue(doFileCreationContextKey{}, true)

		err := file_system.RegisterAbsolutePrefix(file_model.UsersTreeKey, filepath.Join(cnf.DataPath, "users"))
		if err != nil {
			return err
		}

		if err := fs.makeRoot(appCtx, file_model.UsersRootPath); err != nil {
			return err
		}

		err = loadFsCore(appCtx)
		if err != nil {
			return err
		}
	} else if tower_model.Role(cnf.InitRole) == tower_model.RoleBackup {
		err := file_system.RegisterAbsolutePrefix(file_model.BackupTreeKey, filepath.Join(cnf.DataPath, "backup"))
		if err != nil {
			return err
		}

		if err := appCtx.FileService.(*ServiceImpl).makeRoot(ctx, file_model.BackupRootPath); err != nil {
			return err
		}

		return db.WithTransaction(ctx, func(ctx context.Context) error {
			err := loadFsTransactionBackup(ctx)
			if err != nil {
				return err
			}

			return nil
		})
	}

	appCtx.Log().Debug().Msgf("File system initialized in %s", time.Since(start))

	return nil
}

func handleFileCreation(ctx context_service.AppContext, filepath file_system.Filepath, pathMap map[file_system.Filepath]history.FileAction, doFileCreation bool) (*file_model.WeblensFileImpl, error) {
	if f, err := ctx.FileService.GetFileByFilepath(ctx, filepath, true); err == nil {
		ctx.Log().Trace().Msgf("File [%s] found by filepath in file service", filepath)

		return f, nil
	} else if !wlerrors.Is(err, file_model.ErrFileNotFound) {
		return nil, err
	}

	f := file_model.NewWeblensFile(file_model.NewFileOptions{Path: filepath})
	action, ok := pathMap[filepath]

	if !doFileCreation && !ok {
		return nil, wlerrors.ReplaceStack(file_model.ErrFileHistoryMissing)
	}

	if !ok {
		existingAction, err := history.DoesFileExistInHistory(ctx, filepath)
		if err != nil {
			return nil, err
		}

		if existingAction != nil {
			return nil, wlerrors.Errorf("File [%s] not found in pathMap but exists in history, inconsistent state", filepath)
		}

		ctx.Log().Trace().Msgf("File [%s] not found in history, creating new file", filepath)

		if needsContentID(f) {
			newContentID, err := file_model.GenerateContentID(ctx, f)
			if err != nil {
				return nil, err
			}

			f.SetContentID(newContentID)
		}

		action = history.NewCreateAction(ctx, f)

		err = history.SaveAction(ctx, &action)
		if err != nil {
			return nil, err
		}

		pathMap[filepath] = action
	} else if doFileCreation && action.ContentID == "" && needsContentID(f) {
		newContentID, err := file_model.GenerateContentID(ctx, f)
		if err != nil {
			return nil, err
		}

		action.ContentID = newContentID

		ctx.Log().Debug().Msgf("File [%s] found in history, but contentID is empty, generating new one: %s", filepath, newContentID)

		err = history.UpdateAction(ctx, &action)
		if err != nil {
			return nil, err
		}

		pathMap[filepath] = action
	} else {
		existing, ok := ctx.FileService.(*ServiceImpl).getFileInternal(action.FileID)
		if ok {
			return existing, nil
		}
	}

	if action.FileID == "" {
		return nil, wlerrors.Errorf("File [%s] has no action file ID in history, inconsistent state", filepath)
	}

	f.SetID(action.FileID)
	f.SetContentID(action.ContentID)

	return f, nil
}

func loadOneDirectory(ctx context_service.AppContext, dir *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error) {
	if dir.ChildrenLoaded() {
		return nil, wlerrors.WithStack(ErrChildrenAlreadyLoaded)
	}

	if !dir.IsDir() {
		return nil, wlerrors.WithStack(file_model.ErrDirectoryRequired)
	}

	pathMap, err := loadLifetimes(ctx)
	if err != nil {
		return nil, err
	}

	children := []*file_model.WeblensFileImpl{}

	childPaths, err := getChildFilepaths(dir.GetPortablePath())
	if err != nil {
		return nil, err
	}

	doFileCreation, _ := ctx.Value(doFileCreationContextKey{}).(bool)

	for _, childPath := range childPaths {
		child, err := handleFileCreation(ctx, childPath, pathMap, doFileCreation)
		if err != nil && wlerrors.Is(err, file_model.ErrFileHistoryMissing) {
			ctx.Log().Warn().Msgf("File [%s] missing history entry, skipping", childPath)

			continue
		} else if err != nil {
			return nil, wlerrors.Errorf("failed to handle file creation for [%s]: %w", childPath, err)
		}

		err = child.SetParent(dir)
		if err != nil {
			return nil, err
		}

		err = dir.AddChild(child)
		if err != nil && wlerrors.Is(err, file_model.ErrFileAlreadyExists) {
			continue
		} else if err != nil {
			return nil, err
		}

		err = ctx.FileService.AddFile(ctx, child)
		if err != nil {
			return nil, wlerrors.Errorf("failed to add child [%s] to tree: %w", child.GetPortablePath(), err)
		}

		children = append(children, child)
	}

	return children, nil
}

func loadFsCore(ctx context.Context) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.New("not an app context")
	}

	appCtx.Log().Debug().Msg("Loading file system")

	start := time.Now()

	root, err := appCtx.FileService.GetFileByID(ctx, file_model.UsersTreeKey)
	if err != nil {
		return err
	}

	_, err = loadOneDirectory(appCtx, root)
	if err != nil {
		return err
	}

	appCtx.Log().Debug().Msgf("fs load of %s complete in %s", file_model.UsersRootPath, time.Since(start))
	start = time.Now()

	_, err = root.LoadStat()
	if err != nil {
		return err
	}

	appCtx.Log().Trace().Msgf("Computed tree size in %s", time.Since(start))

	for _, userHome := range root.GetChildren() {
		_, err := user_model.GetUserByUsername(ctx, userHome.GetPortablePath().Filename())
		if err != nil {
			continue // Skip if user does not exist
		}

		children, err := appCtx.FileService.GetChildren(appCtx, userHome)
		if err != nil {
			return err
		}

		for _, child := range children {
			_, err = appCtx.TaskService.DispatchJob(appCtx, job_model.LoadFilesystemTask, job_model.LoadFilesystemMeta{File: child}, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func loadLifetimes(ctx context_service.AppContext) (map[file_system.Filepath]history.FileAction, error) {
	fpMapCache := ctx.GetCache("filepath_map")
	if fpMapI, ok := fpMapCache.Get("filepath_map"); ok {
		fpMap, ok := fpMapI.(map[file_system.Filepath]history.FileAction)
		if ok {
			return fpMap, nil
		}

		return nil, wlerrors.Errorf("invalid filepath_map in context")
	}

	fpMap := make(map[file_system.Filepath]history.FileAction)

	localTower, err := tower_model.GetLocal(ctx)
	if err != nil {
		return nil, err
	}

	isBackup := localTower.Role == tower_model.RoleBackup

	var lifetimes []history.FileLifetime
	// Load all active lifetimes, for backups we load all, for cores only those relevant to the local tower
	if isBackup {
		lifetimes, err = history.GetLifetimes(ctx, history.GetLifetimesOptions{ActiveOnly: true})
	} else {
		lifetimes, err = history.GetLifetimes(ctx, history.GetLifetimesOptions{ActiveOnly: true, TowerID: ctx.LocalTowerID})
	}

	if err != nil {
		return nil, err
	}

	for _, lt := range lifetimes {
		a := lt.Actions[len(lt.Actions)-1]

		if a.ActionType == history.FileDelete {
			continue
		}

		a.ContentID = lt.Actions[0].ContentID

		path := a.GetRelevantPath()

		// For backups, translate the path to the local tower's structure
		if isBackup {
			remoteTower, err := tower_model.GetTowerByID(ctx, a.TowerID)
			if err != nil {
				return nil, err
			}

			path, err = TranslateBackupPath(ctx, path, remoteTower)
			if err != nil {
				return nil, err
			}
		}

		fpMap[path] = a
	}

	fpMapCache.Set("filepath_map", fpMap)

	return fpMap, nil
}

func needsContentID(f *file_model.WeblensFileImpl) bool {
	return !f.IsDir() && f.Size() != 0 && f.GetContentID() == ""
}
