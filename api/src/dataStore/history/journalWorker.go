package history

import (
	"errors"
	"sync"

	"github.com/ethrousseau/weblens/api/types"
)

type eventMask struct {
	actionType types.FileActionType
	path       string
	backupId   string
}

var jeStream = make(chan types.FileAction, 20)

var eventMasks []eventMask
var masksLock = &sync.Mutex{}

func MaskEvent(action types.FileActionType, path string) {
	em := eventMask{
		actionType: action,
		path:       path,
	}

	masksLock.Lock()
	defer masksLock.Unlock()
	eventMasks = append(eventMasks, em)
}

var journalWorkerRunning = false

func (j *journalService) JournalWorker() {
	if journalWorkerRunning {
		panic(errors.New("attempted to start duplicate journal worker"))
	}

	journalWorkerRunning = true
	defer func() { journalWorkerRunning = false }()

	for _ = range jeStream {
		// masksLock.Lock()
		// index := slices.IndexFunc(eventMasks, func(m eventMask) bool { return m.path == action.GetDestinationPath() })
		// var m eventMask
		// if index != -1 {
		// 	eventMasks, m = util.Yoink(eventMasks, index)
		// 	action.SetActionType(m.actionType)
		// }
		// masksLock.Unlock()
		//
		// switch action.GetActionType() {
		// case FileCreate:
		// 	isD, err := util.IsDirByPath(action.GetDestinationPath())
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 		continue
		// 	}
		//
		// 	if isD && !strings.HasSuffix(action.GetDestinationPath(), "/") {
		// 		action.SetDestinationPath(action.GetDestinationPath() + "/")
		// 	}
		//
		// 	newFileId := ft.GenerateFileId(action.GetDestinationPath())
		// 	newFile := ft.Get(newFileId)
		// 	if newFile == nil {
		// 		isDir, err := util.IsDirByPath(action.GetDestinationPath())
		// 		if err != nil {
		// 			util.ErrTrace(err)
		// 			continue
		// 		}
		// 		parent := ft.Get(ft.GenerateFileId(filepath.Dir(action.GetDestinationPath()) + "/"))
		// 		if parent == nil {
		// 			util.ErrTrace(errors.New("No parent"))
		// 			continue
		// 		}
		//
		// 		newFile = ft.NewFile(parent, filepath.Base(action.GetDestinationPath()), isDir)
		// 		err = ft.Add(newFile, parent)
		// 		if err != nil {
		// 			util.ErrTrace(err)
		// 			continue
		// 		}
		//
		// 		// TODO - add controllers to history package
		// 		// tasker.ScanFile(newFile)
		// 	}
		// 	if newFile == nil {
		// 		util.Error.Printf("failed to find file at %s while writing a file create journal", action.postFilePath)
		// 		continue
		// 	}
		//
		// 	je := journalFileCreate(newFile)
		// 	if je == nil {
		// 		util.Warning.Println("Skipping journal of file", newFile.GetAbsPath())
		// 		continue
		// 	}
		// 	backup, err := newBackupFile(newFile, je)
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 		continue
		// 	}
		//
		// 	err = j.dbServer.WriteFileEvent() dbServer.newBackupFileRecord(backup)
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 		continue
		// 	}
		//
		// 	if !newFile.IsDir() && newFile.GetContentId() == "" {
		// 		tasker.HashFile(newFile)
		// 	}
		//
		// case FileMove:
		// 	isD, err := isDirByPath(action.GetDestinationPath())
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 		continue
		// 	}
		//
		// 	if isD && !strings.HasSuffix(action.GetDestinationPath(), "/") {
		// 		action.postFilePath += "/"
		// 	}
		//
		// 	newFileId := generateFileId(action.GetDestinationPath())
		// 	newFile := ft.Get(newFileId)
		//
		// 	if newFile == nil {
		// 		util.Error.Printf("failed to find file with ID %s (%s) while writing a file move journal", newFileId, action.postFilePath)
		// 		continue
		// 	}
		//
		// 	je := journalFileMove(action.preFilePath, newFile)
		// 	err = dbServer.backupFileAddHist(newFileId, je.FromFileId, []types.FileJournalEntry{je})
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 		return
		// 	}
		// case FileDelete:
		// 	je := journalFileDelete(action.preFilePath)
		// 	err := dbServer.backupFileAddHist("", je.FromFileId, []types.FileJournalEntry{je})
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 		return
		// 	}
		// 	// globalCaster.PushFileDelete()
		// case FileRestore:
		// 	je := journalFileRestore(action.postFilePath)
		// 	err := dbServer.backupRestoreFile(je.FileId, m.backupId, []types.FileJournalEntry{je})
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 		return
		// 	}
		// }
	}
}
