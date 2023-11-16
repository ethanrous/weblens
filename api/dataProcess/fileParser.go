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

func ProcessMediaFile(path, owner string, db dataStore.Weblensdb) (error) {
	defer func() {
		if err := recover(); err != nil {
			util.Error.Printf("Recovered panic while parsing new file (%s): %v", path, err)
		}
	}()

	m, _ := db.GetMediaByFilepath(path, true)

	var parseAnyway bool = false
	filled, _ := m.IsFilledOut(false)
	if filled && !parseAnyway {
		return nil
	}

	if m.Filepath == "" {
		m.Filepath = path
	}

	if m.MediaType.FriendlyName == "" {
		err := m.ExtractExif()
		if err != nil {
			return err
		}
	}

	if m.MediaType.FriendlyName != "File" {
		thumb := m.GenerateThumbnail()
		if m.BlurHash == "" {
			m.GenerateBlurhash(thumb)
		}
	}

	m.GenerateFileHash()

	if (m.Owner == "") {
		m.Owner = owner
	}

	// util.Debug.Println("Finished parse of file: ", filepath)

	db.DbAddMedia(&m)

	PushItemUpdate(dataStore.GuaranteeUserAbsolutePath(m.Filepath, owner), owner, db)

	return nil
}

func ScanDirectory(scanDir, username string, recursive bool) (WorkerPool) {
	absoluteScanDir := dataStore.GuaranteeAbsolutePath(scanDir, username)
	util.Debug.Printf("Beginning directory scan (recursive: %t): %s\n", recursive, absoluteScanDir)

	db := dataStore.NewDB(username)
	ms := db.GetMediaInDirectory(absoluteScanDir, recursive)
	mediaMap := map[string]bool{}
	for _, m := range ms {
		mediaMap[dataStore.GuaranteeUserAbsolutePath(m.Filepath, username)] = true
	}

	wp := NewWorkerPool(10)
	wp.Run()

	if recursive {
		filepath.WalkDir(absoluteScanDir, func(path string, d fs.DirEntry,  err error) error {
			if !(d.IsDir() || strings.HasSuffix(path, ".DS_Store") || strings.HasSuffix(path, ".thumb.jpeg") || mediaMap[path]) {
				RequestTask("scan_file", ScanMetadata{Path: path, Username: username})
			}
			return nil
		})
	} else {
		files, err := os.ReadDir(absoluteScanDir)
		if err != nil {
			panic(err)
		}
		for _, d := range files {
			tmpPath := filepath.Join(absoluteScanDir, d.Name())
			if !(d.IsDir() || strings.HasSuffix(tmpPath, ".DS_Store") || strings.HasSuffix(tmpPath, ".thumb.jpeg") || mediaMap[tmpPath]) {
				RequestTask("scan_file", ScanMetadata{Path: tmpPath, Username: username})
			}
		}
	}

	for _, m := range ms {
		_, err := os.Stat(dataStore.GuaranteeUserAbsolutePath(m.Filepath, username))
		if errors.Is(err, os.ErrNotExist) {
			util.Warning.Println("ERR: ", err)
			util.Debug.Println("Remove: ", m.Filepath)
			db.RemoveMediaByFilepath(m.Filepath)
		}
	}

	userPath, _ := dataStore.GuaranteeUserRelativePath(scanDir, username)
	if userPath == "/" && recursive {
		trashedFiles := db.GetTrashedFiles()
		for _, trashed := range trashedFiles {
			_, err := os.Stat(dataStore.GuaranteeAbsolutePath(trashed.TrashPath, username))
			if errors.Is(err, os.ErrNotExist) {
				db.RemoveTrashEntry(trashed)
			}
		}
	}

	return wp

}