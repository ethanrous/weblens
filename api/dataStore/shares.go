package dataStore

import (
	"errors"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/mongo"
)

type fileShareData struct {
	ShareId   types.ShareId    `bson:"_id" json:"shareId"`
	FileId    types.FileId     `bson:"fileId" json:"fileId"`
	ShareName string           `bson:"shareName"`
	Owner     types.Username   `bson:"owner"`
	Accessors []types.Username `bson:"accessors"`
	Public    bool             `bson:"public"`
	Wormhole  bool             `bson:"wormhole"`
	Enabled   bool             `bson:"enabled"`
	Expires   time.Time        `bson:"expires"`
	ShareType types.ShareType  `bson:"shareType"`
}

func (s *fileShareData) GetShareId() types.ShareId     { return s.ShareId }
func (s *fileShareData) GetShareType() types.ShareType { return FileShare }
func (s *fileShareData) GetContentId() string          { return s.FileId.String() }
func (s *fileShareData) SetItemId(fileId string)       { s.FileId = types.FileId(fileId) }
func (s *fileShareData) GetAccessors() []types.User {
	return util.Map(s.Accessors, func(un types.Username) types.User { return GetUser(un) })
}
func (s *fileShareData) SetAccessors(newUsers []types.Username, c ...types.BroadcasterAgent) {
	userDiff := util.Diff(s.Accessors, newUsers)
	s.Accessors = newUsers
	for _, u := range userDiff {
		util.Each(c, func(caster types.BroadcasterAgent) {
			caster.PushShareUpdate(u, s)
		})
	}
}
func (s *fileShareData) GetOwner() types.User { return GetUser(s.Owner) }
func (s *fileShareData) IsPublic() bool       { return s.Public }
func (s *fileShareData) SetPublic(pub bool)   { s.Public = pub }

func (s *fileShareData) IsEnabled() bool        { return s.Enabled }
func (s *fileShareData) SetEnabled(enable bool) { s.Enabled = enable }

// LoadAllShares should only be called once per execution of weblens, on initialization
func LoadAllShares(ft types.FileTree) {
	shares, err := dbServer.getAllShares()
	if err != nil {
		panic(err)
	}

	for _, s := range shares {
		if s.GetShareType() != FileShare {
			continue
		}

		fs := s.(*fileShareData)
		file := ft.Get(fs.FileId)
		if file == nil {
			err = dbServer.removeFileShare(fs.ShareId)
			if err != nil {
				panic(err)
			}
			continue
		} else if !IsFileInTrash(file) && !s.IsEnabled() {
			s.SetEnabled(true)
			err = UpdateFileShare(s, ft)
			if err != nil {
				panic(err)
			}
		}
		file.AppendShare(fs)
	}
}

func DeleteShare(s types.Share, ft types.FileTree, c ...types.BroadcasterAgent) (err error) {
	switch s.GetShareType() {
	case FileShare:
		err = dbServer.removeFileShare(s.GetShareId())
		if err != nil {
			return
		}
		f := ft.Get(types.FileId(s.GetContentId()))
		err = f.RemoveShare(s.GetShareId())
		if err != nil {
			return
		}

		util.Each(c, func(caster types.BroadcasterAgent) {
			caster.PushFileUpdate(f)
		})

	default:
		err = ErrBadShareType
	}

	return
}

func CreateFileShare(file types.WeblensFile, owner types.Username, users []types.Username, public, wormhole bool, c ...types.BroadcasterAgent) (newShare types.Share, err error) {
	shareId := types.ShareId(util.GlobbyHash(8, file.ID(), public))

	newShare = &fileShareData{
		ShareId:   shareId,
		FileId:    file.ID(),
		ShareName: file.Filename(), // This is temporary, we want to be able to rename shares for obfuscation
		Owner:     owner,
		Accessors: users,
		Public:    public || wormhole,
		Enabled:   true,
		Wormhole:  wormhole,
		Expires:   time.Unix(0, 0),
		ShareType: FileShare,
	}
	file.AppendShare(newShare)
	err = dbServer.newFileShare(*newShare.(*fileShareData))
	if err != nil {
		return
	}
	util.Each(c, func(caster types.BroadcasterAgent) {
		caster.PushFileUpdate(file)
	})
	return
}

func UpdateFileShare(s types.Share, ft types.FileTree) (err error) {
	switch s.(type) {
	case *fileShareData:
	default:
		return ErrBadShareType
	}
	sObj := s.(*fileShareData)

	file := ft.Get(sObj.FileId)
	if file == nil {
		return ErrNoFile
	}
	err = file.UpdateShare(sObj)

	return
}

func GetShare(shareId types.ShareId, shareType types.ShareType, ft types.FileTree) (s types.Share, err error) {
	switch shareType {
	case FileShare:
		var sObj fileShareData
		sObj, err = dbServer.getFileShare(shareId)
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = ErrNoShare
		}
		if err != nil {
			return
		}
		file := ft.Get(sObj.FileId)
		if file == nil {
			err = ErrNoFile
			return
		}
		return file.GetShare(sObj.ShareId)
	default:
		err = errors.New("unexpected share type")
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		err = ErrNoShare
	}

	return
}

func GetSharedWithUser(u types.User) []types.Share {
	return dbServer.GetSharedWith(u.GetUsername())
}
