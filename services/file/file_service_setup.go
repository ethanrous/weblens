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
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/fs"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/startup"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/journal"
)

var ErrChildrenAlreadyLoaded = errors.Errorf("children already loaded")

type doFileCreationContextKey struct{}

func init() {
	startup.RegisterStartup(loadFs)
}

func LoadFilesRecursively(ctx context_service.AppContext, root *file_model.WeblensFileImpl) error {
	appCtx, _ := context_service.FromContext(ctx)
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
			return errors.Errorf("Failed to load directory %s: %w", file.GetPortablePath(), err)
		}

		searchFiles = append(searchFiles, newChildren...)
	}

	appCtx.Log().Trace().Msgf("Loaded file %s [%s] in %s", root.GetPortablePath().String(), root.ID(), time.Since(start))

	return nil
}

func (fs *FileServiceImpl) makeRoot(ctx context.Context, rootPath file_system.Filepath) error {
	var f *file_model.WeblensFileImpl

	f = file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       rootPath,
		CreateNow:  true,
		GenerateId: true,
	})

	if f == nil {
		return errors.New("failed to create root file")
	}

	if rootPath.RelPath == "" {
		fs.setFileInternal(rootPath.RootName(), f)
	} else {
		parent, err := fs.GetFileById(ctx, rootPath.RootName())
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

func loadFs(ctx context.Context, cnf config.ConfigProvider) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.WithStack(context_service.ErrNoContext)
	}

	err := file_system.RegisterAbsolutePrefix(file_model.RestoreTreeKey, filepath.Join(cnf.DataPath, ".restore"))
	if err != nil {
		return err
	}

	err = file_system.RegisterAbsolutePrefix(file_model.CachesTreeKey, cnf.CachePath)
	if err != nil {
		return err
	}

	fs := appCtx.FileService.(*FileServiceImpl)

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

		if err := appCtx.FileService.(*FileServiceImpl).makeRoot(ctx, file_model.BackupRootPath); err != nil {
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
	if f, _ := ctx.FileService.GetFileByFilepath(ctx, filepath, true); f != nil {
		return f, nil
	}

	f := file_model.NewWeblensFile(file_model.NewFileOptions{Path: filepath})
	action, ok := pathMap[filepath]

	if !doFileCreation && !ok {
		ctx.Log().Warn().Msgf("File [%s] not found in history, and doFileDiscovery is false", filepath)

		return f, nil
	}

	if !ok {
		ctx.Log().Trace().Msgf("File [%s] not found in history, creating new file", filepath)

		if needsContentId(f) {
			// tsk, err := ctx.TaskService.DispatchJob(ctx, job_model.HashFileTask, job_model.HashFileMeta{File: f}, taskPool)
			//
			// if err != nil {
			// 	return nil, err
			// }
			//
			// tsk.Wait()
			//
			// contentId, ok := tsk.GetResult()["contentId"].(string)
			// if !ok {
			// 	return nil, errors.New("failed to get contentId from task result")
			// }

			newContentId, err := file_model.GenerateContentId(ctx, f)
			if err != nil {
				return nil, err
			}

			f.SetContentId(newContentId)
		}

		action = history.NewCreateAction(ctx, f)

		err := history.SaveAction(ctx, &action)
		if err != nil {
			return nil, err
		}

		pathMap[filepath] = action
	} else if doFileCreation && action.ContentId == "" && needsContentId(f) {
		newContentId, err := file_model.GenerateContentId(ctx, f)
		if err != nil {
			return nil, err
		}

		action.ContentId = newContentId

		ctx.Log().Debug().Msgf("File [%s] found in history, but contentId is empty, generating new one: %s", filepath, newContentId)

		err = history.UpdateAction(ctx, &action)
		if err != nil {
			return nil, err
		}

		pathMap[filepath] = action
	} else {
		existing, ok := ctx.FileService.(*FileServiceImpl).getFileInternal(action.FileId)
		if ok {
			return existing, nil
		}
	}

	f.SetId(action.FileId)
	f.SetContentId(action.ContentId)

	return f, nil
}

func loadOneDirectory(ctx context_service.AppContext, dir *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error) {
	if dir.ChildrenLoaded() {
		return nil, errors.WithStack(ErrChildrenAlreadyLoaded)
	}

	if !dir.IsDir() {
		return nil, errors.WithStack(file_model.ErrDirectoryRequired)
	}

	pathMap, err := loadLifetimes(ctx, dir.GetPortablePath())
	if err != nil {
		return nil, err
	}

	children := []*file_model.WeblensFileImpl{}

	childPaths, err := getChildFilepaths(dir.GetPortablePath())
	if err != nil {
		return nil, err
	}

	doFileCreation, _ := ctx.Value(doFileCreationContextKey{}).(bool)

	err = db.WithTransaction(ctx, func(ctx context.Context) error {
		appCtx, _ := context_service.FromContext(ctx)

		for _, childPath := range childPaths {
			child, err := handleFileCreation(appCtx, childPath, pathMap, doFileCreation)
			if err != nil {
				return err
			}

			err = child.SetParent(dir)
			if err != nil {
				return err
			}

			err = dir.AddChild(child)
			if err != nil {
				return err
			}

			err = appCtx.FileService.AddFile(appCtx, child)
			if err != nil {
				return errors.Errorf("failed to add child [%s] to tree: %w", child.GetPortablePath(), err)
			}

			children = append(children, child)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return children, nil
}

func loadFsCore(ctx context.Context) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("not an app context")
	}

	appCtx.Log().Debug().Msg("Loading file system")

	start := time.Now()

	root, err := appCtx.FileService.GetFileById(ctx, file_model.UsersTreeKey)
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

var fpMap = make(map[file_system.Filepath]history.FileAction)

func loadLifetimes(ctx context_service.AppContext, pathPrefix fs.Filepath) (map[file_system.Filepath]history.FileAction, error) {
	if len(fpMap) != 0 {
		// Already loaded
		return fpMap, nil
	}

	lifetimes, err := journal.GetLifetimesByTowerId(ctx, ctx.LocalTowerId, journal.GetLifetimesOptions{ActiveOnly: true})
	if err != nil {
		return nil, err
	}

	// fpMap := make(map[file_system.Filepath]history.FileAction, len(lifetimes))

	// Might be faster than map
	// lifetimes = slices.SortFunc(lifetimes, func(a, b history.FileLifetime) int {
	// 	return strings.Compare(a.Actions[0].GetRelevantPath().String(), b.Actions[0].GetRelevantPath().String())
	// })
	//
	// slices.BinarySearchFunc(lifetimes, pathPrefix, func(lt history.FileLifetime, t file_system.Filepath) int {
	// 	return strings.Compare(lt.Actions[0].GetRelevantPath().String(), t.String())
	// })

	for _, lt := range lifetimes {
		a := lt.Actions[len(lt.Actions)-1]

		if a.ActionType == history.FileDelete {
			continue
		}

		a.ContentId = lt.Actions[0].ContentId

		fpMap[a.GetRelevantPath()] = a
	}

	return fpMap, nil
}

func needsContentId(f *file_model.WeblensFileImpl) bool {
	return !f.IsDir() && f.Size() != 0 && f.GetContentId() == ""
}
