package dataProcess

import (
	"errors"
	"slices"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

// Global exiftool
var gexift *exiftool.Exiftool
var gexiftBufferSize int64

func InitGExif(bufSize int64) *exiftool.Exiftool {
	if bufSize <= gexiftBufferSize {
		return gexift
	}
	if gexift != nil {
		err := gexift.Close()
		util.ErrTrace(err)
		gexift = nil
		gexiftBufferSize = 0
	}
	gbuf := make([]byte, int(bufSize))
	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"), exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(gbuf, int(bufSize)))
	if err != nil {
		util.ErrTrace(err)
		return nil
	}
	gexift = et

	return gexift
}

func processMediaFile(t *task) {
	meta := t.metadata.(ScanMetadata)
	m := meta.partialMedia
	file := meta.file

	file.AddTask(t)
	defer file.RemoveTask(t.TaskId())

	if m == nil {
		t.ErrorAndExit(errors.New("attempted to process nil media"))
		return
	}

	d, err := file.IsDisplayable()
	if err != nil && err != dataStore.ErrNoMedia {
		t.ErrorAndExit(err)
		return
	}
	if !d {
		return
	}

	defer m.Clean()

	m, err = m.LoadFromFile(file, t)
	if err != nil {
		t.ErrorAndExit(err)
		return
	}

	t.CheckExit()

	err = m.Save()
	if err != nil {
		t.ErrorAndExit(err)
		return
	}
	m.SetImported(true)

	t.CheckExit()

	t.caster.PushFileUpdate(file)
	t.taskPool.NotifyTaskComplete(t, t.caster)
	if t.caster.IsBuffered() {
		t.caster.(types.BufferedBroadcasterAgent).Flush()
	}
	t.success()
}

func scanDirectory(t *task) {
	meta := t.metadata.(ScanMetadata)
	scanDir := meta.file

	if scanDir.Filename() == ".user_trash" {
		t.taskPool.NotifyTaskComplete(t, t.caster, "No media to scan")
		globalCaster.PushTaskUpdate(t.taskId, "scan_complete", types.TaskResult{"execution_time": t.ExeTime()}) // Let any client subscribers know we are done
		t.success()
		return
	}

	// "deepScan" defines wether or not we should go sync
	// with the real filesystem for this scan. Otherwise,
	// just handle processing media we already know about
	if meta.deepScan {
		scanDir.ReadDir()
	}

	// Wait for all tasks on this directory to be finished
	// before getting children. Avoids issues with, say, scaning a directory
	// while we are uploading into it. We might grab only the first half of
	// the files that are uploaded, and lose the scan on the second half of the files.
	// Upload tasks attach themselves to the parent dir, so we would see that here.
	for {
		tasks := scanDir.GetTasks()
		if len(tasks) == 0 {
			break
		}
		util.Warning.Println("Waiting on tasks before scanning directory!")
		if tasks[len(tasks)-1] == t {
			util.Error.Println("BAD BAD! WAITING ON CURRENT TASK! THIS IS A DEADLOCK! WEBLENS IS DEAD!")
		}
		tasks[len(tasks)-1].Wait()
	}
	scanDir.AddTask(t)
	defer scanDir.RemoveTask(t.TaskId())

	mediaToScan := []types.WeblensFile{}
	dirsToScan := []types.WeblensFile{}

	children := scanDir.GetChildren()
	for _, c := range children {
		// If this file is already being procecced, don't queue it again
		if slices.ContainsFunc(c.GetTasks(), func(t types.Task) bool { return t.TaskType() == "scan_file" || t.TaskType() == "scan_directory" }) {
			continue
		}
		if !c.IsDir() {
			d, err := c.IsDisplayable()
			if err != nil && err != dataStore.ErrNoMedia {
				util.ErrTrace(err)
				continue
			}
			if !d {
				continue
			}

			m, err := c.GetMedia()
			if err != nil && err != dataStore.ErrNoMedia {
				util.ErrTrace(err)
				continue
			}
			if m == nil || !m.IsImported() {
				mediaToScan = append(mediaToScan, c)
				continue
			}

			filled, _ := m.IsFilledOut()
			if !filled {
				mediaToScan = append(mediaToScan, c)
				continue
			}

		} else if meta.recursive && c.Owner() != dataStore.WEBLENS_ROOT_USER && c != scanDir {
			dirsToScan = append(dirsToScan, c)
		}
	}

	if len(mediaToScan) == 0 && len(dirsToScan) == 0 {
		t.success("No media to scan: ", scanDir.GetAbsPath())
		t.taskPool.NotifyTaskComplete(t, t.caster, "No media to scan")
		globalCaster.PushTaskUpdate(t.taskId, "scan_complete", types.TaskResult{"execution_time": t.ExeTime()}) // Let any client subscribers know we are done
		return
	}

	tp := NewTaskPool(true, t)
	util.Info.Printf("Beginning directory scan for %s [M %d][R %t][D %t]\n", scanDir.GetAbsPath(), len(mediaToScan), meta.recursive, meta.deepScan)

	for _, dir := range dirsToScan {
		tp.ScanDirectory(dir, meta.recursive, meta.deepScan, t.caster)
	}

	if len(mediaToScan) != 0 {
		// Wait on all tasks on files as not to be destructive
		util.Each(mediaToScan, func(wf types.WeblensFile) {
			util.Each(wf.GetTasks(), func(fTask types.Task) {
				if fTask.TaskId() != t.TaskId() {
					fTask.Wait()
				}
			})
		})

		t.SwLap("Pre-scan complete")

		slices.SortFunc(mediaToScan, func(a, b types.WeblensFile) int { return strings.Compare(a.Filename(), b.Filename()) })

		for _, file := range mediaToScan {
			m, err := file.GetMedia()
			if err != nil && err != dataStore.ErrNoMedia {
				util.ErrTrace(err)
				continue
			}
			tp.ScanFile(file, m, t.caster)
		}
	}

	t.SwLap("Queued all tasks")
	tp.SignalAllQueued()
	tp.Wait(true) // Park this thread and wait for the media to finish computing

	t.SwLap("All sub-scans finished")

	globalCaster.PushTaskUpdate(t.taskId, "scan_complete", types.TaskResult{"execution_time": t.ExeTime()}) // Let any client subscribers know we are done
	t.success()
}

func getScanResult(t *task) types.TaskResult {
	var tp *virtualTaskPool

	if t.taskPool != nil {
		tp = t.taskPool.GetRootPool()
	}

	result := types.TaskResult{
		"filename": t.metadata.(ScanMetadata).file.Filename(),
	}

	if tp != nil {
		complete, total, progress := tp.status()
		result["percent_progress"] = progress
		result["tasks_complete"] = complete
		result["tasks_total"] = total
		result["task_job_name"] = tp.createdBy.TaskType()
		result["task_job_target"] = tp.createdBy.metadata.(ScanMetadata).file.Filename()
	} else {
		result["task_job_name"] = t.TaskType()
		result["task_job_target"] = t.metadata.(ScanMetadata).file.Filename()
	}

	return result
}
