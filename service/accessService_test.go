package service_test

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	. "github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessServiceImpl_CanUserAccessFile(t *testing.T) {
	t.Parallel()

	keysCol := mondb.Collection(string(database.ApiKeysCollectionKey) + "-" + t.Name())
	err := keysCol.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	defer func() { log.ErrTrace(keysCol.Drop(context.Background())) }()

	userCol := mondb.Collection(string(database.UsersCollectionKey) + "-" + t.Name())
	userService, err := NewUserService(userCol)
	if err != nil {
		t.Fatal(err)
	}

	acc, err := NewAccessService(userService, keysCol)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", false, true)
	require.NoError(t, err)

	dipperUser, err := models.NewUser("dipperpines", "ivegotabook", false, true)
	require.NoError(t, err)

	// Make file tree
	ft := mock.NewMemFileTree("USERS")
	// Make bills home in tree
	billHome, err := ft.MkDir(ft.GetRoot(), string(billUser.Username), nil)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	// Make dippers home
	dipperHome, err := ft.MkDir(ft.GetRoot(), string(dipperUser.Username), nil)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	// Bill can access his home folder, but nobody else can without a share
	assert.True(t, acc.CanUserAccessFile(billUser, billHome, nil))
	assert.False(t, acc.CanUserAccessFile(dipperUser, billHome, nil))

	// Make a share for bills home, and check dipper can now access it
	// if he is using the share
	billHomeShare := models.NewFileShare(billHome, billUser, []*models.User{dipperUser}, false, false)
	assert.True(t, acc.CanUserAccessFile(dipperUser, billHome, billHomeShare))

	// Check public share grants access, even if user is not in share
	dipperHomeShare := models.NewFileShare(dipperHome, dipperUser, []*models.User{}, true, false)
	assert.True(t, acc.CanUserAccessFile(billUser, dipperHome, dipperHomeShare))

	// Make sure public user can't access private share
	public := userService.GetPublicUser()
	assert.False(t, acc.CanUserAccessFile(public, billHome, billHomeShare))

	// Make sure public user can access public share
	assert.True(t, acc.CanUserAccessFile(public, dipperHome, dipperHomeShare))

	// Make sure root user can access file
	weblensRootUser := userService.GetRootUser()
	assert.True(t, acc.CanUserAccessFile(weblensRootUser, billHome, nil))
}

func TestAccessServiceImpl_GenerateApiKey(t *testing.T) {
	t.Parallel()

	keysCol := mondb.Collection(string(database.ApiKeysCollectionKey) + "-" + t.Name())
	err := keysCol.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	defer func() { log.ErrTrace(keysCol.Drop(context.Background())) }()

	userCol := mondb.Collection(string(database.UsersCollectionKey) + "-" + t.Name())
	userService, err := NewUserService(userCol)
	if err != nil {
		t.Fatal(err)
	}

	acc, err := NewAccessService(userService, keysCol)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", false, true)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	local := models.NewInstance("", "test-instance", "", models.CoreServerRole, true, "", "")

	key, err := acc.GenerateApiKey(billUser, local, "test-key")
	require.NoError(t, err)
	assert.Equal(t, billUser.Username, key.Owner)

	fetchedKey, err := acc.GetApiKey(key.Key)
	require.NoError(t, err)

	if !assert.NotNil(t, fetchedKey) {
		t.FailNow()
	}

	err = acc.DeleteApiKey(key.Key)
	require.NoError(t, err)

	_, err = acc.GetApiKey(key.Key)
	assert.Error(t, err)
}

func TestAccessServiceImpl_SetKeyUsedBy(t *testing.T) {
	t.Parallel()

	keysCol := mondb.Collection(string(database.ApiKeysCollectionKey) + "-" + t.Name())
	err := keysCol.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	defer func() { log.ErrTrace(keysCol.Drop(context.Background())) }()

	userCol := mondb.Collection(string(database.UsersCollectionKey) + "-" + t.Name())
	userService, err := NewUserService(userCol)
	if err != nil {
		t.Fatal(err)
	}

	acc, err := NewAccessService(userService, keysCol)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", true, true)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	local := models.NewInstance("", "test-instance", "", models.CoreServerRole, true, "", "")

	key, err := acc.GenerateApiKey(billUser, local, "test-key")
	require.NoError(t, err)

	backupServer := models.NewInstance("", "test-instance", key.Key, models.BackupServerRole, false, "", t.Name())

	err = acc.SetKeyUsedBy(key.Key, backupServer)
	require.NoError(t, err)

	fetchedKey, err := acc.GetApiKey(key.Key)
	require.NoError(t, err)

	assert.Equal(t, backupServer.ServerId(), fetchedKey.RemoteUsing)

}
