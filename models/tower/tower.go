package tower

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TowerRole string

const TowerCollectionKey = "towers"

const (
	InitServerRole    TowerRole = "init"
	CoreServerRole    TowerRole = "core"
	BackupServerRole  TowerRole = "backup"
	RestoreServerRole TowerRole = "restore"
)

var ErrTowerNotFound = errors.New("no tower found")
var ErrServerNotInitialized = errors.New("server not initialized")
var ErrServerIsBackup = errors.New("server is a backup")

// A "Tower" is a single Weblens server.
// For clarity: Core and Backup are "absolute" server roles, and each server
// will fit into one of these categories once initialized. Local vs Remote
// are RELATIVE terms, meaning a core tower is "remote" to a backup tower, but
// "local" to itself, and vice versa.
type Instance struct {

	// The public ID of the tower
	TowerId string `bson:"towerId"`
	Name    string `bson:"name"`

	// Core or Backup
	Role TowerRole `bson:"serverRole"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `bson:"coreAddress"`

	// The ID of the server in which this remote instance is in reference from
	CreatedBy string `bson:"createdBy"`

	// The time of the latest backup, in milliseconds since epoch
	LastBackup int64 `bson:"lastBackup"`

	// The private ID of the tower only in the local database
	DbId primitive.ObjectID `bson:"_id"`

	// Only one of the following 2 will be set, depending on the role of the local tower

	// The API Key the remote is expected to use to authenticate with the local tower
	IncomingKey string `bson:"incomingKey"`
	// The API Key the remote is expecting the local tower to use to authenticate with the remote tower
	OutgoingKey string `bson:"outgoingKey"`

	// If this tower instance represents the local tower
	IsThisTower bool `bson:"isThisTower"`

	// The role the tower is currently reporting. This is used to determine if the tower is online (and functional) or not
	reportedRole TowerRole `bson:"-"`
}

func CreateTower(ctx context.Context, tower *Instance) error {
	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.InsertOne(ctx, tower)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func GetTowerById(ctx context.Context, towerId string) (tower *Instance, err error) {
	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return nil, err
	}

	tower = &Instance{}
	err = col.FindOne(ctx, bson.M{"towerId": towerId}).Decode(tower)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTowerNotFound
		}
		return nil, errors.WithStack(err)
	}

	return
}

func GetLocal(ctx context.Context) (*Instance, error) {
	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return nil, err
	}

	tower := &Instance{}
	err = col.FindOne(ctx, bson.M{"isThisServer": true}).Decode(tower)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tower, nil
}

func SetLastBackup(ctx context.Context, towerId string, lastBackup time.Time) error {
	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.M{"towerId": towerId}, bson.M{"$set": bson.M{"lastBackup": lastBackup}})
	if err != nil {
		return err
	}

	return nil
}

func GetRemotes(ctx context.Context) ([]*Instance, error) {
	return nil, nil
}

func (t *Instance) IsCore() bool {
	return t.Role == CoreServerRole
}

func (t *Instance) GetReportedRole() TowerRole {
	return t.reportedRole
}

func (t *Instance) SocketType() websocket_mod.ClientType {
	return websocket_mod.TowerClient
}
