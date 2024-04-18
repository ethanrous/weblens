package dataStore

import (
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//////////
// File //
//////////

const (
	FileCreate types.JournalAction = "fileCreate"
	FileMove   types.JournalAction = "fileMove"
	FileDelete types.JournalAction = "fileDelete"
	FileWrite  types.JournalAction = "fileWrite"
	Backup     types.JournalAction = "backup"
)

type fileJournalEntry struct {
	Timestamp types.SafeTime      `bson:"timestamp"`
	Action    types.JournalAction `bson:"action"`
	FileId    types.FileId        `bson:"fileId"`

	// For move actions, the id of the new file before moving
	FromFileId types.FileId `bson:"fromFileId,omitempty"`

	// Portable path string. Location of a new file if action is
	// fileCreate, or location of destination file if action is fileMove
	Path string `bson:"path,omitempty"`

	// Origin file if action is move
	FromPath string `bson:"fromPath,omitempty"`

	// For write actions //
	// Size of new data written
	Size int64 `bson:"size,omitempty"`
	// Start pos of new data
	At int64 `bson:"at,omitempty"`

	SnapshotId string `bson:"snapshotId,omitempty"`
}

func (je *fileJournalEntry) GetFileId() types.FileId {
	return je.FileId
}

func (je *fileJournalEntry) GetFromFileId() types.FileId {
	return je.FromFileId
}

func (je *fileJournalEntry) JournaledAt() time.Time {
	return time.Time(je.Timestamp)
}

func (je *fileJournalEntry) GetAction() types.JournalAction {
	return je.Action
}

func (je *fileJournalEntry) SetSnapshot(snapshotId string) {
	je.SnapshotId = snapshotId
}

func FileJournalEntrySort(a, b types.FileJournalEntry) int {
	return a.JournaledAt().Compare(b.JournaledAt())
}

////////////
// Backup //
////////////

type backupJournalEntry struct {
	Action    types.JournalAction `bson:"action"`
	Timestamp time.Time           `bson:"timestamp"`
	Snapshot  snapshot            `bson:"snapshot"`
}

func (je *backupJournalEntry) JournaledAt() time.Time {
	return je.Timestamp
}

func (je *backupJournalEntry) GetAction() types.JournalAction {
	return Backup
}

type backupFile struct {
	LocalId primitive.ObjectID `bson:"_id" json:"backupId"` // not a file id, the id of this servers representation of the remote file
	IsDir   bool               `bson:"isDir" json:"isDir"`
	FileId  types.FileId       `bson:"fileId" json:"fileId"`

	// ContentId is the hash of the file contents. Conveniently this is also the mediaId
	ContentId string `bson:"contentId,omitempty" json:"contentId"`

	LastUpdate types.SafeTime      `bson:"lastUpdate" json:"lastUpdate"`
	Events     []*fileJournalEntry `bson:"events" json:"events"`
}

type snapshot struct {
	Id     string             `bson:"_id"`
	Events []fileJournalEntry `bson:"events"`
}

func (s *snapshot) AddEvent(je types.FileJournalEntry) {
	s.Events = append(s.Events, *je.(*fileJournalEntry))
}

func (s *snapshot) GetId() string {
	return s.Id
}
