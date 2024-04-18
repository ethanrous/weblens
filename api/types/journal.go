package types

import "time"

type JournalEntry interface {
	JournaledAt() time.Time
	GetAction() JournalAction
}

type FileJournalEntry interface {
	GetFileId() FileId
	GetFromFileId() FileId
	SetSnapshot(snapshotId string)
	JournalEntry
}

const JOURNAL_BUFFER_SIZE = 100

type JournalAction string

type Snapshot interface {
	AddEvent(je FileJournalEntry)
	GetId() string
}

// type Journaler interface {
// 	JournalFileCreate(newFile WeblensFile)
// 	JournalFileMove(oldId FileId, newFile WeblensFile)
// 	JournalFileDelete(deletedId FileId)
// 	JournalFileWrite(file WeblensFile, wroteSize, startPos int64)
// 	JournalBackup()
// }
