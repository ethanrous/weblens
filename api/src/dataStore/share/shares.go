package share

import (
	"encoding/json"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileShare struct {
	ShareId   types.ShareId   `bson:"_id" json:"shareId"`
	FileId    types.FileId    `bson:"fileId" json:"fileId"`
	ShareName string          `bson:"shareName"`
	Owner     types.User      `bson:"owner"`
	Accessors []types.User    `bson:"accessors"`
	Public    bool            `bson:"public"`
	Wormhole  bool            `bson:"wormhole"`
	Enabled   bool            `bson:"enabled"`
	Expires   time.Time       `bson:"expires"`
	ShareType types.ShareType `bson:"shareType"`
}

func NewFileShare(f types.WeblensFile, u types.User, accessors []types.User, public bool, wormhole bool) types.Share {
	return &FileShare{
		ShareId:   types.ShareId(primitive.NewObjectID().Hex()),
		FileId:    f.ID(),
		Owner:     u,
		Accessors: accessors,
		Public:    public,
		Wormhole:  wormhole,
		ShareType: types.FileShare,
		Enabled:   true,
	}
}

func (s *FileShare) GetShareId() types.ShareId     { return s.ShareId }
func (s *FileShare) GetShareType() types.ShareType { return types.FileShare }
func (s *FileShare) GetItemId() string             { return s.FileId.String() }
func (s *FileShare) SetItemId(fileId string)       { s.FileId = types.FileId(fileId) }
func (s *FileShare) GetAccessors() []types.User    { return s.Accessors }

func (s *FileShare) AddUsers(newUsers []types.User) error {
	usernames := util.Map(
		newUsers, func(u types.User) types.Username {
			return u.GetUsername()
		},
	)
	err := types.SERV.StoreService.AddUsersToShare(s, usernames)
	if err != nil {
		return err
	}

	s.Accessors = util.AddToSet(s.Accessors, newUsers)
	return nil
}

func (s *FileShare) GetOwner() types.User { return s.Owner }
func (s *FileShare) IsPublic() bool       { return s.Public }
func (s *FileShare) IsWormhole() bool     { return s.Wormhole }
func (s *FileShare) SetPublic(pub bool) {
	s.Public = pub
}

func (s *FileShare) IsEnabled() bool { return s.Enabled }
func (s *FileShare) SetEnabled(enable bool) error {
	s.Enabled = enable
	return types.SERV.StoreService.SetShareEnabledById(s.ShareId, enable)
}

func (s *FileShare) UnmarshalBSON(bs []byte) error {
	data := map[string]any{}
	err := bson.Unmarshal(bs, &data)
	if err != nil {
		return err
	}

	s.ShareId = types.ShareId(data["_id"].(string))
	s.FileId = types.FileId(data["fileId"].(string))
	s.ShareName = data["shareName"].(string)
	s.Owner = types.SERV.UserService.Get(types.Username(data["owner"].(string)))

	s.Accessors = util.Map(
		util.SliceConvert[string](data["accessors"].(primitive.A)), func(un string) types.User {
			return types.SERV.UserService.Get(types.Username(un))
		},
	)

	s.Public = data["public"].(bool)
	s.Wormhole = data["wormhole"].(bool)
	s.Enabled = data["enabled"].(bool)
	s.Expires = data["expires"].(primitive.DateTime).Time()
	s.ShareType = types.ShareType(data["shareType"].(string))

	return nil
}

func (s *FileShare) MarshalBSON() ([]byte, error) {
	data := map[string]any{
		"_id":       s.ShareId,
		"fileId":    s.FileId,
		"shareName": s.ShareName,
		"owner":     s.Owner.GetUsername(),
		"accessors": util.Map(
			s.Accessors, func(u types.User) types.Username {
				return u.GetUsername()
			},
		),
		"public":    s.Public,
		"wormhole":  s.Wormhole,
		"enabled":   s.Enabled,
		"expires":   s.Expires,
		"shareType": s.ShareType,
	}

	return bson.Marshal(data)
}

func (s *FileShare) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"id":        s.ShareId,
		"fileId":    s.FileId,
		"shareName": s.ShareName,
		"owner":     s.Owner.GetUsername(),
		"accessors": util.Map(
			s.Accessors, func(u types.User) types.Username {
				return u.GetUsername()
			},
		),
		"public":    s.Public,
		"wormhole":  s.Wormhole,
		"enabled":   s.Enabled,
		"expires":   s.Expires,
		"shareType": s.ShareType,
	}

	return json.Marshal(data)
}
