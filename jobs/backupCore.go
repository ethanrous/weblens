package jobs

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	proxy2 "github.com/ethrousseau/weblens/service/proxy"
	"github.com/ethrousseau/weblens/task"
)

func BackupD(
	interval time.Duration, instanceService models.InstanceService, taskService task.TaskService,
	fileService models.FileService, userService models.UserService, websocketService models.ClientManager,
	caster models.Broadcaster,
) {
	if instanceService.GetLocal().ServerRole() != models.BackupServer {
		log.Error.Println("Backup service cannot be run on non-backup instance")
		return
	}
	for {
		for _, remote := range instanceService.GetRemotes() {
			if remote.IsLocal() {
				continue
			}

			// wsConn := websocketService.GetClientByInstanceId(remote.ServerId())
			// if wsConn == nil {
			// 	time.Sleep(time.Millisecond * 100)
			// }

			meta := models.BackupMeta{
				RemoteId:            remote.ServerId(),
				FileService:         fileService,
				UserService:         userService,
				WebsocketService:    websocketService,
				InstanceService:     instanceService,
				TaskService:         taskService,
				Caster:              caster,
				ProxyFileService:    &proxy2.ProxyFileService{Core: remote},
				ProxyJournalService: &proxy2.ProxyJournalService{Core: remote},
				ProxyUserService:    &proxy2.ProxyUserService{Core: remote},
				ProxyMediaService:   &proxy2.ProxyMediaService{Core: remote},
			}
			_, err := taskService.DispatchJob(models.BackupTask, meta, nil)
			if err != nil {
				log.ErrTrace(err)
			}
		}

		now := time.Now()
		sleepFor := now.Truncate(interval).Add(interval).Sub(now)
		log.Debug.Println("BackupD going to sleep for", sleepFor)
		time.Sleep(sleepFor)
	}
}

func DoBackup(t *task.Task) {
	meta := t.GetMeta().(models.BackupMeta)
	localRole := meta.InstanceService.GetLocal().ServerRole()

	if localRole == models.InitServer {
		t.ErrorAndExit(werror.ErrServerNotInitialized)
	} else if localRole != models.BackupServer {
		t.ErrorAndExit(werror.ErrServerIsBackup)
	}

	coreClient := meta.WebsocketService.GetClientByInstanceId(meta.RemoteId)
	if coreClient == nil {
		t.ErrorAndExit(werror.Errorf("Core websocket not connected"))
	}

	users, err := meta.ProxyUserService.GetAll()
	if err != nil {
		t.ErrorAndExit(err)
	}
	for user := range users {
		err = meta.UserService.Add(user)
		if err != nil {
			t.ErrorAndExit(err)
		}
	}

	latest, err := meta.FileService.GetMediaJournal().GetLatestAction()
	if err != nil {
		t.ErrorAndExit(err)
	}

	// Get new history updates
	updatedLifetimes, err := meta.ProxyJournalService.GetLifetimesSince(latest.GetTimestamp())
	if err != nil {
		t.ErrorAndExit(err)
	}

	slices.SortFunc(
		updatedLifetimes, func(a, b *fileTree.Lifetime) int {
			aActions := a.GetActions()
			bActions := b.GetActions()
			return len(aActions[len(aActions)-1].GetDestinationPath()) - len(bActions[len(bActions)-1].GetDestinationPath())
		},
	)

	var newFileIds []fileTree.FileId
	if len(updatedLifetimes) > 0 {
		for _, lt := range updatedLifetimes {
			existLt := meta.FileService.GetMediaJournal().Get(lt.ID())
			existFile, err := meta.FileService.GetFile(lt.ID())
			if err != nil && !errors.Is(err, werror.ErrNoFile) {
				t.ErrorAndExit(err)
			}

			if existLt == nil && existFile == nil {
				newFileIds = append(newFileIds, lt.ID())
				// _, err = proxyService.GetFile(lt.GetLatestFileId())
				// if err != nil {
				// 	t.ErrorAndExit(err)
				// }
			} else {
				// log.Debug.Println("Uhh... should this even happen?")
			}

			err = meta.FileService.GetMediaJournal().Add(lt)
			if err != nil {
				t.ErrorAndExit(err)
			}
		}
	}

	if len(newFileIds) == 0 {
		t.Success()
		return
	}

	newFiles, err := meta.ProxyFileService.GetFiles(newFileIds)

	// files := internal.FilterMap(
	// 	SERV.FileTree.GetJournal().GetActiveLifetimes(), func(lt Lifetime) (*fileTree.WeblensFileImpl, bool) {
	// 		f := SERV.FileTree.Get(lt.GetLatestFileId())
	// 		if f == nil && lt.GetLatestAction().GetActionType() != FileDelete {
	// 			f, err = proxyService.GetFile(lt.GetLatestFileId())
	// 			if err != nil {
	// 				wlog.ShowErr(err)
	// 				wlog.Debug.Println("Failed to get file at", lt.GetLatestAction().GetDestinationPath())
	// 				return nil, false
	// 			}
	// 			err = SERV.FileTree.Add(f)
	// 			if err != nil {
	// 				t.ErrorAndExit(err)
	// 			}
	// 		}
	//
	// 		return f, true
	// 	},
	// )

	slices.SortFunc(
		newFiles, func(a, b *fileTree.WeblensFileImpl) int {
			return len(a.GetAbsPath()) - len(b.GetAbsPath())
		},
	)

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	t.SetChildTaskPool(pool)
	meta.WebsocketService.TaskSubToPool(t.TaskId(), pool.GetRootPool().ID())

	for _, f := range newFiles {
		_, err := meta.FileService.GetFile(f.ID())
		if err == nil {
			log.Debug.Println("Already have file??")
			continue
		}

		err = meta.FileService.ImportFile(f)
		if err != nil {
			t.ErrorAndExit(err)
		}

		if f.IsDir() {
			err = f.CreateSelf()
			if err != nil && !errors.Is(err, werror.ErrFileAlreadyExists) {
				t.ErrorAndExit(err)
			}
			continue
		}

		if !coreClient.IsOpen() {
			coreClient = meta.WebsocketService.GetClientByInstanceId(meta.RemoteId)
		}

		copyFileMeta := models.BackupCoreFileMeta{
			ProxyFileService: meta.ProxyFileService,
			File:             f,
			Caster:           meta.Caster,
		}
		_, err = meta.TaskService.DispatchJob(models.CopyFileFromCoreTask, copyFileMeta, pool)
		if err != nil {
			t.ErrorAndExit(err)
		}
	}

	pool.SignalAllQueued()
	pool.Wait(true)

	if len(pool.Errors()) != 0 {
		t.ErrorAndExit(errors.New(fmt.Sprintf("%d backup file copies have failed", len(pool.Errors()))))
	}

	t.Success()
}

func CopyFileFromCore(t *task.Task) {
	meta := t.GetMeta().(models.BackupCoreFileMeta)

	writeFile, err := meta.File.Writeable()
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer writeFile.Close()

	fileReader, err := meta.ProxyFileService.ReadFile(meta.File)
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer fileReader.Close()

	_, err = io.Copy(writeFile, fileReader)
	if err != nil {
		t.ErrorAndExit(err)
	}

	poolProgress := getScanResult(t)
	poolProgress["filename"] = meta.File.Filename()
	meta.Caster.PushPoolUpdate(t.GetTaskPool(), models.SubTaskCompleteEvent, poolProgress)
	// if meta..IsOpen() {
	// 	meta.core.PushPoolUpdate(t.GetTaskPool(), models.SubTaskCompleteEvent, poolProgress)
	// }

	t.Success()
}
