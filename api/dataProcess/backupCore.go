package dataProcess

import (
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
)

func BackupD(interval time.Duration, r types.Requester) {
	srvInfo := dataStore.GetServerInfo()
	if srvInfo == nil || srvInfo.ServerRole() != types.Backup {
		return
	}
	return
	// for {
	// 	time.Sleep(interval)
	// 	newTask(BackupTask, nil, globalCaster, r).Q(nil)
	// }
}

func doBackup(t *task) {
	srvInfo := dataStore.GetServerInfo()
	if srvInfo == nil {
		t.ErrorAndExit(types.ErrServerNotInit)
	}
	if srvInfo.ServerRole() == types.Core {
		packageBackup(t)
	} else if srvInfo.ServerRole() == types.Backup {
		receiveBackup(t)
	}
}

// task run on backup server to query for and download updated data
func receiveBackup(t *task) {
	// meta := t.metadata.(backupMeta)
	//
	// jes, err := t.requester.RequestCoreSnapshot()
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }
	//
	// if len(jes) == 0 {
	// 	t.success()
	// 	return
	// }
	//
	// snap := dataStore.NewSnapshot()
	//
	// // sort journal entries by timestamp, so they can be processed in order
	// slices.SortFunc(jes, dataStore.FileJournalEntrySort)
	//
	// // The map of event "chains", following the file through its history.
	// // These are pointers to slices for "historical" reasons, might be able to
	// // squash down to just slices in the future
	// eventMap := map[types.FileId]*[]types.FileJournalEntry{}
	//
	// completeChains := [][]types.FileJournalEntry{}
	//
	// for _, je := range jes {
	//
	// 	je.SetSnapshot(snap.GetId())
	//
	// 	switch je.GetAction() {
	// 	case dataStore.FileCreate:
	// 		chain, ok := eventMap[je.GetFileId()]
	// 		if !ok {
	// 			eventMap[je.GetFileId()] = &[]types.FileJournalEntry{je}
	// 			continue
	// 		} else {
	// 			// If the file id of the destination is "occupied", we have an issue
	// 			chainStr := strings.Join(util.Map(*chain, func(j types.FileJournalEntry) string { return string(j.GetAction()) }), "->")
	// 			err := fmt.Errorf("file create found existing file. Snapshot chain: %s", chainStr)
	// 			t.ErrorAndExit(err)
	// 		}
	//
	// 	case dataStore.FileMove:
	// 		chain, ok := eventMap[je.GetFileId()]
	// 		if !ok {
	// 			// If the file id of the destination is "occupied", we have an issue
	// 			if _, ok := eventMap[je.GetFromFileId()]; ok {
	// 				t.ErrorAndExit(fmt.Errorf("backup map collide attempting to insert move, which was thought to be new"))
	// 			}
	// 			eventMap[je.GetFromFileId()] = &[]types.FileJournalEntry{je}
	// 			continue
	// 		}
	//
	// 		// if *chain == nil || (*chain)[len(*chain)-1].GetAction() == dataStore.FileDelete {
	// 		// 	chainStr := strings.Join(util.Map(*chain, func(j types.FileJournalEntry) string { return string(j.GetAction()) }), "->")
	// 		// 	err := fmt.Errorf("got unexpected file move. Snapshot chain: %s", chainStr)
	// 		// 	t.ErrorAndExit(err)
	// 		// }
	//
	// 		*chain = append(*chain, je)
	//
	// 		// remove old fileId from the map, so a new file could inhabit it
	// 		delete(eventMap, je.GetFileId())
	//
	// 		// Add new file id to point to the chain, since
	// 		// future actions will use that new id
	// 		eventMap[je.GetFromFileId()] = chain
	//
	// 	case dataStore.FileDelete:
	// 		chain, ok := eventMap[je.GetFileId()]
	// 		if !ok {
	// 			eventMap[je.GetFileId()] = &[]types.FileJournalEntry{je}
	// 			continue
	// 		}
	//
	// 		if (*chain) == nil {
	// 			err := fmt.Errorf("got unexpected file delete. Snapshot chain is nil")
	// 			t.ErrorAndExit(err)
	// 		}
	//
	// 		// create -> ... -> delete == noop, since if the file was created and deleted
	// 		// between snapshots, we won't have the file bytes or any other data between,
	// 		// so we throw away the chain
	// 		if (*chain)[0].GetAction() != dataStore.FileCreate {
	// 			// otherwise, log delete to chain and save chain to array of complete chains
	// 			*chain = append(*chain, je)
	// 			completeChains = append(completeChains, *chain)
	// 		}
	//
	// 		// remove chain from event map, since nothing can happen after the delete,
	// 		// and if another file is created in the same place, the key must be free
	// 		delete(eventMap, je.GetFileId())
	//
	// 	default:
	// 		util.Error.Println("Unexpected backup journal action:", je.GetAction())
	// 	}
	// }
	//
	// // Add chains in map to array of completed chains
	// completeChains = append(completeChains, util.Map(util.MapToSlicePure(eventMap), func(j *[]types.FileJournalEntry) []types.FileJournalEntry { return *j })...)
	//
	// newIds := util.FilterMap(completeChains, func(chain []types.FileJournalEntry) (types.FileId, bool) {
	// 	if chain[0].GetAction() == dataStore.FileCreate {
	// 		finalId, err := dataStore.GetFinalFileId(chain)
	// 		if err != nil {
	// 			t.ErrorAndExit(err)
	// 		}
	// 		return finalId, true
	// 	} else {
	// 		return "", false
	// 	}
	// })
	//
	// newFilesInfo, err := t.requester.GetCoreFileInfos(newIds)
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }
	// if len(newFilesInfo) != len(newIds) {
	// 	err = errors.New("file info count does not match what was asked for")
	// 	t.ErrorAndExit(err)
	// }
	//
	// // Yoink re-orders the slice when pulling from anywhere but the
	// // end, so we must reverse the slice so we only pull from the end
	// slices.Reverse(newFilesInfo)
	//
	// // Save chains of new events to database file history
	// for _, chain := range completeChains {
	// 	if chain[0].GetAction() == dataStore.FileCreate {
	// 		var newF types.WeblensFile
	// 		newFilesInfo, newF = util.Yoink(newFilesInfo, len(newFilesInfo)-1)
	//
	// 		finalId, err := dataStore.GetFinalFileId(chain)
	// 		if err != nil {
	// 			t.ErrorAndExit(err)
	// 		}
	// 		if finalId != newF.ID() {
	// 			err = errors.New("did not get expected file info")
	// 			t.ErrorAndExit(err)
	// 		}
	//
	// 		var bs [][]byte
	// 		if !newF.IsDir() {
	// 			bs, err = t.requester.GetCoreFileBin(newF)
	// 			if err != nil {
	// 				t.ErrorAndExit(err)
	// 			}
	// 		}
	// 		m := dataStore.MediaMapGet(newF.GetContentId())
	//
	// 		// baseF, err := dataStore.NewBackupFile(chain, meta.remoteId, bs, newF.IsDir(), mId)
	// 		_, baseF, err := dataStore.CacheBaseMedia(newF.GetContentId(), bs, meta.tree)
	// 		if err != nil {
	// 			t.ErrorAndExit(err)
	// 		}
	// 		if m != nil {
	// 			m.AddFile(baseF)
	// 			m.SetImported(false)
	//
	// 			err = m.Save()
	// 			if err != nil {
	// 				t.ErrorAndExit(err)
	// 				return
	// 			}
	//
	// 			m.SetImported(true)
	//
	// 			// err = baseF.SetMedia(m)
	// 			// if err != nil {
	// 			// 	t.ErrorAndExit(err)
	// 			// }
	// 		}
	// 	} else {
	// 		// dataStore.BackupFileAddEvents(chain)
	// 	}
	// }
	//
	// dataStore.JournalBackup(snap)

	t.ErrorAndExit(types.NewWeblensError("not impl"))
}

// task run on core server to package backup info to send to backup server
func packageBackup(t *task) {
	// acc := dataStore.NewAccessMeta(dataStore.WEBLENS_ROOT_USER.Username).SetRequestMode(dataStore.BackupFileScan)

	// t.setResult(types.TaskResult{"files": allFiles})

}
