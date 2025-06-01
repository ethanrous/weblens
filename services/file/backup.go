package file

import (
	"context"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/errors"
	file_system "github.com/ethanrous/weblens/modules/fs"
	context_service "github.com/ethanrous/weblens/services/context"
)

func (fs *FileServiceImpl) NewBackupRestoreFile(ctx context.Context, contentId, remoteTowerId string) (*file_model.WeblensFileImpl, error) {
	restoreRoot, err := fs.GetFileById(ctx, file_model.RestoreTreeKey)
	if err != nil {
		return nil, err
	}

	restorePath := restoreRoot.GetPortablePath().Child(remoteTowerId, true).Child(contentId, false)

	if exists(restorePath) {
		return nil, errors.Errorf("restore file [%s] already exists", restorePath)
	}

	f, err := touch(restorePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Backup paths are in the form of BACKUP:<tower_id>/<path>
// So we check if the root name is BACKUP and the parent of the path is the root (i.e. BACKUP:)
func IsBackupTowerRoot(path file_system.Filepath) bool {
	return path.RootName() == file_model.BackupTreeKey && !path.IsRoot() && path.Dir().IsRoot()
}

func TranslateBackupPath(ctx context_service.AppContext, path file_system.Filepath, core tower_model.Instance) (file_system.Filepath, error) {
	if path.RootName() != file_model.UsersTreeKey {
		return file_system.Filepath{}, errors.Errorf("Path %s is not a user path", path)
	}

	newPath, err := path.ReplacePrefix(file_model.UsersRootPath, file_model.BackupRootPath.Child(core.TowerId, true))
	if err != nil {
		return file_system.Filepath{}, err
	}

	ctx.Log().Trace().Msgf("Translating path %s to %s", path, newPath)

	return newPath, nil
}

func loadFsTransactionBackup(ctx context.Context) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("not an app context")
	}

	appCtx.Log().Debug().Msg("Loading backup file system")

	backupRoot, err := appCtx.FileService.GetFileById(ctx, file_model.BackupTreeKey)
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

		lifetimes, err := history.GetLifetimesByTowerId(ctx, remote.TowerId, history.GetLifetimesOptions{ActiveOnly: true})
		if err != nil {
			return err
		}

		fpMap := make(map[file_system.Filepath]history.FileAction, len(lifetimes))

		for _, lt := range lifetimes {
			a := lt.Actions[len(lt.Actions)-1]

			if a.ActionType == history.FileDelete {
				continue
			}

			a.ContentId = lt.Actions[0].ContentId

			translatedPath, err := TranslateBackupPath(appCtx, a.GetRelevantPath(), remote)
			if err != nil {
				return err
			}

			fpMap[translatedPath] = a
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

		restorePath := file_model.RestoreDirPath.Child(remote.TowerId, true)
		if !exists(restorePath) {
			appCtx.Log().Debug().Msgf("Creating restore path [%s]", restorePath.ToAbsolute())

			_, err = mkdir(restorePath)
			if err != nil {
				return err
			}
		}

		err = loadFilesFromPath(appCtx, remoteDir.GetPortablePath(), fpMap, false, true)
		if err != nil {
			return err
		}
	}

	return nil
}
