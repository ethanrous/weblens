package dataStore

import (
	"fmt"
	"slices"
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
)

type accessMeta struct {
	shares     []weblens.Share
	user        types.User
	usingShare weblens.Share
	requestMode types.RequestMode
	accessAt    time.Time

	shareService weblens.ShareService
}

// func NewAccessMeta(u types.User) types.AccessMeta {
// 	return &accessMeta{
// 		user: u,
// 	}
// }

func (acc *accessMeta) Shares() []weblens.Share {
	return acc.shares
}

func (acc *accessMeta) User() types.User {
	return acc.user
}

func (acc *accessMeta) AddShare(s weblens.Share) error {
	if !acc.CanAccessShare(s) {
		return weblens.ErrUserNotAuthorized
	}

	acc.shares = append(acc.shares, s)
	return nil
}

func (acc *accessMeta) SetRequestMode(r types.RequestMode) types.AccessMeta {
	if acc.requestMode != "" {
		wlog.Warning.Printf("Overriding request mode from %s to %s", acc.requestMode, r)
	}
	acc.requestMode = r

	return acc
}

func (acc *accessMeta) SetTime(t time.Time) types.AccessMeta {
	acc.accessAt = t
	return acc
}

func (acc *accessMeta) GetTime() time.Time {
	return acc.accessAt
}

func (acc *accessMeta) RequestMode() types.RequestMode {
	return acc.requestMode
}

func (acc *accessMeta) AddShareId(sId types.ShareId) types.AccessMeta {
	if sId == "" {
		return acc
	}

	s := acc.shareService.Get(sId)
	if s == nil {
		return acc
	}
	acc.shares = append(acc.shares, s)

	return acc
}

func (acc *accessMeta) UsingShare() weblens.Share {
	return acc.usingShare
}

func (acc *accessMeta) SetUsingShare(s weblens.Share) {
	acc.usingShare = s
}

func (acc *accessMeta) CanAccessFile(file *fileTree.WeblensFile) bool {
	if file == nil {
		return false
	}

	switch acc.RequestMode() {
	case WebsocketFileUpdate, MarshalFile:
		return true
	}

	if file.Owner() == acc.User() {
		return true
	} else if file.Owner() == weblens.ExternalRootUser {
		return acc.User().IsAdmin()
	}

	shares := acc.Shares()
	if len(shares) == 0 {
		return false
	}

	if acc.UsingShare() != nil {
		if types.FileId(acc.UsingShare().GetItemId()) == file.ID() ||
			types.SERV.FileTree.Get(types.FileId(acc.UsingShare().GetItemId())).IsParentOf(file) {
			return true
		} else {
			return false
		}
	}

	var foundShare weblens.Share
	shareFileIds := internal.Map(
		shares, func(s weblens.Share) types.FileId {
			return types.FileId(s.GetItemId())
		},
	)
	err := file.BubbleMap(
		func(wf *fileTree.WeblensFile) error {
			if foundShare != nil {
				return nil
			}

			i := slices.Index(shares, wf.GetShare())
			if i != -1 {
				foundShare = shares[i]
				return nil
			}

			i = slices.Index(shareFileIds, wf.ID())
			if i != -1 {
				foundShare = shares[i]
			}

			return nil
		},
	)

	if err != nil {
		wlog.ErrTrace(err)
	}

	if foundShare != nil {
		acc.SetUsingShare(foundShare)
	}

	return foundShare != nil
}

func (acc *accessMeta) CanAccessShare(s weblens.Share) bool {
	if s == nil {
		err := fmt.Errorf("canAccessShare nil share")
		wlog.ErrTrace(err)
		return false
	}

	if !s.IsEnabled() {
		return false
	}

	if s.IsPublic() {
		return true
	}

	if s.GetOwner() == acc.User().GetUsername() {
		return true
	}

	if slices.Contains(s.GetAccessors(), acc.User().GetUsername()) {
		return true
	}

	return false
}

func (acc *accessMeta) CanAccessAlbum(a types.Album) bool {
	return acc.User() == a.GetOwner() || slices.Contains(a.GetSharedWith(), acc.User().GetUsername())
}
