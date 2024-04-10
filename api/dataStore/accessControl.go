package dataStore

import (
	"slices"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

var apiKeyMap map[types.WeblensApiKey]*ApiKeyInfo = map[types.WeblensApiKey]*ApiKeyInfo{}

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

func (acc *accessMeta) RequestMode() types.RequestMode {
	return acc.requestMode
}

func (acc *accessMeta) AddShareId(sId types.ShareId, st types.ShareType) types.AccessMeta {
	if sId == "" {
		return acc
	}

	s, _ := GetShare(sId, st)
	if s == nil {
		return acc
	}
	acc.shares = append(acc.shares, s)

	return acc
}

func (acc *accessMeta) UsingShare() types.Share {
	return acc.usingShare
}

func (acc *accessMeta) setUsingShare(s types.Share) {
	acc.usingShare = s
}

func GetRelevantShare(file types.WeblensFile, acc types.AccessMeta) types.Share {
	if len(acc.Shares()) == 0 {
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

	switch acc.RequestMode() {
	case WebsocketFileUpdate, MarshalFile:
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

func InitApiKeyMap() {
	keys := fddb.getApiKeys()
	for _, keyInfo := range keys {
		apiKeyMap[keyInfo.Key] = &keyInfo
	}
}

func GetApiKeyInfo(key types.WeblensApiKey) *ApiKeyInfo {
	return apiKeyMap[key]
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
	apiKeyMap[hash] = &newKey

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

func CheckApiKey(key types.WeblensApiKey) bool {
	keyInfo := GetApiKeyInfo(key)
	return keyInfo != nil
}

func SetKeyRemote(key types.WeblensApiKey, remoteName string) error {
	kInfo := GetApiKeyInfo(key)
	if kInfo == nil {
		return ErrNoKey
	}
	kInfo.RemoteUsing = remoteName
	fddb.updateApiKey(*kInfo)

	return nil
}
