package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	. "github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	if mondb == nil {
		var err error
		mondb, err = database.ConnectToMongo(internal.GetMongoURI(), internal.GetMongoDBName()+"-test")
		if err != nil {
			panic(err)
		}
	}
}

func TestInstanceServiceImpl_Add(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	is, err := NewInstanceService(col)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NotNil(t, is.GetLocal()) {
		t.FailNow()
	}
	assert.Equal(t, models.InitServer, is.GetLocal().GetRole())

	localInstance := models.NewInstance("", "My server", "", models.CoreServer, true, "")
	assert.NotEmpty(t, localInstance.ServerId())

	err = is.Add(localInstance)
	assert.ErrorIs(t, err, werror.ErrDuplicateLocalServer)

	assert.Nil(t, is.GetCore())
	// assert.Equal(t, localInstance.ServerId(), is.GetLocal().ServerId())

	remoteId := models.InstanceId(primitive.NewObjectID().Hex())
	remoteBackup := models.NewInstance(
		remoteId, "My remote server", "deadbeefdeadbeef", models.BackupServer, false,
		"http://notrighthere.com",
	)

	assert.Equal(t, remoteId, remoteBackup.ServerId())

	err = is.Add(remoteBackup)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	remoteFetch := is.Get(remoteId)
	assert.Equal(t, remoteId, remoteFetch.ServerId())

	noName := models.NewInstance("", "", "deadbeefdeadbeef", models.BackupServer, false, "")
	err = is.Add(noName)
	assert.ErrorIs(t, err, werror.ErrNoServerName)

	noName.UsingKey = ""
	err = is.Add(noName)
	assert.ErrorIs(t, err, werror.ErrNoServerKey)

	noName.Id = ""
	err = is.Add(noName)
	assert.ErrorIs(t, err, werror.ErrNoServerId)

	anotherCore := models.NewInstance("", "Another Core", "deadbeefdeadbeef", models.CoreServer, false, "")
	err = is.Add(anotherCore)
	assert.ErrorIs(t, err, werror.ErrNoCoreAddress)
}

func TestInstanceServiceImpl_Del(t *testing.T) {

}

func TestInstanceServiceImpl_InitCore(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err = col.Drop(context.Background()); err != nil {
		t.Fatalf(err.Error())
	}
	defer col.Drop(context.Background())

	is, err := NewInstanceService(col)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NotNil(t, is.GetLocal()) {
		t.FailNow()
	}
	assert.Equal(t, models.InitServer, is.GetLocal().GetRole())
	assert.Nil(t, is.GetCore())

	err = is.InitCore("My Core Server")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, models.CoreServer, is.GetLocal().GetRole())
	assert.NotNil(t, is.GetCore())

	if err = col.Drop(context.Background()); err != nil {
		t.Fatalf(err.Error())
	}

	badMongo := &mock.MockFailMongoCol{
		RealCol:    col,
		InsertFail: true,
		FindFail:   false,
		UpdateFail: false,
		DeleteFail: false,
	}

	badIs, err := NewInstanceService(badMongo)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = badIs.InitCore("My Core Server")
	assert.Error(t, err)

	assert.Equal(t, models.InitServer, badIs.GetLocal().GetRole())
	assert.Nil(t, badIs.GetCore())
}

func TestInstanceServiceImpl_InitBackup(t *testing.T) {
	if os.Getenv("REMOTE_TESTS") != "true" {
		t.Skip("REMOTE_TESTS not set")
	}

	coreAddress := os.Getenv("CORE_ADDRESS")
	if coreAddress == "" {
		t.Fatalf("CORE_ADDRESS environment variable required for %s", t.Name())
	}
	coreKey := os.Getenv("CORE_API_KEY")
	if coreKey == "" {
		t.Fatalf("CORE_API_KEY environment variable required for %s", t.Name())
	}

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err = col.Drop(context.Background()); err != nil {
		t.Fatalf(err.Error())
	}
	defer col.Drop(context.Background())

	is, err := NewInstanceService(col)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NotNil(t, is.GetLocal()) {
		t.FailNow()
	}
	assert.Equal(t, models.InitServer, is.GetLocal().GetRole())

	err = is.InitBackup("My backup server", coreAddress, models.WeblensApiKey(coreKey))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, models.BackupServer, is.GetLocal().GetRole())

}
