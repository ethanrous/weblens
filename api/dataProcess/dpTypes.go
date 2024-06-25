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
	TaskCreatedEvent     types.TaskEvent = "task_created"
	PoolCreatedEvent     types.TaskEvent = "pool_created"
	PoolCancelledEvent   types.TaskEvent = "pool_cancelled"
	ScanCompleteEvent    types.TaskEvent = "scan_complete"
	TaskCompleteEvent    types.TaskEvent = "task_complete"
	ZipCompleteEvent     types.TaskEvent = "zip_complete"
	TaskFailedEvent      types.TaskEvent = "task_failure"
	TaskCanceledEvent    types.TaskEvent = "task_canceled"
	SubTaskCompleteEvent types.TaskEvent = "sub_task_complete"
	TaskProgressEvent    types.TaskEvent = "task_progress_update"
)

// Internal types //

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

func (m scanMetadata) FormatToResult() types.TaskResult {
	return types.TaskResult{
		"fileName": m.file.Filename(),
	}
}

type zipMetadata struct {
	files    []types.WeblensFile
	username types.Username
	shareId  types.ShareId
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

func (m zipMetadata) FormatToResult() types.TaskResult {
	return types.TaskResult{}
}

type moveMeta struct {
	fileId              types.FileId
	destinationFolderId types.FileId
	newFilename         string
	fileEvent           types.FileEvent
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

func (m moveMeta) FormatToResult() types.TaskResult {
	return types.TaskResult{}
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

func (m writeFileMeta) FormatToResult() types.TaskResult {
	return types.TaskResult{}
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

func (m fsStatMeta) FormatToResult() types.TaskResult {
	return types.TaskResult{}
}

type fileUploadProgress struct {
	file          types.WeblensFile
	bytesWritten  int64
	fileSizeTotal int64
}

type backupMeta struct {
	remoteId types.InstanceId
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

func (m backupMeta) FormatToResult() types.TaskResult {
	return types.TaskResult{}
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

func (m hashFileMeta) FormatToResult() types.TaskResult {
	return types.TaskResult{}
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
