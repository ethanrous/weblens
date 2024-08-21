package weblens

import (
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
)

type ShareServiceImpl struct {
	repo map[ShareId]Share
	db   types.ShareStore
}

func NewShareService() ShareService {
	return &ShareServiceImpl{
		repo: make(map[ShareId]Share),
	}
}

func (ss *ShareServiceImpl) Init(db types.ShareStore) error {
	ss.db = db
	shares, err := ss.db.GetAllShares()
	if err != nil {
		return err
	}

	ss.repo = make(map[ShareId]Share)
	for _, sh := range shares {
		if len(sh.GetAccessors()) == 0 && !sh.IsPublic() && (sh.GetShareType() != SharedFile || !sh.(*FileShare).IsWormhole()) {
			wlog.Debug.Println("Removing share on init...")
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

func (ss *ShareServiceImpl) Add(sh Share) error {
	err := ss.db.CreateShare(sh)
	if err != nil {
		return err
	}

	ss.repo[sh.GetShareId()] = sh

	if sh.GetShareType() == SharedFile {
		err = types.SERV.FileTree.Get(types.FileId(sh.GetItemId())).SetShare(sh)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *ShareServiceImpl) Del(sId ShareId) (err error) {
	if ss.repo[sId] == nil {
		return types.ErrNoShare
	}

	err = ss.db.DeleteShare(sId)
	if err != nil {
		return err
	}

	delete(ss.repo, sId)
	return nil
}

func (ss *ShareServiceImpl) Get(sId ShareId) Share {
	return ss.repo[sId]
}

func (ss *ShareServiceImpl) GetAllShares() []Share {
	return internal.MapToValues(ss.repo)
}

func (ss *ShareServiceImpl) Size() int {
	return len(ss.repo)
}

func (ss *ShareServiceImpl) GetSharedWithUser(u types.User) ([]Share, error) {
	return ss.db.GetSharedWithUser(u.GetUsername())
}

func (ss *ShareServiceImpl) UpdateShareItem(shareId ShareId, newItemId string) error {
	share := ss.Get(shareId)
	share.SetItemId(newItemId)
	err := ss.db.UpdateShare(share)
	if err != nil {
		return err
	}

	return nil
}

func (ss *ShareServiceImpl) EnableShare(share Share, enable bool) error {
	err := ss.db.SetShareEnabledById(share.GetShareId(), enable)
	if err != nil {
		return err
	}

	share.SetEnabled(enable)
	return nil
}
