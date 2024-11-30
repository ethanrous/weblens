package jobs_test

import (
	"context"
	"os"
	"testing"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
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

	if os.Getenv("REMOTE_TESTS") != "true" {
		t.Skipf("skipping %s without REMOTE_TESTS set", t.Name())
	}

	coreAddress := os.Getenv("CORE_ADDRESS")
	require.NotEmpty(t, coreAddress)

	coreApiKey := os.Getenv("CORE_API_KEY")
	require.NotEmpty(t, coreApiKey)

	mondb, err := database.ConnectToMongo(env.GetMongoURI("TEST-BACKUP"), env.GetMongoDBName("TEST-BACKUP"))
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

	wp := task.NewWorkerPool(2, log.GetLogLevel())
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
}
