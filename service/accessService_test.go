package service_test

import (
	"context"
	"testing"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/models"
	. "github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
)

func TestAccessServiceImpl_CanUserAccessFile(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	defer col.Drop(context.Background())

	userCol := mondb.Collection(t.Name() + "-users")
	userService, err := NewUserService(userCol)
	if err != nil {
		t.Fatal(err)
	}

	acc, err := NewAccessService(userService, col)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", false, true)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	dipperUser, err := models.NewUser("dipperpines", "ivegotabook", false, true)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Make file tree
	ft := mock.NewMemFileTree("MEDIA")
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
}

func TestAccessServiceImpl_GenerateApiKey(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	defer col.Drop(context.Background())

	userCol := mondb.Collection(t.Name() + "-users")
	userService, err := NewUserService(userCol)
	if err != nil {
		t.Fatal(err)
	}

	acc, err := NewAccessService(userService, col)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", false, true)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	_, err = acc.GenerateApiKey(billUser)
	assert.Error(t, err)

	billUser.Admin = true

	key, err := acc.GenerateApiKey(billUser)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, billUser.Username, key.Owner)

	fetchedKey, err := acc.GetApiKey(key.Key)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NotNil(t, fetchedKey) {
		t.FailNow()
	}

	err = acc.DeleteApiKey(key.Key)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	_, err = acc.GetApiKey(key.Key)
	assert.Error(t, err)
}

func TestAccessServiceImpl_SetKeyUsedBy(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	defer col.Drop(context.Background())

	userCol := mondb.Collection(t.Name() + "-users")
	userService, err := NewUserService(userCol)
	if err != nil {
		t.Fatal(err)
	}

	acc, err := NewAccessService(userService, col)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", true, true)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	key, err := acc.GenerateApiKey(billUser)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	backupServer := models.NewInstance("", "test-instance", key.Key, models.BackupServer, false, "")

	err = acc.SetKeyUsedBy(key.Key, backupServer)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	fetchedKey, err := acc.GetApiKey(key.Key)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, backupServer.ServerId(), fetchedKey.RemoteUsing)

}
