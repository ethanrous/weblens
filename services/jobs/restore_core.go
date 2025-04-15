package jobs

import (
	"time"

	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RestoreCore(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	meta := t.GetMeta().(job.RestoreCoreMeta)

	ctx, ok := t.Ctx.(context.AppContext)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))
		return
	}

	type restoreInitParams struct {
		Name     string                `json:"name"`
		Role     tower_model.TowerRole `json:"role"`
		RemoteId string                `json:"remoteId"`
		LocalId  string                `json:"localId"`
		Key      *auth.Token           `json:"usingKeyInfo"`
	}

	// Notify client of restore failure, if any
	t.SetErrorCleanup(
		func(errTsk task_mod.Task) {
			t.Ctx.Notify(notify.NewTaskNotification(errTsk.(*task_model.Task), websocket_mod.RestoreFailedEvent, task_mod.TaskResult{"error": errTsk.ReadError().Error()}))
		},
	)

	// Prime server to be restored. This will fail if the server is already initialized

	t.Ctx.Notify(notify.NewTaskNotification(
		t, websocket_mod.RestoreProgressEvent, task_mod.TaskResult{"stage": "Connecting to remote", "timestamp": time.Now().UnixMilli()},
	))

	tokenId, err := primitive.ObjectIDFromHex(meta.Core.OutgoingKey)
	if err != nil {
		t.Fail(err)
	}

	token, err := auth.GetTokenById(ctx, tokenId)
	if err != nil {
		t.Fail(err)
	}

	initParams := restoreInitParams{
		Name: meta.Core.Name, Role: tower_model.RestoreTowerRole, Key: token, RemoteId: meta.Local.TowerId,
		LocalId: meta.Core.TowerId,
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/servers/init").WithBody(initParams).Call()
	if err != nil {
		t.Fail(err)
	}

	// Restore journal
	t.Ctx.Notify(
		notify.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_mod.TaskResult{"stage": "Restoring file history", "timestamp": time.Now().UnixMilli()}),
	)

	actions, err := journal.GetAllActionsByTowerId(ctx, meta.Core.TowerId)
	if err != nil {
		t.Fail(err)
	}
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/history").WithBody(actions).Call()
	t.ReqNoErr(err)

	// Restore users
	t.Ctx.Notify(notify.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_mod.TaskResult{"stage": "Restoring users", "timestamp": time.Now().UnixMilli()}))

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
	t.Ctx.Notify(notify.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_mod.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()}))

	tokens, err := auth.GetAllTokensByTowerId(ctx, meta.Core.TowerId)
	if err != nil {
		t.Fail(err)
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/keys").WithBody(tokens).Call()
	t.ReqNoErr(err)

	// Restore instances
	t.Ctx.Notify(notify.NewTaskNotification(
		t, websocket_mod.RestoreProgressEvent, task_mod.TaskResult{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()},
	))

	towers, err := tower_model.GetAllTowersByTowerId(ctx, meta.Core.TowerId)
	if err != nil {
		t.Fail(err)
	}
	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/instances").WithBody(towers).Call()
	t.ReqNoErr(err)

	// Restore files
	// t.Ctx.Notify(notify.NewTaskNotification(
	// 	t, websocket_mod.RestoreProgressEvent, task_mod.TaskResult{"stage": "Restoring files", "timestamp": time.Now().UnixMilli()},
	// ))
	// for i, lt := range lts {
	// 	latest := lt.GetLatestAction()
	// 	if latest.GetActionType() == history.FileDelete {
	// 		continue
	// 	}
	// 	portable := fs.ParsePortable(latest.GetDestinationPath())
	// 	if portable.IsDir() {
	// 		continue
	// 	}
	//
	// 	f, err := fileService.GetFileByContentId(lt.GetContentId())
	// 	if err != nil {
	// 		t.Ctx.Log().Error().Stack().Err(err).Msg("")
	// 		continue
	// 	}
	// 	if f == nil {
	// 		t.Ctx.Log().Error().Msgf("File not found for contentId [%s]", lt.GetContentId())
	// 		continue
	// 	}
	//
	// 	bs, err := f.ReadAll()
	// 	if err != nil {
	// 		t.ReqNoErr(err)
	// 	}
	// 	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/file").WithBodyBytes(bs).WithQuery(
	// 		"fileId", lt.ID(),
	// 	).Call()
	// 	if err != nil {
	// 		t.ReqNoErr(err)
	// 	}
	//
	// 	t.Ctx.Notify(notify.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, task_mod.TaskResult{"filesTotal": len(lts), "filesRestored": i}))
	// }

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/complete").Call()
	if err != nil {
		t.ReqNoErr(err)
	}

	meta.Core.SetReportedRole(tower_model.CoreTowerRole)
	t.Ctx.Notify(notify.NewTaskNotification(t, websocket_mod.RestoreCompleteEvent, nil))

	// Disconnect the core client to force a reconnection
	// coreClient := meta.Pack.ClientService.GetClientByServerId(meta.Core.TowerId)
	// meta.Pack.ClientService.ClientDisconnect(coreClient)

	t.Success()
}
