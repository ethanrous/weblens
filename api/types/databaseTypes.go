package types

type DatabaseService interface {
	HistoryDbService
	AlbumsDB
	MediaDB
	ShareDB
	UserDB
	InstanceDB
}

type HistoryDbService interface {
	WriteFileEvent(FileEvent) error
	GetAllLifetimes() ([][]FileAction, error)
}

type AlbumsDB interface {
	GetAllAlbums() ([]Album, error)
	RemoveMediaFromAlbum(AlbumId, ContentId) error
	GetAlbumsByMedia(ContentId) ([]Album, error)
	AddMediaToAlbum(aId AlbumId, mIds []ContentId) error
}

type MediaDB interface {
	GetAllMedia() ([]Media, error)
	DeleteMedia(ContentId) error
	HideMedia(ContentId) error
	AddFileToMedia(mId ContentId, fId FileId) error
	RemoveFileFromMedia(mId ContentId, fId FileId) error
}

type ShareDB interface {
	UpdateShare(s Share) error
	GetAllShares() ([]Share, error)
	SetShareEnabledById(sId ShareId, enabled bool) error
}

type UserDB interface {
	GetAllUsers() ([]User, error)
	UpdatePsaswordByUsername(username Username, newPasswordHash string) error
	SetAdminByUsername(username Username, isAdmin bool) error
}

type InstanceDB interface {
	GetAllServers() ([]Instance, error)
	NewServer(id InstanceId, name string, isThisServer bool, role ServerRole) error
}
