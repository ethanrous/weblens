package types

type DatabaseService interface {
	HistoryDbService
	AlbumsDB
	MediaDB
	ShareDB
	UserDB
	InstanceDB
	FilesDB
}

type HistoryDbService interface {
	WriteFileEvent(FileEvent) error
	GetAllLifetimes() ([]Lifetime, error)
	AddOrUpdateLifetime(l Lifetime) error
	GetActionsByPath(WeblensFilepath) ([]FileAction, error)
	DeleteAllFileHistory() error
}

type AlbumsDB interface {
	GetAllAlbums() ([]Album, error)
	CreateAlbum(Album) error

	SetAlbumCover(AlbumId, string, string, ContentId) error
	GetAlbumsByMedia(ContentId) ([]Album, error)

	AddMediaToAlbum(aId AlbumId, mIds []ContentId) error
	RemoveMediaFromAlbum(AlbumId, ContentId) error

	AddUsersToAlbum(aId AlbumId, us []User) error
}

type MediaDB interface {
	CreateMedia(m Media) error
	GetAllMedia() ([]Media, error)
	DeleteMedia(ContentId) error
	HideMedia(ContentId) error
	AddFileToMedia(mId ContentId, fId FileId) error
	RemoveFileFromMedia(mId ContentId, fId FileId) error

	DeleteAllMedia() error
}

type ShareDB interface {
	CreateShare(Share) error
	UpdateShare(Share) error
	GetAllShares() ([]Share, error)
	SetShareEnabledById(sId ShareId, enabled bool) error
	AddUsersToShare(share Share, users []Username) error
	GetSharedWithUser(username Username) ([]Share, error)
}

type UserDB interface {
	GetAllUsers() ([]User, error)
	UpdatePsaswordByUsername(username Username, newPasswordHash string) error
	SetAdminByUsername(Username, bool) error
	CreateUser(User) error
	ActivateUser(Username) error
	AddTokenToUser(username Username, token string) error
	SearchUsers(search string) ([]Username, error)

	DeleteAllUsers() error
}

type InstanceDB interface {
	GetAllServers() ([]Instance, error)
	NewServer(id InstanceId, name string, isThisServer bool, role ServerRole) error
}

type FilesDB interface {
	NewTrashEntry(te TrashEntry) error
	GetTrashEntry(fileId FileId) (TrashEntry, error)
	DeleteTrashEntry(fileId FileId) error
}
