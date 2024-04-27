package dataStore

import (
	"errors"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s fileShareData) GetShareId() types.ShareId     { return s.ShareId }
func (s fileShareData) GetShareType() types.ShareType { return FileShare }
func (s fileShareData) GetContentId() string          { return s.FileId.String() }
func (s *fileShareData) SetContentId(fileId string)   { s.FileId = types.FileId(fileId) }
func (s fileShareData) GetAccessors() []types.User {
	return util.Map(s.Accessors, func(un types.Username) types.User { return GetUser(un) })
}
func (s *fileShareData) SetAccessors(newUsers []types.Username) {
	userDiff := util.Diff(s.Accessors, newUsers)
	s.Accessors = newUsers
	for _, u := range userDiff {
		globalCaster.PushShareUpdate(u, s)
	}
	// s.Accessors = util.AddToSet(s.Accessors, newUsers)
}
func (s fileShareData) GetOwner() types.User { return GetUser(s.Owner) }
func (s fileShareData) IsPublic() bool       { return s.Public }
func (s *fileShareData) SetPublic(pub bool)  { s.Public = pub }

func (s fileShareData) IsEnabled() bool         { return s.Enabled }
func (s *fileShareData) SetEnabled(enable bool) { s.Enabled = enable }

// This should only be called once per execution of weblens, on initialization
func LoadAllShares() {
	shares, err := fddb.getAllShares()
	if err != nil {
		panic(err)
	}

	for _, s := range shares {
		if s.GetShareType() != FileShare {
			continue
		}
		fs := s.(*fileShareData)
		file := FsTreeGet(fs.FileId)
		if file == nil {
			fddb.removeFileShare(fs.ShareId)
			continue
		}
		file.AppendShare(fs)
	}
}

func DeleteShare(s types.Share) (err error) {
	switch s.GetShareType() {
	case FileShare:
		err = fddb.removeFileShare(s.GetShareId())
		if err != nil {
			return
		}
		f := FsTreeGet(types.FileId(s.GetContentId()))
		err = f.RemoveShare(s.GetShareId())
		if err != nil {
			return
		}
		globalCaster.PushFileUpdate(f)

	default:
		err = ErrBadShareType
	}

	return
}

func CreateFileShare(file types.WeblensFile, owner types.Username, users []types.Username, public, wormhole bool) (newShare types.Share, err error) {
	shareId := types.ShareId(util.GlobbyHash(8, file.Id(), public))

	newShare = &fileShareData{
		ShareId:   shareId,
		FileId:    file.Id(),
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
	err = fddb.newFileShare(*newShare.(*fileShareData))
	if err == nil {
		globalCaster.PushFileUpdate(file)
	}
	return
}

func UpdateFileShare(s types.Share) (err error) {
	switch s.(type) {
	case *fileShareData:
	default:
		return ErrBadShareType
	}
	sObj := s.(*fileShareData)

	file := FsTreeGet(sObj.FileId)
	if file == nil {
		return ErrNoFile
	}
	err = file.UpdateShare(sObj)

	return
}

func GetShare(shareId types.ShareId, shareType types.ShareType) (s types.Share, err error) {
	switch shareType {
	case FileShare:
		var sObj fileShareData
		sObj, err = fddb.getFileShare(shareId)
		if err == mongo.ErrNoDocuments {
			err = ErrNoShare
		}
		if err != nil {
			return
		}
		file := FsTreeGet(sObj.FileId)
		if file == nil {
			err = ErrNoFile
			return
		}
		return file.GetShare(sObj.ShareId)
	default:
		err = errors.New("unexpected share type")
	}

	if err == mongo.ErrNoDocuments {
		err = ErrNoShare
	}

	return
}
