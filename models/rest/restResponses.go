package rest

import (
	"encoding/json"
	"errors"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
)

type Option[T any] struct {
	value T
	set   bool
} // @name

func NewOption[T any](value T) Option[T] {
	return Option[T]{value: value, set: true}
}

func (o Option[T]) Set(v T) Option[T] {
	o.value = v
	o.set = true

	return o
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.set {
		return json.Marshal(o.value)
	}
	return nil, nil
}

type WeblensErrorInfo struct {
	Error string `json:"error"`
} // @name Error

// FileInfo is a structure for safely sending file information to the client
type FileInfo struct {
	Id           fileTree.FileId   `json:"id"`
	PortablePath string            `json:"portablePath"`
	Filename     string            `json:"filename"`
	Size         int64             `json:"size"`
	IsDir        bool              `json:"isDir"`
	ModTime      int64             `json:"modifyTimestamp"`
	ParentId     fileTree.FileId   `json:"parentId"`
	Children     []fileTree.FileId `json:"childrenIds"`
	ContentId    models.ContentId  `json:"contentId"`
	Owner        models.Username   `json:"owner"`
	ShareId      models.ShareId    `json:"shareId,omitempty"`
	Modifiable   bool              `json:"modifiable"`
} // @name FileInfo

func WeblensFileToFileInfo(f *fileTree.WeblensFileImpl, pack *models.ServicePack, isParent bool) (FileInfo, error) {
	// Some fields are only needed if the file is the parent file of the request,
	// when the file is a child, these fields are not needed, and can be expensive to fetch,
	// so we conditionally ignore them.
	var owner models.Username
	var children []fileTree.FileId
	owner = pack.FileService.GetFileOwner(f).GetUsername()
	for _, c := range f.GetChildren() {
		children = append(children, c.ID())
	}

	share, err := pack.ShareService.GetFileShare(f)
	if err != nil && !errors.Is(err, werror.ErrNoShare) {
		return FileInfo{}, err
	}
	var shareId models.ShareId
	if share != nil {
		shareId = share.ID()
	}

	if f.IsDir() && f.GetContentId() == "" {
		_, err = pack.FileService.GetFolderCover(f)
		if err != nil {
			return FileInfo{}, err
		}
	}

	modifiable := !pack.FileService.IsFileInTrash(f)

	return FileInfo{
		Id:           f.ID(),
		PortablePath: f.GetPortablePath().ToPortable(),
		Size:         f.Size(),
		IsDir:        f.IsDir(),
		ModTime:      f.ModTime().UnixMilli(),
		ParentId:     f.GetParentId(),
		ContentId:    f.GetContentId(),
		ShareId:      shareId,
		Modifiable:   modifiable,

		Owner:    owner,
		Children: children,
	}, nil
}

type FolderInfoResponse struct {
	Self     FileInfo        `json:"self"`
	Children []FileInfo      `json:"children"`
	Parents  []FileInfo      `json:"parents"`
	Medias   []*models.Media `json:"medias"`
} // @name FolderInfo

type MediaBatchInfo struct {
	Media      []*models.Media `json:"Media"`
	MediaCount int             `json:"mediaCount"`
} // @name MediaBatchInfo

type TakeoutInfo struct {
	TakeoutId string `json:"takeoutId"`
	Single    bool   `json:"single"`
	TaskId    string `json:"taskId"`
	Filename  string `json:"filename"`
} // @name TakeoutInfo

type DispatchInfo struct {
	TaskId string `json:"taskId"`
} // @name DispatchInfo

type ServerInfo struct {
	Id   models.InstanceId `json:"id"`
	Name string            `json:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey models.WeblensApiKey `json:"-"`

	// Core or Backup
	Role models.ServerRole `json:"role"`

	// If this server info represents this local server
	IsThisServer bool `json:"-"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress"`

	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	ReportedRole models.ServerRole `json:"reportedRole"`

	Online bool `json:"online"`

	LastBackup int64 `json:"lastBackup"`

	BackupSize int64 `json:"backupSize"`

	Started bool `json:"started"`

	UserCount int `json:"userCount"`
} // @name ServerInfo

func InstanceToServerInfo(i *models.Instance) ServerInfo {
	return ServerInfo{
		Id:           i.Id,
		Name:         i.Name,
		Role:         i.Role,
		IsThisServer: i.IsThisServer,
		Address:      i.Address,
		LastBackup:   i.LastBackup,
	}
}

type UserInfo struct {
	Username models.Username `json:"username"`
	Admin    bool            `json:"admin"`
	Owner    bool            `json:"owner"`
	HomeId   fileTree.FileId `json:"homeId"`
	TrashId  fileTree.FileId `json:"trashId"`
} // @name UserInfo

type UserInfoArchive struct {
	UserInfo
	Password  string `json:"password" omitEmpty:"true"`
	Activated bool   `json:"activated"`
} // @name UserInfoArchive

func UserToUserInfo(u *models.User) UserInfo {
	if u == nil || u.IsSystemUser() {
		return UserInfo{}
	}
	info := UserInfo{
		Username: u.GetUsername(),
		Admin:    u.IsAdmin(),
		Owner:    u.IsOwner(),
		HomeId:   u.HomeId,
		TrashId:  u.TrashId,
	}

	return info
}

func UserToUserInfoArchive(u *models.User) UserInfoArchive {
	if u == nil || u.IsSystemUser() {
		return UserInfoArchive{}
	}
	info := UserInfoArchive{
		Password:  u.Password,
		Activated: u.IsActive(),
	}
	info.Username = u.GetUsername()
	info.Admin = u.IsAdmin()
	info.Owner = u.IsOwner()
	info.HomeId = u.HomeId
	info.TrashId = u.TrashId

	return info
}

type ApiKeyInfo struct {
	Id          string               `json:"id"`
	Key         models.WeblensApiKey `json:"key"`
	Owner       models.Username      `json:"owner"`
	CreatedTime int64                `json:"createdTime"`
	RemoteUsing models.InstanceId    `json:"remoteUsing"`
	CreatedBy   models.InstanceId    `json:"createdBy"`
} // @name ApiKeyInfo

func ApiKeyToApiKeyInfo(k models.ApiKey) ApiKeyInfo {
	return ApiKeyInfo{
		Id:          k.Id.Hex(),
		Key:         k.Key,
		Owner:       k.Owner,
		CreatedTime: k.CreatedTime.UnixMilli(),
		RemoteUsing: k.RemoteUsing,
		CreatedBy:   k.CreatedBy,
	}
}

type MediaTypeInfo struct {
	MimeMap map[string]models.MediaType `json:"mimeMap"`
	ExtMap  map[string]models.MediaType `json:"extMap"`
} // @name MediaTypeInfo
