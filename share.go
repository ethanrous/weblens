package weblens

import (
	"encoding/json"
	"time"

	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShareId string

type FileShare struct {
	ShareId   ShareId          `bson:"_id" json:"shareId"`
	FileId    fileTree.FileId  `bson:"fileId" json:"fileId"`
	ShareName string          `bson:"shareName"`
	Owner     types.Username   `bson:"owner"`
	Accessors []types.Username `bson:"accessors"`
	Public    bool            `bson:"public"`
	Wormhole  bool            `bson:"wormhole"`
	Enabled   bool            `bson:"enabled"`
	Expires   time.Time        `bson:"expires"`
}

func NewFileShare(
	f *fileTree.WeblensFile, u types.User, accessors []types.User, public bool, wormhole bool,
) Share {
	return &FileShare{
		ShareId: ShareId(primitive.NewObjectID().Hex()),
		FileId:    f.ID(),
		Owner:   u.GetUsername(),
		Accessors: internal.Map(
			accessors, func(u types.User) types.Username {
				return u.GetUsername()
			},
		),
		Public:    public,
		Wormhole:  wormhole,
		Enabled:   true,
	}
}

func (s *FileShare) GetShareId() ShareId            { return s.ShareId }
func (s *FileShare) GetShareType() ShareType        { return SharedFile }
func (s *FileShare) GetItemId() string              { return string(s.FileId) }
func (s *FileShare) SetItemId(fileId string)        { s.FileId = fileTree.FileId(fileId) }
func (s *FileShare) GetAccessors() []types.Username { return s.Accessors }

func (s *FileShare) AddUsers(newUsers []types.User) error {
	usernames := internal.Map(
		newUsers, func(u types.User) types.Username {
			return u.GetUsername()
		},
	)
	err := types.SERV.StoreService.AddUsersToShare(s, usernames)
	if err != nil {
		return err
	}

	s.Accessors = internal.AddToSet(s.Accessors, usernames...)
	return nil
}

func (s *FileShare) GetOwner() types.Username { return s.Owner }
func (s *FileShare) IsPublic() bool           { return s.Public }
func (s *FileShare) IsWormhole() bool         { return s.Wormhole }
func (s *FileShare) SetPublic(pub bool) {
	s.Public = pub
}

func (s *FileShare) IsEnabled() bool { return s.Enabled }
func (s *FileShare) SetEnabled(enable bool) {
	s.Enabled = enable
}

func (s *FileShare) UnmarshalBSON(bs []byte) error {
	data := map[string]any{}
	err := bson.Unmarshal(bs, &data)
	if err != nil {
		return err
	}

	s.ShareId = ShareId(data["_id"].(string))
	s.FileId = fileTree.FileId(data["fileId"].(string))
	s.ShareName = data["shareName"].(string)
	s.Owner = types.Username(data["owner"].(string))

	s.Accessors = internal.Map(
		internal.SliceConvert[string](data["accessors"].(primitive.A)), func(un string) types.Username {
			return types.Username(un)
		},
	)

	s.Public = data["public"].(bool)
	s.Wormhole = data["wormhole"].(bool)
	s.Enabled = data["enabled"].(bool)
	s.Expires = data["expires"].(primitive.DateTime).Time()

	return nil
}

func (s *FileShare) MarshalBSON() ([]byte, error) {
	data := map[string]any{
		"_id":       s.ShareId,
		"fileId":    s.FileId,
		"shareName": s.ShareName,
		"owner":     s.Owner,
		"accessors": s.Accessors,
		"public":    s.Public,
		"wormhole":  s.Wormhole,
		"enabled":   s.Enabled,
		"expires":   s.Expires,
	}

	return bson.Marshal(data)
}

func (s *FileShare) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"id":        s.ShareId,
		"fileId":    s.FileId,
		"shareName": s.ShareName,
		"owner":     s.Owner,
		"accessors": s.Accessors,
		"public":    s.Public,
		"wormhole":  s.Wormhole,
		"enabled":   s.Enabled,
		"expires":   s.Expires,
	}

	return json.Marshal(data)
}

type Share interface {
	GetShareId() ShareId
	GetShareType() ShareType
	GetItemId() string
	IsPublic() bool
	IsEnabled() bool
	GetAccessors() []types.Username
	GetOwner() types.Username
	AddUsers(newUsers []types.User) error

	SetItemId(string)
	SetPublic(bool)
	SetEnabled(bool)
}

type ShareService interface {
	Init(store types.ShareStore) error
	Size() int

	Get(id ShareId) Share
	Add(share Share) error
	Del(id ShareId) error

	GetSharedWithUser(u types.User) ([]Share, error)
	GetAllShares() []Share
	UpdateShareItem(ShareId, string) error
}

type ShareType string

const (
	SharedFile ShareType = "file"
	ShareAlbum ShareType = "album"
)

var ErrUserNotAuthorized = werror.NewWeblensError("user does not have access the requested resource")
