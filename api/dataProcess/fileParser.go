package dataProcess

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

// Global exiftool
var gexift *exiftool.Exiftool

func InitGExif(bufSize int64) *exiftool.Exiftool {
	gbuf := make([]byte, int(bufSize))
	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"), exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(gbuf, int(bufSize)))
	if err != nil {
		util.DisplayError(err)
		return nil
	}
	gexift = et

	return gexift
}

func processMediaFile(t *Task) {
	meta := t.metadata.(ScanMetadata)
	m := meta.PartialMedia
	file := meta.File

	defer file.RemoveTask(t.TaskId())

	if m == nil {
		t.error(errors.New("attempted to process nil media"))
		return
	}

	defer m.Clean()

	var parseAnyway bool = false
	filled, _ := m.IsFilledOut(false)
	if filled && !parseAnyway {
		util.Warning.Println("Tried to process media file that already exists in the database")
		t.success()
		return
	}

	m.FileId = file.Id()

	if m.MediaType == nil {
		err := m.ComputeExif()
		util.FailOnError(err, "Failed to extract exif data")
	}

	m.Id()

	// Files that are not "media" (jpeg, png, mov, etc.) should not be stored in the media database
	if m.MediaType == nil || !m.MediaType.IsDisplayable {
		caster.PushFileUpdate(file)
		t.success()
		return
	}

	thumb, err := m.GenerateThumbnail()
	if err != nil {
		t.error(err)
		return
	}

	if m.BlurHash == "" {
		m.GenerateBlurhash(thumb)
	}

	if m.Owner == "" {
		m.Owner = file.Owner()
	}

	db := dataStore.NewDB()

	err = db.AddMedia(m)
	if err != nil {
		t.error(err)
		return
	}

	m.SetImported()
	file.SetMedia(m)

	caster.PushFileUpdate(file)
	t.success()

}

func scanDirectory(t *Task) {
	meta := t.metadata.(ScanMetadata)
	scanDir := meta.File

	if scanDir.Filename() == ".user_trash" {
		caster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
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
			if err != nil && !errors.Is(err, dataStore.ErrNoMedia) {
				util.DisplayError(err)
				continue
			}

			if !m.IsImported() {
				mediaToScan = append(mediaToScan, c)
			}
		} else if recursive && c != scanDir {
			dirsToScan = append(dirsToScan, c)
		}
	}

	for _, dir := range dirsToScan {
		GetGlobalQueue().ScanDirectory(dir, recursive, deepScan)
	}

	if len(mediaToScan) == 0 {
		caster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
		t.success("No media to scan:", scanDir.String())
		return
	}

	// Wait on all tasks on files as not to be destructive
	util.Each(mediaToScan, func(wf *dataStore.WeblensFile) { util.Each(wf.GetTasks(), func(t dataStore.Task) { t.Wait() }) })

	util.Info.Printf("Beginning directory scan for %s [M %d][R %t][D %t]\n", scanDir.String(), len(mediaToScan), recursive, deepScan)

	sw := util.NewStopwatch(fmt.Sprint("Directory scan ", scanDir.String()))

	for _, file := range mediaToScan {
		newMedia := &dataStore.Media{}
		newMedia.FileId = file.Id()
		file.SetMedia(newMedia)
	}

	sw.Lap("Loaded medias")

	wq := NewWorkQueue()

	slices.SortFunc(mediaToScan, func(a, b *dataStore.WeblensFile) int { return strings.Compare(a.Filename(), b.Filename()) })

	for _, file := range mediaToScan {
		m, err := file.GetMedia()
		if err != nil {
			util.DisplayError(err)
			continue
		}
		wq.ScanFile(file, m)
	}

	sw.Lap("Queued all tasks")

	wq.SignalAllQueued()
	wq.Wait(true) // Park this thread and wait for the media to finish computing

	sw.Lap("All tasks finished")
	sw.Stop()
	sw.PrintResults()

	caster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
	t.success()
}
