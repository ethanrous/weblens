// Package jobs implements background job tasks.
package jobs

import (
	"context"
	"encoding/json"
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
	"github.com/ethanrous/weblens/modules/set"
	"github.com/ethanrous/weblens/modules/startup"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/rs/zerolog"
)

func init() {
	startup.RegisterHook(func(ctx context.Context, cp config.Provider) error {
		go BackupD(context_mod.ToZ(ctx), cp.BackupInterval)

		return nil
	})
}

// BackupD runs the backup daemon that periodically backs up all connected core servers.
func BackupD(ctx context_mod.Z, interval time.Duration) {
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
				ctx.Log().Error().Stack().Err(wlerrors.Errorf("remote \"%s\" [%s] has no address", remote.Name, remote.TowerID)).Msgf("Skipping backup for remote \"%s\"", remote.Name)

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

// BackupOne initiates a backup task for a single core server.
func BackupOne(ctx context.Context, core tower_model.Instance) (*task.Task, error) {
	meta := job.BackupMeta{
		Core: core,
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, wlerrors.New("Failed to cast context to AppContext")
	}

	return appCtx.DispatchJob(job.BackupTask, meta, nil)
}

// DoBackup executes the backup task for a core server, downloading all changed data and metadata.
func DoBackup(tsk *task.Task) {
	meta := tsk.GetMeta().(job.BackupMeta)

	ctx, ok := context_service.FromContext(tsk.Ctx)
	if !ok {
		tsk.Fail(wlerrors.New("Failed to cast context to FilerContext"))

		return
	}

	if meta.Core.Role != tower_model.RoleCore || (meta.Core.GetReportedRole() != "" && meta.Core.GetReportedRole() != tower_model.RoleCore) {
		tsk.Fail(wlerrors.Errorf("Remote role is [%s -- %s], expected core", meta.Core.Role, meta.Core.GetReportedRole()))
	}

	tsk.Log().Info().Msgf("Starting backup of [%s] with address [%s]", meta.Core.Name, meta.Core.Address)

	notif := notify.NewTaskNotification(tsk, websocket_mod.BackupStartedEvent, task.Result{"coreID": meta.Core.TowerID})
	ctx.Notify(ctx, notif)

	tsk.SetResult(task.Result{
		"coreID": meta.Core.TowerID,
	})

	tsk.OnResult(
		func(r task.Result) {
			notif := notify.NewTaskNotification(
				tsk,
				websocket_mod.BackupProgressEvent,
				r,
			)
			ctx.Notify(ctx, notif)
		},
	)

	tsk.SetErrorCleanup(
		func(errTsk *task.Task) {
			err := errTsk.ReadError()
			notif := notify.NewTaskNotification(tsk, websocket_mod.BackupFailedEvent, task.Result{"coreID": meta.Core.TowerID, "error": err.Error()})
			ctx.Notify(errTsk.Ctx, notif)
		},
	)

	local, err := tower_model.GetLocal(tsk.Ctx)
	if err != nil {
		tsk.Fail(err)
	}

	if local.Role != tower_model.RoleBackup {
		tsk.Fail(tower_model.ErrTowerNotBackup)

		return
	}

	// Check if the core is reachable
	_, err = tower_service.Ping(ctx, meta.Core)
	if err != nil {
		tsk.Fail(err)

		return
	}

	// Find most recent action timestamp
	latestTime := time.UnixMilli(0)

	latestAction, err := history_model.GetLatestActionByTowerID(ctx, meta.Core.TowerID)
	if err != nil && !db.IsNotFound(err) {
		tsk.Fail(err)
	} else if err == nil {
		latestTime = latestAction.Timestamp
	}

	tsk.Log().Trace().Msgf("Backup latest action was at [%s]", latestTime)

	backupResponse, err := tower_service.GetBackup(ctx, meta.Core, latestTime)
	if err != nil {
		tsk.Fail(err)

		return
	}

	ctx.Log().Debug().Func(func(e *zerolog.Event) {
		res, _ := json.Marshal(backupResponse)
		e.Msgf("Received backup response from core [%s]: %s",
			meta.Core.Name,
			res,
		)
	})

	err = db.WithTransaction(ctx, func(ctx context.Context) error {
		appCtx, ok := context_service.FromContext(ctx)
		if !ok {
			return wlerrors.New("Failed to cast context to AppContext")
		}

		appCtx.Log().Trace().Msgf("Got %d users from core", len(backupResponse.Users))
		// Write new the users to db
		for _, userInfo := range backupResponse.Users {
			u := reshape.UserInfoArchiveToUser(userInfo)

			_, err = user_model.GetUserByUsername(tsk.Ctx, u.Username)
			if err == nil {
				continue
			}

			u.CreatedBy = meta.Core.TowerID

			err = user_model.SaveUser(tsk.Ctx, u)
			if err != nil {
				return err
			}
		}

		appCtx.Log().Trace().Msgf("Got %d tokens from core", len(backupResponse.Tokens))
		// Write new keys to db
		for _, key := range backupResponse.Tokens {
			token, err := reshape.TokenInfoToToken(tsk.Ctx, key)
			if err != nil {
				return err
			}

			// Check if token already exists
			_, err = token_model.GetTokenByID(tsk.Ctx, token.ID)
			if err != nil {
				if wlerrors.Is(err, token_model.ErrTokenNotFound) {
					continue
				}

				return err
			}

			err = token_model.SaveToken(tsk.Ctx, token)
			if err != nil {
				return err
			}
		}

		appCtx.Log().Trace().Msgf("Got %d towers from core", len(backupResponse.Instances))
		// Write new towers to db
		for _, serverInfo := range backupResponse.Instances {
			// Check if we already have this tower
			_, err := tower_model.GetBackupTowerByID(tsk.Ctx, serverInfo.ID, meta.Core.TowerID)
			if err == nil {
				continue
			} else if !db.IsNotFound(err) {
				return err
			}

			instance := reshape.TowerInfoToTower(serverInfo)
			instance.CreatedBy = meta.Core.TowerID

			err = tower_model.SaveTower(tsk.Ctx, instance)
			if err != nil {
				return err
			}
		}

		appCtx.Log().Trace().Msgf("Got %d actions from core", len(backupResponse.FileHistory))

		actions := make([]history_model.FileAction, 0, len(backupResponse.FileHistory))

		for _, action := range backupResponse.FileHistory {
			newAction := reshape.FileActionInfoToFileAction(action)
			newAction.TowerID = meta.Core.TowerID
			actions = append(actions, newAction)
		}

		err = history.SaveActions(tsk.Ctx, actions)
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
				if events.Has(a.EventID) {
					continue
				}

				events.Add(a.EventID)
			}

			filteredActions = append(filteredActions, a)
		}

		remoteDataDir, err := appCtx.FileService.InitBackupDirectory(tsk.Ctx, meta.Core)
		if err != nil {
			return err
		}

		// Create a task pool to keep track of copy file tasks,
		// and set it as a child of the main backup task (this task)
		pool, err := appCtx.TaskService.NewTaskPool(true, tsk)
		if err != nil {
			return err
		}

		tsk.SetChildTaskPool(pool)

		for _, a := range filteredActions {
			// Check if the file already exists on the server and copy/move/delete it if it is in the wrong place
			err = handleFileAction(appCtx, a, meta.Core, tsk, pool)
			if err != nil {
				return err
			}
		}

		tsk.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Waiting for %d copy file tasks", pool.Status().Total) })

		// Wait for all copy file tasks to finish
		pool.SignalAllQueued()
		pool.Wait(true)

		if len(pool.Errors()) != 0 {
			return wlerrors.Errorf("%d of %d backup file copies have failed", len(pool.Errors()), pool.Status().Total)
		}

		_, err = remoteDataDir.LoadStat()
		if err != nil {
			return err
		}

		err = tower_model.SetLastBackup(tsk.Ctx, meta.Core.TowerID, time.Now(), remoteDataDir.Size())
		if err != nil {
			return err
		}

		r := tsk.GetResult()
		r["backupSize"] = remoteDataDir.Size()
		r["totalTime"] = tsk.ExeTime()
		r["complete"] = true

		notif := notify.NewTaskNotification(
			tsk,
			websocket_mod.BackupCompleteEvent,
			r,
		)
		appCtx.Notify(ctx, notif)

		return nil
	})
	if err != nil {
		tsk.Fail(err)

		return
	}

	tsk.Success()
}

func getExistingFile(ctx context_service.AppContext, a history_model.FileAction, core tower_model.Instance) (*file_model.WeblensFileImpl, error) {
	path, err := file_service.TranslateBackupPath(ctx, a.GetOriginPath(), core)
	if err != nil {
		return nil, err
	}

	existingFile, err := ctx.FileService.GetFileByFilepath(ctx, path)
	if err != nil && !wlerrors.Is(err, file_model.ErrFileNotFound) {
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

func handleFileAction(ctx context_service.AppContext, a history_model.FileAction, core tower_model.Instance, callingTask *task.Task, pool *task.Pool) error {
	existingFile, err := getExistingFile(ctx, a, core)
	if err != nil {
		return err
	}

	backupFilePath, err := file_service.TranslateBackupPath(ctx, a.GetDestinationPath(), core)
	if err != nil {
		return err
	}

	ctx.Log().Trace().Msgf("Handling backup action [%s] for file [%s]", a.ActionType, backupFilePath)

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

	// Handle directory creation, no need to continue afterwards since directories have no content to copy
	if backupFilePath.IsDir() && a.ActionType == history_model.FileCreate {
		parent, err := ctx.FileService.GetFileByFilepath(ctx, backupFilePath.Dir())
		if err != nil {
			return fmt.Errorf("failed to get new backup directory parent [%s]: %w", backupFilePath.Dir(), err)
		}

		_, err = ctx.FileService.CreateFolder(ctx, parent, backupFilePath.Filename())
		if wlerrors.Is(err, file_model.ErrDirectoryAlreadyExists) {
			// Directory already exists, derive the file from the action to ensure correct representation
			f, err := file_service.DeriveFileFromAction(ctx, a, core)
			if err != nil {
				return wlerrors.Errorf("failed to derive existing backup directory [%s]: %w", backupFilePath, err)
			}

			err = f.SetParent(parent)
			if err != nil {
				return wlerrors.Errorf("failed to set parent for existing backup directory [%s]: %w", backupFilePath, err)
			}

			err = ctx.FileService.AddFile(ctx, f)
			if err != nil {
				return wlerrors.Errorf("failed to add existing backup directory [%s]: %w", backupFilePath, err)
			}
		} else if err != nil {
			return err
		}

		ctx.Log().Trace().Msgf("Created backup directory [%s]", backupFilePath)

		return nil
	}

	// If the action has no contentID, it means the file was created but has no content
	if a.ContentID == "" {
		ctx.Log().Trace().Msgf("File %s has no contentID", a.GetRelevantPath())

		return nil
	}

	restoreFile := file_model.NewWeblensFile(file_model.NewFileOptions{
		FileID:    a.FileID,
		Path:      backupFilePath,
		ContentID: a.ContentID,
	})

	parentDir, err := ctx.FileService.GetFileByFilepath(ctx, restoreFile.GetPortablePath().Dir())
	if err != nil {
		return err
	}

	err = restoreFile.SetParent(parentDir)
	if err != nil {
		return err
	}

	ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Queuing copy file task for %s", restoreFile.GetPortablePath()) })

	// Spawn subtask to copy the file from the core server
	copyFileMeta := job.BackupCoreFileMeta{
		File:       restoreFile,
		Core:       core,
		CoreFileID: a.FileID,
		Filename:   backupFilePath.Filename(),
	}

	// Launch the copy file task in the provided pool to copy the file from the core server
	tsk, err := ctx.DispatchJob(job.CopyFileFromCoreTask, copyFileMeta, pool)
	if err != nil {
		return err
	}

	callingTask.AtomicSetResult(func(currentResult task.Result) task.Result {
		currentBytesToCopyI, ok := currentResult["totalBytesToCopy"]

		var currentBytesToCopy int64
		if !ok {
			currentBytesToCopy = 0
		} else {
			currentBytesToCopy = currentBytesToCopyI.(int64)
		}

		currentResult["totalBytesToCopy"] = currentBytesToCopy + a.Size

		return currentResult
	})

	tsk.SetCleanup(func(tsk *task.Task) {
		copyTaskResult := tsk.GetResult()

		callingTask.AtomicSetResult(func(currentResult task.Result) task.Result {
			currentBytesCopiedI, ok := currentResult["totalBytesCopied"]

			var currentBytesCopied int64
			if !ok {
				currentBytesCopied = 0
			} else {
				currentBytesCopied = currentBytesCopiedI.(int64)
			}

			currentResult["totalBytesCopied"] = currentBytesCopied + copyTaskResult["bytesCopied"].(int64)

			return currentResult
		})
	})

	return nil
}
