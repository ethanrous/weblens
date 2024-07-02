package album

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Album struct {
	Id             types.AlbumId `bson:"_id"`
	Name           string        `bson:"name"`
	Owner          types.User    `bson:"owner"`
	Medias         []types.Media `bson:"medias"`
	Cover          types.Media   `bson:"cover"`
	PrimaryColor   string        `bson:"primaryColor"`
	SecondaryColor string        `bson:"secondaryColor"`
	SharedWith     []types.User  `bson:"sharedWith"`
	ShowOnTimeline bool          `bson:"showOnTimeline"`
}

func New(albumName string, owner types.User) types.Album {
	albumId := types.AlbumId(util.GlobbyHash(12, fmt.Sprintln(albumName, owner.GetUsername())))
	return &Album{
		Id:             albumId,
		Name:           albumName,
		Owner:          owner,
		Medias:         []types.Media{},
		ShowOnTimeline: true,
	}
}

func (a *Album) ID() types.AlbumId {
	return a.Id
}

func (a *Album) GetName() string {
	return a.Name
}

func (a *Album) AddMedia(ms ...types.Media) error {
	if ms == nil {
		return types.ErrNoMedia
	}

	err := types.SERV.Database.AddMediaToAlbum(
		a.Id, util.Map(
			ms, func(m types.Media) types.ContentId {
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

func (a *Album) RemoveMedia(m ...types.ContentId) error {
	for _, m := range m {
		a.Medias, _, _ = util.YoinkFunc(
			a.Medias, func(f types.Media) bool {
				return f.ID() == m
			},
		)
	}

	util.ShowErr(types.NewWeblensError("not impl"))

	return nil
}

func (a *Album) Rename(newName string) error {
	return types.NewWeblensError("not impl - album rename")
}

func (a *Album) SetCover(cover types.Media) error {
	colors, err := cover.GetProminentColors()
	if err != nil {
		return err
	}

	err = types.SERV.Database.SetAlbumCover(a.Id, colors[0], colors[1], cover.ID())
	if err != nil {
		return err
	}

	a.Cover = cover
	return nil
}

func (a *Album) GetCover() types.Media {
	return a.Cover
}

func (a *Album) GetMedias() []types.Media {
	if a.Medias == nil {
		a.Medias = []types.Media{}
	}
	return a.Medias
}

func (a *Album) GetOwner() types.User {
	return a.Owner
}

func (a *Album) GetPrimaryColor() string {
	return a.PrimaryColor
}

func (a *Album) AddUsers(us ...types.User) error {
	err := types.SERV.Database.AddUsersToAlbum(a.ID(), us)
	if err != nil {
		return err
	}

	a.SharedWith = util.AddToSet(a.SharedWith, us)

	return nil
}

func (a *Album) GetUsers() []types.User {
	return a.SharedWith
}

func (a *Album) RemoveUsers(uns ...types.Username) error {
	a.SharedWith = util.Filter(
		a.SharedWith, func(u types.User) bool {
			return !slices.Contains(uns, u.GetUsername())
		},
	)
	return types.NewWeblensError("not impl - album remove users")
}

func (a *Album) UnmarshalBSON(bs []byte) error {

	var data map[string]any
	if err := bson.Unmarshal(bs, &data); err != nil {
		return err
	}

	a.Id = types.AlbumId(data["_id"].(string))
	a.Name = data["name"].(string)
	a.Owner = types.SERV.UserService.Get(types.Username(data["owner"].(string)))

	a.Medias = util.FilterMap(
		data["medias"].(primitive.A), func(mId any) (types.Media, bool) {
			m := types.SERV.MediaRepo.Get(types.ContentId(mId.(string)))
			if m == nil {
				// util.Error.Printf("Nil media [%s] while loading album (%s)", mId, a.Name)
				return nil, false
			}
			return m, true
		},
	)

	a.Cover = types.SERV.MediaRepo.Get(types.ContentId(data["cover"].(string)))
	a.PrimaryColor = data["primaryColor"].(string)
	a.SecondaryColor = data["secondaryColor"].(string)
	a.SharedWith = util.Map(
		data["sharedWith"].(primitive.A), func(un any) types.User {
			return types.SERV.UserService.Get(types.Username(un.(string)))
		},
	)
	a.ShowOnTimeline = data["showOnTimeline"].(bool)

	return nil
}

func (a *Album) MarshalJSON() ([]byte, error) {
	data := map[string]any{}

	data["id"] = a.Id
	data["name"] = a.Name
	data["owner"] = a.Owner.GetUsername()
	data["medias"] = util.Map(
		a.Medias, func(m types.Media) types.ContentId {
			return m.ID()
		},
	)
	data["cover"] = ""
	if a.Cover != nil {
		data["cover"] = a.Cover.ID()
	}
	data["primaryColor"] = a.PrimaryColor
	data["secondaryColor"] = a.SecondaryColor
	data["sharedWith"] = util.Map(
		a.SharedWith, func(u types.User) types.Username {
			return u.GetUsername()
		},
	)
	data["showOnTimeline"] = a.ShowOnTimeline

	return json.Marshal(data)
}

func (a *Album) MarshalBSON() ([]byte, error) {
	data := map[string]any{}

	data["_id"] = a.Id
	data["name"] = a.Name
	data["owner"] = a.Owner.GetUsername()
	data["medias"] = util.Map(
		a.Medias, func(m types.Media) types.ContentId {
			return m.ID()
		},
	)
	data["cover"] = ""
	if a.Cover != nil {
		data["cover"] = a.Cover.ID()
	}
	data["primaryColor"] = a.PrimaryColor
	data["secondaryColor"] = a.SecondaryColor
	data["sharedWith"] = util.Map(
		a.SharedWith, func(u types.User) types.Username {
			return u.GetUsername()
		},
	)
	data["showOnTimeline"] = a.ShowOnTimeline

	return bson.Marshal(data)
}

// func GetAlbumById(albumId types.AlbumId) (a *AlbumData, err error) {
// 	a, err = dbServer.GetAlbumById(albumId)
// 	return
// }
//
// func DeleteAlbum(albumId types.AlbumId) (err error) {
// 	err = dbServer.DeleteAlbum(albumId)
// 	return
// }
//
// func (a *AlbumData) CanUserAccess(username types.Username) bool {
// 	return a.Owner == username || slices.Contains(a.SharedWith, username)
// }
//
// func (a *AlbumData) AddMedia(ms []types.Media) (addedCount int, err error) {
// 	mIds := util.Map(ms, func(m types.Media) types.ContentId { return m.ID() })
// 	a.Medias = util.AddToSet(a.Medias, mIds)
// 	addedCount, err = dbServer.addMediaToAlbum(a.Id, mIds)
// 	return
// }
//
// func (a *AlbumData) Rename(newName string) (err error) {
// 	err = dbServer.setAlbumName(a.Id, newName)
// 	return
// }
//
// func (a *AlbumData) SetCover(mediaId types.ContentId, ft types.FileTree) error {
// 	m := a. media.MediaMapGet(mediaId)
// 	if m == nil {
// 		return ErrNoMedia
// 	}
// 	colors, err := m.(*media.media).getProminentColors(ft)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = dbServer.SetAlbumCover(a.Id, mediaId, colors[0], colors[1])
// 	if err != nil {
// 		return fmt.Errorf("failed to set album cover")
// 	}
// 	return nil
// }
//
// func (a *AlbumData) AddUsers(usernames []types.Username) (err error) {
// 	a.SharedWith = util.AddToSet(a.SharedWith, usernames)
// 	err = dbServer.shareAlbum(a.Id, usernames)
// 	return
// }
//
// func (a *AlbumData) RemoveUsers(usernames []types.Username) (err error) {
// 	a.SharedWith = util.Filter(a.SharedWith, func(s types.Username) bool { return !slices.Contains(usernames, s) })
// 	err = dbServer.unshareAlbum(a.Id, usernames)
// 	return
// }
