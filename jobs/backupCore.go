package jobs

import (
	"errors"
	"io"
	"slices"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/service/proxy"
	"github.com/ethanrous/weblens/task"
)

func BackupD(interval time.Duration, pack *models.ServicePack) {
	if pack.InstanceService.GetLocal().GetRole() != models.BackupServerRole {
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
		Core:             core,
		FileService:      pack.FileService,
		UserService:      pack.UserService,
		WebsocketService: pack.ClientService,
		InstanceService:  pack.InstanceService,
		TaskService:      pack.TaskService,
		Caster:           pack.Caster,
		AccessService:    pack.AccessService,
	}
	return pack.TaskService.DispatchJob(models.BackupTask, meta, nil)
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
				errTsk, models.BackupFailedEvent,
				task.TaskResult{"coreId": meta.Core.ServerId(), "error": err.Error()},
			)
		},
	)

	localRole := meta.InstanceService.GetLocal().GetRole()
	if localRole == models.InitServerRole {
		t.ReqNoErr(werror.ErrServerNotInitialized)
	} else if localRole != models.BackupServerRole {
		t.ReqNoErr(werror.ErrServerIsBackup)
	}

	// Get the active websocket client for the core
	// coreClient := meta.WebsocketService.GetClientByServerId(meta.Core.ServerId())
	// if coreClient == nil {
	// 	t.Fail(werror.Errorf("Core websocket not connected"))
	// }

	// log.Debug.Printf("Starting backup of [%s]", coreClient.GetRemote().GetName())

	stages := models.NewBackupTaskStages()

	stages.StartStage("connecting")
	t.SetResult(task.TaskResult{"stages": stages, "coreId": meta.Core.ServerId()})

	// Read core server info and check if it is really a core server
	req := proxy.NewCoreRequest(meta.Core, "GET", "/info")
	infoRes, err := proxy.CallHomeStruct[rest.ServerInfo](req)
	if err != nil {
		t.ReqNoErr(err)
	}

	meta.Core.SetReportedRole(infoRes.Role)
	if infoRes.Role != models.CoreServerRole {
		t.ReqNoErr(werror.Errorf("Remote role is [%s] expected core", infoRes.Role))
	}

	// Find most recent action timestamp
	latest, err := meta.FileService.GetJournalByTree(meta.Core.Id).GetLatestAction()
	if err != nil {
		t.ReqNoErr(err)
	}

	var latestTime = time.UnixMilli(0)
	if latest != nil {
		latestTime = latest.GetTimestamp()
	}

	log.Trace.Func(func(l log.Logger) { l.Printf("Backup latest action is %s", latestTime.String()) })

	stages.StartStage("fetching_backup_data")
	t.SetResult(task.TaskResult{"stages": stages})

	backupRq := proxy.NewCoreRequest(meta.Core, "GET", "/servers/backup").WithQuery("timestamp", strconv.FormatInt(latestTime.UnixMilli(), 10))
	backupResponse, err := proxy.CallHomeStruct[rest.BackupInfo](backupRq)
	t.ReqNoErr(err)

	stages.StartStage("writing_users")
	t.SetResult(task.TaskResult{"stages": stages})

	// Write the users to the users service
	for _, userInfo := range backupResponse.Users {
		user := rest.UserInfoArchiveToUser(userInfo)
		err = meta.UserService.Add(user)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	stages.StartStage("writing_keys")
	t.SetResult(task.TaskResult{"stages": stages})

	// Write keys to access service
	for _, key := range backupResponse.ApiKeys {
		if _, err := meta.AccessService.GetApiKey(key.Key); err == nil {
			continue
		}

		err = meta.AccessService.AddApiKey(key)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	stages.StartStage("writing_instances")
	t.SetResult(task.TaskResult{"stages": stages})

	// Write instances to access service
	for _, serverInfo := range backupResponse.Instances {
		instance := rest.ServerInfoToInstance(serverInfo)
		if err := meta.InstanceService.Get(instance.ServerId()); err == nil {
			continue
		}

		err = meta.InstanceService.Add(instance)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	stages.StartStage("sync_journal")
	t.SetResult(task.TaskResult{"stages": stages})

	log.Trace.Func(func(l log.Logger) {
		l.Printf("Backup got %d updated lifetimes from core", len(backupResponse.FileHistory))
	})

	stages.StartStage("sync_fs")
	t.SetResult(task.TaskResult{"stages": stages})

	// Sort lifetimes so that files created or moved most recently are updated last.
	slices.SortFunc(backupResponse.FileHistory, fileTree.LifetimeSorter)

	journal := meta.FileService.GetJournalByTree(meta.Core.Id)
	if len(backupResponse.FileHistory) > 0 {
		for _, lt := range backupResponse.FileHistory {
			// Add the lifetime to the journal, or update it if it already exists
			err = journal.Add(lt)
			t.ReqNoErr(err)
		}
	}

	if len(journal.GetAllLifetimes()) < backupResponse.LifetimesCount {
		log.Debug.Func(func(l log.Logger) {
			l.Printf("Backup journal is missing %d lifetimes", backupResponse.LifetimesCount-len(journal.GetAllLifetimes()))
		})

		req := proxy.NewCoreRequest(meta.Core, "GET", "/journal").WithQuery("timestamp", "0")
		allLts, err := proxy.CallHomeStruct[[]*fileTree.Lifetime](req)
		t.ReqNoErr(err)
		for _, lt := range allLts {
			if journal.Get(lt.ID()) == nil {
				err = journal.Add(lt)
				t.ReqNoErr(err)
			}
		}
	}

	// Get all lifetimes we currently know about and find which files are new
	// and therefore need to be created or copied from the core
	activeLts := meta.FileService.GetJournalByTree(meta.Core.Id).GetActiveLifetimes()

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	t.SetChildTaskPool(pool)

	// Sort lifetimes so that files created or moved most recently are updated last.
	// This is to make sure parent directories are created before their children
	slices.SortFunc(activeLts, fileTree.LifetimeSorter)

	// Check if the file already exists on the server and copy it if it doesn't
	for _, lt := range activeLts {
		latestMove := lt.GetLatestMove()

		existingFile, err := meta.FileService.GetFileByTree(lt.ID(), meta.Core.ServerId())
		log.Trace.Printf("File %s exists: %v", lt.GetLatestPath(), existingFile != nil)

		// If the file already exists, but is the wrong size, an earlier copy most likely failed. Delete it and copy it again.
		if existingFile != nil && !existingFile.IsDir() && existingFile.Size() != lt.Actions[0].Size {
			err = meta.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{existingFile}, meta.Core.ServerId(), meta.Caster)
			t.ReqNoErr(err)
			existingFile = nil
		}

		if err == nil && existingFile != nil {
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

		if restoreFile == nil || restoreFile.Size() != 0 {
			continue
		}

		log.Trace.Printf("Queuing copy file task for %s", restoreFile.GetPortablePath())

		// Spawn subtask to copy the file from the core server
		copyFileMeta := models.BackupCoreFileMeta{
			FileService: meta.FileService,
			File:        restoreFile,
			Caster:      meta.Caster,
			Core:        meta.Core,
			CoreFileId:  lt.ID(),
			Filename:    filename,
		}

		_, err = meta.TaskService.DispatchJob(models.CopyFileFromCoreTask, copyFileMeta, pool)
		t.ReqNoErr(err)
	}

	log.Debug.Printf("Waiting for %d copy file tasks", pool.Status().Total)

	pool.SignalAllQueued()
	pool.Wait(true)

	if len(pool.Errors()) != 0 {
		t.Fail(werror.Errorf("%d of %d backup file copies have failed", len(pool.Errors()), pool.Status().Total))
	}

	stages.FinishStage("sync_fs")
	t.SetResult(task.TaskResult{"stages": stages})

	root, err := meta.FileService.GetFileByTree("ROOT", meta.Core.ServerId())
	t.ReqNoErr(err)

	err = meta.FileService.ResizeDown(root, nil, meta.Caster)
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
		meta.Caster.PushTaskUpdate(t, models.CopyFileFailedEvent, task.TaskResult{"filename": meta.Filename, "coreId": meta.Core.ServerId()})
		rmErr := meta.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{meta.File}, meta.Core.ServerId(), meta.Caster)
		if rmErr != nil {
			log.ErrTrace(rmErr)
		}
	})

	filename := meta.Filename
	if filename == "" {
		filename = meta.File.Filename()
	}

	meta.Caster.PushPoolUpdate(
		t.GetTaskPool(), models.CopyFileStartedEvent, task.TaskResult{
			"filename": filename, "coreId": meta.Core.ServerId(), "timestamp": time.Now().UnixMilli(),
		},
	)

	log.Trace.Func(func(l log.Logger) { l.Printf("Copying file from core [%s]", meta.File.Filename()) })

	if meta.File.GetContentId() == "" {
		t.ReqNoErr(werror.WithStack(werror.ErrNoContentId))
	}

	writeFile, err := meta.File.Writeable()
	if err != nil {
		t.ReqNoErr(err)
	}
	defer writeFile.Close()

	res, err := proxy.NewCoreRequest(meta.Core, "GET", "/files/"+meta.CoreFileId+"/download").Call()
	t.ReqNoErr(err)

	defer res.Body.Close()

	_, err = io.Copy(writeFile, res.Body)
	t.ReqNoErr(err)

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
		RemoteId string            `json:"remoteId"`
		LocalId  string            `json:"localId"`
		Key      models.ApiKey     `json:"usingKeyInfo"`
	}

	// Notify client of restore failure, if any
	t.SetErrorCleanup(
		func(errTsk *task.Task) {
			meta.Pack.Caster.PushTaskUpdate(
				errTsk, models.RestoreFailedEvent, task.TaskResult{"error": errTsk.ReadError().Error()},
			)
		},
	)

	// Prime server to be restored. This will fail if the server is already initialized
	meta.Pack.Caster.PushTaskUpdate(
		t, models.RestoreProgressEvent, task.TaskResult{"stage": "Connecting to remote", "timestamp": time.Now().UnixMilli()},
	)

	key, err := meta.Pack.AccessService.GetApiKey(meta.Core.UsingKey)
	t.ReqNoErr(err)

	initParams := restoreInitParams{
		Name: meta.Core.Name, Role: models.RestoreServerRole, Key: key, RemoteId: meta.Local.Id,
		LocalId: meta.Core.Id,
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/servers/init").WithBody(initParams).Call()
	if err != nil {
		t.ReqNoErr(err)
	}

	// Restore journal
	meta.Pack.Caster.PushTaskUpdate(
		t, models.RestoreProgressEvent, task.TaskResult{"stage": "Restoring file history", "timestamp": time.Now().UnixMilli()},
	)
	lts := meta.Pack.FileService.GetJournalByTree(meta.Core.Id).GetAllLifetimes()
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/history").WithBody(lts).Call()
	t.ReqNoErr(err)

	// Restore users
	meta.Pack.Caster.PushTaskUpdate(
		t, models.RestoreProgressEvent, task.TaskResult{"stage": "Restoring users", "timestamp": time.Now().UnixMilli()},
	)

	usersIter, err := meta.Pack.UserService.GetAll()
	t.ReqNoErr(err)

	var users []rest.UserInfo
	for u := range usersIter {
		users = append(users, rest.UserToUserInfo(u))
	}
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/users").WithBody(users).Call()
	t.ReqNoErr(err)

	// Restore keys
	meta.Pack.Caster.PushTaskUpdate(
		t, models.RestoreProgressEvent, task.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()},
	)
	rootUser := meta.Pack.UserService.GetRootUser()
	keys, err := meta.Pack.AccessService.GetAllKeysByServer(rootUser, meta.Core.ServerId())
	t.ReqNoErr(err)

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/keys").WithBody(keys).Call()
	t.ReqNoErr(err)

	// Restore instances
	meta.Pack.Caster.PushTaskUpdate(
		t, models.RestoreProgressEvent, task.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()},
	)

	instances := meta.Pack.InstanceService.GetAllByOriginServer(meta.Core.ServerId())
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/instances").WithBody(instances).Call()
	t.ReqNoErr(err)

	// Restore files
	meta.Pack.Caster.PushTaskUpdate(
		t, models.RestoreProgressEvent, task.TaskResult{"stage": "Restoring files", "timestamp": time.Now().UnixMilli()},
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
			t, models.RestoreProgressEvent, task.TaskResult{"filesTotal": len(lts), "filesRestored": i},
		)
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/complete").Call()
	if err != nil {
		t.ReqNoErr(err)
	}

	meta.Core.SetReportedRole(models.CoreServerRole)
	meta.Pack.Caster.PushTaskUpdate(t, models.RestoreCompleteEvent, nil)

	// Disconnect the core client to force a reconnection
	coreClient := meta.Pack.ClientService.GetClientByServerId(meta.Core.ServerId())
	meta.Pack.ClientService.ClientDisconnect(coreClient)

	t.Success()
}
