package file

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/startup"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/pkg/errors"
)

func init() {
	startup.RegisterStartup(loadFs)
}

func needsContentId(f *file_model.WeblensFileImpl) bool {
	return !f.IsDir() && f.Size() != 0 && f.GetContentId() == ""
}

func (fs *FileServiceImpl) makeRoot(rootPath file_system.Filepath) error {
	var f *file_model.WeblensFileImpl

	var err error
	if !exists(rootPath) {
		f, err = mkdir(rootPath)
	} else {
		f = file_model.NewWeblensFile(file_model.NewFileOptions{Path: rootPath})
	}

	if err != nil {
		return err
	}

	if rootPath.RelPath == "" {
		fs.files[rootPath.RootName()] = f
	}

	return nil
}

func loadFs(ctx context.Context, cnf config.ConfigProvider) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("not an app context")
	}

	err := file_system.RegisterAbsolutePrefix(file_model.RestoreTreeKey, filepath.Join(cnf.DataPath, ".restore"))
	if err != nil {
		return err
	}

	err = file_system.RegisterAbsolutePrefix(file_model.CachesTreeKey, cnf.CachePath)
	if err != nil {
		return err
	}

	for _, root := range []file_system.Filepath{file_model.CacheRootPath, file_model.ThumbsDirPath, file_model.RestoreDirPath} {
		if err := appCtx.FileService.(*FileServiceImpl).makeRoot(root); err != nil {
			return err
		}
	}

	if tower_model.Role(cnf.InitRole) == tower_model.RoleCore {
		err := file_system.RegisterAbsolutePrefix(file_model.UsersTreeKey, filepath.Join(cnf.DataPath, "users"))
		if err != nil {
			return err
		}

		if err := appCtx.FileService.(*FileServiceImpl).makeRoot(file_model.UsersRootPath); err != nil {
			return err
		}

		return db.WithTransaction(ctx, func(ctx context.Context) error {
			err := loadFsTransaction(ctx)
			if err != nil {
				return err
			}

			return nil
		})
	} else if tower_model.Role(cnf.InitRole) == tower_model.RoleBackup {
		err := file_system.RegisterAbsolutePrefix(file_model.BackupTreeKey, filepath.Join(cnf.DataPath, "backup"))
		if err != nil {
			return err
		}

		if err := appCtx.FileService.(*FileServiceImpl).makeRoot(file_model.BackupRootPath); err != nil {
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

	return nil
}

func loadFsTransaction(ctx context.Context) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("not an app context")
	}

	appCtx.Log().Debug().Msg("Loading file system")

	start := time.Now()

	lifetimes, err := history.GetLifetimes(ctx, true)
	if err != nil {
		return err
	}

	appCtx.Log().Debug().Msgf("Found %v lifetimes", len(lifetimes))

	fpMap := make(map[file_system.Filepath]history.FileAction, len(lifetimes))

	for _, lt := range lifetimes {
		a := lt.Actions[len(lt.Actions)-1]

		if a.ActionType == history.FileDelete {
			continue
		}

		a.ContentId = lt.Actions[0].ContentId

		fpMap[a.GetRelevantPath()] = a
	}

	root, err := appCtx.FileService.GetFileById(file_model.UsersTreeKey)
	if err != nil {
		return err
	}

	searchFiles := []file_system.Filepath{root.GetPortablePath()}

	for len(searchFiles) != 0 {
		var fp file_system.Filepath
		fp, searchFiles = searchFiles[0], searchFiles[1:]
		appCtx.Log().Trace().Msgf("Loading file %s", fp.ToAbsolute())

		start := time.Now()
		f := file_model.NewWeblensFile(file_model.NewFileOptions{Path: fp})

		action, ok := fpMap[fp]
		if !ok {
			appCtx.Log().Trace().Msgf("File [%s] not found in history, creating new file", fp)

			if needsContentId(f) {
				_, err = GenerateContentId(ctx, f)
				if err != nil {
					return err
				}
			}

			action = history.NewCreateAction(ctx, f)

			err = history.SaveAction(ctx, &action)
			if err != nil {
				return err
			}

			fpMap[fp] = action
		} else if action.ContentId == "" && needsContentId(f) {
			newContentId, err := GenerateContentId(ctx, f)
			if err != nil {
				return err
			}

			action.ContentId = newContentId

			err = history.UpdateAction(ctx, &action)
			if err != nil {
				return err
			}

			fpMap[fp] = action
		}

		f.SetId(action.FileId)
		f.SetContentId(action.ContentId)

		parentId := ""
		if fp.Dir() == file_model.UsersRootPath {
			parentId = file_model.UsersTreeKey
		} else if fp.Dir() == file_model.BackupRootPath {
			parentId = file_model.BackupTreeKey
		} else {
			parentAction := fpMap[fp.Dir()]
			parentId = parentAction.FileId
		}

		parent, err := appCtx.FileService.GetFileById(parentId)
		if err != nil {
			return fmt.Errorf("Failed to find parent directory [%s]: %w", fp.Dir(), err)
		}

		if parent == nil {
			return fmt.Errorf("Parent directory [%s] not found", fp.Dir())
		}

		err = f.SetParent(parent)
		if err != nil {
			return err
		}

		err = parent.AddChild(f)
		if err != nil {
			return err
		}

		err = appCtx.FileService.AddFile(ctx, f)
		if err != nil {
			return err
		}

		if f.IsDir() {
			children, err := getChildFilepaths(fp)
			if err != nil {
				return err
			}

			appCtx.Log().Trace().Msgf("Found %d children for %s", len(children), fp.ToAbsolute())
			searchFiles = append(searchFiles, children...)
		}

		appCtx.Log().Trace().Msgf("Loaded file %s in %s", fp.ToAbsolute(), time.Since(start))
	}

	appCtx.Log().Debug().Msgf("fs load of %s complete in %s", file_model.UsersRootPath, time.Since(start))
	start = time.Now()

	_, err = root.LoadStat()
	if err != nil {
		return err
	}

	appCtx.Log().Trace().Msgf("Computed tree size in %s", time.Since(start))

	return nil
}

func loadFsTransactionBackup(ctx context.Context) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("not an app context")
	}

	appCtx.Log().Debug().Msg("Loading backup file system")

	backupRoot, err := appCtx.FileService.GetFileById(file_model.BackupTreeKey)
	if err != nil {
		return err
	}

	remotes, err := tower_model.GetRemotes(ctx)
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		if remote.Role != tower_model.RoleCore {
			continue
		}

		remoteDir, err := appCtx.FileService.CreateFolder(ctx, backupRoot, remote.TowerId)
		if err != nil && !errors.Is(err, file_model.ErrDirectoryAlreadyExists) {
			return err
		} else if err != nil {
			remoteDir = file_model.NewWeblensFile(file_model.NewFileOptions{
				Path:   backupRoot.GetPortablePath().Child(remote.TowerId, true),
				FileId: remote.TowerId,
			})

			err = remoteDir.SetParent(backupRoot)
			if err != nil {
				return err
			}

			err = appCtx.FileService.AddFile(ctx, remoteDir)
			if err != nil {
				return err
			}
		}

		children, err := getChildFilepaths(remoteDir.GetPortablePath())
		if err != nil {
			return err
		}

		for _, childPath := range children {
			f := file_model.NewWeblensFile(file_model.NewFileOptions{
				FileId:    childPath.Filename(),
				ContentId: childPath.Filename(),
				Path:      childPath,
			})

			err = f.SetParent(remoteDir)
			if err != nil {
				return err
			}

			err = appCtx.FileService.AddFile(ctx, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
