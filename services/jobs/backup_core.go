package jobs

import (
	"context"
	"fmt"
	"slices"
	"time"

	token_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	history_model "github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/set"
	"github.com/ethanrous/weblens/modules/startup"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_service "github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/rs/zerolog"
)

func init() {
	startup.RegisterStartup(func(ctx context.Context, cp config.ConfigProvider) error {
		go BackupD(context_mod.ToZ(ctx), cp.BackupInterval)

		return nil
	})
}

func BackupD(ctx context_mod.ContextZ, interval time.Duration) {
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to get local instance for backup service")

		return
	}

	if local.Role != tower_model.RoleBackup {
		return
	}

	for {
		remotes, err := tower_model.GetRemotes(ctx)
		if err != nil {
			return
		}

		for _, remote := range remotes {
			if remote.Role != tower_model.RoleCore {
				continue
			} else if remote.Address == "" {
				ctx.Log().Error().Stack().Err(errors.Errorf("remote \"%s\" [%s] has no address", remote.Name, remote.TowerId)).Msgf("Skipping backup for remote \"%s\"", remote.Name)

				continue
			}

			_, err := BackupOne(ctx, remote)
			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msg("")
			}
		}

		now := time.Now()
		sleepFor := now.Truncate(interval).Add(interval).Sub(now)

		select {
		case <-time.Tick(sleepFor):
		case <-ctx.Done():
			ctx.Log().Debug().Msg("BackupD exiting")

			return
		}
	}
}

func BackupOne(ctx context_mod.DispatcherContext, core tower_model.Instance) (task_mod.Task, error) {
	meta := job.BackupMeta{
		Core: core,
	}

	return ctx.DispatchJob(job.BackupTask, meta, nil)
}

func DoBackup(tsk task_mod.Task) {
	t := tsk.(*task.Task)
	meta := t.GetMeta().(job.BackupMeta)

	ctx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))

		return
	}

	if meta.Core.Role != tower_model.RoleCore || (meta.Core.GetReportedRole() != "" && meta.Core.GetReportedRole() != tower_model.RoleCore) {
		t.Fail(errors.Errorf("Remote role is [%s -- %s], expected core", meta.Core.Role, meta.Core.GetReportedRole()))
	}

	t.Log().Debug().Msgf("Starting backup of [%s] with adddress [%s] using key [%s]", meta.Core.Name, meta.Core.Address, meta.Core.OutgoingKey)

	t.OnResult(
		func(r task_mod.TaskResult) {
			notif := notify.NewTaskNotification(t, websocket_mod.BackupProgressEvent, r)
			ctx.Notify(ctx, notif)
		},
	)

	t.SetErrorCleanup(
		func(errTsk task_mod.Task) {
			err := errTsk.ReadError()
			notif := notify.NewTaskNotification(t, websocket_mod.BackupFailedEvent, task_mod.TaskResult{"coreId": meta.Core.TowerId, "error": err.Error()})
			ctx.Notify(ctx, notif)
		},
	)

	local, err := tower_model.GetLocal(t.Ctx)
	if err != nil {
		t.Fail(err)
	}

	if local.Role != tower_model.RoleBackup {
		t.Fail(tower_model.ErrTowerNotBackup)

		return
	}

	// Check if the core is reachable
	_, err = tower_service.Ping(ctx, meta.Core)
	if err != nil {
		t.Fail(err)
		return
	}

	// Find most recent action timestamp
	latestTime := time.UnixMilli(0)

	latestAction, err := history_model.GetLatestAction(ctx)
	if err != nil && !db.IsNotFound(err) {
		t.Fail(err)
	} else if err == nil {
		latestTime = latestAction.Timestamp
	}

	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Backup latest action is %s", latestTime) })

	backupResponse, err := tower_service.GetBackup(ctx, meta.Core, latestTime)
	if err != nil {
		t.Fail(err)

		return
	}

	err = db.WithTransaction(ctx, func(ctx context.Context) error {
		appCtx, ok := context_service.FromContext(ctx)
		if !ok {
			return errors.New("Failed to cast context to AppContext")
		}

		appCtx.Log().Trace().Msgf("Got %d users from core", len(backupResponse.Users))
		// Write new the users to db
		for _, userInfo := range backupResponse.Users {
			u := reshape.UserInfoArchiveToUser(userInfo)

			_, err = user_model.GetUserByUsername(t.Ctx, u.Username)
			if err == nil {
				continue
			}

			u.CreatedBy = meta.Core.TowerId

			err = user_model.SaveUser(t.Ctx, u)
			if err != nil {
				return err
			}
		}

		appCtx.Log().Trace().Msgf("Got %d tokens from core", len(backupResponse.Tokens))
		// Write new keys to db
		for _, key := range backupResponse.Tokens {
			token, err := reshape.TokenInfoToToken(t.Ctx, key)
			if err != nil {
				return err
			}

			// Check if token already exists
			_, err = token_model.GetTokenById(t.Ctx, token.Id)
			if err != nil {
				if errors.Is(err, token_model.ErrTokenNotFound) {
					continue
				}

				return err
			}

			err = token_model.SaveToken(t.Ctx, token)
			if err != nil {
				return err
			}
		}

		appCtx.Log().Trace().Msgf("Got %d towers from core", len(backupResponse.Instances))
		// Write new towers to db
		for _, serverInfo := range backupResponse.Instances {
			// Check if we already have this tower
			_, err := tower_model.GetBackupTowerById(t.Ctx, serverInfo.Id, meta.Core.TowerId)
			if err == nil {
				continue
			} else if !db.IsNotFound(err) {
				return err
			}

			instance := reshape.ApiTowerInfoToTower(serverInfo)
			instance.CreatedBy = meta.Core.TowerId

			err = tower_model.SaveTower(t.Ctx, &instance)
			if err != nil {
				return err
			}
		}

		appCtx.Log().Trace().Msgf("Got %d actions from core", len(backupResponse.FileHistory))

		actions := make([]history_model.FileAction, 0, len(backupResponse.FileHistory))

		for _, action := range backupResponse.FileHistory {
			newAction := reshape.FileActionInfoToFileAction(action)
			newAction.TowerId = meta.Core.TowerId
			actions = append(actions, newAction)
		}

		err = history.SaveActions(t.Ctx, actions)
		if err != nil {
			return err
		}

		// Sort lifetimes so that files created or moved most recently are updated last.
		// This is to make sure parent directories are created before their children
		slices.SortFunc(actions, history_model.ActionSorter)

		filteredActions := make([]history_model.FileAction, 0, len(actions))
		events := set.New[string]()

		for _, a := range actions {
			if a.ActionType == history_model.FileMove {
				if events.Has(a.EventId) {
					continue
				}

				events.Add(a.EventId)
			}

			filteredActions = append(filteredActions, a)
		}

		appCtx.WithValue(file_service.SkipJournalKey, true)

		remoteDataDir, err := appCtx.FileService.InitBackupDirectory(t.Ctx, meta.Core)
		if err != nil {
			return err
		}

		pool := appCtx.TaskService.NewTaskPool(true, t)
		t.SetChildTaskPool(pool)

		for _, a := range filteredActions {
			// Check if the file already exists on the server and copy/move/delete it if it is in the wrong place
			err = handleFileAction(appCtx, a, meta.Core, pool)
			if err != nil {
				return err
			}
		}

		t.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Waiting for %d copy file tasks", pool.Status().Total) })

		// Wait for all copy file tasks to finish
		pool.SignalAllQueued()
		pool.Wait(true)

		if len(pool.Errors()) != 0 {
			return errors.Errorf("%d of %d backup file copies have failed", len(pool.Errors()), pool.Status().Total)
		}

		err = tower_model.SetLastBackup(t.Ctx, meta.Core.TowerId, time.Now())
		if err != nil {
			return err
		}

		// Don't broadcast this last event set
		t.OnResult(nil)

		_, err = remoteDataDir.LoadStat()
		if err != nil {
			return err
		}

		t.SetResult(task_mod.TaskResult{
			"backupSize": remoteDataDir.Size(),
			"totalTime":  t.ExeTime(),
		})

		endNotif := notify.NewTaskNotification(t, websocket_mod.BackupCompleteEvent, t.GetResults())
		appCtx.Notify(ctx, endNotif)

		return nil
	})

	if err != nil {
		t.Fail(err)

		return
	}

	t.Success()
}

func getExistingFile(ctx context_service.AppContext, a history_model.FileAction, core tower_model.Instance) (*file_model.WeblensFileImpl, error) {
	path, err := file_service.TranslateBackupPath(ctx, a.GetOriginPath(), core)
	if err != nil {
		return nil, err
	}

	existingFile, err := ctx.FileService.GetFileByFilepath(ctx, path)
	if err != nil && !errors.Is(err, file_model.ErrFileNotFound) {
		return nil, err
	}

	ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("File %s exists: %v", path, existingFile != nil) })

	// If the file already exists, but is the wrong size, an earlier copy most likely failed. Delete it and copy it again.
	if existingFile != nil && !existingFile.IsDir() && existingFile.Size() != a.Size {
		err = ctx.FileService.DeleteFiles(ctx, existingFile)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	return existingFile, nil
}

func handleFileAction(ctx context_service.AppContext, a history_model.FileAction, core tower_model.Instance, pool task_mod.Pool) error {
	existingFile, err := getExistingFile(ctx, a, core)
	if err != nil {
		return err
	}

	backupFilePath, err := file_service.TranslateBackupPath(ctx, a.GetDestinationPath(), core)
	if err != nil {
		return err
	}

	if existingFile != nil {
		if a.ActionType == history_model.FileDelete {
			return ctx.FileService.DeleteFiles(ctx, existingFile)
		}

		if backupFilePath != existingFile.GetPortablePath() {
			newParent, err := ctx.FileService.GetFileByFilepath(ctx, backupFilePath.Dir())
			if err != nil {
				return err
			}

			return ctx.FileService.MoveFiles(ctx, []*file_model.WeblensFileImpl{existingFile}, newParent)
		}

		return nil
	}

	if backupFilePath.IsDir() && a.ActionType == history_model.FileCreate {
		parent, err := ctx.FileService.GetFileByFilepath(ctx, backupFilePath.Dir())
		if err != nil {
			return fmt.Errorf("failed to get new backup directory parent [%s]: %w", backupFilePath.Dir(), err)
		}

		// If the file is a directory, create it and return
		_, err = ctx.FileService.CreateFolder(ctx, parent, backupFilePath.Filename())
		if err != nil && !errors.Is(err, file_model.ErrDirectoryAlreadyExists) {
			return err
		}

		ctx.Log().Trace().Msgf("Created backup directory [%s]", backupFilePath)

		return nil
	}

	if a.ContentId == "" {
		ctx.Log().Trace().Msgf("File %s has no contentId", a.GetRelevantPath())

		return nil
	}

	restoreFile := file_model.NewWeblensFile(file_model.NewFileOptions{
		FileId:    a.FileId,
		Path:      backupFilePath,
		ContentId: a.ContentId,
	})

	ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Queuing copy file task for %s", restoreFile.GetPortablePath()) })

	// Spawn subtask to copy the file from the core server
	copyFileMeta := job.BackupCoreFileMeta{
		File:       restoreFile,
		Core:       core,
		CoreFileId: a.FileId,
		Filename:   backupFilePath.Filename(),
	}

	_, err = ctx.DispatchJob(job.CopyFileFromCoreTask, copyFileMeta, pool)
	if err != nil {
		return err
	}

	return nil
}
