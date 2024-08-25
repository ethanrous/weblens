package database

//
// import (

// 	"github.com/ethrousseau/weblens/api/internal"
// 	error2 "github.com/ethrousseau/weblens/api/internal/werror"
// 	"github.com/ethrousseau/weblens/api/types"
// 	"go.mongodb.org/mongo-driver/bson"
// )
//
// func (db *databaseService) GetAllServers() ([]types.Instance, error) {
// 	ret, err := db.servers.Find(db.ctx, bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	var servers []*weblens.Instance
// 	err = ret.All(db.ctx, &servers)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return internal.SliceConvert[types.Instance](servers), nil
// }
//
// func (db *databaseService) NewServer(i types.Instance) error {
// 	_, err := db.servers.InsertOne(db.ctx, i)
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
// 	return nil
// }
//
// func (db *databaseService) AttachToCore(i types.Instance, core types.Instance) (types.Instance, error) {
// 	return nil, error2.NotImplemented("AttachToCore local")
// }
//
// func (db *databaseService) DeleteServer(id types.InstanceId) error {
// 	_, err := db.servers.DeleteOne(db.ctx, bson.M{"_id": id})
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
//
// 	filter := bson.M{"remoteUsing": id}
// 	update := bson.M{"$set": bson.M{"remoteUsing": ""}}
// 	_, err = db.apiKeys.UpdateMany(db.ctx, filter, update)
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
// 	return nil
// }
