package share

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type shareService struct {
	repo map[types.ShareId]types.Share
	db   types.ShareStore
}

func NewService() types.ShareService {
	return &shareService{
		repo: make(map[types.ShareId]types.Share),
	}
}

func (ss *shareService) Init(db types.ShareStore) error {
	ss.db = db
	shares, err := ss.db.GetAllShares()
	if err != nil {
		return err
	}

	ss.repo = make(map[types.ShareId]types.Share)
	for _, sh := range shares {
		if len(sh.GetAccessors()) == 0 && !sh.IsPublic() && (sh.GetShareType() != types.FileShare || !sh.(*FileShare).IsWormhole()) {
			util.Debug.Println("Removing share on init...")
			err = db.DeleteShare(sh.GetShareId())
			if err != nil {
				return err
			}
			continue
		}

		ss.repo[sh.GetShareId()] = sh
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
		err = types.SERV.FileTree.Get(types.FileId(sh.GetItemId())).SetShare(sh)
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

func (ss *shareService) GetAllShares() []types.Share {
	return util.MapToValues(ss.repo)
}

func (ss *shareService) Size() int {
	return len(ss.repo)
}

func (ss *shareService) GetSharedWithUser(u types.User) ([]types.Share, error) {
	return types.SERV.StoreService.GetSharedWithUser(u.GetUsername())
}
