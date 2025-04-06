package jobs

import (
	"slices"
	"strconv"
	"time"

	token_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	history_model "github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/structs"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func BackupD(ctx context_mod.ContextZ, interval time.Duration) {
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to get local instance for backup service")
		return
	}
	if local.Role != tower_model.BackupServerRole {
		ctx.Log().Error().Msg("Backup service cannot be run on non-backup instance")
		return
	}
	for {
		now := time.Now()
		sleepFor := now.Truncate(interval).Add(interval).Sub(now)
		ctx.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("BackupD going to sleep for %s", sleepFor) })
		time.Sleep(sleepFor)

		remotes, err := tower_model.GetRemotes(ctx)
		if err != nil {
			return
		}
		for _, remote := range remotes {
			_, err := BackupOne(ctx, remote)
			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msg("")
			}
		}
	}
}

func BackupOne(ctx context_mod.DispatcherContext, core *tower_model.Instance) (task_mod.Task, error) {
	meta := job.BackupMeta{
		Core: core,
	}
	return ctx.DispatchJob(job.BackupTask, meta, nil)
}

func DoBackup(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	meta := t.GetMeta().(job.BackupMeta)

	ctx, ok := t.Ctx.(*context.AppContext)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))
		return
	}

	if meta.Core.Role != tower_model.CoreServerRole || (meta.Core.GetReportedRole() != "" && meta.Core.GetReportedRole() != tower_model.CoreServerRole) {
		t.Fail(errors.Errorf("Remote role is [%s -- %s], expected core", meta.Core.Role, meta.Core.GetReportedRole()))
	}

	t.Ctx.Log().Debug().Msgf("Starting backup of [%s] with adddress [%s] using key [%s]", meta.Core.Name, meta.Core.Address, meta.Core.OutgoingKey)

	t.OnResult(
		func(r task_mod.TaskResult) {
			notif := client.NewTaskNotification(t, websocket_mod.BackupCompleteEvent, r)
			t.Ctx.Notify(notif)
		},
	)

	t.SetErrorCleanup(
		func(errTsk task_mod.Task) {
			err := errTsk.ReadError()
			notif := client.NewTaskNotification(t, websocket_mod.BackupFailedEvent, task_mod.TaskResult{"coreId": meta.Core.TowerId, "error": err.Error()})
			t.Ctx.Notify(notif)
		},
	)

	local, err := tower_model.GetLocal(t.Ctx)
	if err != nil {
		t.Fail(err)
	}
	if local.Role == tower_model.InitServerRole {
		t.ReqNoErr(tower_model.ErrServerNotInitialized)
	} else if local.Role != tower_model.BackupServerRole {
		t.ReqNoErr(tower_model.ErrServerIsBackup)
	}

	stages := ""
	t.SetResult(task_mod.TaskResult{"stages": stages, "coreId": meta.Core.TowerId})

	// Find most recent action timestamp
	latestTime, err := history_model.GetLatestAction(t.Ctx)
	if err != nil {
		t.Fail(err)
	}

	t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Backup latest action is %s", latestTime) })

	// stages.StartStage("fetching_backup_data")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	backupRq := proxy.NewCoreRequest(meta.Core, "GET", "/servers/backup").
		WithQuery("timestamp", strconv.FormatInt(latestTime.Timestamp.UnixMilli(), 10)).
		WithHeader("Wl-Server-Id", local.TowerId)
	backupResponse, err := proxy.CallHomeStruct[structs.BackupInfo](backupRq)
	t.ReqNoErr(err)

	// stages.StartStage("writing_users")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	// Write new the users to db
	for _, userInfo := range backupResponse.Users {
		u := reshape.UserInfoArchiveToUser(userInfo)
		err = user_model.CreateUser(t.Ctx, u)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	// stages.StartStage("writing_keys")
	t.SetResult(task_mod.TaskResult{"stages": stages})

	// Write new keys to db
	for _, key := range backupResponse.Tokens {
		token := reshape.TokenInfoToToken(t.Ctx, key)

		// Check if token already exists
		_, err := token_model.GetTokenById(t.Ctx, token.Id)
		if err != nil {
			if errors.Is(err, token_model.ErrTokenNotFound) {
				continue
			}
			t.Fail(err)
		}

		token_model.SaveToken(t.Ctx, token)
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

		instance := reshape.TowerInfoToTower(serverInfo)
		instance.CreatedBy = meta.Core.TowerId

		err = tower_model.CreateTower(t.Ctx, instance)
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

	req := proxy.NewCoreRequest(meta.Core, "GET", "/journal").WithQuery("timestamp", "0")
	allActions, err := proxy.CallHomeStruct[[]*history_model.FileAction](req)
	t.ReqNoErr(err)
	for _, action := range allActions {
		err := history_model.SaveAction(t.Ctx, action)
		target := &db.AlreadyExistsError{}
		if errors.As(err, &target) {
			// This can happen if the lifetime already exists in our journal
			// We can ignore this error safely
			continue
		} else if err != nil {
			t.Fail(err)
		}
	}

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

	restoreRoot, err := ctx.FileService.GetFileById(file_service.RestoreDirPath.ToPortable())
	if err != nil {
		t.Fail(err)
	}

	// Check if the file already exists on the server and copy it if it doesn't
	for _, a := range actions {
		existingFile, err := ctx.FileService.GetFileByFilepath(a.GetRelevantPath())
		t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("File %s exists: %v", a.GetRelevantPath(), existingFile != nil) })

		// If the file already exists, but is the wrong size, an earlier copy most likely failed. Delete it and copy it again.
		if existingFile != nil && !existingFile.IsDir() && existingFile.Size() != a.Size {
			err = ctx.FileService.DeleteFiles(t.Ctx, []*file_model.WeblensFileImpl{existingFile})
			t.ReqNoErr(err)
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

				err = ctx.FileService.MoveFiles(t.Ctx, []*file_model.WeblensFileImpl{existingFile}, newParent, meta.Core.TowerId)
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

		restoreFile, err := ctx.FileService.CreateFile(restoreRoot, a.ContentId)
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

	endNotif := client.NewTaskNotification(t, websocket_mod.BackupCompleteEvent, t.GetResults())
	t.Ctx.Notify(endNotif)

	t.Success()
}
