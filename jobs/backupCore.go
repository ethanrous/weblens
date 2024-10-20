package jobs

import (
	"errors"
	"io"
	"slices"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service/proxy"
	"github.com/ethanrous/weblens/task"
)

func BackupD(interval time.Duration, pack *models.ServicePack) {
	if pack.InstanceService.GetLocal().GetRole() != models.BackupServer {
		log.Error.Println("Backup service cannot be run on non-backup instance")
		return
	}
	for {
		now := time.Now()
		sleepFor := now.Truncate(interval).Add(interval).Sub(now)
		log.Debug.Println("BackupD going to sleep for", sleepFor)
		time.Sleep(sleepFor)

		for _, remote := range pack.InstanceService.GetRemotes() {
			if remote.IsLocal() {
				continue
			}

			_, err := BackupOne(remote, pack)
			if err != nil {
				log.ErrTrace(err)
			}
		}
	}
}

func BackupOne(core *models.Instance, pack *models.ServicePack) (*task.Task, error) {
	meta := models.BackupMeta{
		Core:                core,
		FileService:         pack.FileService,
		UserService:         pack.UserService,
		WebsocketService:    pack.ClientService,
		InstanceService:     pack.InstanceService,
		TaskService:         pack.TaskService,
		Caster:              pack.Caster,
		AccessService:       pack.AccessService,
		ProxyFileService:    &proxy.ProxyFileService{Core: core},
		ProxyJournalService: &proxy.ProxyJournalService{Core: core},
		ProxyUserService:    proxy.NewProxyUserService(core),
		ProxyMediaService:   &proxy.ProxyMediaService{Core: core},
	}
	return pack.TaskService.DispatchJob(models.BackupTask, meta, nil)
}

type serverInfoResponse struct {
	Info models.ServerInfo `json:"info"`
}

func DoBackup(t *task.Task) {
	meta := t.GetMeta().(models.BackupMeta)

	t.OnResult(
		func(r task.TaskResult) {
			meta.Caster.PushTaskUpdate(t, models.BackupProgressEvent, r)
		},
	)

	t.SetErrorCleanup(
		func(errTsk *task.Task) {
			err := errTsk.ReadError()
			if err == nil {
				log.Error.Println("Trying to show error in backup task, but error is nil")
				return
			}

			meta.Caster.PushTaskUpdate(
				errTsk, "backup_failed",
				task.TaskResult{"coreId": meta.Core.ServerId(), "error": err.Error()},
			)
		},
	)

	localRole := meta.InstanceService.GetLocal().GetRole()
	if localRole == models.InitServer {
		t.ReqNoErr(werror.ErrServerNotInitialized)
	} else if localRole != models.BackupServer {
		t.ReqNoErr(werror.ErrServerIsBackup)
	}

	coreClient := meta.WebsocketService.GetClientByServerId(meta.Core.ServerId())
	if coreClient == nil {
		t.ReqNoErr(werror.Errorf("Core websocket not connected"))
		// Dead code
		return
	}

	log.Debug.Printf("Starting backup of [%s]", coreClient.GetRemote().GetName())

	stages := models.NewBackupTaskStages()

	stages.StartStage("connecting")
	t.SetResult(task.TaskResult{"stages": stages, "coreId": meta.Core.ServerId()})

	// Read core server info and check if it is really a core server
	req := proxy.NewCoreRequest(meta.Core, "GET", "").OverwriteEndpoint("/api/info")
	infoRes, err := proxy.CallHomeStruct[serverInfoResponse](req)
	if err != nil {
		t.ReqNoErr(err)
	}

	meta.Core.SetReportedRole(infoRes.Info.Role)
	if infoRes.Info.Role != models.CoreServer {
		t.ReqNoErr(werror.Errorf("Remote role is [%s] expected core", infoRes.Info.Role))
	}

	stages.StartStage("fetching_users")
	t.SetResult(task.TaskResult{"stages": stages})
	// Backup users
	users, err := meta.ProxyUserService.GetAll()
	if err != nil {
		t.ReqNoErr(err)
	}

	stages.StartStage("writing_users")
	t.SetResult(task.TaskResult{"stages": stages})
	for user := range users {
		err = meta.UserService.Add(user)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	stages.StartStage("fetching_keys")
	t.SetResult(task.TaskResult{"stages": stages})
	// Fetch keys from core
	keysRq := proxy.NewCoreRequest(meta.Core, "GET", "/backup/keys")
	keys, err := proxy.CallHomeStruct[[]models.ApiKeyInfo](keysRq)
	if err != nil {
		t.ReqNoErr(err)
	}

	stages.StartStage("writing_keys")
	t.SetResult(task.TaskResult{"stages": stages})
	// Add keys to access service
	for _, key := range keys {
		if _, err := meta.AccessService.GetApiKey(key.Key); err == nil {
			continue
		}

		err = meta.AccessService.AddApiKey(key)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	stages.StartStage("fetching_instances")
	t.SetResult(task.TaskResult{"stages": stages})
	// Fetch instances from core
	instancesRq := proxy.NewCoreRequest(meta.Core, "GET", "/backup/instances")
	instances, err := proxy.CallHomeStruct[[]*models.Instance](instancesRq)
	if err != nil {
		t.ReqNoErr(err)
	}

	stages.StartStage("writing_instances")
	t.SetResult(task.TaskResult{"stages": stages})
	// Add instances to access service
	for _, r := range instances {
		if err := meta.InstanceService.Get(r.ServerId()); err == nil {
			continue
		}

		err = meta.InstanceService.Add(r)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	stages.StartStage("sync_journal")
	t.SetResult(task.TaskResult{"stages": stages})
	// Find most recent action timestamp
	latest, err := meta.FileService.GetJournalByTree(meta.Core.Id).GetLatestAction()
	if err != nil {
		t.ReqNoErr(err)
	}

	var latestTime = time.UnixMilli(0)
	if latest != nil {
		latestTime = latest.GetTimestamp()
	}

	log.Trace.Printf("Backup latest action is %s", latestTime.String())

	// Get new history updates
	updatedLifetimes, err := meta.ProxyJournalService.GetLifetimesSince(latestTime)
	t.ReqNoErr(err)

	log.Trace.Printf("Backup got %d updated lifetimes from core", len(updatedLifetimes))

	stages.StartStage("sync_fs")
	t.SetResult(task.TaskResult{"stages": stages})

	// Sort lifetimes so that files created or moved most recently are updated last.
	slices.SortFunc(updatedLifetimes, fileTree.LifetimeSorter)

	journal := meta.FileService.GetJournalByTree(meta.Core.Id)
	if len(updatedLifetimes) > 0 {
		for _, lt := range updatedLifetimes {
			// Add the lifetime to the journal, or update it if it already exists
			err = journal.Add(lt)
			t.ReqNoErr(err)
		}
	}

	// Get all lifetimes we currently know about and find which files are new
	// and therefore need to be created or copied from the core
	activeLts := meta.FileService.GetJournalByTree(meta.Core.Id).GetActiveLifetimes()

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	t.SetChildTaskPool(pool)
	// meta.WebsocketService.TaskSubToPool(t.TaskId(), pool.GetRootPool().ID())

	// Sort lifetimes so that files created or moved most recently are updated last.
	// This is to make sure parent directories are created before their children
	slices.SortFunc(activeLts, fileTree.LifetimeSorter)

	for _, lt := range activeLts {
		latestMove := lt.GetLatestMove()

		existingFile, err := meta.FileService.GetFileByTree(lt.ID(), meta.Core.ServerId())
		if err == nil && existingFile.Size() == lt.Actions[0].Size {
			if latestMove.ActionType == fileTree.FileDelete {
				err = meta.FileService.DeleteFiles(
					[]*fileTree.WeblensFileImpl{existingFile}, meta.Core.ServerId(), meta.Caster,
				)
				t.ReqNoErr(err)
			} else if latestMove.DestinationPath != existingFile.GetPortablePath().OverwriteRoot("USERS").ToPortable() {
				newParent, err := meta.FileService.GetFileByTree(latestMove.ParentId, meta.Core.ServerId())
				if err != nil {
					t.ReqNoErr(err)
				}

				err = meta.FileService.MoveFiles(
					[]*fileTree.WeblensFileImpl{existingFile}, newParent, meta.Core.ServerId(), meta.Caster,
				)
				t.ReqNoErr(err)
			}
			continue

		} else if err != nil && !errors.Is(err, werror.ErrNoFile) {
			t.Fail(err)
		} else if !existingFile.IsDir() && existingFile.Size() != lt.Actions[0].Size {
			err = meta.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{existingFile}, meta.Core.ServerId(), meta.Caster)
			t.ReqNoErr(err)
		}

		if lt.GetLatestAction().ActionType == fileTree.FileDelete {
			if lt.GetIsDir() {
				continue
			}

			f, err := meta.FileService.GetFileByContentId(lt.ContentId)
			if err == nil && f.Size() == lt.Actions[0].Size {
				continue
			} else if !errors.Is(err, werror.ErrNoFile) {
				t.Fail(err)
			}
		}

		filename := lt.GetLatestPath().Filename()

		restoreFile, err := meta.FileService.NewBackupFile(lt)
		t.ReqNoErr(err)

		if restoreFile == nil {
			continue
		}

		if !coreClient.IsOpen() {
			coreClient = meta.WebsocketService.GetClientByServerId(meta.Core.ServerId())
		}

		copyFileMeta := models.BackupCoreFileMeta{
			ProxyFileService: meta.ProxyFileService,
			FileService:      meta.FileService,
			File:             restoreFile,
			Caster:           meta.Caster,
			Core:             meta.Core,
			Filename:         filename,
		}

		_, err = meta.TaskService.DispatchJob(models.CopyFileFromCoreTask, copyFileMeta, pool)
		t.ReqNoErr(err)
	}

	pool.SignalAllQueued()
	pool.Wait(true)

	if len(pool.Errors()) != 0 {
		t.ReqNoErr(werror.Errorf("%d backup file copies have failed", len(pool.Errors())))
	}

	stages.FinishStage("sync_fs")
	t.SetResult(task.TaskResult{"stages": stages})

	root, err := meta.FileService.GetFileByTree("ROOT", meta.Core.ServerId())
	t.ReqNoErr(err)

	err = meta.FileService.ResizeDown(root, meta.Caster)
	t.ReqNoErr(err)

	err = meta.InstanceService.SetLastBackup(meta.Core.ServerId(), time.Now())
	t.ReqNoErr(err)

	// Don't broadcast this last event set
	t.OnResult(nil)
	t.SetResult(task.TaskResult{"backupSize": root.Size(), "totalTime": t.ExeTime()})

	meta.Caster.PushTaskUpdate(t, models.BackupCompleteEvent, t.GetResults())
	t.Success()
}

func CopyFileFromCore(t *task.Task) {
	meta := t.GetMeta().(models.BackupCoreFileMeta)
	t.SetErrorCleanup(func(t *task.Task) {
		meta.Caster.PushTaskUpdate(t, "copy_file_failed", task.TaskResult{"filename": meta.Filename, "coreId": meta.Core.ServerId()})
	})

	filename := meta.Filename
	if filename == "" {
		filename = meta.File.Filename()
	}

	meta.Caster.PushPoolUpdate(
		t.GetTaskPool(), "copy_file_started", task.TaskResult{
			"filename": filename, "coreId": meta.Core.ServerId(), "timestamp": time.Now().UnixMilli(),
		},
	)
	log.Trace.Printf("Copying file from core [%s]", meta.File.Filename())

	if meta.File.GetContentId() == "" {
		t.ReqNoErr(werror.WithStack(werror.ErrNoContentId))
	}

	writeFile, err := meta.File.Writeable()
	if err != nil {
		t.ReqNoErr(err)
	}
	defer writeFile.Close()

	fileReader, err := meta.ProxyFileService.ReadFile(meta.File)
	if err != nil {
		t.ReqNoErr(err)
	}
	defer fileReader.Close()

	_, err = io.Copy(writeFile, fileReader)
	if err != nil {
		rmErr := meta.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{meta.File}, meta.Core.ServerId(), meta.Caster)
		if rmErr != nil {
			t.ReqNoErr(
				werror.Errorf(
					"Failed to write to file: %s\nThis Occoured while cleaning up from another error: %s",
					rmErr, err,
				),
			)
			t.ReqNoErr(rmErr)
		}
		t.ReqNoErr(err)
	}

	poolProgress := getScanResult(t)
	poolProgress["filename"] = filename
	poolProgress["coreId"] = meta.Core.ServerId()
	meta.Caster.PushPoolUpdate(t.GetTaskPool(), models.CopyFileCompleteEvent, poolProgress)

	t.Success()
}

func RestoreCore(t *task.Task) {
	meta := t.GetMeta().(models.RestoreCoreMeta)

	type restoreInitParams struct {
		Name     string            `json:"name"`
		Role     models.ServerRole `json:"role"`
		Key      models.ApiKeyInfo `json:"usingKeyInfo"`
		RemoteId string            `json:"remoteId"`
		LocalId  string            `json:"localId"`
	}

	// Notify client of restore failure, if any
	t.SetErrorCleanup(
		func(errTsk *task.Task) {
			meta.Pack.Caster.PushTaskUpdate(
				errTsk, "restore_failed", task.TaskResult{"error": errTsk.ReadError().Error()},
			)
		},
	)

	// Prime server to be restored. This will fail if the server is already initialized
	meta.Pack.Caster.PushTaskUpdate(
		t, "restore_progress", task.TaskResult{"stage": "Connecting to remote", "timestamp": time.Now().UnixMilli()},
	)

	key, err := meta.Pack.AccessService.GetApiKey(meta.Core.UsingKey)
	t.ReqNoErr(err)

	initParams := restoreInitParams{
		Name: meta.Core.Name, Role: models.RestoreServer, Key: key, RemoteId: meta.Local.Id,
		LocalId: meta.Core.Id,
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "").OverwriteEndpoint("/api/init").WithBody(initParams).Call()
	if err != nil {
		t.ReqNoErr(err)
	}

	// Restore journal
	meta.Pack.Caster.PushTaskUpdate(
		t, "restore_progress", task.TaskResult{"stage": "Restoring file history", "timestamp": time.Now().UnixMilli()},
	)
	lts := meta.Pack.FileService.GetJournalByTree(meta.Core.Id).GetAllLifetimes()
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/history").WithBody(lts).Call()
	t.ReqNoErr(err)

	// Restore users
	meta.Pack.Caster.PushTaskUpdate(
		t, "restore_progress", task.TaskResult{"stage": "Restoring users", "timestamp": time.Now().UnixMilli()},
	)

	usersIter, err := meta.Pack.UserService.GetAll()
	t.ReqNoErr(err)

	var users []map[string]any
	for u := range usersIter {
		formatted, err := u.FormatArchive()
		if err != nil {
			t.ReqNoErr(err)
		}
		users = append(users, formatted)
	}
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/users").WithBody(users).Call()
	t.ReqNoErr(err)

	// Restore keys
	meta.Pack.Caster.PushTaskUpdate(
		t, "restore_progress", task.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()},
	)
	rootUser := meta.Pack.UserService.GetRootUser()
	keys, err := meta.Pack.AccessService.GetAllKeysByServer(rootUser, meta.Core.ServerId())
	t.ReqNoErr(err)

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/keys").WithBody(keys).Call()
	t.ReqNoErr(err)

	// Restore instances
	meta.Pack.Caster.PushTaskUpdate(
		t, "restore_progress", task.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()},
	)

	instances := meta.Pack.InstanceService.GetAllByOriginServer(meta.Core.ServerId())
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/instances").WithBody(instances).Call()
	t.ReqNoErr(err)

	// Restore files
	meta.Pack.Caster.PushTaskUpdate(
		t, "restore_progress", task.TaskResult{"stage": "Restoring files", "timestamp": time.Now().UnixMilli()},
	)
	for i, lt := range lts {
		latest := lt.GetLatestAction()
		if latest.GetActionType() == fileTree.FileDelete {
			continue
		}
		portable := fileTree.ParsePortable(latest.GetDestinationPath())
		if portable.IsDir() {
			continue
		}

		f, err := meta.Pack.FileService.GetFileByContentId(lt.GetContentId())
		if err != nil {
			log.ErrTrace(err)
			continue
		}
		if f == nil {
			log.Error.Printf("File not found for contentId [%s]", lt.GetContentId())
			continue
		}

		bs, err := f.ReadAll()
		if err != nil {
			t.ReqNoErr(err)
		}
		_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/file").WithBodyBytes(bs).WithQuery(
			"fileId", lt.ID(),
		).Call()
		if err != nil {
			t.ReqNoErr(err)
		}

		meta.Pack.Caster.PushTaskUpdate(
			t, "restore_progress", task.TaskResult{"files_total": len(lts), "files_restored": i},
		)
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/complete").Call()
	if err != nil {
		t.ReqNoErr(err)
	}

	meta.Core.SetReportedRole(models.CoreServer)
	meta.Pack.Caster.PushTaskUpdate(t, "restore_complete", nil)

	// Disconnect the core client to force a reconnection
	coreClient := meta.Pack.ClientService.GetClientByServerId(meta.Core.ServerId())
	meta.Pack.ClientService.ClientDisconnect(coreClient)

	t.Success()
}
