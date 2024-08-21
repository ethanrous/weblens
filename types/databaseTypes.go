package types

import (
	"context"
	"io"
	"time"

	"github.com/ethrousseau/weblens/api"
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
	WriteFileEvent(FileEvent) error
	GetAllLifetimes() ([]Lifetime, error)
	GetLifetimesSince(time.Time) ([]Lifetime, error)
	UpsertLifetime(l Lifetime) error
	InsertManyLifetimes([]Lifetime) error
	GetActionsByPath(WeblensFilepath) ([]FileAction, error)
	DeleteAllFileHistory() error
	GetLatestAction() (FileAction, error)

	NewTrashEntry(te TrashEntry) error
	GetTrashEntry(fileId FileId) (TrashEntry, error)
	DeleteTrashEntry(fileId FileId) error
	GetAllFiles() ([]WeblensFile, error)
	StatFile(WeblensFile) (FileStat, error)
	ReadFile(WeblensFile) ([]byte, error)
	ReadDir(WeblensFile) ([]FileStat, error)
	TouchFile(WeblensFile) error
	GetFile(FileId) (WeblensFile, error)
	StreamFile(WeblensFile) (io.ReadCloser, error)

	GetAllUsers() ([]User, error)
	UpdatePasswordByUsername(username Username, newPasswordHash string) error
	SetAdminByUsername(Username, bool) error
	CreateUser(User) error
	ActivateUser(Username) error
	AddTokenToUser(username Username, token string) error
	SearchUsers(search string) ([]Username, error)

	DeleteUserByUsername(Username) error
	DeleteAllUsers() error

	GetAllServers() ([]Instance, error)
	NewServer(Instance) error
	DeleteServer(InstanceId) error
	AttachToCore(this Instance, core Instance) (Instance, error)
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

	SetAlbumCover(AlbumId, string, string, weblens.ContentId) error
	GetAlbumsByMedia(weblens.ContentId) ([]Album, error)

	AddMediaToAlbum(aId AlbumId, mIds []weblens.ContentId) error
	RemoveMediaFromAlbum(AlbumId, weblens.ContentId) error

	AddUsersToAlbum(aId AlbumId, us []User) error

	DeleteAlbum(aId AlbumId) error
}

type MediaStore interface {
	CreateMedia(m Media) error
	GetAllMedia() ([]Media, error)
	DeleteMedia(weblens.ContentId) error
	SetMediaHidden(id weblens.ContentId, hidden bool) error
	AddFileToMedia(mId weblens.ContentId, fId FileId) error
	RemoveFileFromMedia(mId weblens.ContentId, fId FileId) error
	GetFetchMediaCacheImage(ctx context.Context) ([]byte, error)
	AddLikeToMedia(weblens.ContentId, Username, bool) error

	DeleteAllMedia() error
}

type ShareStore interface {
	CreateShare(weblens.Share) error
	UpdateShare(weblens.Share) error
	GetAllShares() ([]weblens.Share, error)
	SetShareEnabledById(sId ShareId, enabled bool) error
	AddUsersToShare(share weblens.Share, users []Username) error
	GetSharedWithUser(username Username) ([]weblens.Share, error)
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
	StreamFile(WeblensFile) (io.ReadCloser, error)
}
