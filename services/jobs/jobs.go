package jobs

import (
	"maps"
	"strconv"
	"strings"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	job_model "github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/errors"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/rs/zerolog"
)

func parseRangeHeader(contentRange string) (min, max, total int64, err error) {
	contentRange = strings.TrimPrefix(contentRange, "bytes=")

	rangeAndSize := strings.Split(contentRange, "/")
	rangeParts := strings.Split(rangeAndSize[0], "-")

	min, err = strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return
	}

	max, err = strconv.ParseInt(rangeParts[1], 10, 64)
	if err != nil {
		return
	}

	total, err = strconv.ParseInt(rangeAndSize[1], 10, 64)
	if err != nil {
		return
	}

	return
}

type extSize struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func GatherFilesystemStats(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	meta := t.GetMeta().(job.FsStatMeta)

	filetypeSizeMap := map[string]int64{}
	folderCount := 0

	// media := dataStore.GetMediaDir()
	// external := dataStore.GetExternalDir()
	// dataStore.ResizeDown(media)

	sizeFunc := func(wf *file_model.WeblensFileImpl) error {
		if wf.IsDir() {
			folderCount++

			return nil
		}
		filename := wf.GetPortablePath().Filename()
		index := strings.LastIndex(filename, ".")
		size := wf.Size()
		if index == -1 {
			filetypeSizeMap["other"] += size
		} else {
			filetypeSizeMap[filename[index+1:]] += size
		}

		return nil
	}

	err := meta.RootDir.RecursiveMap(sizeFunc)
	if err != nil {
		t.ReqNoErr(err)
	}

	returned := make([]extSize, 0, len(filetypeSizeMap))
	for name, value := range maps.All(filetypeSizeMap) {
		returned = append(returned, extSize{Name: name, Value: value})
	}

	// freeSpace := dataStore.GetFreeSpace(meta.rootDir.GetAbsPath())
	freeSpace := 0

	t.SetResult(task_mod.TaskResult{"sizesByExtension": returned, "bytesFree": freeSpace})
	t.Success()
}

func HashFile(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	meta := t.GetMeta().(job.HashFileMeta)

	contentId, err := file_model.GenerateContentId(t.Ctx, meta.File)
	t.ReqNoErr(err)

	if contentId == "" && meta.File.Size() != 0 {
		t.Fail(file_model.ErrNoContentId)
	}

	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Hashed file [%s] to [%s]", meta.File.GetPortablePath(), contentId) })

	// TODO - sync database content id if this file is created before being added to db (i.e upload)
	// err = dataStore.SetContentId(meta.file, contentId)
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }

	t.SetResult(task_mod.TaskResult{"contentId": contentId})

	poolStatus := t.GetTaskPool().Status()
	notif := notify.NewTaskNotification(
		t, websocket.TaskCompleteEvent, task_mod.TaskResult{
			"filename":      meta.File.GetPortablePath().Filename(),
			"tasksTotal":    poolStatus.Total,
			"tasksComplete": poolStatus.Complete,
		},
	)
	appCtx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(errors.New("failed to get context"))

		return
	}

	appCtx.Notify(t.Ctx, notif)

	t.Success()
}

func RegisterJobs(workerPool *task.WorkerPool) {
	workerPool.RegisterJob(job_model.ScanDirectoryTask, ScanDirectory)
	workerPool.RegisterJob(job_model.ScanFileTask, ScanFile)
	workerPool.RegisterJob(job_model.UploadFilesTask, HandleFileUploads, task.TaskOptions{Persistent: true, Unique: true})
	workerPool.RegisterJob(job_model.CreateZipTask, CreateZip, task.TaskOptions{Persistent: true, Unique: false})
	workerPool.RegisterJob(job_model.GatherFsStatsTask, GatherFilesystemStats)
	workerPool.RegisterJob(job_model.BackupTask, DoBackup)
	workerPool.RegisterJob(job_model.CopyFileFromCoreTask, CopyFileFromCore)
	workerPool.RegisterJob(job_model.RestoreCoreTask, RestoreCore)
	workerPool.RegisterJob(job_model.HashFileTask, HashFile)
	workerPool.RegisterJob(job_model.LoadFilesystemTask, LoadAtPath)
}
