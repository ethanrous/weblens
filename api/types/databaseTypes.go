package types

import "go.mongodb.org/mongo-driver/mongo"

type DatabaseService interface {
	HistoryDbService
	AlbumsDB
}

type HistoryDbService interface {
	WriteFileEvent(fe FileEvent) error
	GetAllFileEvents([]FileEvent) (*mongo.Cursor, error)
}

type AlbumsDB interface {
	GetAllAlbums() ([]Album, error)
}
