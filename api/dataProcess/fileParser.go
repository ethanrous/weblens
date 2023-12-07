package dataProcess

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

func ProcessMediaFile(file *dataStore.WeblensFileDescriptor, m *dataStore.Media, db *dataStore.Weblensdb) error {
	defer util.RecoverPanic("Panic caught while processing media file:", file.String())

	var parseAnyway bool = false
	filled, _ := m.IsFilledOut(false)
	if filled && !parseAnyway {
		// util.Warning.Println("Tried to process media file that already exists in the database")
		return nil
	}

	if m.ParentFolder == "" {
		m.ParentFolder = file.ParentFolderId
	}

	if m.Filename == "" {
		m.Filename = file.Filename
	}

	if m.MediaType.FriendlyName == "" {
		err := m.ComputeExif()
		util.FailOnError(err, "Failed to extract exif data")
	}

	m.FileHash = util.HashOfString(8, file.String())
	// m.GenerateFileHash(file)

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

	util.Debug.Printf("Beginning directory scan (recursive: %t): %s\n", recursive, scanDir.String())

	db := dataStore.NewDB(username)
	ms := db.GetMediaInDirectory(scanDir.String(), recursive)
	mediaMap := map[string]bool{}
	for _, m := range ms {
		mFile := m.GetBackingFile()
		if mFile.Err() != nil {
			util.DisplayError(mFile.Err())
			continue
		}
		if !mFile.Exists() {
			db.RemoveMediaByFilepath(m.ParentFolder, m.Filename)
			continue
		}
		mediaMap[mFile.String()] = true
	}

	files, err := os.ReadDir(scanDir.String())
	if err != nil {
		panic(err)
	}

	newFilesMap := map[string]*dataStore.WeblensFileDescriptor{}

	NewWorkSubQueue(t.TaskId)

	for _, d := range files {
		file := scanDir.JoinStr(d.Name())
		util.FailOnError(file.Err(), "")
		if recursive && file.IsDir() {
			RequestTask("scan_directory", "", ScanMetadata{File: file, Username: username, Recursive: recursive})
		} else if (file.Filename != ".DS_Store" && !strings.HasSuffix(d.Name(), ".thumb.jpeg") && !mediaMap[file.String()] && !file.IsDir()) {
			if (!file.IsDisplayable()) {
				RequestTask("scan_file", t.TaskId, ScanMetadata{File: file, Username: username})
			} else {
				_, err := file.GetMedia()
				if err != nil {
					newFilesMap[file.String()] = file
				}
			}
		}
	}

	et, err := exiftool.NewExiftool()
	if err != nil {
		panic(err)
	}

	start := time.Now()
	allMeta := et.ExtractMetadata(util.MapToSlice(newFilesMap, func(absPath string, f *dataStore.WeblensFileDescriptor) string {return absPath})...)
	thumbsReader := readRawThumbBytesFromDir(util.Map(allMeta, func(meta exiftool.FileMetadata) string {return meta.File})...)
	offset := 0
	for _, meta := range allMeta {
		file := newFilesMap[meta.File]
		m := dataStore.Media{}
		m.Filename = file.Filename
		m.ParentFolder = file.ParentFolderId

		rawJpegLen := meta.Fields["JpgFromRawLength"]
		if rawJpegLen != nil {
			thumbLen := int(rawJpegLen.(float64))
			m.DumpThumbBytes(thumbsReader[offset:offset+thumbLen])
			offset += thumbLen
		}

		m.DumpRawExif(meta.Fields)
		RequestTask("scan_file", t.TaskId, ScanMetadata{File: file, Username: username, PartialMedia: &m})
	}

	MainNotifyAllQueued(t.TaskId)
	util.Debug.Println("Pre-import meta collection", time.Since(start))

	start = time.Now()
	MainWorkQueueWait(t.TaskId)
	util.Debug.Printf("Completed %d scanning images in %v", len(allMeta), time.Since(start))
}

func readRawThumbBytesFromDir(paths ...string) []byte {
	paths = util.Map(paths, func(path string) string {return strings.ReplaceAll(path, " ", "\\ ")})
	allPathsStr := strings.Join(paths, " ")
	cmdString := fmt.Sprintf("exiftool -a -b -JpgFromRaw %s", allPathsStr)
	cmd := exec.Command("/bin/bash", "-c", cmdString)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		util.FailOnError(err, "Failed to run exiftool extract command")
	}

	bytes := out.Bytes()
	return bytes
}