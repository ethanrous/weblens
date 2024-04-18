package dataStore

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var journal []types.JournalEntry
var journalLock *sync.Mutex = &sync.Mutex{}

// Journal a file event on the correct backup file
func handleFile(je *fileJournalEntry, postFile, prefile types.WeblensFile) {
	postId := types.FileId("")
	switch je.Action {
	case FileCreate:
		backup, err := newBackupFile(postFile, je)
		if err != nil {
			util.ShowErr(err)
			return
		}
		fddb.newBackupFileRecord(backup)
	case FileMove:
		postId = postFile.Id()
		fallthrough
	case FileDelete:
		fddb.backupFileAddHist(postId, prefile.Id(), []types.FileJournalEntry{je})
	}
}

func newBackupFile(f types.WeblensFile, createEntry *fileJournalEntry) (*backupFile, error) {
	if createEntry.Action != FileCreate {
		return nil, ErrBadJournalAction
	}

	contentId, err := handleFileContent(f)
	if err != nil {
		return nil, err
	}

	return &backupFile{
		LocalId:    primitive.NewObjectID(),
		IsDir:      f.IsDir(),
		FileId:     f.Id(),
		ContentId:  contentId,
		LastUpdate: createEntry.Timestamp,
		Events:     []*fileJournalEntry{createEntry},
	}, nil
}

func JournalSince(since time.Time) ([]types.JournalEntry, error) {
	journalFlush()
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

func journalFlush() {
	journalLock.Lock()
	err := fddb.writeJournalEntries(journal)
	if err != nil {
		util.ErrTrace(err)
	}
	journal = []types.JournalEntry{}
	journalLock.Unlock()
}

func addJE(newJe types.JournalEntry) {
	journalLock.Lock()
	journal = append(journal, newJe)
	if len(journal) > types.JOURNAL_BUFFER_SIZE {
		journalLock.Unlock()
		journalFlush()
		return
	}
	journalLock.Unlock()
}

func JournalFileCreate(newFile types.WeblensFile) {
	if newFile.Owner() == WEBLENS_ROOT_USER {
		return
	}
	newJe := &fileJournalEntry{
		Timestamp: types.SafeTime(time.Now()),
		Action:    FileCreate,
		FileId:    newFile.Id(),
		Path:      AbsToPortable(newFile.GetAbsPath()).PortableString(),
	}

	go handleFile(newJe, newFile, nil)
}

func JournalFileMove(oldFile types.WeblensFile, newFile types.WeblensFile) {
	if newFile.Owner() == WEBLENS_ROOT_USER {
		return
	}
	newJe := &fileJournalEntry{
		Timestamp: types.SafeTime(time.Now()),
		Action:    FileMove,

		FileId:     newFile.Id(),
		FromFileId: oldFile.Id(),

		Path:     AbsToPortable(newFile.GetAbsPath()).PortableString(),
		FromPath: AbsToPortable(oldFile.GetAbsPath()).PortableString(),
	}

	go handleFile(newJe, newFile, oldFile)
}

func JournalFileDelete(deletedF types.WeblensFile) {
	newJe := &fileJournalEntry{
		Timestamp: types.SafeTime(time.Now()),
		Action:    FileDelete,
		FileId:    deletedF.Id(),
	}

	go handleFile(newJe, nil, deletedF)
}

func JournalFileWrite(file types.WeblensFile, wroteSize, startPos int64) {
	// if file.Owner() == WEBLENS_ROOT_USER {
	// 	return
	// }
	// newJe := &fileJournalEntry{
	// 	Timestamp: types.SafeTime(time.Now()),
	// 	Action:    FileWrite,
	// 	FileId:    file.Id(),
	// 	Size:      wroteSize,
	// 	At:        startPos,
	// }

	// handleFile(newJe)
}

func NewSnapshot() types.Snapshot {
	newId := util.GlobbyHash(8, time.Now())
	return &snapshot{
		Id: newId,
	}
}

func JournalBackup(snap types.Snapshot) {
	newJe := &backupJournalEntry{
		Action:    Backup,
		Timestamp: time.Now(),
		Snapshot:  *snap.(*snapshot),
	}

	addJE(newJe)
	journalFlush()
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

// func NewBackupFile(jes []types.FileJournalEntry, remoteId string, data [][]byte, isDir bool, mediaId types.MediaId) (types.WeblensFile, error) {
// 	if jes[0].GetAction() != FileCreate {
// 		return nil, ErrBadJournalAction
// 	}

// 	var base types.WeblensFile
// 	var baseId types.FileId
// 	if !isDir {
// 		var err error
// 		base, err = BackupBaseFile(remoteId, data[0])
// 		if err != nil && err != ErrDirAlreadyExists && err != ErrFileAlreadyExists {
// 			return nil, err
// 		}

// 		baseId = base.Id()

// 		if len(data) == 3 && mediaId != "" {
// 			CacheBaseMedia(mediaId, data[1:])
// 		}
// 	}

// 	finalId, err := GetFinalFileId(jes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	bf := backupFile{
// 		LocalId:       primitive.NewObjectID(),
// 		IsDir:         isDir,
// 		RemoteFileId:  finalId,
// 		BaseContentId: baseId,
// 		Events:        util.SliceConvert[*fileJournalEntry](jes),
// 		MediaId:       mediaId,
// 	}

// 	return base, fddb.newBackupFileRecord(bf)
// }

// func BackupFileAddEvents(jes []types.FileJournalEntry) error {
// 	lookup := jes[0].GetFileId()
// 	setId, err := GetFinalFileId(jes)
// 	if err != nil {
// 		return err
// 	}
// 	return
// }

func GetPastFileInfo(folder types.WeblensFile, acc types.AccessMeta) ([]types.FileInfo, error) {
	path := AbsToPortable(folder.GetAbsPath()).PortableString()
	backups, err := fddb.getFilesPathAndTime(path, acc.GetTime())
	if err != nil {
		return nil, err
	}

	infos := util.Map(backups, func(b backupFile) types.FileInfo {
		m, err := MediaMapGet(types.MediaId(b.ContentId))
		if err != nil {
			util.ShowErr(err)
		}
		absPath := portableFromString(b.Events[len(b.Events)-1].Path).Abs()
		tmpF := newWeblensFile(folder, filepath.Base(absPath), b.IsDir)
		tmpF.pastFile = true
		tmpF.contentId = b.ContentId
		tmpF.media = m

		info, err := tmpF.FormatFileInfo(acc)
		if err != nil {
			util.ShowErr(err)
			return info
		}

		return info
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
		case FileCreate, FileMove:
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
			// find the first create or move event relevent to us
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
