package dataStore

import (
	"errors"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

func (s fileShareData) GetShareType() shareType { return FileShare }
func (s fileShareData) GetContentId() string    { return s.FileId }
func (s fileShareData) IsPublic() bool          { return s.Public }

func CreateWormhole(f *WeblensFile) {

	wormholeName := f.Filename()
	whId := util.GlobbyHash(8, f.Id(), wormholeName)
	wormhole := shareData{
		ShareId:   whId,
		ShareName: wormholeName,
		Public:    true,
		Wormhole:  true,
		Expires:   time.Unix(0, 0),
	}

	fddb.addShareToFolder(f, wormhole)
}

func GetWormhole(shareId string) (sd shareData, folderId string, err error) {
	fd, err := fddb.getFolderByShare(shareId)
	if err != nil {
		return
	}
	folderId = fd.FolderId

	_, sd, e := util.YoinkFunc(fd.Shares, func(s shareData) bool { return s.ShareId == shareId })
	if !e {
		err = ErrNoShare
	}

	return
}

func RemoveWormhole(shareId string) (err error) {
	err = fddb.removeShare(shareId)

	return
}

func CreatePublicFileShare(fileId string) (shareId string, err error) {
	f := FsTreeGet(fileId)
	if f == nil {
		err = ErrNoFile
		return
	}
	shareId = util.GlobbyHash(8, f.Id(), true)

	newShare := fileShareData{
		ShareId:   shareId,
		FileId:    fileId,
		ShareName: f.Filename(), // This is temporary, we want to be able to rename shares for obfuscation
		Public:    true,
		Wormhole:  false,
		Expires:   time.Unix(0, 0),
		ShareType: FileShare,
	}
	err = fddb.newFileShare(newShare)
	if err == nil {
		globalCaster.PushFileUpdate(f)
	}
	return
}

func GetShare(shareId string, shareType shareType) (s Share, err error) {
	switch shareType {
	case FileShare:
		s, err = fddb.getFileShare(shareId)
	default:
		err = errors.New("unexpected share type")
	}

	return
}
