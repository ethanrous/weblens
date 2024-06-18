package share

import (
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

	shareService *shareService
}

func NewFileShare(f types.WeblensFile, u types.User, accessors []types.User, public bool, wormhole bool) types.Share {
	return &FileShare{
		FileId:    f.ID(),
		Owner:     u,
		Accessors: accessors,
		Public:    public,
		Wormhole:  wormhole,
	}
}

func (s *FileShare) GetShareId() types.ShareId     { return s.ShareId }
func (s *FileShare) GetShareType() types.ShareType { return types.FileShare }
func (s *FileShare) GetContentId() string          { return s.FileId.String() }
func (s *FileShare) SetItemId(fileId string)       { s.FileId = types.FileId(fileId) }
func (s *FileShare) GetAccessors() []types.User    { return s.Accessors }
func (s *FileShare) SetAccessors(newUsers []types.User, c ...types.BroadcasterAgent) {
	userDiff := util.Diff(s.Accessors, newUsers)
	s.Accessors = newUsers
	for _, u := range userDiff {
		util.Each(c, func(caster types.BroadcasterAgent) {
			caster.PushShareUpdate(u.GetUsername(), s)
		})
	}
}
func (s *FileShare) GetOwner() types.User { return s.Owner }
func (s *FileShare) IsPublic() bool       { return s.Public }
func (s *FileShare) SetPublic(pub bool)   { s.Public = pub }

func (s *FileShare) IsEnabled() bool { return s.Enabled }
func (s *FileShare) SetEnabled(enable bool) error {
	s.Enabled = enable
	return s.shareService.db.SetShareEnabledById(s.ShareId, enable)
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

	s.Accessors = util.Map(util.SliceConvert[string](data["accessors"].(primitive.A)), func(un string) types.User {
		return types.SERV.UserService.Get(types.Username(un))
	})

	s.Public = data["public"].(bool)
	s.Wormhole = data["wormhole"].(bool)
	s.Enabled = data["enabled"].(bool)
	s.Expires = data["expires"].(primitive.DateTime).Time()
	s.ShareType = types.ShareType(data["shareType"].(string))

	return nil
}

// LoadAllShares should only be called once per execution of weblens, on initialization
// func LoadAllShares(ft types.FileTree) {
// 	shares, err := dataStore.dbServer.getAllShares()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	for _, s := range shares {
// 		if s.GetShareType() != dataStore.FileShare {
// 			continue
// 		}
//
// 		fs := s.(*FileShare)
// 		file := ft.Get(fs.FileId)
// 		if file == nil {
// 			err = dataStore.dbServer.removeFileShare(fs.ShareId)
// 			if err != nil {
// 				panic(err)
// 			}
// 			continue
// 		} else if !dataStore.IsFileInTrash(file) && !s.IsEnabled() {
// 			s.SetEnabled(true)
// 			err = UpdateFileShare(s, ft)
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 		file.AppendShare(fs)
// 	}
// }

// func CreateFileShare(file types.WeblensFile, owner types.Username, users []types.Username, public, wormhole bool, c ...types.BroadcasterAgent) (newShare types.Share, err error) {
// 	shareId := types.ShareId(util.GlobbyHash(8, file.ID(), public))
//
// 	newShare = &FileShare{
// 		ShareId:   shareId,
// 		FileId:    file.ID(),
// 		ShareName: file.Filename(), // This is temporary, we want to be able to rename shares for obfuscation
// 		Owner:     owner,
// 		Accessors: users,
// 		Public:    public || wormhole,
// 		Enabled:   true,
// 		Wormhole:  wormhole,
// 		Expires:   time.Unix(0, 0),
// 		ShareType: dataStore.FileShare,
// 	}
// 	file.AppendShare(newShare)
// 	err = dataStore.dbServer.newFileShare(*newShare.(*FileShare))
// 	if err != nil {
// 		return
// 	}
// 	util.Each(c, func(caster types.BroadcasterAgent) {
// 		caster.PushFileUpdate(file)
// 	})
// 	return
// }

// func UpdateFileShare(s types.Share, ft types.FileTree) (err error) {
// 	switch s.(type) {
// 	case *FileShare:
// 	default:
// 		return dataStore.ErrBadShareType
// 	}
// 	sObj := s.(*FileShare)
//
// 	file := ft.Get(sObj.FileId)
// 	if file == nil {
// 		return dataStore.ErrNoFile
// 	}
// 	err = file.UpdateShare(sObj)
//
// 	return
// }

func GetSharedWithUser(u types.User) []types.Share {
	util.ShowErr(types.NewWeblensError("not impl"))
	return []types.Share{}

	// return dataStore.dbServer.GetSharedWith(u.GetUsername())
}
