package share

import (
	"github.com/ethrousseau/weblens/api/types"
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
	}

	return nil
}

func (ss *shareService) Add(s types.Share) error {
	ss.repo[s.GetShareId()] = s

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
