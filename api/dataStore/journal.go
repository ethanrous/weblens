package dataStore

//
// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"regexp"
// 	"slices"
// 	"time"
//
// 	"github.com/ethrousseau/weblens/api/dataStore/history"
// 	"github.com/ethrousseau/weblens/api/types"
// 	"github.com/ethrousseau/weblens/api/util"
// )
//
// // func newBackupFile(f types.WeblensFile, createEntry *fileJournalEntry) (*backupFile, error) {
// // 	if createEntry.Action != FileCreate {
// // 		return nil, ErrBadJournalAction
// // 	}
// //
// // 	return &backupFile{
// // 		LocalId: primitive.NewObjectID(),
// // 		IsDir:   f.IsDir(),
// // 		FileId:  f.ID(),
// // 		// ContentId:  contentId,
// // 		LastUpdate: createEntry.Timestamp,
// // 		Events:     []*fileJournalEntry{createEntry},
// // 	}, nil
// // }
//
// // func SetContentId(f types.WeblensFile, contentId types.ContentId) error {
// // 	f.(*weblensFile).contentId = contentId
// // 	return dbServer.setContentId(f.ID(), contentId)
// // }
//
// // func journalFileCreate(newFile types.WeblensFile) *fileJournalEntry {
// // 	if newFile.Owner() == WeblensRootUser {
// // 		return nil
// // 	}
// // 	newJe := &fileJournalEntry{
// // 		Timestamp: types.SafeTime(time.Now()),
// // 		Action:    FileCreate,
// // 		FileId:    newFile.ID(),
// // 		Path:      absToPortable(newFile.GetAbsPath()).PortableString(),
// // 	}
// //
// // 	return newJe
// // }
//
// // func journalFileRestore(restoreToPath string) *fileJournalEntry {
// // 	newJe := &fileJournalEntry{
// // 		Timestamp: types.SafeTime(time.Now()),
// // 		Action:    FileRestore,
// // 		FileId:    generateFileId(restoreToPath),
// // 		Path:      absToPortable(restoreToPath).PortableString(),
// // 	}
// //
// // 	return newJe
// // }
//
// // func journalFileMove(oldFilePath string, newFile types.WeblensFile) *fileJournalEntry {
// // 	if newFile.Owner() == WeblensRootUser {
// // 		return nil
// // 	}
// // 	newJe := &fileJournalEntry{
// // 		Timestamp: types.SafeTime(time.Now()),
// // 		Action:    FileMove,
// //
// // 		FileId:     newFile.ID(),
// // 		FromFileId: generateFileId(oldFilePath),
// //
// // 		Path:     absToPortable(newFile.GetAbsPath()).PortableString(),
// // 		FromPath: absToPortable(oldFilePath).PortableString(),
// // 	}
// //
// // 	return newJe
// //
// // 	// go handleFile(newJe, newFile, oldFile)
// // }
//
// // func journalFileDelete(deletedPath string) *fileJournalEntry {
// // 	newJe := &fileJournalEntry{
// // 		Timestamp:  types.SafeTime(time.Now()),
// // 		Action:     FileDelete,
// // 		FromFileId: generateFileId(deletedPath),
// // 		FromPath:   absToPortable(deletedPath).PortableString(),
// // 	}
// //
// // 	return newJe
// //
// // 	// go handleFile(newJe, nil, deletedF)
// // }
//
// // func NewSnapshot() types.Snapshot {
// // 	newId := util.GlobbyHash(8, time.Now())
// // 	return &snapshot{
// // 		Id: newId,
// // 	}
// // }
//
// // func JournalBackup(snap types.Snapshot) {
// 	// newJe := &backupJournalEntry{
// 	// 	Action:    Backup,
// 	// 	Timestamp: time.Now(),
// 	// 	Snapshot:  *snap.(*snapshot),
// 	// }
//
// 	// addJE(newJe)
// // }
//
// // func GetSnapshots() ([]types.JournalEntry, error) {
// // 	return dbServer.getSnapshots()
// // }
//
// func GetPastFileInfo(folder types.WeblensFile, acc types.AccessMeta) ([]types.FileInfo, error) {
// 	path := FilepathFromAbs(folder.GetAbsPath()).ToPortable()
// 	backups, err := dbServer.getFilesPathAndTime(path, acc.GetTime())
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	infos := util.FilterMap(backups, func(b backupFile) (types.FileInfo, bool) {
// 		// m, err := MediaMapGet(types.ContentId(b.ContentId[:8]))
// 		if err != nil {
// 			util.ShowErr(err)
// 		}
// 		absPath := FilepathFromPortable(b.Events[len(b.Events)-1].Path).ToAbsPath()
// 		tmpF := folder.GetTree().NewFile(folder, filepath.Base(absPath), b.IsDir).(*weblensFile)
// 		tmpF.pastFile = true
// 		tmpF.contentId = b.ContentId
// 		// tmpF.media = m
// 		if b.FileId != "" {
// 			tmpF.currentId = b.FileId
// 		} else {
// 			child, err := contentRoot.GetChild(string(b.ContentId))
// 			if err != nil {
// 				util.ErrTrace(err)
// 				return types.FileInfo{}, false
// 			}
// 			tmpF.currentId = child.ID()
// 		}
//
// 		info, err := tmpF.FormatFileInfo(acc)
// 		if err != nil {
// 			util.ShowErr(err, tmpF.absolutePath)
// 			return info, false
// 		}
//
// 		return info, true
// 	})
// 	return infos, nil
// }
//
// func GetFinalFileId(chain []types.FileJournalEntry) (types.FileId, error) {
// 	last := chain[len(chain)-1]
// 	switch last.GetAction() {
// 	case FileCreate, FileDelete:
// 		return last.GetFileId(), nil
// 	case FileMove:
// 		return last.GetFromFileId(), nil
// 	default:
// 		return "", fmt.Errorf("got unexpected final action type in chain: %s", last.GetAction())
// 	}
// }
//
// // func matchEventsPath(r *regexp.Regexp, invert bool) func(*fileJournalEntry) bool {
// // 	return func(fje *fileJournalEntry) bool {
// // 		switch fje.Action {
// // 		case FileCreate, FileMove, FileRestore:
// // 			return r.MatchString(fje.Path) != invert
// // 		default:
// // 			return false
// // 		}
// // 	}
// // }
//
// func GetFileHistory(f types.WeblensFile) ([]types.FileJournalEntry, error) {
// 	if f == nil {
// 		return nil, ErrNoFile
// 	}
//
// 	filePath := FilepathFromAbs(f.GetAbsPath()).ToPortable()
//
// 	files, err := dbServer.fileEventsByPath(filePath)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	regexStr := "^" + filePath + "[^/]*$"
// 	r, err := regexp.Compile(regexStr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var events []*fileJournalEntry
// 	for _, be := range files {
// 		for len(be.Events) != 0 {
// 			// find the first create or move event relevant to us
// 			goodIndex := slices.IndexFunc(be.Events, matchEventsPath(r, false))
// 			if goodIndex == -1 {
// 				break
// 			}
// 			be.Events = be.Events[goodIndex:]
//
// 			badIndex := slices.IndexFunc(be.Events, matchEventsPath(r, true))
// 			// Bad index matches the next (move) event where the destination path is not what we want,
// 			// but that means we are moving out of a directory that we do care about, so we include the
// 			// event at the bad index
// 			if badIndex != -1 {
// 				badIndex++
// 			} else {
// 				events = append(events, be.Events...)
// 				break
// 			}
//
// 			events = append(events, be.Events[:badIndex]...)
// 			be.Events = be.Events[badIndex:]
// 		}
// 	}
//
// 	iEvents := util.SliceConvert[types.FileJournalEntry](events)
// 	slices.SortFunc(iEvents, FileJournalEntrySort)
// 	return iEvents, nil
// }
//
// func JournalSince(since time.Time) ([]types.JournalEntry, error) {
// 	jes, err := dbServer.journalSince(since)
// 	if err != nil {
// 		util.ShowErr(err)
// 		return nil, err
// 	}
// 	return jes, nil
// }
//
// func GetLatestBackup() (t time.Time, err error) {
// 	latest, err := dbServer.getLatestBackup()
// 	if err != nil && !errors.Is(err, ErrNoBackup) {
// 		return
// 	} else if errors.Is(err, ErrNoBackup) {
// 		return time.Unix(0, 0), nil
// 	}
//
// 	events := util.SliceConvert[types.FileJournalEntry](latest.Events)
// 	slices.SortFunc(events, FileJournalEntrySort)
// 	t = events[len(events)-1].JournaledAt()
//
// 	return
// }
//
// func RestoreFiles(fileIds []types.FileId, timestamp time.Time, ft types.FileTree) error {
// 	for _, fId := range fileIds {
// 		backupFiles, err := dbServer.findPastFile(fId, timestamp)
// 		if err != nil {
// 			return err
// 		}
//
// 		var latestEvent *fileJournalEntry
// 		var backupFile backupFile
// 		for _, b := range backupFiles {
// 			es := util.Filter(b.Events, func(e *fileJournalEntry) bool {
// 				return e.FileId == fId && time.Time(e.Timestamp).Unix() == timestamp.Unix() || time.Time(e.Timestamp).Before(timestamp)
// 			})
// 			if len(es) != 0 {
// 				latestEvent = es[len(es)-1]
// 				backupFile = b
// 				break
// 			}
// 		}
// 		if backupFile.FileId != "" {
// 			util.Warning.Println("Skipping file restore because it already has an ID")
// 			continue
// 		}
//
// 		if latestEvent == nil {
// 			err = errors.New("could not get latest event")
// 			return err
// 		}
//
// 		abs := FilepathFromPortable(latestEvent.Path).ToAbsPath()
// 		parentId := ft.GenerateFileId(filepath.Dir(abs) + "/")
// 		parent := ft.Get(parentId)
//
// 		restoredF := ft.NewFile(parent, filepath.Base(abs), backupFile.IsDir)
//
// 		history.MaskEvent(FileRestore, abs)
// 		// addEventMask(eventMask{action: FileRestore, path: abs, backupId: backupFile.LocalId.Hex()})
//
// 		err = os.Rename(filepath.Join(contentRoot.absolutePath, string(backupFile.ContentId)), abs)
// 		if err != nil {
// 			return err
// 		}
//
// 		err = ft.Add(restoredF, parent)
// 		if err != nil {
// 			return err
// 		}
//
// 		util.Debug.Println(latestEvent, backupFile.ContentId)
// 	}
//
// 	return nil
// }
