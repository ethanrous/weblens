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

// FileInfo is a structure for safely sending file information to the client.
type FileInfo struct {
	Id           string           `json:"id"`
	PortablePath string           `json:"portablePath"`
	Filename     string           `json:"filename"`
	ParentId     string           `json:"parentId"`
	ContentId    models.ContentId `json:"contentId"`
	Owner        models.Username  `json:"owner"`
	ShareId      models.ShareId   `json:"shareId,omitempty"`
	PastId       string           `json:"currentId"`
	Children     []string         `json:"childrenIds"`
	Size         int64            `json:"size"`
	ModTime      int64            `json:"modifyTimestamp"`
	IsDir        bool             `json:"isDir"`
	Modifiable   bool             `json:"modifiable"`
	PastFile     bool             `json:"pastFile"`
} // @name FileInfo

func WeblensFileToFileInfo(f *fileTree.WeblensFileImpl, pack *models.ServicePack, isPastFile bool) (FileInfo, error) {
	// Some fields are only needed if the file is the parent file of the request,
	// when the file is a child, these fields are not needed, and can be expensive to fetch,
	// so we conditionally ignore them.
	var ownerName models.Username
	var children []fileTree.FileId
	owner := pack.FileService.GetFileOwner(f)
	if owner == nil {
		return FileInfo{}, werror.WithStack(werror.ErrNoUser)
	}
	ownerName = owner.GetUsername()

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

	modifiable := !isPastFile && !pack.FileService.IsFileInTrash(f)

	return FileInfo{
		Id:           f.ID(),
		PortablePath: f.GetPortablePath().ToPortable(),
		Filename:     f.Filename(),
		Size:         f.Size(),
		IsDir:        f.IsDir(),
		ModTime:      f.ModTime().UnixMilli(),
		ParentId:     f.GetParentId(),
		ContentId:    f.GetContentId(),
		ShareId:      shareId,
		Modifiable:   modifiable,
		PastFile:     isPastFile,
		PastId:       f.GetPastId(),

		Owner:    ownerName,
		Children: children,
	}, nil
}

type FolderInfoResponse struct {
	Children []FileInfo  `json:"children"`
	Parents  []FileInfo  `json:"parents"`
	Medias   []MediaInfo `json:"medias"`
	Self     FileInfo    `json:"self"`
} // @name FolderInfo

type MediaBatchInfo struct {
	Media      []MediaInfo `json:"Media"`
	MediaCount int         `json:"mediaCount"`
} // @name MediaBatchInfo

func NewMediaBatchInfo(m []*models.Media) MediaBatchInfo {
	if len(m) == 0 {
		return MediaBatchInfo{
			Media:      []MediaInfo{},
			MediaCount: 0,
		}
	}
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
	TaskId    string `json:"taskId"`
	Filename  string `json:"filename"`
	Single    bool   `json:"single"`
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

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress"`

	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	ReportedRole models.ServerRole `json:"reportedRole"`

	LastBackup int64 `json:"lastBackup"`

	BackupSize int64 `json:"backupSize"`

	UserCount int `json:"userCount"`

	// If this server info represents this local server
	IsThisServer bool `json:"-"`

	Online bool `json:"online"`

	Started bool `json:"started"`
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
	HomeId    string          `json:"homeId"`
	TrashId   string          `json:"trashId"`
	Token     string          `json:"token" omitEmpty:"true"`
	HomeSize  int64           `json:"homeSize"`
	TrashSize int64           `json:"trashSize"`
	Admin     bool            `json:"admin"`
	Owner     bool            `json:"owner"`
} // @name UserInfo

type UserInfoArchive struct {
	Password string `json:"password" omitEmpty:"true"`
	UserInfo
	Activated bool `json:"activated"`
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
	RemoteUsing models.InstanceId    `json:"remoteUsing"`
	CreatedBy   models.InstanceId    `json:"createdBy"`
	CreatedTime int64                `json:"createdTime"`
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
	ActionType      string `json:"actionType" validate:"required"`
	OriginPath      string `json:"originPath" validate:"required"`
	DestinationPath string `json:"destinationPath" validate:"required"`
	LifeId          string `json:"lifeId" validate:"required"`
	EventId         string `json:"eventId" validate:"required"`
	ParentId        string `json:"parentId" validate:"required"`
	ServerId        string `json:"serverId" validate:"required"`
	Timestamp       int64  `json:"timestamp" validate:"required"`
	Size            int64  `json:"size" validate:"required"`
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
}

type MediaInfo struct {
	MediaId string `json:"-" example:"5f9b3b3b7b4f3b0001b3b3b7"`

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentId string `json:"contentId"`

	// User who owns the file that resulted in this media being created
	Owner string `json:"owner"`

	// Mime-type key of the media
	MimeType string `json:"mimeType"`

	// Slices of files whos content hash to the contentId
	FileIds []string `json:"fileIds"`

	// Tags from the ML image scan so searching for particular objects in the images can be done
	RecognitionTags []string `json:"recognitionTags"`

	LikedBy []string `json:"likedBy,omitempty"`

	CreateDate int64 `json:"createDate"`

	// Full-res image dimensions
	Width  int `json:"width"`
	Height int `json:"height"`

	// Number of pages (typically 1, 0 in not a valid page count)
	PageCount int `json:"pageCount"`

	// Total time, in milliseconds, of a video
	Duration int `json:"duration"`

	// If the media is hidden from the timeline
	// TODO - make this per user
	Hidden bool `json:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `json:"enabled"`

	Imported bool `json:"imported"`
} // @Name MediaInfo

func MediaToMediaInfo(m *models.Media) MediaInfo {
	return MediaInfo{
		MediaId:         m.MediaID.Hex(),
		ContentId:       m.ContentID,
		FileIds:         m.FileIDs,
		CreateDate:      m.CreateDate.UnixMilli(),
		Owner:           m.Owner,
		Width:           m.Width,
		Height:          m.Height,
		PageCount:       m.PageCount,
		Duration:        m.Duration,
		MimeType:        m.MimeType,
		RecognitionTags: m.GetRecognitionTags(),
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
	ShareType string   `json:"shareType"`
	Accessors []string `json:"accessors"`
	Expires   int64    `json:"expires"`
	Updated   int64    `json:"updated"`
	Public    bool     `json:"public"`
	Wormhole  bool     `json:"wormhole"`
	Enabled   bool     `json:"enabled"`
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
	Cover          string   `json:"cover"`
	PrimaryColor   string   `json:"primaryColor"`
	SecondaryColor string   `json:"secondaryColor"`
	Medias         []string `json:"medias"`
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
