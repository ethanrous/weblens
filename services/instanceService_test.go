package services

// import (
// 	"context"
// 	"testing"
//
// 	"github.com/ethanrous/weblens/database"
// 	"github.com/ethanrous/weblens/internal/env"
// 	"github.com/ethanrous/weblens/internal/log"
// 	"github.com/ethanrous/weblens/internal/tests"
// 	"github.com/ethanrous/weblens/internal/werror"
// 	"github.com/ethanrous/weblens/models"
// 	"github.com/ethanrous/weblens/service/mock"
// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )
//
// func init() {
// 	if mondb == nil {
// 		logger := log.NewZeroLogger()
// 		var err error
// 		mondb, err = database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName(env.Config{})+"-test", logger)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }
//
// func TestInstanceServiceImpl_Add(t *testing.T) {
// 	t.Parallel()
//
// 	logger := log.NewZeroLogger()
//
// 	col := mondb.Collection(t.Name())
// 	err := col.Drop(context.Background())
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer tests.CheckDropCol(col, logger)
//
// 	is, err := NewInstanceService(col, logger)
// 	require.NoError(t, err)
//
// 	if !assert.NotNil(t, is.GetLocal()) {
// 		t.FailNow()
// 	}
// 	assert.Equal(t, models.InitServerRole, is.GetLocal().GetRole())
//
// 	localInstance := models.NewInstance("", "My server", "", models.CoreServerRole, true, "", t.Name())
// 	assert.NotEmpty(t, localInstance.ServerId())
//
// 	err = is.Add(localInstance)
// 	assert.NoError(t, err)
// 	// assert.ErrorIs(t, err, werror.ErrDuplicateLocalServer)
//
// 	remoteId := models.InstanceId(primitive.NewObjectID().Hex())
// 	remoteBackup := models.NewInstance(
// 		remoteId, "My remote server", "deadbeefdeadbeef", models.BackupServerRole, false,
// 		"http://notrighthere.com", t.Name(),
// 	)
//
// 	assert.Equal(t, remoteId, remoteBackup.ServerId())
//
// 	err = is.Add(remoteBackup)
// 	require.NoError(t, err)
//
// 	assert.False(t, remoteBackup.DbId.IsZero())
//
// 	remoteFetch := is.Get(remoteBackup.DbId.Hex())
// 	require.NotNil(t, remoteFetch)
// 	assert.Equal(t, remoteId, remoteFetch.ServerId())
//
// 	badServer := models.NewInstance(
// 		"", "", "deadbeefdeadbeef", models.BackupServerRole, false, "", is.GetLocal().ServerId(),
// 	)
// 	err = is.Add(badServer)
// 	assert.ErrorIs(t, err, werror.ErrNoServerName)
//
// 	badServer.UsingKey = ""
// 	badServer.Name = "test server name"
// 	err = is.Add(badServer)
// 	assert.ErrorIs(t, err, werror.ErrNoServerKey)
//
// 	badServer.UsingKey = "deadbeefdeadbeef"
// 	badServer.Id = ""
// 	err = is.Add(badServer)
// 	assert.ErrorIs(t, err, werror.ErrNoServerId)
//
// 	anotherCore := models.NewInstance(
// 		"", "Another Core", "deadbeefdeadbeef", models.CoreServerRole, false, "", is.GetLocal().ServerId(),
// 	)
// 	err = is.Add(anotherCore)
// 	assert.ErrorIs(t, err, werror.ErrNoCoreAddress)
// }
//
// func TestInstanceServiceImpl_InitCore(t *testing.T) {
// 	t.Parallel()
//
// 	logger := log.NewZeroLogger()
//
// 	col := mondb.Collection(t.Name())
// 	err := col.Drop(context.Background())
// 	if err != nil {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.Fatal(err)
// 	}
// 	if err = col.Drop(context.Background()); err != nil {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.Fatal(err)
// 	}
// 	defer tests.CheckDropCol(col, logger)
//
// 	is, err := NewInstanceService(col, logger)
// 	require.NoError(t, err)
//
// 	if !assert.NotNil(t, is.GetLocal()) {
// 		t.FailNow()
// 	}
// 	assert.Equal(t, models.InitServerRole, is.GetLocal().GetRole())
//
// 	err = is.InitCore("My Core Server")
// 	require.NoError(t, err)
//
// 	assert.Equal(t, models.CoreServerRole, is.GetLocal().GetRole())
//
// 	if err = col.Drop(context.Background()); err != nil {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.Fatal(err)
// 	}
//
// 	badMongo := &mock.MockFailMongoCol{
// 		RealCol:    col,
// 		InsertFail: true,
// 		FindFail:   false,
// 		UpdateFail: false,
// 		DeleteFail: false,
// 	}
//
// 	badIs, err := NewInstanceService(badMongo, logger)
// 	require.NoError(t, err)
//
// 	err = badIs.InitCore("My Core Server")
// 	assert.Error(t, err)
//
// 	assert.Equal(t, models.InitServerRole, badIs.GetLocal().GetRole())
// }
//
// func TestInstanceServiceImpl_InitBackup(t *testing.T) {
// 	t.Parallel()
//
// 	logger := log.NewZeroLogger()
// 	nop := zerolog.Nop()
//
// 	coreServices, err := tests.NewWeblensTestInstance(t.Name(), env.Config{
// 		Role: string(models.CoreServerRole),
// 	}, &nop)
//
// 	require.NoError(t, err)
//
// 	keys, err := coreServices.AccessService.GetKeysByUser(coreServices.UserService.Get("test-username"))
// 	if err != nil {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.FailNow()
// 	}
// 	logger.Debug().Func(func(e *zerolog.Event) { e.Msgf("Key count: %d", len(keys)) })
//
// 	coreAddress := env.GetProxyAddress(coreServices.Cnf)
// 	coreApiKey := keys[0].Key
//
// 	owner := coreServices.UserService.Get("test-username")
// 	if owner == nil {
// 		t.Fatalf("No owner")
// 	}
//
// 	col := mondb.Collection(t.Name())
// 	err = col.Drop(context.Background())
// 	if err != nil {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.FailNow()
// 	}
//
// 	if err = col.Drop(context.Background()); err != nil {
// 		t.Fatal(err)
// 	}
// 	defer tests.CheckDropCol(col, logger)
//
// 	is, err := NewInstanceService(col, logger)
// 	require.NoError(t, err)
//
// 	if !assert.NotNil(t, is.GetLocal()) {
// 		t.FailNow()
// 	}
// 	assert.Equal(t, models.InitServerRole, is.GetLocal().GetRole())
//
// 	_, err = is.InitBackup("My backup server", coreAddress, coreApiKey)
// 	require.NoError(t, err)
//
// 	assert.Equal(t, models.BackupServerRole, is.GetLocal().GetRole())
// }
