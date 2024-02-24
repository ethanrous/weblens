package dataStore

import (
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

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
