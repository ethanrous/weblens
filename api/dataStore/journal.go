package dataStore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fileEvent struct {
	action       types.JournalAction
	preFilePath  string
	postFilePath string
}

type eventMask struct {
	action   types.JournalAction
	path     string
	backupId string
}

var jeStream = make(chan (fileEvent), 20)

var eventMasks = []eventMask{}
var masksLock = &sync.Mutex{}

func journalWorker() {
	// var log *os.File
	// var err error
	// if util.IsDevMode() {
	// 	log, err = os.OpenFile("/Users/ethan/Downloads/journal.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
	for e := range jeStream {
		masksLock.Lock()
		index := slices.IndexFunc(eventMasks, func(m eventMask) bool { return m.path == e.postFilePath })
		var m eventMask
		if index != -1 {
			eventMasks, m = util.Yoink(eventMasks, index)
			e.action = m.action
		}
		masksLock.Unlock()

		// if log != nil {
		// 	_, err := log.Write([]byte(fmt.Sprintf("[%s] %s %s -> %s\n", time.Now().Format("2-1-06 15:04:05"), e.action, e.preFilePath, e.postFilePath)))
		// 	if err != nil {
		// 		util.ShowErr(err)
		// 	}
		// }

		switch e.action {
		case FileCreate:
			isD, err := isDirByPath(e.postFilePath)
			if err != nil {
				util.ShowErr(err)
				continue
			}

			if isD && !strings.HasSuffix(e.postFilePath, "/") {
				e.postFilePath += "/"
			}

			newFileId := generateFileId(e.postFilePath)
			newFile := FsTreeGet(newFileId)
			if newFile == nil {
				isDir, err := isDirByPath(e.postFilePath)
				if err != nil {
					util.ErrTrace(err)
					continue
				}
				parent := FsTreeGet(generateFileId(filepath.Dir(e.postFilePath) + "/"))
				if parent == nil {
					util.ErrTrace(ErrNoFile)
					continue
				}

				newFile = newWeblensFile(parent, filepath.Base(e.postFilePath), isDir)
				fsTreeInsert(newFile, parent, globalCaster)
				tasker.ScanFile(newFile, globalCaster)
			}
			if newFile == nil {
				util.Error.Printf("failed to find file at %s while writing a file create journal", e.postFilePath)
				continue
			}

			je := journalFileCreate(newFile)
			if je == nil {
				util.Warning.Println("Skipping journal of file", newFile.GetAbsPath())
				continue
			}
			backup, err := newBackupFile(newFile, je)
			if err != nil {
				util.ShowErr(err)
				continue
			}

			err = fddb.newBackupFileRecord(backup)
			if err != nil {
				util.ShowErr(err)
				continue
			}

			if !newFile.IsDir() && newFile.GetContentId() == "" {
				tasker.HashFile(newFile)
			}

		case FileMove:
			isD, err := isDirByPath(e.postFilePath)
			if err != nil {
				util.ShowErr(err)
				continue
			}

			if isD && !strings.HasSuffix(e.postFilePath, "/") {
				e.postFilePath += "/"
			}

			newFileId := generateFileId(e.postFilePath)
			newFile := FsTreeGet(newFileId)

			if newFile == nil {
				util.Error.Printf("failed to find file with Id %s (%s) while writing a file move journal", newFileId, e.postFilePath)
				continue
			}

			je := journalFileMove(e.preFilePath, newFile)
			fddb.backupFileAddHist(newFileId, je.FromFileId, []types.FileJournalEntry{je})
			// if (oldFile != nil) {
			// 	FsTreeMove()
			// 	globalCaster.PushFileMove(oldFile, newFile)
			// }
		case FileDelete:
			je := journalFileDelete(e.preFilePath)
			fddb.backupFileAddHist("", je.FromFileId, []types.FileJournalEntry{je})
			// globalCaster.PushFileDelete()
		case FileRestore:
			je := journalFileRestore(e.postFilePath)
			fddb.backupRestoreFile(je.FileId, m.backupId, []types.FileJournalEntry{je})
		}
	}
}

func addEventMask(m eventMask) {
	masksLock.Lock()
	eventMasks = append(eventMasks, m)
	masksLock.Unlock()
}

func newBackupFile(f types.WeblensFile, createEntry *fileJournalEntry) (*backupFile, error) {
	if createEntry.Action != FileCreate {
		return nil, ErrBadJournalAction
	}

	return &backupFile{
		LocalId: primitive.NewObjectID(),
		IsDir:   f.IsDir(),
		FileId:  f.Id(),
		// ContentId:  contentId,
		LastUpdate: createEntry.Timestamp,
		Events:     []*fileJournalEntry{createEntry},
	}, nil
}

func SetContentId(f types.WeblensFile, contentId types.ContentId) error {
	f.(*weblensFile).contentId = contentId
	return fddb.setContentId(f.Id(), contentId)
}

func journalFileCreate(newFile types.WeblensFile) *fileJournalEntry {
	if newFile.Owner() == WEBLENS_ROOT_USER {
		return nil
	}
	newJe := &fileJournalEntry{
		Timestamp: types.SafeTime(time.Now()),
		Action:    FileCreate,
		FileId:    newFile.Id(),
		Path:      AbsToPortable(newFile.GetAbsPath()).PortableString(),
	}

	return newJe
}

func journalFileRestore(restoreToPath string) *fileJournalEntry {
	newJe := &fileJournalEntry{
		Timestamp: types.SafeTime(time.Now()),
		Action:    FileRestore,
		FileId:    generateFileId(restoreToPath),
		Path:      AbsToPortable(restoreToPath).PortableString(),
	}

	return newJe
}

func journalFileMove(oldFilePath string, newFile types.WeblensFile) *fileJournalEntry {
	if newFile.Owner() == WEBLENS_ROOT_USER {
		return nil
	}
	newJe := &fileJournalEntry{
		Timestamp: types.SafeTime(time.Now()),
		Action:    FileMove,

		FileId:     newFile.Id(),
		FromFileId: generateFileId(oldFilePath),

		Path:     AbsToPortable(newFile.GetAbsPath()).PortableString(),
		FromPath: AbsToPortable(oldFilePath).PortableString(),
	}

	return newJe

	// go handleFile(newJe, newFile, oldFile)
}

func journalFileDelete(deletedPath string) *fileJournalEntry {
	newJe := &fileJournalEntry{
		Timestamp:  types.SafeTime(time.Now()),
		Action:     FileDelete,
		FromFileId: generateFileId(deletedPath),
		FromPath:   AbsToPortable(deletedPath).PortableString(),
	}

	return newJe

	// go handleFile(newJe, nil, deletedF)
}

func NewSnapshot() types.Snapshot {
	newId := util.GlobbyHash(8, time.Now())
	return &snapshot{
		Id: newId,
	}
}

func JournalBackup(snap types.Snapshot) {
	// newJe := &backupJournalEntry{
	// 	Action:    Backup,
	// 	Timestamp: time.Now(),
	// 	Snapshot:  *snap.(*snapshot),
	// }

	// addJE(newJe)
}

func GetSnapshots() ([]types.JournalEntry, error) {
	return fddb.getSnapshots()
}

func JeFileId(je types.JournalEntry) (types.FileId, error) {
	switch je.GetAction() {
	case FileCreate, FileMove, FileWrite, FileDelete:
		return je.(*fileJournalEntry).FileId, nil
	default:
		return "", ErrBadJournalAction
	}
}

func GetPastFileInfo(folder types.WeblensFile, acc types.AccessMeta) ([]types.FileInfo, error) {
	path := AbsToPortable(folder.GetAbsPath()).PortableString()
	backups, err := fddb.getFilesPathAndTime(path, acc.GetTime())
	if err != nil {
		return nil, err
	}

	infos := util.FilterMap(backups, func(b backupFile) (types.FileInfo, bool) {
		// m, err := MediaMapGet(types.ContentId(b.ContentId[:8]))
		if err != nil {
			util.ShowErr(err)
		}
		absPath := portableFromString(b.Events[len(b.Events)-1].Path).Abs()
		tmpF := newWeblensFile(folder, filepath.Base(absPath), b.IsDir)
		tmpF.pastFile = true
		tmpF.contentId = b.ContentId
		// tmpF.media = m
		if b.FileId != "" {
			tmpF.currentId = b.FileId
		} else {
			child, err := contentRoot.GetChild(string(b.ContentId))
			if err != nil {
				util.ErrTrace(err)
				return types.FileInfo{}, false
			}
			tmpF.currentId = child.Id()
		}

		info, err := tmpF.FormatFileInfo(acc)
		if err != nil {
			util.ShowErr(err, tmpF.absolutePath)
			return info, false
		}

		return info, true
	})
	return infos, nil
}

func GetFinalFileId(chain []types.FileJournalEntry) (types.FileId, error) {
	last := chain[len(chain)-1]
	switch last.GetAction() {
	case FileCreate, FileDelete:
		return last.GetFileId(), nil
	case FileMove:
		return last.GetFromFileId(), nil
	default:
		return "", fmt.Errorf("got unexpected final action type in chain: %s", last.GetAction())
	}
}

func matchEventsPath(r *regexp.Regexp, invert bool) func(*fileJournalEntry) bool {
	return func(fje *fileJournalEntry) bool {
		switch fje.Action {
		case FileCreate, FileMove, FileRestore:
			return r.MatchString(fje.Path) != invert
		default:
			return false
		}
	}
}

func GetFileHistory(fileId types.FileId) ([]types.FileJournalEntry, error) {
	f := FsTreeGet(fileId)
	if f == nil {
		return nil, ErrNoFile
	}

	filePath := AbsToPortable(f.GetAbsPath()).PortableString()

	files, err := fddb.fileEventsByPath(filePath)
	if err != nil {
		return nil, err
	}

	regexStr := "^" + filePath + "[^/]*$"
	r, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}
	events := []*fileJournalEntry{}
	for _, be := range files {
		for len(be.Events) != 0 {
			// find the first create or move event relevant to us
			goodIndex := slices.IndexFunc(be.Events, matchEventsPath(r, false))
			if goodIndex == -1 {
				break
			}
			be.Events = be.Events[goodIndex:]

			badIndex := slices.IndexFunc(be.Events, matchEventsPath(r, true))
			// Bad index matches the next (move) event where the destination path is not what we want,
			// but that means we are moving out of a directory that we do care about, so we include the
			// event at the bad index
			if badIndex != -1 {
				badIndex++
			} else {
				events = append(events, be.Events...)
				break
			}

			events = append(events, be.Events[:badIndex]...)
			be.Events = be.Events[badIndex:]
		}
	}

	iEvents := util.SliceConvert[types.FileJournalEntry](events)
	slices.SortFunc(iEvents, FileJournalEntrySort)
	return iEvents, nil
}

func JournalSince(since time.Time) ([]types.JournalEntry, error) {
	jes, err := fddb.journalSince(since)
	if err != nil {
		util.ShowErr(err)
		return nil, err
	}
	return jes, nil
}

func GetLatestBackup() (t time.Time, err error) {
	latest, err := fddb.getLatestBackup()
	if err != nil && err != ErrNoBackup {
		return
	} else if err == ErrNoBackup {
		return time.Unix(0, 0), nil
	}

	events := util.SliceConvert[types.FileJournalEntry](latest.Events)
	slices.SortFunc(events, FileJournalEntrySort)
	t = events[len(events)-1].JournaledAt()

	return
}

func RestoreFiles(fileIds []types.FileId, timestamp time.Time) error {
	for _, fId := range fileIds {
		backupFiles, err := fddb.findPastFile(fId, timestamp)
		if err != nil {
			return err
		}

		var latestEvent *fileJournalEntry
		var backupFile backupFile
		for _, b := range backupFiles {
			es := util.Filter(b.Events, func(e *fileJournalEntry) bool {
				return e.FileId == fId && time.Time(e.Timestamp).Unix() == timestamp.Unix() || time.Time(e.Timestamp).Before(timestamp)
			})
			if len(es) != 0 {
				latestEvent = es[len(es)-1]
				backupFile = b
				break
			}
		}
		if backupFile.FileId != "" {
			util.Warning.Println("Skipping file restore because it already has an ID")
			continue
		}
		abs := portableFromString(latestEvent.Path).Abs()
		parentId := generateFileId(filepath.Dir(abs) + "/")
		parent := FsTreeGet(parentId)

		restoredF := newWeblensFile(parent, filepath.Base(abs), backupFile.IsDir)

		addEventMask(eventMask{action: FileRestore, path: abs, backupId: backupFile.LocalId.Hex()})

		os.Rename(filepath.Join(contentRoot.absolutePath, string(backupFile.ContentId)), abs)
		fsTreeInsert(restoredF, parent)
		util.Debug.Println(latestEvent, backupFile.ContentId)
	}

	return nil
}
