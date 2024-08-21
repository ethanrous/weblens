package database

import (
	"io"
	"os"

	"github.com/ethrousseau/weblens/api/internal"
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) NewTrashEntry(te types.TrashEntry) error {
	_, err := db.trash.InsertOne(db.ctx, te)
	return err
}

func (db *databaseService) DeleteTrashEntry(fileId types.FileId) error {
	res, err := db.trash.DeleteOne(db.ctx, bson.M{"trashFileId": fileId})
	if res.DeletedCount == 0 {
		return error2.NewWeblensError("delete trash entry did not get expected delete count")
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

func (db *databaseService) GetAllFiles() ([]*fileTree.WeblensFile, error) {
	return []*fileTree.WeblensFile{}, nil
}

func (db *databaseService) StatFile(f *fileTree.WeblensFile) (types.FileStat, error) {
	fileInfo, err := os.Stat(f.GetAbsPath())
	if err != nil {
		return types.FileStat{}, error2.Wrap(err)
	}

	return types.FileStat{
		Exists:  true,
		Name:    fileInfo.Name(),
		Size:    fileInfo.Size(),
		IsDir:   fileInfo.IsDir(),
		ModTime: fileInfo.ModTime(),
	}, nil
}

func (db *databaseService) ReadFile(f *fileTree.WeblensFile) ([]byte, error) {
	if f.IsDir() {
		return nil, error2.WErrMsg("trying to read directory as regular file")
	}

	bs, err := os.ReadFile(f.GetAbsPath())
	if err != nil {
		return nil, error2.Wrap(err)
	}
	return bs, nil
}

func (db *databaseService) StreamFile(f *fileTree.WeblensFile) (io.ReadCloser, error) {
	return f.Readable()
}

func (db *databaseService) ReadDir(f *fileTree.WeblensFile) ([]types.FileStat, error) {
	if !f.IsDir() {
		return nil, error2.WErrMsg("trying to read regular file as directory")
	}

	entries, err := os.ReadDir(f.GetAbsPath())
	if err != nil {
		return nil, error2.Wrap(err)
	}

	children := internal.Map(
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

func (db *databaseService) TouchFile(f *fileTree.WeblensFile) error {
	stat, _ := db.StatFile(f)
	if stat.Exists {
		return types.ErrFileAlreadyExists(f.GetAbsPath())
	}

	var err error
	if f.IsDir() {
		err = os.Mkdir(f.GetAbsPath(), os.FileMode(0777))
		if err != nil {
			return error2.Wrap(err)
		}
	} else {
		var osFile *os.File
		if f.IsDetached() {
			osFile, err = os.Create("/tmp/" + f.Filename())
		} else {
			osFile, err = os.Create(f.GetAbsPath())
		}
		if err != nil {
			return error2.Wrap(err)
		}
		err = osFile.Close()
		if err != nil {
			return error2.Wrap(err)
		}
	}

	return nil
}

func (db *databaseService) GetFile(fileId types.FileId) (*fileTree.WeblensFile, error) {
	return nil, error2.NotImplemented("GetFile filesDB")
}
