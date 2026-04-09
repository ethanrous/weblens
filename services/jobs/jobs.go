package jobs

import (
	"maps"
	"strconv"
	"strings"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	job_model "github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
)

func parseRangeHeader(contentRange string) (rangeMin, rangeMax, total int64, err error) {
	contentRange = strings.TrimPrefix(contentRange, "bytes=")

	rangeAndSize := strings.Split(contentRange, "/")
	rangeParts := strings.Split(rangeAndSize[0], "-")

	rangeMin, err = strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return
	}

	rangeMax, err = strconv.ParseInt(rangeParts[1], 10, 64)
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

// GatherFilesystemStats collects statistics about file sizes grouped by extension in a directory tree.
func GatherFilesystemStats(tsk *task.Task) {
	meta := tsk.GetMeta().(job.FsStatMeta)

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
		tsk.ReqNoErr(err)
	}

	returned := make([]extSize, 0, len(filetypeSizeMap))
	for name, value := range maps.All(filetypeSizeMap) {
		returned = append(returned, extSize{Name: name, Value: value})
	}

	// freeSpace := dataStore.GetFreeSpace(meta.rootDir.GetAbsPath())
	freeSpace := 0

	tsk.SetResult(task.Result{"sizesByExtension": returned, "bytesFree": freeSpace})
	tsk.Success()
}

// RegisterJobs registers all available job handlers with the worker pool.
func RegisterJobs(workerPool *task.WorkerPool) {
	workerPool.RegisterJob(job_model.ScanDirectoryTask, ScanDirectory)
	workerPool.RegisterJob(job_model.ScanFileTask, ScanFile)
	workerPool.RegisterJob(job_model.UploadFilesTask, HandleFileUploads, task.Options{Persistent: true, Unique: true})
	workerPool.RegisterJob(job_model.CreateZipTask, CreateZip, task.Options{Persistent: true, Unique: false})
	workerPool.RegisterJob(job_model.GatherFsStatsTask, GatherFilesystemStats)
	workerPool.RegisterJob(job_model.BackupTask, DoBackup)
	workerPool.RegisterJob(job_model.CopyFileFromCoreTask, CopyFileFromCore)
	workerPool.RegisterJob(job_model.RestoreCoreTask, RestoreCore)
	workerPool.RegisterJob(job_model.LoadFilesystemTask, LoadAtPath)
}
