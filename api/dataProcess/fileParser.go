package dataProcess

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

func ProcessMediaFile(file *dataStore.WeblensFileDescriptor, db *dataStore.Weblensdb) error {
	defer util.RecoverPanic("Panic caught while processing media file:", file.String())

	m, _ := db.GetMediaByFile(file, true)

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
		err := m.ExtractExif()
		util.FailOnError(err, "Failed to extract exif data")
	}

	m.GenerateFileHash(file)

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

	db.DbAddMedia(&m)

	PushItemCreate(file)

	return nil
}

func ScanDirectory(scanDir, username string, recursive bool) {
	absoluteScanDir := dataStore.GuaranteeAbsolutePath(scanDir)
	util.Debug.Printf("Beginning directory scan (recursive: %t): %s\n", recursive, absoluteScanDir)

	db := dataStore.NewDB(username)
	ms := db.GetMediaInDirectory(absoluteScanDir, recursive)
	mediaMap := map[string]bool{}
	for _, m := range ms {
		mFile := m.GetBackingFile()
		if mFile.Err() != nil {
			util.DisplayError(mFile.Err())
			continue
		}
		mediaMap[mFile.String()] = true
	}

	if recursive {
		filepath.WalkDir(absoluteScanDir, func(path string, d fs.DirEntry,  err error) error {
			file := dataStore.WFDByPath(path)
			util.FailOnError(file.Err(), "")
			if !(file.IsDir() || strings.HasSuffix(path, ".DS_Store") || strings.HasSuffix(path, ".thumb.jpeg") || mediaMap[path]) {
				RequestTask("scan_file", ScanMetadata{File: file, Username: username})
			}
			return nil
		})
	} else {
		files, err := os.ReadDir(absoluteScanDir)
		if err != nil {
			panic(err)
		}
		for _, d := range files {
			file := dataStore.WFDByPath(filepath.Join(absoluteScanDir, d.Name()))
			util.FailOnError(file.Err(), "")
			if (!file.IsDir() && file.Filename != ".DS_Store" && !strings.HasSuffix(d.Name(), ".thumb.jpeg") && !mediaMap[file.String()]) {
				RequestTask("scan_file", ScanMetadata{File: file, Username: username})
			}
		}
	}

	for _, m := range ms {
		mFile := m.GetBackingFile()
		if mFile.Err() != nil {
			util.DisplayError(mFile.Err())
			continue
		}
		if !mFile.Exists() {
			db.RemoveMediaByFilepath(m.ParentFolder, m.Filename)
		}
	}

	userPath, _ := dataStore.GuaranteeUserRelativePath(scanDir, username)
	if userPath == "/" && recursive {
		trashedFiles := db.GetTrashedFiles()
		for _, trashed := range trashedFiles {
			_, err := os.Stat(dataStore.GuaranteeAbsolutePath(trashed.TrashPath))
			if errors.Is(err, os.ErrNotExist) {
				db.RemoveTrashEntry(trashed)
			}
		}
	}

}