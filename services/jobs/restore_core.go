package jobs

import (
	"time"

	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/job"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/structs"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
)

func RestoreCore(t *task_model.Task) {
	meta := t.GetMeta().(job.RestoreCoreMeta)

	filerCtx, ok := t.Ctx.(context.FilerContext)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))
		return
	}
	fileService := filerCtx.FileService()

	type restoreInitParams struct {
		Name     string                `json:"name"`
		Role     tower_model.TowerRole `json:"role"`
		RemoteId string                `json:"remoteId"`
		LocalId  string                `json:"localId"`
		Key      models.ApiKey         `json:"usingKeyInfo"`
	}

	// Notify client of restore failure, if any
	t.SetErrorCleanup(
		func(errTsk *task_model.Task) {
			t.Ctx.Notify(client.NewTaskNotification(errTsk, websocket_mod.RestoreFailedEvent, task_model.TaskResult{"error": errTsk.ReadError().Error()}))
		},
	)

	// Prime server to be restored. This will fail if the server is already initialized

	t.Ctx.Notify(client.NewTaskNotification(
		t, websocket_mod.RestoreProgressEvent, task_model.TaskResult{"stage": "Connecting to remote", "timestamp": time.Now().UnixMilli()},
	))

	key, err := meta.Pack.AccessService.GetApiKey(meta.Core.UsingKey)
	t.ReqNoErr(err)

	initParams := restoreInitParams{
		Name: meta.Core.Name, Role: tower_model.RestoreServerRole, Key: key, RemoteId: meta.Local.Id,
		LocalId: meta.Core.TowerId,
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/servers/init").WithBody(initParams).Call()
	if err != nil {
		t.ReqNoErr(err)
	}

	// Restore journal
	t.Ctx.Notify(
		client.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_model.TaskResult{"stage": "Restoring file history", "timestamp": time.Now().UnixMilli()}),
	)

	lts := meta.Pack.FileService.GetJournalByTree(meta.Core.TowerId).GetAllLifetimes()
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/history").WithBody(lts).Call()
	t.ReqNoErr(err)

	// Restore users
	t.Ctx.Notify(client.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_model.TaskResult{"stage": "Restoring users", "timestamp": time.Now().UnixMilli()}))

	users, err := user_model.GetAllUsers(t.Ctx)
	if err != nil {
		t.Fail(err)
	}

	var userInfos []structs.UserInfo
	for _, u := range users {
		userInfos = append(userInfos, reshape.UserToUserInfo(t.Ctx, u))
	}
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/users").WithBody(users).Call()
	t.ReqNoErr(err)

	// Restore keys
	t.Ctx.Notify(client.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_model.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()}))

	rootUser := meta.Pack.UserService.GetRootUser()
	keys, err := meta.Pack.AccessService.GetAllKeysByServer(rootUser, meta.Core.TowerId)
	t.ReqNoErr(err)

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/keys").WithBody(keys).Call()
	t.ReqNoErr(err)

	// Restore instances
	t.Ctx.Notify(client.NewTaskNotification(
		t, websocket_mod.RestoreProgressEvent, task_model.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()},
	))

	instances := meta.Pack.InstanceService.GetAllByOriginServer(meta.Core.TowerId)
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/instances").WithBody(instances).Call()
	t.ReqNoErr(err)

	// Restore files
	t.Ctx.Notify(client.NewTaskNotification(
		t, websocket_mod.RestoreProgressEvent, task_model.TaskResult{"stage": "Restoring files", "timestamp": time.Now().UnixMilli()},
	))
	for i, lt := range lts {
		latest := lt.GetLatestAction()
		if latest.GetActionType() == history.FileDelete {
			continue
		}
		portable := fs.ParsePortable(latest.GetDestinationPath())
		if portable.IsDir() {
			continue
		}

		f, err := fileService.GetFileByContentId(lt.GetContentId())
		if err != nil {
			t.Ctx.Log().Error().Stack().Err(err).Msg("")
			continue
		}
		if f == nil {
			t.Ctx.Log().Error().Msgf("File not found for contentId [%s]", lt.GetContentId())
			continue
		}

		bs, err := f.ReadAll()
		if err != nil {
			t.ReqNoErr(err)
		}
		_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/file").WithBodyBytes(bs).WithQuery(
			"fileId", lt.ID(),
		).Call()
		if err != nil {
			t.ReqNoErr(err)
		}

		t.Ctx.Notify(client.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_model.TaskResult{"filesTotal": len(lts), "filesRestored": i}))
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/complete").Call()
	if err != nil {
		t.ReqNoErr(err)
	}

	meta.Core.SetReportedRole(tower_model.CoreServerRole)
	t.Ctx.Notify(client.NewTaskNotification(t, websocket_mod.RestoreCompleteEvent, nil))

	// Disconnect the core client to force a reconnection
	coreClient := meta.Pack.ClientService.GetClientByServerId(meta.Core.TowerId)
	meta.Pack.ClientService.ClientDisconnect(coreClient)

	t.Success()
}
