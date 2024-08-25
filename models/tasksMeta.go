package models

import (
	"encoding/json"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/task"
)

const (
	ScanDirectoryTask = "scan_directory"
	ScanFileTask      = "scan_file"
	MoveFileTask      = "move_file"
	WriteFileTask     = "write_file"
	CreateZipTask     = "create_zip"
	GatherFsStatsTask = "gather_filesystem_stats"
	BackupTask        = "do_backup"
	HashFile          = "hash_file"
	CopyFileFromCore  = "copy_file_from_core"
)

type TaskSubscriber interface {
	FolderSubToPool(fileTree.FileId, task.TaskId)
}

type ScanMeta struct {
	File         *fileTree.WeblensFile
	FileBytes    []byte
	PartialMedia *Media

	FileService  FileService
	MediaService MediaService
	TaskService  task.TaskService
	TaskSubber   TaskSubscriber
	Caster       FileCaster
}

// MetaString overrides marshal function for scanMetadata
func (m ScanMeta) MetaString() string {
	data := map[string]any{
		"JobName": ScanFileTask,
		"FileId":  m.File.ID(),
		// "Recursive": m.recursive,
		// "Deep":      m.deepScan,
	}
	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m ScanMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{
		"filename": m.File.Filename(),
	}
}

func (m ScanMeta) JobName() string {
	if m.File.IsDir() {
		return ScanDirectoryTask
	} else {
		return ScanFileTask
	}
}

type ZipMeta struct {
	Files    []*fileTree.WeblensFile
	Username Username
	Share    *FileShare
	Owner    *User

	FileService FileService
	Caster      FileCaster
}

func (m ZipMeta) MetaString() string {
	ids := internal.Map(
		m.Files, func(f *fileTree.WeblensFile) fileTree.FileId {
			return f.ID()
		},
	)

	data := map[string]any{
		"JobName":  CreateZipTask,
		"Files":    ids,
		"Username": m.Username,
		"Share":    m.Share,
	}
	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m ZipMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{
		"filenames": internal.Map(
			m.Files, func(f *fileTree.WeblensFile) string {
				return f.Filename()
			},
		),
	}
}

func (m ZipMeta) JobName() string {
	return CreateZipTask
}

type MoveMeta struct {
	FileId              fileTree.FileId
	DestinationFolderId fileTree.FileId
	NewFilename         string
	FileEvent           *fileTree.FileEvent
	Caster              FileCaster

	User        *User
	FileService FileService
}

func (m MoveMeta) MetaString() string {
	data := map[string]any{
		"JobName":     MoveFileTask,
		"FileId":      m.FileId,
		"DestId":      m.DestinationFolderId,
		"NewFileName": m.NewFilename,
	}
	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m MoveMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m MoveMeta) JobName() string {
	return MoveFileTask
}

type FileChunk struct {
	FileId       fileTree.FileId
	Chunk        []byte
	ContentRange string

	NewFile *fileTree.WeblensFile
}

type WriteFileMeta struct {
	ChunkStream  chan FileChunk
	RootFolderId fileTree.FileId
	ChunkSize    int64
	TotalSize    int64

	FileService  FileService
	MediaService MediaService
	TaskService  task.TaskService
	TaskSubber   TaskSubscriber
	User         *User
	Share        *FileShare
	Caster       FileCaster
}

func (m WriteFileMeta) MetaString() string {
	data := map[string]any{
		"JobName":    WriteFileTask,
		"RootFolder": m.RootFolderId,
		"ChunkSize":  m.ChunkSize,
		"TotalSize":  m.TotalSize,
	}
	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m WriteFileMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m WriteFileMeta) JobName() string {
	return WriteFileTask
}

type FsStatMeta struct {
	RootDir *fileTree.WeblensFile
}

func (m FsStatMeta) MetaString() string {
	data := map[string]any{
		"JobName":    GatherFsStatsTask,
		"RootFolder": m.RootDir.ID(),
	}
	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m FsStatMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m FsStatMeta) JobName() string {
	return GatherFsStatsTask
}

type FileUploadProgress struct {
	File          *fileTree.WeblensFile
	BytesWritten  int64
	FileSizeTotal int64
}

type BackupMeta struct {
	RemoteId        InstanceId
	InstanceService InstanceService
}

func (m BackupMeta) MetaString() string {
	data := map[string]any{
		"JobName":  BackupTask,
		"remoteId": m.RemoteId,
	}
	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m BackupMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m BackupMeta) JobName() string {
	return BackupTask
}

type HashFileMeta struct {
	File *fileTree.WeblensFile
}

func (m HashFileMeta) MetaString() string {
	data := map[string]any{
		"JobName": HashFile,
		"fileId":  m.File.ID(),
	}
	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m HashFileMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m HashFileMeta) JobName() string {
	return HashFile
}

type BackupCoreFileMeta struct {
	File *fileTree.WeblensFile
	// Client comm.Client
}

func (m BackupCoreFileMeta) MetaString() string {
	data := map[string]any{
		"JobName":    CopyFileFromCore,
		"backupFile": m.File.ID(),
	}

	bs, err := json.Marshal(data)
	log.ErrTrace(err)

	return string(bs)
}

func (m BackupCoreFileMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m BackupCoreFileMeta) JobName() string {
	return CopyFileFromCore
}
