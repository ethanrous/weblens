package models

import (
	"fmt"
	"iter"
	"slices"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/internal/log"
)

type AlbumId string

type Album struct {
	Id             AlbumId     `bson:"_id" json:"id"`
	Name           string      `bson:"name" json:"name"`
	Owner          Username    `bson:"owner" json:"owner"`
	Medias         []ContentId `bson:"medias" json:"medias"`
	Cover          ContentId   `bson:"cover" json:"cover"`
	PrimaryColor   string      `bson:"primaryColor" json:"primaryColor"`
	SecondaryColor string      `bson:"secondaryColor" json:"secondaryColor"`
	SharedWith     []Username  `bson:"sharedWith" json:"sharedWith"`
	ShowOnTimeline bool        `bson:"showOnTimeline" json:"showOnTimeline"`
}

func NewAlbum(albumName string, owner *User) *Album {
	albumId := AlbumId(internal.GlobbyHash(12, fmt.Sprintln(albumName, owner.GetUsername())))
	return &Album{
		Id:             albumId,
		Name:           albumName,
		Owner:          owner.GetUsername(),
		Medias:         []ContentId{},
		ShowOnTimeline: true,
		SharedWith:     []Username{},
	}
}

func (a *Album) ID() AlbumId {
	return a.Id
}

func (a *Album) GetName() string {
	return a.Name
}

func (a *Album) RemoveMedia(toRemoveIds ...ContentId) {
	a.Medias = internal.Filter(
		a.Medias, func(mediaId ContentId) bool {
			return !slices.Contains(toRemoveIds, mediaId)
		},
	)
}

func (a *Album) Rename(newName string) error {
	return werror.New("not impl - album rename")
}

func (a *Album) SetCover(cover ContentId, color1, color2 string) {
	a.Cover = cover
	a.PrimaryColor = color1
	a.PrimaryColor = color2
}

func (a *Album) GetCover() ContentId {
	return a.Cover
}

func (a *Album) GetMedias() []ContentId {
	return a.Medias
}

func (a *Album) GetOwner() Username {
	if a.Owner == "" {
		log.Error.Println("No owner for Album")
	}
	return a.Owner
}

func (a *Album) GetPrimaryColor() string {
	return a.PrimaryColor
}

func (a *Album) GetSharedWith() []Username {
	if a.SharedWith == nil {
		a.SharedWith = []Username{}
	}

	return a.SharedWith
}

func (a *Album) RemoveUsers(uns ...Username) error {
	a.SharedWith = internal.Filter(
		a.SharedWith, func(un Username) bool {
			return !slices.Contains(uns, un)
		},
	)
	return werror.New("not impl - album remove users")
}

// func (a *Album) UnmarshalBSON(bs []byte) error {
//
// 	var data map[string]any
// 	if err := bson.Unmarshal(bs, &data); err != nil {
// 		return err
// 	}
//
// 	a.Id = AlbumId(data["_id"].(string))
// 	a.Name = data["name"].(string)
// 	// a.Owner = Username(data["owner"].(string))
//
// 	a.Medias = internal.SliceConvert[ContentId](data["medias"].(primitive.A))
//
// 	a.Cover = ContentId(data["cover"].(string))
// 	a.PrimaryColor = data["primaryColor"].(string)
// 	a.SecondaryColor = data["secondaryColor"].(string)
// 	a.SharedWith = internal.Map(
// 		data["sharedWith"].(primitive.A), func(username any) Username {
// 			return Username(username.(string))
// 		},
// 	)
// 	a.ShowOnTimeline = data["showOnTimeline"].(bool)
//
// 	return nil
// }

// func (a *Album) MarshalJSON() ([]byte, error) {
// 	data := map[string]any{}
//
// 	data["id"] = a.Id
// 	data["name"] = a.Name
// 	data["owner"] = a.Owner
// 	data["medias"] = a.Medias
//
// 	data["cover"] = ""
// 	data["cover"] = a.Cover
// 	data["primaryColor"] = a.PrimaryColor
// 	data["secondaryColor"] = a.SecondaryColor
// 	data["sharedWith"] = a.SharedWith
// 	data["showOnTimeline"] = a.ShowOnTimeline
//
// 	return json.Marshal(data)
// }
//
// func (a *Album) MarshalBSON() ([]byte, error) {
// 	data := map[string]any{}
//
// 	data["_id"] = a.Id
// 	data["name"] = a.Name
// 	data["owner"] = a.Owner
// 	data["medias"] = a.Medias
// 	data["cover"] = ""
// 	data["cover"] = a.Cover
//
// 	data["primaryColor"] = a.PrimaryColor
// 	data["secondaryColor"] = a.SecondaryColor
// 	data["sharedWith"] = a.SharedWith
// 	data["showOnTimeline"] = a.ShowOnTimeline
//
// 	return bson.Marshal(data)
// }

type AlbumService interface {
	Init() error
	Size() int
	Get(id AlbumId) *Album
	Add(album *Album) error
	Del(id AlbumId) error
	GetAllByUser(u *User) []*Album

	GetAlbumMedias(album *Album) iter.Seq[*Media]

	SetAlbumCover(albumId AlbumId, cover *Media) error
	AddUsersToAlbum(album *Album, us ...*User) error
	AddMediaToAlbum(album *Album, media ...*Media) error
	RemoveMediaFromAlbum(album *Album, mediaIds ...ContentId) error

	RemoveMediaFromAny(ContentId) error
}
