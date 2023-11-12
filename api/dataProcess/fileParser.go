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

func HandleNewImage(path, owner string, db dataStore.Weblensdb) (*dataStore.Media, error) {
	defer func() {
		if err := recover(); err != nil {
			util.Error.Printf("Recovered panic while parsing new file (%s): %v", path, err)
		}
	}()

	m, _ := db.GetMediaByFilepath(path, true)

	var parseAnyway bool = false
	filled, _ := m.IsFilledOut(false)
	if filled && !parseAnyway {
		return &m, nil
	}

	if m.Filepath == "" {
		m.Filepath = path
	}

	if m.MediaType.FriendlyName == "" {
		err := m.ExtractExif()
		if err != nil {
			return nil, err
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

	return &m, nil
}

func middleware(path, username string, d fs.DirEntry, wp WorkerPool, db dataStore.Weblensdb) error {
	if d.IsDir() {
		return nil
	}
	if strings.HasSuffix(path, ".DS_Store") || strings.HasSuffix(path, ".thumb.jpeg") {
		return nil
	}

	wp.AddTask(func() {
		_, err:= HandleNewImage(path, username ,db)
		util.DisplayError(err, "Error attempting to import new media")
	})

	return nil
}

func ScanDirectory(scanDir, username string, recursive bool) (WorkerPool) {
	absoluteScanDir := dataStore.GuaranteeAbsolutePath(scanDir, username)
	util.Debug.Printf("Beginning directory scan (recursive: %t): %s\n", recursive, absoluteScanDir)

	db := dataStore.NewDB(username)

	wp := NewWorkerPool(10)
	wp.Run()

	if recursive {
		filepath.WalkDir(absoluteScanDir, func(path string, d fs.DirEntry,  err error) error {
			e := middleware(path, username, d, wp, db)
			return e
		})
	} else {
		files, err := os.ReadDir(absoluteScanDir)
		if err != nil {
			panic(err)
		}
		for _, d := range files {
			middleware(filepath.Join(absoluteScanDir, d.Name()), username, d, wp, db)
		}
	}

	ms := db.GetMediaInDirectory(absoluteScanDir, recursive)
	util.Debug.Println(absoluteScanDir)


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