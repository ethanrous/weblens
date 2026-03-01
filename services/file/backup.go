package file

import (
	"context"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/tower"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/wlerrors"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// NewBackupRestoreFile creates a new restore file in the backup restore tree for a specific content and tower.
func (fs *ServiceImpl) NewBackupRestoreFile(ctx context.Context, contentID, remoteTowerID string) (*file_model.WeblensFileImpl, error) {
	restoreRoot, err := fs.GetFileByID(ctx, file_model.RestoreTreeKey)
	if err != nil {
		return nil, err
	}

	restorePath := restoreRoot.GetPortablePath().Child(remoteTowerID, true).Child(contentID, false)

	if exists(restorePath) {
		f := file_model.NewWeblensFile(file_model.NewFileOptions{
			ContentID: contentID,
			Path:      restorePath,
		})

		return f, wlerrors.Errorf("not creating restore file at [%s]: %w", restorePath, file_model.ErrFileAlreadyExists)
	}

	f, err := touch(restorePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// IsBackupTowerRoot checks if the given path is a backup tower root directory.
// Backup paths are in the form of BACKUP:<tower_id>/<path>
// So we check if the root name is BACKUP and the parent of the path is the root (i.e. BACKUP:)
func IsBackupTowerRoot(path file_system.Filepath) bool {
	return path.RootName() == file_model.BackupTreeKey && !path.IsRoot() && path.Dir().IsRoot()
}

// TranslateBackupPath translates a user path to its corresponding backup path for a given tower.
func TranslateBackupPath(ctx context_service.AppContext, path file_system.Filepath, core tower_model.Instance) (file_system.Filepath, error) {
	if path.RootName() == file_model.BackupTreeKey {
		// Already a backup path
		return path, nil
	}

	if path.RootName() != file_model.UsersTreeKey {
		return file_system.Filepath{}, wlerrors.Errorf("Path %s is not a user path", path)
	}

	newPath, err := path.ReplacePrefix(file_model.UsersRootPath, file_model.BackupRootPath.Child(core.TowerID, true))
	if err != nil {
		return file_system.Filepath{}, err
	}

	ctx.Log().Trace().Msgf("Translating path %s -> %s", path, newPath)

	return newPath, nil
}

func loadFsTransactionBackup(ctx context.Context) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.New("not an app context")
	}

	appCtx.Log().Debug().Msg("Loading backup file system")

	backupRoot, err := appCtx.FileService.GetFileByID(ctx, file_model.BackupTreeKey)
	if err != nil {
		return err
	}

	remotes, err := tower_model.GetRemotes(ctx)
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		// Only load core remotes
		if remote.Role != tower_model.RoleCore {
			continue
		}

		lifetimes, err := history.GetLifetimes(ctx, history.GetLifetimesOptions{ActiveOnly: true, TowerID: remote.TowerID})
		if err != nil {
			return err
		}

		fpMap := make(map[file_system.Filepath]history.FileAction, len(lifetimes))

		for _, lt := range lifetimes {
			a := lt.Actions[len(lt.Actions)-1]

			if a.ActionType == history.FileDelete {
				continue
			}

			a.ContentID = lt.Actions[0].ContentID

			translatedPath, err := TranslateBackupPath(appCtx, a.GetRelevantPath(), remote)
			if err != nil {
				return err
			}

			fpMap[translatedPath] = a
		}

		remoteDir, err := appCtx.FileService.CreateFolder(ctx, backupRoot, remote.TowerID)
		if err != nil && !wlerrors.Is(err, file_model.ErrDirectoryAlreadyExists) {
			return err
		} else if err != nil {
			remoteDir = file_model.NewWeblensFile(file_model.NewFileOptions{
				Path:   backupRoot.GetPortablePath().Child(remote.TowerID, true),
				FileID: remote.TowerID,
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

		restorePath := file_model.RestoreDirPath.Child(remote.TowerID, true)
		if !exists(restorePath) {
			appCtx.Log().Debug().Msgf("Creating restore path [%s]", restorePath.ToAbsolute())

			_, err = mkdir(restorePath)
			if err != nil {
				return err
			}
		}

		err = LoadFilesRecursively(appCtx, remoteDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeriveFileFromAction creates a WeblensFileImpl from a FileAction, translating the path to the backup location.
// Useful for deriving the correct file object representation when restoring files from backup.
func DeriveFileFromAction(ctx context_service.AppContext, action history.FileAction, core tower.Instance) (*file_model.WeblensFileImpl, error) {
	filePath, err := TranslateBackupPath(ctx, action.GetRelevantPath(), core)
	if err != nil {
		return nil, err
	}

	f := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:      filePath,
		FileID:    action.FileID,
		ContentID: action.ContentID,
	})

	return f, nil
}
