package tower

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServerRole string

const TowerCollectionKey = "towers"

const (
	InitServerRole    ServerRole = "init"
	CoreServerRole    ServerRole = "core"
	BackupServerRole  ServerRole = "backup"
	RestoreServerRole ServerRole = "restore"
)

// A "Tower" is a single Weblens server.
// For clarity: Core vs Backup are absolute server roles, and each server
// will fit into one of these categories once initialized. Local vs Remote
// are RELATIVE terms, meaning one core servers "remote" is the backup
// servers "local".
type Instance struct {

	// The public ID of the tower
	TowerId string `bson:"towerId"`
	Name    string `bson:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	// UsingKey WeblensApiKey `bson:"usingKey"`

	// Core or BackupServer
	Role ServerRole `bson:"serverRole"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `bson:"coreAddress"`

	// The ID of the server in which this remote instance is in reference from
	CreatedBy string `bson:"createdBy"`

	// TODO: Move to structs package
	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	// reportedRole ServerRole

	// The time of the latest backup, in milliseconds since epoch
	LastBackup int64 `bson:"lastBackup"`

	// The private ID of the server only in the local database
	DbId primitive.ObjectID `bson:"_id"`

	// If this server info represents this local server
	IsThisServer bool `bson:"isThisServer"`
}

func GetTowerById(ctx context.Context, towerId string) (*Instance, error) {
	col, err := db.GetCollection(ctx, TowerCollectionKey)
	if err != nil {
		return nil, err
	}

	res := col.FindOne(ctx, bson.M{"towerId": towerId})
	if res.Err() != nil {
		return nil, res.Err()
	}

	tower := &Instance{}
	err = res.Decode(tower)
	if err != nil {
		return nil, err
	}

	return tower, nil
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

func SetLastBackup(ctx context.Context, towerId string, lastBackup int64) error {
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

func (t *Instance) IsCore() bool {
	return t.Role == CoreServerRole
}
