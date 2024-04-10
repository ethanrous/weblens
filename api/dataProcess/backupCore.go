package dataProcess

import (
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
)

func BackupD(interval time.Duration, r types.Requester) {
	srvInfo := dataStore.GetServerInfo()
	if srvInfo == nil || srvInfo.ServerRole() != types.BackupMode {
		return
	}
	return
	for {
		time.Sleep(interval)
		newTask(BackupTask, nil, globalCaster, r).Q(nil)

	}
}

func doBackup(t *task) {
	srvInfo := dataStore.GetServerInfo()
	if srvInfo == nil {
		t.ErrorAndExit(types.ErrServerNotInit)
	}
	if srvInfo.ServerRole() == types.CoreMode {
		packageBackup(t)
	} else if srvInfo.ServerRole() == types.BackupMode {
		receiveBackup(t)
	}
}

// task run on backup server to query for and download updated data
func receiveBackup(t *task) {
	// tp := NewTaskPool(true, t)
	// fsStatT := tp.GatherFsStats(dataStore.GetMediaDir(), globalCaster).Q(tp)
	// fsStatT.Wait()

	err := t.requester.GetCoreSnapshot()
	if err != nil {
		t.ErrorAndExit(err)
		// util.ShowErr(err)
	}
	t.setResult(types.TaskResult{"yay": "cool"})

	// util.Debug.Println("Free Space", humanize.Bytes(fsStatT.GetResult("bytesFree")["bytesFree"].(uint64)))

	t.success()
}

// task run on core server to package backup info to send to backup server
func packageBackup(t *task) {
	// acc := dataStore.NewAccessMeta(dataStore.WEBLENS_ROOT_USER.Username).SetRequestMode(dataStore.BackupFileScan)

	allFiles := dataStore.GetAllFiles()

	t.setResult(types.TaskResult{"files": allFiles})

}
