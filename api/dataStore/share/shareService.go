package share

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type shareService struct {
	repo map[types.ShareId]types.Share
	db   types.ShareDB
}

func NewService() types.ShareService {
	return &shareService{
		repo: make(map[types.ShareId]types.Share),
	}
}

func (ss *shareService) Init(db types.DatabaseService) error {
	ss.db = db
	shares, err := ss.db.GetAllShares()
	if err != nil {
		return err
	}

	ss.repo = make(map[types.ShareId]types.Share)
	for _, sh := range shares {
		ss.repo[sh.GetShareId()] = sh
		switch sh.GetShareType() {
		case types.FileShare:
			sharedFile := types.SERV.FileTree.Get(types.FileId(sh.GetItemId()))
			if sharedFile != nil {
				err := sharedFile.SetShare(sh)
				if err != nil {
					return err
				}
			} else {
				util.Warning.Println("Ignoring possibly no longer existing file in share init")
			}
		}
	}

	return nil
}

func (ss *shareService) Add(sh types.Share) error {
	err := ss.db.CreateShare(sh)
	if err != nil {
		return err
	}

	ss.repo[sh.GetShareId()] = sh

	if sh.GetShareType() == types.FileShare {
		err := types.SERV.FileTree.Get(types.FileId(sh.GetItemId())).SetShare(sh)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *shareService) Del(sId types.ShareId) (err error) {
	if ss.repo[sId] == nil {
		return types.ErrNoShare
	}
	delete(ss.repo, sId)
	return types.NewWeblensError("not impl - delete from share repo")

	// switch s.GetShareType() {
	// case dataStore.FileShare:
	// 	err = dataStore.dbServer.removeFileShare(s.GetShareId())
	// 	if err != nil {
	// 		return
	// 	}
	// 	f := ft.Get(types.FileId(s.GetContentId()))
	// 	err = f.RemoveShare(s.GetShareId())
	// 	if err != nil {
	// 		return
	// 	}
	//
	// 	util.Each(c, func(caster types.BroadcasterAgent) {
	// 		caster.PushFileUpdate(f)
	// 	})
	//
	// default:
	// 	err = dataStore.ErrBadShareType
	// }
	//
	// return
}

func (ss *shareService) Get(sId types.ShareId) types.Share {
	return ss.repo[sId]
}

func (ss *shareService) Size() int {
	return len(ss.repo)
}
