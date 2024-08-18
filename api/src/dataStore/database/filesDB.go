package database

import (
	"io"
	"os"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) NewTrashEntry(te types.TrashEntry) error {
	_, err := db.trash.InsertOne(db.ctx, te)
	return err
}

func (db *databaseService) DeleteTrashEntry(fileId types.FileId) error {
	res, err := db.trash.DeleteOne(db.ctx, bson.M{"trashFileId": fileId})
	if res.DeletedCount == 0 {
		return types.NewWeblensError("delete trash entry did not get expected delete count")
	}
	return err
}

func (db *databaseService) GetTrashEntry(fileId types.FileId) (te types.TrashEntry, err error) {
	filter := bson.D{{"trashFileId", fileId}}
	ret := db.trash.FindOne(db.ctx, filter)
	if err = ret.Err(); err != nil {
		return
	}

	err = ret.Decode(&te)
	return
}

func (db *databaseService) GetAllFiles() ([]types.WeblensFile, error) {
	return []types.WeblensFile{}, nil
}

func (db *databaseService) StatFile(f types.WeblensFile) (types.FileStat, error) {
	fileInfo, err := os.Stat(f.GetAbsPath())
	if err != nil {
		return types.FileStat{}, types.WeblensErrorFromError(err)
	}

	return types.FileStat{
		Exists:  true,
		Name:    fileInfo.Name(),
		Size:    fileInfo.Size(),
		IsDir:   fileInfo.IsDir(),
		ModTime: fileInfo.ModTime(),
	}, nil
}

func (db *databaseService) ReadFile(f types.WeblensFile) ([]byte, error) {
	if f.IsDir() {
		return nil, types.WeblensErrorMsg("trying to read directory as regular file")
	}

	bs, err := os.ReadFile(f.GetAbsPath())
	if err != nil {
		return nil, types.WeblensErrorFromError(err)
	}
	return bs, nil
}

func (db *databaseService) StreamFile(f types.WeblensFile) (io.ReadCloser, error) {
	return f.Readable()
}

func (db *databaseService) ReadDir(f types.WeblensFile) ([]types.FileStat, error) {
	if !f.IsDir() {
		return nil, types.WeblensErrorMsg("trying to read regular file as directory")
	}

	entries, err := os.ReadDir(f.GetAbsPath())
	if err != nil {
		return nil, types.WeblensErrorFromError(err)
	}

	children := util.Map(
		entries, func(e os.DirEntry) types.FileStat {
			info, _ := e.Info()
			return types.FileStat{
				Name:    e.Name(),
				Size:    info.Size(),
				IsDir:   e.IsDir(),
				ModTime: info.ModTime(),
				Exists:  true,
			}
		},
	)

	return children, nil
}

func (db *databaseService) TouchFile(f types.WeblensFile) error {
	stat, _ := db.StatFile(f)
	if stat.Exists {
		return types.ErrFileAlreadyExists(f.GetAbsPath())
	}

	var err error
	if f.IsDir() {
		err = os.Mkdir(f.GetAbsPath(), os.FileMode(0777))
		if err != nil {
			return types.WeblensErrorFromError(err)
		}
	} else {
		var osFile *os.File
		if f.IsDetached() {
			osFile, err = os.Create("/tmp/" + f.Filename())
		} else {
			osFile, err = os.Create(f.GetAbsPath())
		}
		if err != nil {
			return types.WeblensErrorFromError(err)
		}
		err = osFile.Close()
		if err != nil {
			return types.WeblensErrorFromError(err)
		}
	}

	return nil
}

func (db *databaseService) GetFile(fileId types.FileId) (types.WeblensFile, error) {
	return nil, types.ErrNotImplemented("GetFile filesDB")
}
