package jobs

import (
	"context"
	"slices"
	"time"

	token_model "github.com/ethanrous/weblens/models/auth"
	file_model "github.com/ethanrous/weblens/models/file"
	history_model "github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/startup"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/pkg/errors"
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

	ctx, ok := t.Ctx.(context_service.AppContext)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))

		return
	}

	if meta.Core.Role != tower_model.RoleCore || (meta.Core.GetReportedRole() != "" && meta.Core.GetReportedRole() != tower_model.RoleCore) {
		t.Fail(errors.Errorf("Remote role is [%s -- %s], expected core", meta.Core.Role, meta.Core.GetReportedRole()))
	}

	t.Ctx.Log().Debug().Msgf("Starting backup of [%s] with adddress [%s] using key [%s]", meta.Core.Name, meta.Core.Address, meta.Core.OutgoingKey)

	t.OnResult(
		func(r task_mod.TaskResult) {
			notif := notify.NewTaskNotification(t, websocket_mod.BackupCompleteEvent, r)
			t.Ctx.Notify(ctx, notif)
		},
	)

	t.SetErrorCleanup(
		func(errTsk task_mod.Task) {
			err := errTsk.ReadError()
			notif := notify.NewTaskNotification(t, websocket_mod.BackupFailedEvent, task_mod.TaskResult{"coreId": meta.Core.TowerId, "error": err.Error()})
			t.Ctx.Notify(ctx, notif)
		},
	)

	local, err := tower_model.GetLocal(t.Ctx)
	if err != nil {
		t.Fail(err)
	}

	if local.Role == tower_model.RoleInit {
		t.ReqNoErr(tower_model.ErrTowerNotInitialized)
	} else if local.Role != tower_model.RoleBackup {
		t.ReqNoErr(tower_model.ErrTowerNotFound)
	}

	stages := ""
	t.SetResult(task_mod.TaskResult{"stages": stages, "coreId": meta.Core.TowerId})

	// Find most recent action timestamp
	latestTime, err := history_model.GetLatestAction(t.Ctx)
	if err != nil {
		t.Fail(err)
	}

	t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Backup latest action is %s", latestTime.Timestamp) })

	// stages.StartStage("fetching_backup_data")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	backupResponse, err := tower_service.GetBackup(t.Ctx, meta.Core, time.UnixMilli(0))
	if err != nil {
		t.Fail(err)

		return
	}

	// stages.StartStage("writing_users")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	// Write new the users to db
	for _, userInfo := range backupResponse.Users {
		u := reshape.UserInfoArchiveToUser(userInfo)

		err = user_model.SaveUser(t.Ctx, u)
		if err != nil {
			t.Fail(err)

			return
		}
	}

	// stages.StartStage("writing_keys")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	// Write new keys to db
	for _, key := range backupResponse.Tokens {
		token, err := reshape.TokenInfoToToken(t.Ctx, key)
		if err != nil {
			t.Fail(err)

			return
		}

		// Check if token already exists
		_, err = token_model.GetTokenById(t.Ctx, token.Id)
		if err != nil {
			if errors.Is(err, token_model.ErrTokenNotFound) {
				continue
			}

			t.Fail(err)

			return
		}

		err = token_model.SaveToken(t.Ctx, token)
		if err != nil {
			t.Fail(err)

			return
		}
	}

	// stages.StartStage("writing_instances")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	// Write new towers to db
	for _, serverInfo := range backupResponse.Instances {
		// Check if we already have this tower
		_, err := tower_model.GetTowerById(t.Ctx, serverInfo.Id)
		if err != nil {
			if errors.Is(err, tower_model.ErrTowerNotFound) {
				continue
			}

			t.Fail(err)
		}

		instance := reshape.ApiTowerInfoToTower(serverInfo)
		instance.CreatedBy = meta.Core.TowerId

		err = tower_model.SaveTower(t.Ctx, &instance)
		if err != nil {
			t.Fail(err)
		}
	}

	// stages.StartStage("sync_journal")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	t.Ctx.Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Backup got %d updated actions from core", len(backupResponse.FileHistory))
	})

	// stages.StartStage("sync_fs")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	// TODO: Do something with these actions
	var actions []*history_model.FileAction
	for _, action := range backupResponse.FileHistory {
		actions = append(actions, reshape.FileActionInfoToFileAction(action))
	}

	// Sort lifetimes so that files created or moved most recently are updated last.
	slices.SortFunc(actions, history_model.ActionSorter)

	// req := proxy.NewCoreRequest(&meta.Core, "GET", "/journal").WithQuery("timestamp", "0")
	// allActions, err := proxy.CallHomeStruct[[]*history_model.FileAction](req)
	// t.ReqNoErr(err)
	// for _, action := range allActions {
	// 	err := history_model.SaveAction(t.Ctx, action)
	// 	target := &db.AlreadyExistsError{}
	// 	if errors.As(err, &target) {
	// 		// This can happen if the lifetime already exists in our journal
	// 		// We can ignore this error safely
	// 		continue
	// 	} else if err != nil {
	// 		t.Fail(err)
	// 	}
	// }

	// Get all lifetimes we currently know about and find which files are new
	// and therefore need to be created or copied from the core
	actions, err = history_model.GetActionsByTowerId(t.Ctx, meta.Core.TowerId)
	if err != nil {
		t.Fail(err)
	}

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	t.SetChildTaskPool(pool)

	// Sort lifetimes so that files created or moved most recently are updated last.
	// This is to make sure parent directories are created before their children
	slices.SortFunc(actions, history_model.ActionSorter)

	restoreRoot, err := ctx.FileService.GetFileById(file_model.RestoreDirPath.ToPortable())
	if err != nil {
		t.Fail(err)
	}

	// Check if the file already exists on the server and copy it if it doesn't
	for _, a := range actions {
		existingFile, err := ctx.FileService.GetFileByFilepath(ctx, a.GetRelevantPath())

		t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("File %s exists: %v", a.GetRelevantPath(), existingFile != nil) })

		// If the file already exists, but is the wrong size, an earlier copy most likely failed. Delete it and copy it again.
		if existingFile != nil && !existingFile.IsDir() && existingFile.Size() != a.Size {
			err = ctx.FileService.DeleteFiles(t.Ctx, []*file_model.WeblensFileImpl{existingFile})
			if err != nil {
				t.Fail(err)

				return
			}

			existingFile = nil
		}

		if err == nil && existingFile != nil {
			if a.ActionType == history_model.FileDelete {
				err = ctx.FileService.DeleteFiles(t.Ctx, []*file_model.WeblensFileImpl{existingFile})
				t.ReqNoErr(err)
			} else if a.DestinationPath != existingFile.GetPortablePath() {
				parentPath := fs.BuildFilePath(meta.Core.TowerId, a.DestinationPath.Dir().RelPath)

				newParent, err := ctx.FileService.GetFileById(parentPath.ToPortable())
				if err != nil {
					t.ReqNoErr(err)
				}

				err = ctx.FileService.MoveFiles(t.Ctx, []*file_model.WeblensFileImpl{existingFile}, newParent)
				t.ReqNoErr(err)
			}

			continue
		} else if err != nil && !errors.Is(err, file_model.ErrFileNotFound) {
			t.Fail(err)
		}

		if a.ActionType == history_model.FileDelete {
			path := a.GetRelevantPath()
			if path.IsDir() {
				continue
			}

			f, err := ctx.FileService.GetFileByContentId(a.ContentId)
			if err == nil && f.Size() == a.Size {
				continue
			} else if !errors.Is(err, file_model.ErrFileNotFound) {
				t.Fail(err)
			}
		}

		restoreFile, err := ctx.FileService.CreateFile(ctx, restoreRoot, a.ContentId)
		if err != nil {
			t.Fail(err)
		}

		if restoreFile == nil || restoreFile.Size() != 0 {
			continue
		}

		t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Queuing copy file task for %s", restoreFile.GetPortablePath()) })

		// Spawn subtask to copy the file from the core server
		copyFileMeta := job.BackupCoreFileMeta{
			File:       restoreFile,
			Core:       meta.Core,
			CoreFileId: a.GetDestinationPath().ToPortable(),
			Filename:   a.GetRelevantPath().Filename(),
		}

		_, err = t.Ctx.DispatchJob(job.CopyFileFromCoreTask, copyFileMeta, nil)
		if err != nil {
			t.Fail(err)
		}
	}

	t.Ctx.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Waiting for %d copy file tasks", pool.Status().Total) })

	pool.SignalAllQueued()
	pool.Wait(true)

	if len(pool.Errors()) != 0 {
		t.Fail(errors.Errorf("%d of %d backup file copies have failed", len(pool.Errors()), pool.Status().Total))
	}

	// stages.FinishStage("sync_fs")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	// coreTree, err := ctx.FileService.GetFileTreeByName(meta.Core.TowerId)
	// if err != nil {
	// 	t.Fail(err)
	// }
	//
	// root := coreTree.GetRoot()
	//
	// err = ctx.FileService.ResizeDown(root, nil)
	// t.ReqNoErr(err)

	tower_model.SetLastBackup(t.Ctx, meta.Core.TowerId, time.Now())
	t.ReqNoErr(err)

	// Don't broadcast this last event set
	t.OnResult(nil)
	t.SetResult(task_mod.TaskResult{
		// "backupSize": root.GetSize(),
		"backupSize": -1,
		"totalTime":  t.ExeTime(),
	})

	endNotif := notify.NewTaskNotification(t, websocket_mod.BackupCompleteEvent, t.GetResults())
	t.Ctx.Notify(ctx, endNotif)

	t.Success()
}
