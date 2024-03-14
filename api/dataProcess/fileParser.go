package dataProcess

import (
	"errors"
	"slices"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
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
		util.DisplayError(err)
		gexift = nil
		gexiftBufferSize = 0
	}
	gbuf := make([]byte, int(bufSize))
	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"), exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(gbuf, int(bufSize)))
	if err != nil {
		util.DisplayError(err)
		return nil
	}
	gexift = et

	return gexift
}

func processMediaFile(t *task) {
	meta := t.metadata.(ScanMetadata)
	m := meta.PartialMedia
	file := meta.File

	defer file.RemoveTask(t.TaskId())

	if m == nil {
		t.error(errors.New("attempted to process nil media"))
		return
	}

	d, err := file.IsDisplayable()
	if err != nil && err != dataStore.ErrNoMedia {
		t.error(err)
		return
	}
	if !d {
		return
	}

	defer m.Clean()

	m, err = m.LoadFromFile(file, t)
	if err != nil {
		t.error(err)
		return
	}

	err = m.WriteToDb()
	if err != nil {
		t.error(err)
		return
	}
	if !m.IsImported() {
		m.SetImported()
	}

	globalCaster.PushFileUpdate(file)
	t.success()
}

func scanDirectory(t *task) {
	meta := t.metadata.(ScanMetadata)
	scanDir := meta.File

	if scanDir.Filename() == ".user_trash" {
		globalCaster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
		t.success()
		return
	}

	recursive := meta.Recursive
	deepScan := meta.DeepScan

	mediaToScan := []*dataStore.WeblensFile{}
	dirsToScan := []*dataStore.WeblensFile{}

	// "deepScan" defines wether or not we should go sync
	// with the real filesystem for this scan. Otherwise,
	// just handle processing media we already know about
	if deepScan {
		scanDir.ReadDir()
	}

	children := scanDir.GetChildren()
	for _, c := range children {
		// If this file is already being procecced, don't queue it again
		if slices.ContainsFunc(c.GetTasks(), func(t dataStore.Task) bool { return t.TaskType() == "scan_file" || t.TaskType() == "scan_directory" }) {
			continue
		}
		if !c.IsDir() {
			m, err := c.GetMedia()
			if err != nil && err != dataStore.ErrNoMedia {
				util.DisplayError(err)
				continue
			}

			d, err := c.IsDisplayable()
			if err != nil && err != dataStore.ErrNoMedia {
				util.DisplayError(err)
				continue
			}

			if m == nil {
				mediaToScan = append(mediaToScan, c)
				continue
			}

			if !m.IsImported() && d {
				mediaToScan = append(mediaToScan, c)
				continue
			}

			filled, _ := m.IsFilledOut()
			if !filled && d {
				mediaToScan = append(mediaToScan, c)
				continue
			}

		} else if recursive && c != scanDir {
			dirsToScan = append(dirsToScan, c)
		}
	}

	for _, dir := range dirsToScan {
		GetGlobalQueue().ScanDirectory(dir, recursive, deepScan, t.caster)
	}

	if len(mediaToScan) == 0 {
		globalCaster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
		t.success("No media to scan: ", scanDir.String())
		return
	}

	// Wait on all tasks on files as not to be destructive
	util.Each(mediaToScan, func(wf *dataStore.WeblensFile) { util.Each(wf.GetTasks(), func(t dataStore.Task) { t.Wait() }) })

	util.Info.Printf("Beginning directory scan for %s [M %d][R %t][D %t]\n", scanDir.String(), len(mediaToScan), recursive, deepScan)

	t.SwLap("Pre-scan complete")

	wq := NewWorkQueue()

	slices.SortFunc(mediaToScan, func(a, b *dataStore.WeblensFile) int { return strings.Compare(a.Filename(), b.Filename()) })

	for _, file := range mediaToScan {
		m, err := file.GetMedia()
		if err != nil && err != dataStore.ErrNoMedia {
			util.DisplayError(err)
			continue
		}
		wq.ScanFile(file, m, t.caster)
	}

	t.SwLap("Queued all tasks")

	wq.SignalAllQueued()
	wq.Wait(true) // Park this thread and wait for the media to finish computing

	t.SwLap("All sub-scans finished")

	globalCaster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
	t.success()
}
