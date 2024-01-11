package dataProcess

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

func ProcessMediaFile(file *dataStore.WeblensFileDescriptor, m *dataStore.Media, db *dataStore.Weblensdb) error {
	defer util.RecoverPanic("Panic caught while processing media file:", file.String())

	if m == nil {
		return errors.New("attempted to process nil media")
	}

	defer m.Clean()

	var parseAnyway bool = false
	filled, _ := m.IsFilledOut(false)
	if filled && !parseAnyway {
		// util.Warning.Println("Tried to process media file that already exists in the database")
		return nil
	}

	m.FileId = file.Id()

	if m.MediaType.FriendlyName == "" {
		err := m.ComputeExif()
		util.FailOnError(err, "Failed to extract exif data")
	}

	m.FileHash = util.HashOfString(8, file.String())
	// Files that are not "media" (jpeg, png, mov, etc.) should not be stored in the media database
	if (!m.MediaType.IsDisplayable) {
		PushItemCreate(file)
		return nil
	}

	thumb := m.GenerateThumbnail()
	if m.BlurHash == "" {
		m.GenerateBlurhash(thumb)
	}

	if (m.Owner == "") {
		m.Owner = file.Owner()
	}

	db.DbAddMedia(m)

	PushItemCreate(file)

	return nil
}

func ScanDirectory(t *task) {
	meta := t.metadata.(ScanMetadata)
	scanDir := meta.File
	recursive := meta.Recursive
	username := meta.Username
	deepScan := meta.DeepScan

	util.Info.Printf("Beginning directory scan (recursive: %t, deep: %t): %s\n", recursive, deepScan, scanDir.String())

	mediaToScan := []*dataStore.WeblensFileDescriptor{}
	dirsToScan := []*dataStore.WeblensFileDescriptor{}

	if deepScan {
		scanDir.ReadDir()
	}

	children := scanDir.GetChildren()
	for _, c := range children {
		if !c.IsDir() {
			if m, _ := c.GetMedia(); m == nil && c.IsDisplayable() {
				mediaToScan = append(mediaToScan, c)
			}
		} else if recursive && c != scanDir {
			dirsToScan = append(dirsToScan, c)
		}
	}

	for _, dir := range dirsToScan {
		t := NewTask("scan_directory", ScanMetadata{File: dir, Username: username, Recursive: recursive, DeepScan: deepScan})
		ttInstance.globalQueue.QueueTask(t)
	}

	if len(mediaToScan) == 0 {
		t.BroadcastComplete("scan_complete")
		util.Info.Printf("No media to scan: %s", scanDir.String())
		return
	}

	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"))
	if err != nil {
		util.DisplayError(err)
		t.err = err
	}

	start := time.Now()
	filesStrings := util.Map(mediaToScan, func(f *dataStore.WeblensFileDescriptor) string {return f.String()})
	allMeta := et.ExtractMetadata(filesStrings...)
	thumbsReader := readRawThumbBytesFromDir(util.Map(allMeta, func(meta exiftool.FileMetadata) string {return meta.File})...)

	wq := NewWorkQueue()

	offset := 0
	for i, meta := range allMeta {
		file := mediaToScan[i]
		m := &dataStore.Media{}
		m.FileId = file.Id()

		rawJpegLen := meta.Fields["JpgFromRawLength"]
		if rawJpegLen != nil && thumbsReader != nil {
			thumbLen := int(rawJpegLen.(float64))
			m.DumpThumbBytes(thumbsReader[offset:offset+thumbLen])
			offset += thumbLen
		}

		err := file.SetMedia(m)
		if err != nil {
			util.DisplayError(err)
			continue
		}

		m.DumpRawExif(meta.Fields)
		t := NewTask("scan_file", ScanMetadata{File: file, Username: username, PartialMedia: m})
		wq.QueueTask(t)
	}

	wq.AllQueued()
	util.Info.Println("Pre-import meta collection", time.Since(start))

	// start = time.Now()
	wq.Wait()
	t.BroadcastComplete("scan_complete")
	// t.Complete(fmt.Sprintf("Completed scanning %d images in %v", len(allMeta), time.Since(start)))
	runtime.GC()

	PushItemUpdate(scanDir, scanDir)
}

func readRawThumbBytesFromDir(paths ...string) []byte {
	if len(paths) == 0 {
		return nil
	}

	allPathsStr := strings.Join(util.Map(paths, func(path string) string {return fmt.Sprintf("'%s'", path)}), " ")
	cmdString := fmt.Sprintf("exiftool -a -b -JpgFromRaw %s", allPathsStr)
	cmd := exec.Command("/bin/bash", "-c", cmdString)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		util.Warning.Printf("Failed to bulk read thumbnails: %s %s", err, stderr.String())
		return nil
	}

	bytes := out.Bytes()
	return bytes
}