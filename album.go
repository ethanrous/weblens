package weblens

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/modern-go/reflect2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumId string

type Album struct {
	Id             AlbumId     `bson:"_id"`
	Name           string      `bson:"name"`
	Owner          Username    `bson:"owner"`
	Medias         []ContentId `bson:"medias"`
	Cover          ContentId   `bson:"cover"`
	PrimaryColor   string      `bson:"primaryColor"`
	SecondaryColor string      `bson:"secondaryColor"`
	SharedWith     []Username  `bson:"sharedWith"`
	ShowOnTimeline bool        `bson:"showOnTimeline"`
}

func NewAlbum(albumName string, owner *User) *Album {
	albumId := AlbumId(internal.GlobbyHash(12, fmt.Sprintln(albumName, owner.GetUsername())))
	return &Album{
		Id:             albumId,
		Name:           albumName,
		Owner:      owner.GetUsername(),
		Medias:     []ContentId{},
		ShowOnTimeline: true,
		SharedWith: []Username{},
	}
}

func (a *Album) ID() AlbumId {
	return a.Id
}

func (a *Album) GetName() string {
	return a.Name
}

func (a *Album) AddMedia(ms ...*Media) error {
	if ms == nil {
		return werror.ErrNoMedia
	}

	err := types.SERV.StoreService.AddMediaToAlbum(
		a.Id, internal.Map(
			ms, func(m *Media) ContentId {
				return m.ID()
			},
		),
	)

	if err != nil {
		return err
	}

	a.Medias = append(a.Medias, ms...)

	return err
}

func (a *Album) RemoveMedia(m ...ContentId) werror.WErr {
	for _, m := range m {
		a.Medias, _, _ = internal.YoinkFunc(
			a.Medias, func(f *Media) bool {
				return f.ID() == m
			},
		)
	}

	return werror.NotImplemented("Remove media from album")
}

func (a *Album) Rename(newName string) error {
	return werror.NewWeblensError("not impl - album rename")
}

func (a *Album) setCover(cover ContentId, color1, color2 string) {
	a.Cover = cover
	a.PrimaryColor = color1
	a.PrimaryColor = color2
}

func (a *Album) GetCover() ContentId {
	return a.Cover
}

func (a *Album) GetMedias() []*Media {
	if a.Medias == nil {
		a.Medias = []*Media{}
	}
	return a.Medias
}

func (a *Album) GetOwner() Username {
	if a.Owner == "" {
		wlog.Error.Println("No owner for Album")
	}
	return a.Owner
}

func (a *Album) GetPrimaryColor() string {
	return a.PrimaryColor
}

func (a *Album) AddUsers(us ...*User) error {
	err := types.SERV.StoreService.AddUsersToAlbum(a.ID(), us)
	if err != nil {
		return err
	}

	a.SharedWith = internal.AddToSet(
		a.SharedWith, internal.Map(
			us, func(u *User) Username {
				return u.GetUsername()
			},
		)...,
	)

	return nil
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
	return werror.NewWeblensError("not impl - album remove users")
}

func (a *Album) UnmarshalBSON(bs []byte) error {

	var data map[string]any
	if err := bson.Unmarshal(bs, &data); err != nil {
		return err
	}

	a.Id = AlbumId(data["_id"].(string))
	a.Name = data["name"].(string)
	// a.Owner = Username(data["owner"].(string))

	a.Medias = internal.FilterMap(
		data["medias"].(primitive.A), func(mId any) (*Media, bool) {
			m := types.SERV.MediaRepo.Get(ContentId(mId.(string)))
			if reflect2.IsNil(m) {
				// util.Error.Printf("Nil media [%s] while loading album (%s)", mId, a.Name)
				return nil, false
			}
			return m, true
		},
	)

	a.Cover = ContentId(data["cover"].(string))
	a.PrimaryColor = data["primaryColor"].(string)
	a.SecondaryColor = data["secondaryColor"].(string)
	a.SharedWith = internal.Map(
		data["sharedWith"].(primitive.A), func(username any) Username {
			return Username(username.(string))
		},
	)
	a.ShowOnTimeline = data["showOnTimeline"].(bool)

	return nil
}

func (a *Album) MarshalJSON() ([]byte, error) {
	data := map[string]any{}

	data["id"] = a.Id
	data["name"] = a.Name
	data["owner"] = a.Owner
	data["medias"] = internal.FilterMap(
		a.Medias, func(m *Media) (ContentId, bool) {
			if reflect2.IsNil(m) {
				return "", false
			}
			return m.ID(), true
		},
	)

	data["cover"] = ""
	data["cover"] = a.Cover
	data["primaryColor"] = a.PrimaryColor
	data["secondaryColor"] = a.SecondaryColor
	data["sharedWith"] = a.SharedWith
	data["showOnTimeline"] = a.ShowOnTimeline

	return json.Marshal(data)
}

func (a *Album) MarshalBSON() ([]byte, error) {
	data := map[string]any{}

	data["_id"] = a.Id
	data["name"] = a.Name
	data["owner"] = a.Owner
	data["medias"] = internal.Map(
		a.Medias, func(m *Media) ContentId {
			return m.ID()
		},
	)
	data["cover"] = ""
	data["cover"] = a.Cover

	data["primaryColor"] = a.PrimaryColor
	data["secondaryColor"] = a.SecondaryColor
	data["sharedWith"] = a.SharedWith
	data["showOnTimeline"] = a.ShowOnTimeline

	return bson.Marshal(data)
}

type AlbumService interface {
	Init(store types.AlbumsStore) error
	Size() int
	Get(id AlbumId) *Album
	Add(album *Album) error
	Del(id AlbumId) error
	GetAllByUser(u User) []*Album
	RemoveMediaFromAny(ContentId) error
	SetAlbumCover(albumId AlbumId, cover Media) error
}
