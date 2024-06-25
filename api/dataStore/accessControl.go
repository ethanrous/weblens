package dataStore

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type accessMeta struct {
	shares      []types.Share
	user        types.User
	usingShare  types.Share
	requestMode types.RequestMode
	accessAt    time.Time

	fileTree     types.FileTree
	shareService types.ShareService
}

var apiKeyMap = map[types.WeblensApiKey]*ApiKeyInfo{}
var keyMapMu = &sync.Mutex{}

func NewAccessMeta(u types.User, ft types.FileTree) types.AccessMeta {
	return &accessMeta{
		user:     u,
		fileTree: ft,
	}
}

func (acc *accessMeta) Shares() []types.Share {
	return acc.shares
}

func (acc *accessMeta) User() types.User {
	return acc.user
}

func (acc *accessMeta) AddShare(s types.Share) error {
	if !acc.CanAccessShare(s) {
		return ErrUserNotAuthorized
	}

	acc.shares = append(acc.shares, s)
	return nil
}

func (acc *accessMeta) SetRequestMode(r types.RequestMode) types.AccessMeta {
	if acc.requestMode != "" {
		util.Warning.Printf("Overriding request mode from %s to %s", acc.requestMode, r)
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

func (acc *accessMeta) UsingShare() types.Share {
	return acc.usingShare
}

func (acc *accessMeta) SetUsingShare(s types.Share) {
	acc.usingShare = s
}

func (acc *accessMeta) CanAccessFile(file types.WeblensFile) bool {
	if file == nil {
		return false
	}

	switch acc.RequestMode() {
	case WebsocketFileUpdate, MarshalFile:
		return true
	}

	if file.Owner() == acc.User() {
		return true
	} else if file.Owner() == ExternalRootUser {
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

	var foundShare types.Share
	shareFileIds := util.Map(shares, func(s types.Share) types.FileId {
		return types.FileId(s.GetItemId())
	})
	err := file.BubbleMap(
		func(wf types.WeblensFile) error {
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
		util.ErrTrace(err)
	}

	if foundShare != nil {
		acc.SetUsingShare(foundShare)
	}

	return foundShare != nil
}

func (acc *accessMeta) CanAccessShare(s types.Share) bool {
	if s == nil {
		err := fmt.Errorf("canAccessShare nil share")
		util.ErrTrace(err)
		return false
	}

	if !s.IsEnabled() {
		return false
	}

	if s.IsPublic() {
		return true
	}

	if s.GetOwner() == acc.User() {
		return true
	}

	if slices.Contains(s.GetAccessors(), acc.User()) {
		return true
	}

	return false
}

func (acc *accessMeta) CanAccessAlbum(a types.Album) bool {
	return acc.User() == a.GetOwner() || slices.Contains(a.GetUsers(), acc.User())
}

func InitApiKeyMap() {
	keys := dbServer.getApiKeys()
	keyMapMu.Lock()
	defer keyMapMu.Unlock()
	for _, keyInfo := range keys {
		apiKeyMap[keyInfo.Key] = &keyInfo
	}
}

func GetApiKeyInfo(key types.WeblensApiKey) *ApiKeyInfo {
	keyMapMu.Lock()
	defer keyMapMu.Unlock()
	return apiKeyMap[key]
}

func GenerateApiKey(acc types.AccessMeta) (key *ApiKeyInfo, err error) {
	if !acc.User().IsAdmin() {
		err = ErrUserNotAuthorized
		return
	} else if acc.RequestMode() != ApiKeyCreate {
		err = ErrBadRequestMode
		return
	}

	createTime := time.Now()
	hash := types.WeblensApiKey(util.GlobbyHash(0, acc.User().GetUsername(), strconv.Itoa(int(createTime.Unix()))))

	newKey := &ApiKeyInfo{
		Key:         hash,
		Owner:       acc.User().GetUsername(),
		CreatedTime: createTime,
	}

	err = dbServer.newApiKey(*newKey)
	if err != nil {
		return nil, err
	}
	keyMapMu.Lock()
	apiKeyMap[hash] = newKey
	keyMapMu.Unlock()

	return newKey, nil
}

func GetApiKeys(acc types.AccessMeta) ([]ApiKeyInfo, error) {
	if acc.RequestMode() != ApiKeyGet {
		return nil, ErrBadRequestMode
	}
	keys := dbServer.getApiKeysByUser(acc.User().GetUsername())
	if keys == nil {
		keys = []ApiKeyInfo{}
	}
	return keys, nil
}

func CheckApiKey(key types.WeblensApiKey) bool {
	keyInfo := GetApiKeyInfo(key)
	return keyInfo != nil
}

func DeleteApiKey(key types.WeblensApiKey) {
	keyMapMu.Lock()
	delete(apiKeyMap, key)
	keyMapMu.Unlock()
	dbServer.removeApiKey(key)
}
