// Package tower manages tower instances and their connections in the Weblens distributed system.
package tower

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Role represents a tower's role in the system.
type Role string

// TowerCollectionKey is the MongoDB collection name for towers.
const TowerCollectionKey = "towers"

// Tower role constants.
const (
	RoleInit    Role = "init"
	RoleCore    Role = "core"
	RoleBackup  Role = "backup"
	RoleRestore Role = "restore"
)

// ErrTowerNotFound is returned when a tower cannot be found.
var ErrTowerNotFound = wlerrors.New("no tower found")

// ErrTowerNotInitialized is returned when a tower has not been initialized.
var ErrTowerNotInitialized = wlerrors.New("tower not initialized")

// ErrTowerAlreadyInitialized is returned when attempting to initialize an already initialized tower.
var ErrTowerAlreadyInitialized = wlerrors.New("tower is already initialized")

// ErrTowerIsBackup is returned when a tower is a backup but shouldn't be.
var ErrTowerIsBackup = wlerrors.New("tower is a backup")

// ErrTowerNotBackup is returned when a tower was expected to be a backup but is not.
var ErrTowerNotBackup = wlerrors.New("tower was expected to be a backup tower, but is not")

// ErrNotCore is returned when a tower was expected to be a core but is not.
var ErrNotCore = wlerrors.New("tower was expected to be a core tower, but is not")

// Instance represents a Weblens tower.
// For clarity: Core and Backup are "absolute" tower roles, and each tower
// will fit into one of these categories once initialized. Local vs Remote
// are RELATIVE terms, meaning a core tower is "remote" to a backup tower, but
// "local" to itself, and vice versa.
type Instance struct {

	// The public ID of the tower
	TowerID string `bson:"towerID"`
	Name    string `bson:"name"`

	// Core or Backup
	Role Role `bson:"towerRole"`

	// Address of the remote tower, only if the instance is a core.
	// Not set for any remotes/backups on core tower, as it IS the core
	Address string `bson:"coreAddress"`

	// The ID of the tower in which this remote instance is in reference from
	CreatedBy string `bson:"createdBy"`

	// The time of the latest backup, in milliseconds since epoch
	LastBackup int64 `bson:"lastBackup"`

	// The private ID of the tower only in the local database
	DbID primitive.ObjectID `bson:"_id"`

	// Only one of the following 2 will be set, depending on the role of the local tower

	// The API Key the remote is expected to use to authenticate with the local tower
	IncomingKey string `bson:"incomingKey"`
	// The API Key the remote is expecting the local tower to use to authenticate with the remote tower
	OutgoingKey string `bson:"outgoingKey"`

	// If this tower instance represents the local tower
	IsThisTower bool `bson:"isThisTower"`

	// The role the tower is currently reporting. This is used to determine if the tower is online (and functional) or not
	reportedRole Role `bson:"-"`
}

// CreateLocal creates a new local tower instance.
func CreateLocal(ctx context.Context) (t Instance, err error) {
	t = Instance{
		DbID:        primitive.NewObjectID(),
		TowerID:     primitive.NewObjectID().Hex(),
		Role:        RoleInit,
		IsThisTower: true,
	}

	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return t, wlerrors.WithStack(err)
	}

	_, err = col.InsertOne(ctx, t)
	if err != nil {
		return t, wlerrors.WithStack(err)
	}

	return t, nil
}

// ResetLocal deletes and recreates the local tower instance.
func ResetLocal(ctx context.Context) (t Instance, err error) {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return t, wlerrors.WithStack(err)
	}

	_, err = col.DeleteMany(ctx, bson.M{"isThisTower": true})
	if err != nil {
		return t, wlerrors.WithStack(err)
	}

	return CreateLocal(ctx)
}

// SaveTower persists a tower instance to the database.
func SaveTower(ctx context.Context, tower *Instance) error {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return err
	}

	if tower.DbID.IsZero() {
		tower.DbID = primitive.NewObjectID()
	} else {
		_, err = col.ReplaceOne(ctx, bson.M{"_id": tower.DbID}, tower)

		return db.WrapError(err, "failed to update tower")
	}

	err = validateNewTower(ctx, tower)
	if err != nil {
		return err
	}

	_, err = col.InsertOne(ctx, tower)
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return nil
}

// GetTowerByID retrieves a tower by its public ID.
func GetTowerByID(ctx context.Context, towerID string) (tower Instance, err error) {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return tower, err
	}

	err = col.FindOne(ctx, bson.M{"towerID": towerID}).Decode(&tower)
	if err != nil {
		if wlerrors.Is(err, mongo.ErrNoDocuments) {
			return tower, ErrTowerNotFound
		}

		return tower, wlerrors.WithStack(err)
	}

	return
}

// GetBackupTowerByID retrieves a backup tower by its ID and remote ID.
func GetBackupTowerByID(ctx context.Context, towerID string, remoteID string) (tower Instance, err error) {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return tower, err
	}

	err = col.FindOne(ctx, bson.M{"towerID": towerID, "createdBy": remoteID}).Decode(&tower)
	if err != nil {
		return tower, db.WrapError(err, "failed to get backup tower")
	}

	return
}

// DeleteTowerByID removes a tower by its public ID.
func DeleteTowerByID(ctx context.Context, towerID string) error {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.DeleteOne(ctx, bson.M{"towerID": towerID})
	if err != nil {
		return db.WrapError(err, "failed to delete tower")
	}

	return nil
}

// GetLocal retrieves the local tower instance.
func GetLocal(ctx context.Context) (t Instance, err error) {
	col, err := db.GetCollection[*Instance](ctx, TowerCollectionKey)
	if err != nil {
		return t, err
	}

	tower := Instance{}

	err = col.FindOne(ctx, bson.M{"isThisTower": true}).Decode(&tower)
	if err != nil {
		return t, wlerrors.WithStack(err)
	}

	return tower, nil
}

// SetLastBackup updates the last backup timestamp for a tower.
func SetLastBackup(ctx context.Context, towerID string, lastBackup time.Time) error {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.M{"towerID": towerID}, bson.M{"$set": bson.M{"lastBackup": lastBackup.UnixMilli()}})
	if err != nil {
		return err
	}

	return nil
}

// UpdateTower updates a tower instance in the database.
func UpdateTower(ctx context.Context, tower *Instance) error {
	if tower.DbID.IsZero() {
		return wlerrors.New("tower DBID is not set")
	}

	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": tower.DbID}, bson.M{"$set": tower})
	if err != nil {
		return err
	}

	return nil
}

// GetAllTowersByTowerID retrieves all towers created by a specific tower.
func GetAllTowersByTowerID(ctx context.Context, towerID string) ([]Instance, error) {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return nil, err
	}

	cursor, err := col.Find(ctx, bson.M{"createdBy": towerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	var towers []Instance
	if err := cursor.All(ctx, &towers); err != nil {
		return nil, err
	}

	return towers, nil
}

// GetRemotes retrieves all remote tower instances.
func GetRemotes(ctx context.Context) ([]Instance, error) {
	col, err := db.GetCollection[any](ctx, TowerCollectionKey)
	if err != nil {
		return nil, err
	}

	cursor, err := col.Find(ctx, bson.M{"isThisTower": false})
	if err != nil {
		return nil, err
	}

	var remotes []Instance

	err = cursor.All(ctx, &remotes)
	if err != nil {
		return nil, err
	}

	return remotes, nil
}

// IsCore returns true if the tower has the core role.
func (t *Instance) IsCore() bool {
	return t.Role == RoleCore
}

// IsBackup returns true if the tower has the backup role.
func (t *Instance) IsBackup() bool {
	return t.Role == RoleBackup
}

// GetReportedRole returns the role the tower is currently reporting.
func (t *Instance) GetReportedRole() Role {
	return t.reportedRole
}

// SetReportedRole sets the role the tower is currently reporting.
func (t *Instance) SetReportedRole(role Role) {
	t.reportedRole = role
}

// SocketType returns the websocket client type for this tower.
func (t *Instance) SocketType() websocket_mod.ClientType {
	return websocket_mod.TowerClient
}
