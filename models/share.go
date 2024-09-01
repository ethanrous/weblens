package models

import (
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ Share = (*FileShare)(nil)
var _ Share = (*AlbumShare)(nil)

type ShareId string

type FileShare struct {
	ShareId   ShareId         `bson:"_id" json:"shareId"`
	FileId    fileTree.FileId `bson:"fileId" json:"fileId"`
	ShareName string          `bson:"shareName"`
	Owner     Username        `bson:"owner"`
	Accessors []Username      `bson:"accessors"`
	Public    bool            `bson:"public"`
	Wormhole  bool            `bson:"wormhole"`
	Enabled   bool            `bson:"enabled"`
	Expires   time.Time       `bson:"expires"`
	Updated   time.Time `bson:"updated"`
	ShareType ShareType `bson:"shareType"`
}

type AlbumShare struct {
	ShareId   ShareId    `bson:"_id" json:"shareId"`
	AlbumId   AlbumId    `bson:"albumId" json:"albumId"`
	Owner     Username   `bson:"owner"`
	Accessors []Username `bson:"accessors"`
	Public    bool       `bson:"public"`
	Enabled   bool       `bson:"enabled"`
	Expires   time.Time  `bson:"expires"`
	Updated   time.Time  `bson:"updated"`
	ShareType ShareType  `bson:"shareType"`
}

func NewFileShare(
	f *fileTree.WeblensFileImpl, u *User, accessors []*User, public bool, wormhole bool,
) *FileShare {
	return &FileShare{
		ShareId: ShareId(primitive.NewObjectID().Hex()),
		FileId:  f.ID(),
		Owner:   u.GetUsername(),
		Accessors: internal.Map(
			accessors, func(u *User) Username {
				return u.GetUsername()
			},
		),
		Public:    public,
		Wormhole:  wormhole,
		Enabled:   true,
		Updated:   time.Now(),
		ShareType: SharedFile,
	}
}

func NewAlbumShare(
	a *Album, u *User, accessors []*User, public bool,
) Share {
	return &AlbumShare{
		ShareId: ShareId(primitive.NewObjectID().Hex()),
		AlbumId: a.ID(),
		Owner:   u.GetUsername(),
		Accessors: internal.Map(
			accessors, func(u *User) Username {
				return u.GetUsername()
			},
		),
		Public:    public,
		Enabled:   true,
		ShareType: SharedAlbum,
	}
}

func (s *FileShare) GetShareId() ShareId      { return s.ShareId }
func (s *FileShare) GetShareType() ShareType  { return SharedFile }
func (s *FileShare) GetItemId() string        { return string(s.FileId) }
func (s *FileShare) SetItemId(fileId string)  { s.FileId = fileTree.FileId(fileId) }
func (s *FileShare) GetAccessors() []Username { return s.Accessors }

func (s *FileShare) AddUsers(usernames []Username) {
	s.Accessors = internal.AddToSet(s.Accessors, usernames...)
}

func (s *FileShare) RemoveUsers(usernames []Username) {
	s.Accessors = internal.AddToSet(s.Accessors, usernames...)
}

func (s *FileShare) GetOwner() Username { return s.Owner }
func (s *FileShare) IsPublic() bool     { return s.Public }
func (s *FileShare) IsWormhole() bool   { return s.Wormhole }
func (s *FileShare) SetPublic(pub bool) {
	s.Public = pub
}

func (s *FileShare) IsEnabled() bool { return s.Enabled }
func (s *FileShare) SetEnabled(enable bool) {
	s.Enabled = enable
}

func (s *FileShare) UpdatedNow() {
	s.Updated = time.Now()
}

func (s *FileShare) LastUpdated() time.Time {
	return s.Updated
}

// func (s *FileShare) UnmarshalBSON(bs []byte) error {
// 	data := map[string]any{}
// 	err := bson.Unmarshal(bs, &data)
// 	if err != nil {
// 		return err
// 	}
//
// 	s.ShareId = ShareId(data["_id"].(string))
// 	s.FileId = fileTree.FileId(data["fileId"].(string))
// 	s.ShareName = data["shareName"].(string)
// 	s.Owner = Username(data["owner"].(string))
//
// 	s.Accessors = internal.Map(
// 		internal.SliceConvert[string](data["accessors"].(primitive.A)), func(un string) Username {
// 			return Username(un)
// 		},
// 	)
//
// 	s.Public = data["public"].(bool)
// 	s.Wormhole = data["wormhole"].(bool)
// 	s.Enabled = data["enabled"].(bool)
// 	s.Expires = data["expires"].(primitive.DateTime).Time()
//
// 	return nil
// }
//
// func (s *FileShare) MarshalBSON() ([]byte, error) {
// 	data := map[string]any{
// 		"_id":       s.ShareId,
// 		"fileId":    s.FileId,
// 		"shareName": s.ShareName,
// 		"owner":     s.Owner,
// 		"accessors": s.Accessors,
// 		"public":    s.Public,
// 		"wormhole":  s.Wormhole,
// 		"enabled":   s.Enabled,
// 		"expires":   s.Expires,
// 		"shareType": "file",
// 	}
//
// 	return bson.Marshal(data)
// }

func (s *AlbumShare) GetShareId() ShareId      { return s.ShareId }
func (s *AlbumShare) GetShareType() ShareType  { return SharedAlbum }
func (s *AlbumShare) GetItemId() string        { return string(s.AlbumId) }
func (s *AlbumShare) SetItemId(albumId string) { s.AlbumId = AlbumId(albumId) }
func (s *AlbumShare) GetAccessors() []Username { return s.Accessors }

func (s *AlbumShare) AddUsers(usernames []Username) {
	s.Accessors = internal.AddToSet(s.Accessors, usernames...)
}

func (s *AlbumShare) RemoveUsers(usernames []Username) {
	s.Accessors = internal.AddToSet(s.Accessors, usernames...)
}

func (s *AlbumShare) GetOwner() Username { return s.Owner }
func (s *AlbumShare) IsPublic() bool     { return s.Public }
func (s *AlbumShare) SetPublic(pub bool) {
	s.Public = pub
}

func (s *AlbumShare) IsEnabled() bool { return s.Enabled }
func (s *AlbumShare) SetEnabled(enable bool) {
	s.Enabled = enable
}

func (s *AlbumShare) UpdatedNow() {
	s.Updated = time.Now()
}

func (s *AlbumShare) LastUpdated() time.Time {
	return s.Updated
}

func (s *AlbumShare) UnmarshalBSON(bs []byte) error {
	data := map[string]any{}
	err := bson.Unmarshal(bs, &data)
	if err != nil {
		return err
	}

	s.ShareId = ShareId(data["_id"].(string))
	s.AlbumId = AlbumId(data["albumId"].(string))
	s.Owner = Username(data["owner"].(string))

	s.Accessors = internal.Map(
		internal.SliceConvert[string](data["accessors"].(primitive.A)), func(un string) Username {
			return Username(un)
		},
	)

	s.Public = data["public"].(bool)
	s.Enabled = data["enabled"].(bool)
	s.Expires = data["expires"].(primitive.DateTime).Time()

	return nil
}

func (s *AlbumShare) MarshalBSON() ([]byte, error) {
	data := map[string]any{
		"_id":       s.ShareId,
		"albumId":   s.AlbumId,
		"owner":     s.Owner,
		"accessors": s.Accessors,
		"public":    s.Public,
		"enabled":   s.Enabled,
		"expires":   s.Expires,
		"shareType": "album",
	}

	return bson.Marshal(data)
}

type Share interface {
	GetShareId() ShareId
	GetShareType() ShareType
	GetItemId() string
	IsPublic() bool
	IsEnabled() bool
	GetAccessors() []Username
	GetOwner() Username
	AddUsers(usernames []Username)
	RemoveUsers(usernames []Username)

	SetItemId(string)
	SetPublic(bool)
	SetEnabled(bool)

	UpdatedNow()
	LastUpdated() time.Time
}

type ShareService interface {
	Size() int

	Get(id ShareId) Share
	Add(share Share) error
	Del(id ShareId) error
	AddUsers(share Share, newUsers []*User) error
	RemoveUsers(share Share, newUsers []*User) error

	GetAllShares() []Share
	GetFileSharesWithUser(u *User) ([]*FileShare, error)
	GetAlbumSharesWithUser(u *User) ([]*AlbumShare, error)

	GetFileShare(f *fileTree.WeblensFileImpl) (*FileShare, error)
}

type ShareType string

const (
	SharedFile  ShareType = "file"
	SharedAlbum ShareType = "album"
)
