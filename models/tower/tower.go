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
	InitTowerRole    TowerRole = "init"
	CoreTowerRole    TowerRole = "core"
	BackupTowerRole  TowerRole = "backup"
	RestoreTowerRole TowerRole = "restore"
)

var ErrTowerNotFound = errors.New("no tower found")
var ErrTowerNotInitialized = errors.New("tower not initialized")
var ErrTowerAlreadyInitialized = errors.New("tower is already initialized")
var ErrTowerIsBackup = errors.New("tower is a backup")

// A "Tower" is a single Weblens tower.
// For clarity: Core and Backup are "absolute" tower roles, and each tower
// will fit into one of these categories once initialized. Local vs Remote
// are RELATIVE terms, meaning a core tower is "remote" to a backup tower, but
// "local" to itself, and vice versa.
type Instance struct {

	// The public ID of the tower
	TowerId string `bson:"towerId"`
	Name    string `bson:"name"`

	// Core or Backup
	Role TowerRole `bson:"towerRole"`

	// Address of the remote tower, only if the instance is a core.
	// Not set for any remotes/backups on core tower, as it IS the core
	Address string `bson:"coreAddress"`

	// The ID of the tower in which this remote instance is in reference from
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

func CreateLocal(ctx context.Context) (*Instance, error) {
	tower := &Instance{
		DbId:        primitive.NewObjectID(),
		TowerId:     primitive.NewObjectID().Hex(),
		Role:        InitTowerRole,
		IsThisTower: true,
	}

	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, err = col.InsertOne(ctx, tower)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tower, nil
}

func ResetLocal(ctx context.Context) (*Instance, error) {
	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, err = col.DeleteMany(ctx, bson.M{"isThisTower": true})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return CreateLocal(ctx)
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
	err = col.FindOne(ctx, bson.M{"isThisTower": true}).Decode(tower)
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

func UpdateTower(ctx context.Context, tower *Instance) error {
	if tower.DbId.IsZero() {
		return errors.New("tower DBID is not set")
	}

	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": tower.DbId}, bson.M{"$set": tower})
	if err != nil {
		return err
	}

	return nil
}

func GetAllTowersByTowerId(ctx context.Context, towerId string) ([]*Instance, error) {
	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return nil, err
	}

	cursor, err := col.Find(ctx, bson.M{"createdBy": towerId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var towers []*Instance
	for cursor.Next(ctx) {
		var tower Instance
		if err := cursor.Decode(&tower); err != nil {
			return nil, err
		}
		towers = append(towers, &tower)
	}

	return towers, nil
}

func GetRemotes(ctx context.Context) ([]*Instance, error) {
	return nil, nil
}

func (t *Instance) IsCore() bool {
	return t.Role == CoreTowerRole
}

func (t *Instance) IsBackup() bool {
	return t.Role == BackupTowerRole
}

func (t *Instance) GetReportedRole() TowerRole {
	return t.reportedRole
}

func (t *Instance) SetReportedRole(role TowerRole) {
	t.reportedRole = role
}

func (t *Instance) SocketType() websocket_mod.ClientType {
	return websocket_mod.TowerClient
}
