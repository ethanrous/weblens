package jobs_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/tests"
	. "github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/ethanrous/weblens/service/proxy"
	"github.com/ethanrous/weblens/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupCore(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	t.Parallel()

	coreServices, err := tests.NewWeblensTestInstance(t.Name(), env.Config{
		Role: string(models.CoreServerRole),
	})
	if err != nil {
		t.Fatal(err)
	}

	coreKeys, err := coreServices.AccessService.GetAllKeys(coreServices.UserService.GetRootUser())
	require.NoError(t, err)
	coreApiKey := coreKeys[0].Key
	coreAddress := env.GetProxyAddress(coreServices.Cnf)

	cnf := env.Config{
		MongodbUri:  env.GetMongoURI(),
		MongodbName: "weblens-" + t.Name(),
		DataRoot:    filepath.Join(env.GetBuildDir(), "fs/test", t.Name(), "data"),
		CachesRoot:  filepath.Join(env.GetBuildDir(), "fs/test", t.Name(), "cache"),
	}

	mondb, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName)
	require.NoError(t, err)
	err = mondb.Drop(context.Background())
	require.NoError(t, err)

	userService, err := service.NewUserService(mondb.Collection("users"))
	if err != nil {
		t.Fatal(err)
	}

	accessService, err := service.NewAccessService(userService, mondb.Collection("apiKeys"))
	if err != nil {
		t.Fatal(err)
	}

	wp := task.NewWorkerPool(2, log.NewEmptyLogPackage())
	wp.RegisterJob(models.BackupTask, DoBackup)

	instanceService, err := service.NewInstanceService(mondb.Collection("servers"))
	if err != nil {
		t.Fatal(err)
	}

	local := models.NewInstance("test-backup-id", "test-backup-name", "", models.BackupServerRole, true, "", "test-backup-id")
	err = instanceService.Add(local)
	if err != nil {
		t.Fatal(err)
	}

	core, err := instanceService.AttachRemoteCore(coreAddress, coreApiKey)
	if err != nil {
		t.Fatal(err)
	}

	journal := mock.NewHollowJournalService()
	coreTree := mock.NewMemFileTree(core.Id)
	coreTree.SetJournal(journal)
	fileService := mock.NewMockFileService()
	fileService.AddTree(coreTree)

	userRequest := proxy.NewCoreRequest(core, "GET", "/users/me")
	user, err := proxy.CallHomeStruct[rest.UserInfo](userRequest)
	if err != nil {
		t.Fatal(err)
	}

	_, err = proxy.NewCoreRequest(core, "POST", "/folder").WithBody(rest.CreateFolderBody{ParentFolderId: user.HomeId, NewFolderName: "newFolder"}).Call()
	if err != nil {
		t.Fatal(err)
	}

	meta := models.BackupMeta{
		Core:             core,
		FileService:      fileService,
		UserService:      userService,
		WebsocketService: &mock.MockClientService{},
		InstanceService:  instanceService,
		TaskService:      wp,
		AccessService:    accessService,
		Caster:           &mock.MockCaster{},
	}

	wp.Run()
	defer wp.Stop()

	backupTask, err := wp.DispatchJob(models.BackupTask, meta, nil)
	if err != nil {
		t.Fatal(err)
	}

	backupTask.Wait()

	err = backupTask.ReadError()
	assert.NoError(t, err)

	coreServices.Server.Stop()
}
