package jobs

import (
	"slices"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/models"
	token_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/client"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/rest"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
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
	return ctx.DispatchJob(job.BackupTask, meta)
}

func DoBackup(t *task_model.Task) {
	meta := t.GetMeta().(job.BackupMeta)

	filerCtx, ok := t.Ctx.(context.FilerContext)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))
		return
	}
	fileService := filerCtx.FileService()

	if meta.Core.Role != tower_model.CoreServerRole || (meta.Core.GetReportedRole() != "" && meta.Core.GetReportedRole() != tower_model.CoreServerRole) {
		t.Fail(errors.Errorf("Remote role is [%s -- %s], expected core", meta.Core.Role, meta.Core.GetReportedRole()))
	}

	t.Ctx.Log().Debug().Msgf("Starting backup of [%s] with adddress [%s] using key [%s]", meta.Core.Name, meta.Core.Address, meta.Core.OutgoingKey)

	t.OnResult(
		func(r task_model.TaskResult) {
			notif := client.NewTaskNotification(t, websocket_mod.BackupCompleteEvent, r)
			t.Ctx.Notify(notif)
		},
	)

	t.SetErrorCleanup(
		func(errTsk *task_model.Task) {
			err := errTsk.ReadError()
			if err == nil {
				t.Ctx.Log().Error().Msg("Trying to show error in backup task, but error is nil")
				return
			}

			notif := client.NewTaskNotification(t, websocket_mod.BackupFailedEvent, task_model.TaskResult{"coreId": meta.Core.TowerId, "error": err.Error()})
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

	stages := models.NewBackupTaskStages()

	stages.StartStage("connecting")
	t.SetResult(task_model.TaskResult{"stages": stages, "coreId": meta.Core.TowerId})

	// Find most recent action timestamp
	latestTime, err := history.GetLastJournalUpdate(t.Ctx)
	if err != nil {
		t.Fail(err)
	}

	t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Backup latest action is %s", latestTime.String()) })

	stages.StartStage("fetching_backup_data")
	t.SetResult(task_model.TaskResult{"stages": stages})

	backupRq := proxy.NewCoreRequest(meta.Core, "GET", "/servers/backup").
		WithQuery("timestamp", strconv.FormatInt(latestTime.UnixMilli(), 10)).
		WithHeader("Wl-Server-Id", local.TowerId)
	backupResponse, err := proxy.CallHomeStruct[rest.BackupInfo](backupRq)
	t.ReqNoErr(err)

	stages.StartStage("writing_users")
	t.SetResult(task_model.TaskResult{"stages": stages})

	// Write new the users to db
	for _, userInfo := range backupResponse.Users {
		u := reshape.UserInfoArchiveToUser(userInfo)
		err = user_model.CreateUser(t.Ctx, u)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	stages.StartStage("writing_keys")
	t.SetResult(task_model.TaskResult{"stages": stages})

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

	stages.StartStage("writing_instances")
	t.SetResult(task_model.TaskResult{"stages": stages})

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

	stages.StartStage("sync_journal")
	t.SetResult(task_model.TaskResult{"stages": stages})

	t.Ctx.Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Backup got %d updated lifetimes from core", len(backupResponse.FileHistory))
	})

	stages.StartStage("sync_fs")
	t.SetResult(task_model.TaskResult{"stages": stages})

	// Sort lifetimes so that files created or moved most recently are updated last.
	slices.SortFunc(backupResponse.FileHistory, history.LifetimeSorter)

	req := proxy.NewCoreRequest(meta.Core, "GET", "/journal").WithQuery("timestamp", "0")
	allLts, err := proxy.CallHomeStruct[[]*history.Lifetime](req)
	t.ReqNoErr(err)
	for _, lt := range allLts {
		err := history.WriteNewLifetime(t.Ctx, lt)
		if errors.Is(err, history.ErrLifetimeAlreadyExists) {
			// This can happen if the lifetime already exists in our journal
			// We can ignore this error safely
			t.Ctx.Log().Trace().Func(func(e *zerolog.Event) {
				e.Msgf("Ignoring duplicate lifetime [%s]", lt.ID())
			})
			continue
		} else if err != nil {
			t.Fail(err)
		}
	}

	// Get all lifetimes we currently know about and find which files are new
	// and therefore need to be created or copied from the core
	activeLts, err := history.GetActiveLifetimesByTowerId(t.Ctx, meta.Core.TowerId)
	if err != nil {
		t.Fail(err)
	}

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	t.SetChildTaskPool(pool)

	// Sort lifetimes so that files created or moved most recently are updated last.
	// This is to make sure parent directories are created before their children
	slices.SortFunc(activeLts, history.LifetimeSorter)

	// Check if the file already exists on the server and copy it if it doesn't
	for _, lt := range activeLts {
		latestMove := lt.GetLatestMove()

		existingFile, err := fileService.GetFileByTree(lt.ID(), meta.Core.TowerId)
		t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("File %s exists: %v", lt.GetLatestPath(), existingFile != nil) })

		// If the file already exists, but is the wrong size, an earlier copy most likely failed. Delete it and copy it again.
		if existingFile != nil && !existingFile.IsDir() && existingFile.Size() != lt.Actions[0].Size {
			err = fileService.DeleteFiles(t.Ctx, []*file_model.WeblensFileImpl{existingFile}, meta.Core.TowerId)
			t.ReqNoErr(err)
			existingFile = nil
		}

		if err == nil && existingFile != nil {
			if latestMove.ActionType == history.FileDelete {
				err = fileService.DeleteFiles(t.Ctx, []*file_model.WeblensFileImpl{existingFile}, meta.Core.TowerId)
				t.ReqNoErr(err)
			} else if latestMove.DestinationPath != existingFile.GetPortablePath().OverwriteRoot("USERS").ToPortable() {
				newParent, err := fileService.GetFileByTree(latestMove.ParentId, meta.Core.TowerId)
				if err != nil {
					t.ReqNoErr(err)
				}

				err = fileService.MoveFiles([]*file_model.WeblensFileImpl{existingFile}, newParent, meta.Core.TowerId)
				t.ReqNoErr(err)
			}
			continue
		} else if err != nil && !errors.Is(err, file_model.ErrFileNotFound) {
			t.Fail(err)
		}

		if lt.GetLatestAction().ActionType == history.FileDelete {
			if lt.GetIsDir() {
				continue
			}

			f, err := fileService.GetFileByContentId(lt.ContentId)
			if err == nil && f.Size() == lt.Actions[0].Size {
				continue
			} else if !errors.Is(err, file_model.ErrFileNotFound) {
				t.Fail(err)
			}
		}

		filename := lt.GetLatestPath().Filename()

		restoreFile, err := fileService.NewBackupFile(lt)
		t.ReqNoErr(err)

		if restoreFile == nil || restoreFile.Size() != 0 {
			continue
		}

		t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Queuing copy file task for %s", restoreFile.GetPortablePath()) })

		// Spawn subtask to copy the file from the core server
		copyFileMeta := job.BackupCoreFileMeta{
			File:       restoreFile,
			Core:       meta.Core,
			CoreFileId: lt.ID(),
			Filename:   filename,
		}

		_, err = t.Ctx.DispatchJob(job.CopyFileFromCoreTask, copyFileMeta)
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

	stages.FinishStage("sync_fs")
	t.SetResult(task_model.TaskResult{"stages": stages})

	coreTree, err := fileService.GetFileTreeByName(meta.Core.TowerId)
	if err != nil {
		t.Fail(err)
	}

	root := coreTree.GetRoot()

	err = fileService.ResizeDown(root, nil)
	t.ReqNoErr(err)

	tower_model.SetLastBackup(t.Ctx, meta.Core.TowerId, time.Now())
	t.ReqNoErr(err)

	// Don't broadcast this last event set
	t.OnResult(nil)
	t.SetResult(task_model.TaskResult{"backupSize": root.Size(), "totalTime": t.ExeTime()})

	endNotif := client.NewTaskNotification(t, websocket_mod.BackupCompleteEvent, t.GetResults())
	t.Ctx.Notify(endNotif)

	t.Success()
}
