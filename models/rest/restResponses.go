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
} // @name ErrorInfo

// FileInfo is a structure for safely sending file information to the client
type FileInfo struct {
	Id           string           `json:"id"`
	PortablePath string           `json:"portablePath"`
	Filename     string           `json:"filename"`
	Size         int64            `json:"size"`
	IsDir        bool             `json:"isDir"`
	ModTime      int64            `json:"modifyTimestamp"`
	ParentId     string           `json:"parentId"`
	Children     []string         `json:"childrenIds"`
	ContentId    models.ContentId `json:"contentId"`
	Owner        models.Username  `json:"owner"`
	ShareId      models.ShareId   `json:"shareId,omitempty"`
	Modifiable   bool             `json:"modifiable"`
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

	share, err := pack.ShareService.GetFileShare(f.ID())
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
	Self     FileInfo    `json:"self"`
	Children []FileInfo  `json:"children"`
	Parents  []FileInfo  `json:"parents"`
	Medias   []MediaInfo `json:"medias"`
} // @name FolderInfo

type MediaBatchInfo struct {
	Media      []MediaInfo `json:"Media"`
	MediaCount int         `json:"mediaCount"`
} // @name MediaBatchInfo

func NewMediaBatchInfo(m []*models.Media) MediaBatchInfo {
	var mediaInfos []MediaInfo
	for _, media := range m {
		mediaInfos = append(mediaInfos, MediaToMediaInfo(media))
	}
	return MediaBatchInfo{
		Media:      mediaInfos,
		MediaCount: len(m),
	}
}

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

func ServerInfoToInstance(si ServerInfo) *models.Instance {
	return &models.Instance{
		Id:           si.Id,
		Name:         si.Name,
		Role:         si.Role,
		IsThisServer: si.IsThisServer,
		Address:      si.Address,
		LastBackup:   si.LastBackup,
	}
}

type UserInfo struct {
	Username  models.Username `json:"username"`
	Admin     bool            `json:"admin"`
	Owner     bool            `json:"owner"`
	HomeId    string          `json:"homeId"`
	TrashId   string          `json:"trashId"`
	HomeSize  int64           `json:"homeSize"`
	TrashSize int64           `json:"trashSize"`
	Token     string          `json:"token" omitEmpty:"true"`
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

func UserInfoArchiveToUser(uInfo UserInfoArchive) *models.User {
	u := &models.User{
		Username:      uInfo.Username,
		Password:      uInfo.Password,
		Activated:     uInfo.Activated,
		Admin:         uInfo.Admin,
		IsServerOwner: uInfo.Owner,
		HomeId:        uInfo.HomeId,
		TrashId:       uInfo.TrashId,
	}

	return u
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

type FileActionInfo struct {
	Timestamp       int64  `json:"timestamp"`
	ActionType      string `json:"actionType"`
	OriginPath      string `json:"originPath"`
	DestinationPath string `json:"destinationPath"`
	LifeId          string `json:"lifeId"`
	EventId         string `json:"eventId"`
	Size            int64  `json:"size"`
	ParentId        string `json:"parentId"`
	ServerId        string `json:"serverId"`
} // @name FileActionInfo

func FileActionToFileActionInfo(fa *fileTree.FileAction) FileActionInfo {
	return FileActionInfo{
		Timestamp:       fa.Timestamp.UnixMilli(),
		ActionType:      fa.ActionType,
		OriginPath:      fa.OriginPath,
		DestinationPath: fa.DestinationPath,
		LifeId:          fa.LifeId,
		EventId:         fa.EventId,
		Size:            fa.Size,
		ParentId:        fa.ParentId,
		ServerId:        fa.ServerId,
	}
} // @name FileActionInfo

type MediaInfo struct {
	MediaId string `json:"-" example:"5f9b3b3b7b4f3b0001b3b3b7"`

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentId string `json:"contentId"`

	// Slices of files whos content hash to the contentId
	FileIds []string `json:"fileIds"`

	CreateDate int64 `json:"createDate"`

	// User who owns the file that resulted in this media being created
	Owner string `json:"owner"`

	// Full-res image dimensions
	Width  int `json:"width"`
	Height int `json:"height"`

	// Number of pages (typically 1, 0 in not a valid page count)
	PageCount int `json:"pageCount"`

	// Total time, in milliseconds, of a video
	Duration int `json:"duration"`

	// Mime-type key of the media
	MimeType string `json:"mimeType"`

	// Tags from the ML image scan so searching for particular objects in the images can be done
	RecognitionTags []string `json:"recognitionTags"`

	// If the media is hidden from the timeline
	// TODO - make this per user
	Hidden bool `json:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `json:"enabled"`

	LikedBy []string `json:"likedBy,omitempty"`

	Imported bool `json:"imported"`
} // @Name MediaInfo

func MediaToMediaInfo(m *models.Media) MediaInfo {
	return MediaInfo{
		MediaId:         m.MediaId.Hex(),
		ContentId:       m.ContentId,
		FileIds:         m.FileIds,
		CreateDate:      m.CreateDate.UnixMilli(),
		Owner:           m.Owner,
		Width:           m.Width,
		Height:          m.Height,
		PageCount:       m.PageCount,
		Duration:        m.Duration,
		MimeType:        m.MimeType,
		RecognitionTags: m.RecognitionTags,
		Hidden:          m.Hidden,
		Enabled:         m.Enabled,
		LikedBy:         m.LikedBy,
		Imported:        m.IsImported(),
	}
}

type ShareInfo struct {
	ShareId   string   `json:"shareId"`
	FileId    string   `json:"fileId"`
	ShareName string   `json:"shareName"`
	Owner     string   `json:"owner"`
	Accessors []string `json:"accessors"`
	Public    bool     `json:"public"`
	Wormhole  bool     `json:"wormhole"`
	Enabled   bool     `json:"enabled"`
	Expires   int64    `json:"expires"`
	Updated   int64    `json:"updated"`
	ShareType string   `json:"shareType"`
} // @name ShareInfo

func ShareToShareInfo(s *models.FileShare) ShareInfo {
	return ShareInfo{
		ShareId:   s.ID(),
		FileId:    s.FileId,
		ShareName: s.ShareName,
		Owner:     s.GetOwner(),
		Accessors: s.GetAccessors(),
		Public:    s.IsPublic(),
		Wormhole:  s.IsWormhole(),
		Enabled:   s.IsEnabled(),
		Expires:   s.Expires.UnixMilli(),
		Updated:   s.Updated.UnixMilli(),
		ShareType: s.GetShareType(),
	}
}

type NewUploadInfo struct {
	UploadId string `json:"uploadId"`
} // @name NewUploadInfo

type NewFileInfo struct {
	FileId string `json:"fileId"`
} // @name NewFileInfo

type NewFilesInfo struct {
	FileIds []string `json:"fileIds"`
} // @name NewFilesInfo

type AlbumInfo struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Owner          string   `json:"owner"`
	Medias         []string `json:"medias"`
	Cover          string   `json:"cover"`
	PrimaryColor   string   `json:"primaryColor"`
	SecondaryColor string   `json:"secondaryColor"`
	ShowOnTimeline bool     `json:"showOnTimeline"`
} // @name AlbumInfo

func AlbumToAlbumInfo(a *models.Album) AlbumInfo {
	return AlbumInfo{
		Id:             a.Id,
		Name:           a.Name,
		Owner:          a.Owner,
		Medias:         a.Medias,
		Cover:          a.Cover,
		PrimaryColor:   a.PrimaryColor,
		SecondaryColor: a.SecondaryColor,
		ShowOnTimeline: a.ShowOnTimeline,
	}
}
