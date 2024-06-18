package dataProcess

import (
	"encoding/json"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

// Tasks //

type taskHandler func(*task)

const (
	TaskSuccess  types.TaskExitStatus = "success"
	TaskCanceled types.TaskExitStatus = "cancelled"
	TaskError    types.TaskExitStatus = "error"
	TaskNoStatus types.TaskExitStatus = ""
)

const (
	ScanDirectoryTask types.TaskType = "scan_directory"
	ScanFileTask      types.TaskType = "scan_file"
	MoveFileTask      types.TaskType = "move_file"
	WriteFileTask     types.TaskType = "write_file"
	CreateZipTask     types.TaskType = "create_zip"
	GatherFsStatsTask types.TaskType = "gather_filesystem_stats"
	BackupTask        types.TaskType = "do_backup"
	HashFile          types.TaskType = "hash_file"
)

const (
	TaskCreated     types.TaskEvent = "task_created"
	ScanComplete    types.TaskEvent = "scan_complete"
	TaskComplete    types.TaskEvent = "task_complete"
	TaskFailed      types.TaskEvent = "task_failure"
	SubTaskComplete types.TaskEvent = "sub_task_complete"
	TaskProgress    types.TaskEvent = "task_progress_update"
)

// Internal types //

type taskMetadata interface {
	MetaString() string
}

type scanMetadata struct {
	file types.WeblensFile
	// recursive    bool
	// deepScan     bool
	fileBytes    []byte
	partialMedia types.Media
}

// MetaString overrides marshal function for scanMetadata
func (m scanMetadata) MetaString() string {
	data := map[string]any{
		"TaskType": ScanFileTask,
		"FileId":   m.file.ID(),
		// "Recursive": m.recursive,
		// "Deep":      m.deepScan,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type zipMetadata struct {
	files    []types.WeblensFile
	username types.Username
	shareId  types.ShareId
	fileTree types.FileTree
}

func (m zipMetadata) MetaString() string {
	data := map[string]any{
		"TaskType": CreateZipTask,
		"Files":    m.files,
		"Username": m.username,
		"ShareId":  m.shareId,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type moveMeta struct {
	fileId              types.FileId
	destinationFolderId types.FileId
	newFilename         string
	fileTree            types.FileTree
}

func (m moveMeta) MetaString() string {
	data := map[string]any{
		"TaskType":    MoveFileTask,
		"FileId":      m.fileId,
		"DestId":      m.destinationFolderId,
		"NewFileName": m.newFilename,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type fileChunk struct {
	FileId       types.FileId
	Chunk        []byte
	ContentRange string

	newFile types.WeblensFile
}

type writeFileMeta struct {
	chunkStream  chan fileChunk
	rootFolderId types.FileId
	chunkSize    int64
	totalSize    int64
	fileTree     types.FileTree
}

func (m writeFileMeta) MetaString() string {
	data := map[string]any{
		"TaskType":   WriteFileTask,
		"RootFolder": m.rootFolderId,
		"ChunkSize":  m.chunkSize,
		"TotalSize":  m.totalSize,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type fsStatMeta struct {
	rootDir types.WeblensFile
}

func (m fsStatMeta) MetaString() string {
	data := map[string]any{
		"TaskType":   GatherFsStatsTask,
		"RootFolder": m.rootDir.ID(),
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type fileUploadProgress struct {
	file          types.WeblensFile
	bytesWritten  int64
	fileSizeTotal int64
}

type backupMeta struct {
	remoteId types.InstanceId
	tree     types.FileTree
}

func (m backupMeta) MetaString() string {
	data := map[string]any{
		"TaskType": BackupTask,
		"remoteId": m.remoteId,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type hashFileMeta struct {
	file types.WeblensFile
}

func (m hashFileMeta) MetaString() string {
	data := map[string]any{
		"TaskType": HashFile,
		"fileId":   m.file.ID(),
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

// Errors

var ErrNonDisplayable = types.NewWeblensError("attempt to process non-displayable file")
var ErrEmptyZip = types.NewWeblensError("cannot create a zip with no files")
var ErrTaskExit = types.NewWeblensError("task exit")
var ErrTaskError = types.NewWeblensError("task generated an error")
var ErrTaskTimeout = types.NewWeblensError("task timed out")
var ErrBadTaskType = types.NewWeblensError("task metadata type is not supported on attempted operation")
var ErrBadCaster = types.NewWeblensError("task was given the wrong type of caster")
var ErrChildTaskFailed = types.NewWeblensError("a task spawned by this task has failed")
