package dataProcess

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

// Tasks //

type taskTracker struct {
	taskMu      sync.Mutex
	taskMap     map[types.TaskId]types.Task
	wp          *WorkerPool
	globalQueue *taskPool
}

type taskHandler func(*task)

type task struct {
	taskId     types.TaskId
	completed  bool
	taskPool   *taskPool
	caster     types.BroadcasterAgent
	requester  types.Requester
	work       taskHandler
	taskType   types.TaskType
	metadata   any
	result     types.TaskResult
	persistant bool

	err          any
	timeout      time.Time
	exitStatus   types.TaskExitStatus // "success", "error" or "cancelled"
	errorCleanup func()

	sw util.Stopwatch

	// signal is used for signaling a task to change behavior after it has been queued,
	// to exit prematurely, for example. The signalChan serves the same purpose, but is
	// used when a task might block waiting for another channel.
	// Key: 1 is exit,
	signal     int
	signalChan chan int

	waitMu *sync.Mutex
}

const (
	TaskSuccess  types.TaskExitStatus = "success"
	TaskCanceled types.TaskExitStatus = "cancelled"
	TaskError    types.TaskExitStatus = "error"
)

const (
	ScanDirectoryTask types.TaskType = "scan_directory"
	ScanFileTask      types.TaskType = "scan_file"
	MoveFileTask      types.TaskType = "move_file"
	WriteFileTask     types.TaskType = "write_file"
	CreateZipTask     types.TaskType = "create_zip"
	GatherFsStatsTask types.TaskType = "gather_filesystem_stats"

	BackupTask types.TaskType = "do_backup"
)

// Worker pool //

type hit struct {
	time   time.Time
	target *task
}

type workChannel chan *task

type hitChannel chan hit

type taskPool struct {
	treatAsGlobal  bool
	hasQueueThread bool
	totalTasks     *atomic.Int64
	completedTasks *atomic.Int64
	waiterCount    *atomic.Int32
	waiterGate     *sync.Mutex
	exitLock       *sync.Mutex
	allQueuedFlag  bool
	createdBy      *task
	workerPool     *WorkerPool
	parentTaskPool *taskPool
}

type WorkerPool struct {
	maxWorkers     *atomic.Int64 // Max allowed worker count
	currentWorkers *atomic.Int64 // Currnet worker count
	busyCount      *atomic.Int64 // Number of workers currently executing a task

	lifetimeQueuedCount *atomic.Int64

	taskStream   workChannel
	taskBufferMu *sync.Mutex
	taskBuffer   []*task
	hitStream    hitChannel

	exitFlag int
}

// Internal types //

type TaskMetadata interface {
	MetaString() string
}

type ScanMetadata struct {
	file         types.WeblensFile
	recursive    bool
	deepScan     bool
	partialMedia types.Media
}

// Override marshal function for ScanMetadata
func (m ScanMetadata) MetaString() string {
	data := map[string]any{
		"FileId":    m.file.Id(),
		"Recursive": m.recursive,
		"Deep":      m.deepScan,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type ZipMetadata struct {
	files    []types.WeblensFile
	username types.Username
	shareId  types.ShareId
}

func (m ZipMetadata) MetaString() string {
	data := map[string]any{
		"Files":    m.files,
		"Username": m.username,
		"ShareId":  m.shareId,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type MoveMeta struct {
	fileId              types.FileId
	destinationFolderId types.FileId
	newFilename         string
}

func (m MoveMeta) MetaString() string {
	data := map[string]any{
		"FileId":      m.fileId,
		"DestId":      m.destinationFolderId,
		"NewFileName": m.newFilename,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type FileChunk struct {
	FileId       types.FileId
	Chunk        []byte
	ContentRange string
}

type WriteFileMeta struct {
	chunkStream  chan FileChunk `json:"-"`
	rootFolderId types.FileId
	chunkSize    int64
	totalSize    int64
}

func (m WriteFileMeta) MetaString() string {
	data := map[string]any{
		"RootFolder": m.rootFolderId,
		"ChunkSize":  m.chunkSize,
		"TotalSize":  m.totalSize,
	}
	bs, err := json.Marshal(data)
	util.ErrTrace(err)

	return string(bs)
}

type FsStatMeta struct {
	rootDir types.WeblensFile
}

func (m FsStatMeta) MetaString() string {
	data := map[string]any{
		"RootFolder": m.rootDir.Id(),
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

// Misc
type KeyVal struct {
	Key string
	Val string
}

var ErrNonDisplayable = errors.New("attempt to process non-displayable file")
var ErrEmptyZip = errors.New("cannot create a zip with no files")
var ErrTaskExit = errors.New("task exit")
var ErrTaskError = errors.New("task generated an error")
var ErrTaskTimeout = errors.New("task timed out")
var ErrBadTaskMetaType = errors.New("task metadata type is not supported on attempted operation")
var ErrBadCaster = errors.New("task was given the wrong type of caster")
