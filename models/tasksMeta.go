package models

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/task"
)

const (
	ScanDirectoryTask = "scan_directory"
	ScanFileTask      = "scan_file"
	MoveFileTask    = "move_file"
	UploadFilesTask = "write_file"
	CreateZipTask   = "create_zip"
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

func (m ScanMeta) Verify() error {
	return nil
}

type ZipMeta struct {
	Files     []*fileTree.WeblensFile
	Share     *FileShare
	Requester *User

	FileService FileService
	Caster      FileCaster
}

func (m ZipMeta) MetaString() string {
	ids := internal.Map(
		m.Files, func(f *fileTree.WeblensFile) fileTree.FileId {
			return f.ID()
		},
	)

	slices.Sort(ids)
	idsString, err := json.Marshal(ids)
	log.ErrTrace(err)

	data := CreateZipTask + string(idsString) + string(m.Requester.GetUsername()) + string(m.Share.ShareId) + m.Share.LastUpdated().String()
	return data
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

func (m ZipMeta) Verify() error {
	if len(m.Files) == 0 {
		return werror.ErrBadJobMetadata(m.JobName(), "files")
	} else if m.Share == nil {
		log.Warning.Println("No share in zip meta...")
	} else if m.Requester == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "requester")
	} else if m.FileService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "fileService")
	} else if m.Caster == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "caster")
	}

	return nil
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

func (m MoveMeta) Verify() error {
	return nil
}

type FileChunk struct {
	FileId       fileTree.FileId
	Chunk        []byte
	ContentRange string

	NewFile *fileTree.WeblensFile
}

type UploadFilesMeta struct {
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

func (m UploadFilesMeta) MetaString() string {
	return fmt.Sprintf("%s%s%d%d", UploadFilesTask, m.RootFolderId, m.ChunkSize, m.TotalSize)
}

func (m UploadFilesMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m UploadFilesMeta) JobName() string {
	return UploadFilesTask
}

func (m UploadFilesMeta) Verify() error {
	if m.ChunkStream == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "chunkStream")
	} else if m.RootFolderId == "" {
		return werror.ErrBadJobMetadata(m.JobName(), "rootFolderId")
	} else if m.ChunkSize == 0 {
		return werror.ErrBadJobMetadata(m.JobName(), "chunkSize")
	} else if m.TotalSize == 0 {
		return werror.ErrBadJobMetadata(m.JobName(), "totalSize")
	} else if m.FileService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "fileService")
	} else if m.MediaService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "mediaService")
	} else if m.TaskService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "taskService")
	} else if m.TaskSubber == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "taskSubscriber")
	} else if m.User == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "user")
	} else if m.Caster == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "caster")
	}
	
	return nil
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

func (m FsStatMeta) Verify() error {
	return nil
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

func (m BackupMeta) Verify() error {
	return nil
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

func (m HashFileMeta) Verify() error {
	return nil
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

func (m BackupCoreFileMeta) Verify() error {
	return nil
}
