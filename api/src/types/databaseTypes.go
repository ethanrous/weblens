package types

import (
	"context"
	"time"
)

type StoreService interface {
	HistoryStore
	AlbumsStore
	MediaStore
	ShareStore
	UserStore
	InstanceStore
	FilesStore
	AccessStore
}

type ProxyStore interface {
	HistoryStore
	FilesStore
	UserStore
	InstanceStore

	GetLocalStore() StoreService
}

type HistoryStore interface {
	WriteFileEvent(FileEvent) error
	GetAllLifetimes() ([]Lifetime, error)

	// GetLifetimesSince gets all lifetimes that have been updated since the given date
	GetLifetimesSince(time.Time) ([]Lifetime, error)

	UpsertLifetime(l Lifetime) error
	InsertManyLifetimes([]Lifetime) error
	GetActionsByPath(WeblensFilepath) ([]FileAction, error)
	DeleteAllFileHistory() error
	GetLatestAction() (FileAction, error)
}

type AlbumsStore interface {
	GetAllAlbums() ([]Album, error)
	CreateAlbum(Album) error

	SetAlbumCover(AlbumId, string, string, ContentId) error
	GetAlbumsByMedia(ContentId) ([]Album, error)

	AddMediaToAlbum(aId AlbumId, mIds []ContentId) error
	RemoveMediaFromAlbum(AlbumId, ContentId) error

	AddUsersToAlbum(aId AlbumId, us []User) error
}

type MediaStore interface {
	CreateMedia(m Media) error
	GetAllMedia() ([]Media, error)
	DeleteMedia(ContentId) error
	SetMediaHidden(id ContentId, hidden bool) error
	AddFileToMedia(mId ContentId, fId FileId) error
	RemoveFileFromMedia(mId ContentId, fId FileId) error
	GetFetchMediaCacheImage(ctx context.Context) ([]byte, error)

	DeleteAllMedia() error
}

type ShareStore interface {
	CreateShare(Share) error
	UpdateShare(Share) error
	GetAllShares() ([]Share, error)
	SetShareEnabledById(sId ShareId, enabled bool) error
	AddUsersToShare(share Share, users []Username) error
	GetSharedWithUser(username Username) ([]Share, error)
	DeleteShare(shareId ShareId) error
}

type UserStore interface {
	GetAllUsers() ([]User, error)
	UpdatePasswordByUsername(username Username, newPasswordHash string) error
	SetAdminByUsername(Username, bool) error
	CreateUser(User) error
	ActivateUser(Username) error
	AddTokenToUser(username Username, token string) error
	SearchUsers(search string) ([]Username, error)

	DeleteUserByUsername(Username) error
	DeleteAllUsers() error
}

type InstanceStore interface {
	GetAllServers() ([]Instance, error)
	NewServer(Instance) error
	DeleteServer(InstanceId) error
	AttachToCore(this Instance, core Instance) (Instance, error)
}

type AccessStore interface {
	CreateApiKey(ApiKeyInfo) error
	SetRemoteUsing(key WeblensApiKey, remoteId InstanceId) error
	GetApiKeys() ([]ApiKeyInfo, error)
	DeleteApiKey(WeblensApiKey) error
}

type FilesStore interface {
	NewTrashEntry(te TrashEntry) error
	GetTrashEntry(fileId FileId) (TrashEntry, error)
	DeleteTrashEntry(fileId FileId) error
	GetAllFiles() ([]WeblensFile, error)
	StatFile(WeblensFile) (FileStat, error)
	ReadFile(WeblensFile) ([]byte, error)
	ReadDir(WeblensFile) ([]FileStat, error)
	TouchFile(WeblensFile) error
	GetFile(FileId) (WeblensFile, error)
}
