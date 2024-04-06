package dataStore

import (
	"slices"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func (a accessMeta) Shares() []types.Share {
	return a.shares
}

func (a accessMeta) User() types.User {
	return a.user
}

func NewAccessMeta(u types.Username) types.AccessMeta {
	user := GetUser(u)
	return &accessMeta{
		user: user,
	}
}

func (a *accessMeta) AddShare(s types.Share) types.AccessMeta {
	a.shares = append(a.shares, s)
	return a
}

func (a *accessMeta) SetRequestMode(r types.RequestMode) types.AccessMeta {
	if a.requestMode != "" {
		util.Warning.Printf("Overriding request mode from %s to %s", a.requestMode, r)
	}
	a.requestMode = r

	return a
}

func (a *accessMeta) RequestMode() types.RequestMode {
	return a.requestMode
}

func (a *accessMeta) AddShareId(sId types.ShareId, st types.ShareType) types.AccessMeta {
	s, _ := GetShare(sId, st)
	a.shares = append(a.shares, s)

	return a
}

func (a *accessMeta) UsingShare() types.Share {
	return a.usingShare
}

func (a *accessMeta) setUsingShare(s types.Share) {
	a.usingShare = s
}

func GetRelevantShare(file types.WeblensFile, acc types.AccessMeta) types.Share {
	shares := acc.Shares()
	if len(shares) == 0 {
		return nil
	}

	ancestors := []types.FileId{}
	file.BubbleMap(func(wf types.WeblensFile) {
		ancestors = append(ancestors, wf.Id())
	})

	var foundShare types.Share
	if len(ancestors) != 0 {
		for _, s := range acc.Shares() {
			if slices.Contains(ancestors, types.FileId(s.GetContentId())) && (s.IsPublic() || slices.Contains(s.GetAccessors(), acc.User().GetUsername())) {
				foundShare = s
				break
			}
		}
	}

	if foundShare != nil {
		acc.(*accessMeta).setUsingShare(foundShare)
	}
	return foundShare
}

func CanAccessFile(file types.WeblensFile, acc types.AccessMeta) bool {
	if file == nil {
		return false
	}

	if acc.RequestMode() == WebsocketFileUpdate {
		return true
	}

	if file.Owner() == acc.User() {
		return true
	} else if file.Owner() == EXTERNAL_ROOT_USER {
		return acc.User().IsAdmin()
		// util.Debug.Println("REQUEST:", acc.RequestMode(), "ID:", file.Id())

		// Clients are only allowed to subscribe to root folders, nothing else
		// TODO. This is quite a hack, will generalize later. Only allow external folder to be subbed to
		// return acc.RequestMode() == FileSubscribeRequest && file.Id() == "EXTERNAL_ROOT"
	}

	shares := acc.Shares()
	if len(shares) == 0 {
		return false
	}

	using := acc.UsingShare()
	if using != nil {
		if types.FileId(using.GetContentId()) == file.Id() {
			return true
		}
	}
	return GetRelevantShare(file, acc) != nil
}

func CanUserAccessShare(s types.Share, username types.Username) bool {
	return s.IsEnabled() && (s.IsPublic() || s.GetOwner() == username || slices.Contains(s.GetAccessors(), username))
}

func GenerateApiKey(acc types.AccessMeta) (key ApiKeyInfo, err error) {
	if !acc.User().IsAdmin() {
		err = ErrUserNotAuthorized
		return
	} else if acc.RequestMode() != ApiKeyCreate {
		err = ErrBadRequestMode
		return
	}

	createTime := time.Now()
	hash := types.WeblensApiKey(util.GlobbyHash(0, acc.User().GetUsername(), strconv.Itoa(int(createTime.Unix()))))

	newKey := ApiKeyInfo{
		Key:         hash,
		Owner:       acc.User().GetUsername(),
		CreatedTime: createTime,
	}

	fddb.newApiKey(newKey)

	return newKey, nil
}

func GetApiKeys(acc types.AccessMeta) ([]ApiKeyInfo, error) {
	if acc.RequestMode() != ApiKeyGet {
		return nil, ErrBadRequestMode
	}
	keys := fddb.getApiKeysByUser(acc.User().GetUsername())
	if keys == nil {
		keys = []ApiKeyInfo{}
	}
	return keys, nil
}

func CheckApiKey(key string) bool {
	keyInfo := fddb.getApiKey(key)
	return keyInfo.Key != ""
}