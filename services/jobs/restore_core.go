package jobs

import (
	"time"

	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/ethanrous/weblens/services/reshape"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RestoreCore restores a core server from backup data.
func RestoreCore(tsk *task.Task) {

	meta := tsk.GetMeta().(job.RestoreCoreMeta)

	ctx, ok := tsk.Ctx.(ctxservice.AppContext)
	if !ok {
		tsk.Fail(wlerrors.New("Failed to cast context to FilerContext"))

		return
	}

	type restoreInitParams struct {
		Name     string           `json:"name"`
		Role     tower_model.Role `json:"role"`
		RemoteID string           `json:"remoteID"`
		LocalID  string           `json:"localID"`
		Key      *auth.Token      `json:"usingKeyInfo"`
	}

	// Notify client of restore failure, if any
	tsk.SetErrorCleanup(
		func(errTsk *task.Task) {
			ctx.Notify(ctx, notify.NewTaskNotification(errTsk, websocket_mod.RestoreFailedEvent, task.Result{"error": errTsk.ReadError().Error()}))
		},
	)

	// Prime server to be restored. This will fail if the server is already initialized

	ctx.Notify(ctx, notify.NewTaskNotification(
		tsk, websocket_mod.RestoreProgressEvent, task.Result{"stage": "Connecting to remote", "timestamp": time.Now().UnixMilli()},
	))

	tokenID, err := primitive.ObjectIDFromHex(meta.Core.OutgoingKey)
	if err != nil {
		tsk.Fail(err)
	}

	token, err := auth.GetTokenByID(ctx, tokenID)
	if err != nil {
		tsk.Fail(err)
	}

	initParams := restoreInitParams{
		Name: meta.Core.Name, Role: tower_model.RoleRestore, Key: token, RemoteID: meta.Local.TowerID,
		LocalID: meta.Core.TowerID,
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/servers/init").WithBody(initParams).Call()
	if err != nil {
		tsk.Fail(err)
	}

	// Restore journal
	ctx.Notify(ctx,
		notify.NewTaskNotification(tsk, websocket_mod.RestoreProgressEvent, task.Result{"stage": "Restoring file history", "timestamp": time.Now().UnixMilli()}),
	)

	actions, err := journal.GetAllActionsByTowerID(ctx, meta.Core.TowerID)
	if err != nil {
		tsk.Fail(err)
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/history").WithBody(actions).Call()
	tsk.ReqNoErr(err)

	// Restore users
	ctx.Notify(ctx, notify.NewTaskNotification(tsk, websocket_mod.RestoreProgressEvent, task.Result{"stage": "Restoring users", "timestamp": time.Now().UnixMilli()}))

	users, err := user_model.GetAllUsers(tsk.Ctx)
	if err != nil {
		tsk.Fail(err)
	}

	userInfos := make([]structs.UserInfo, 0, len(users))
	for _, u := range users {
		userInfos = append(userInfos, reshape.UserToUserInfo(tsk.Ctx, u))
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/users").WithBody(users).Call()
	tsk.ReqNoErr(err)

	// Restore keys
	ctx.Notify(ctx, notify.NewTaskNotification(tsk, websocket_mod.RestoreProgressEvent, task.Result{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()}))

	tokens, err := auth.GetAllTokensByTowerID(ctx, meta.Core.TowerID)
	if err != nil {
		tsk.Fail(err)
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/keys").WithBody(tokens).Call()
	tsk.ReqNoErr(err)

	// Restore instances
	ctx.Notify(ctx, notify.NewTaskNotification(
		tsk, websocket_mod.RestoreProgressEvent, task.Result{"stage": "Restoring api keys", "timestamp": time.Now().UnixMilli()},
	))

	towers, err := tower_model.GetAllTowersByTowerID(ctx, meta.Core.TowerID)
	if err != nil {
		tsk.Fail(err)
	}

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/instances").WithBody(towers).Call()
	tsk.ReqNoErr(err)

	// Restore files
	// ctx.Notify(ctx, notify.NewTaskNotification(
	// 	t, websocket_mod.RestoreProgressEvent, *task.TaskResult{"stage": "Restoring files", "timestamp": time.Now().UnixMilli()},
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
	// 	f, err := fileService.GetFileByContentID(lt.GetContentID())
	// 	if err != nil {
	// 		t.Log().Error().Stack().Err(err).Msg("")
	// 		continue
	// 	}
	// 	if f == nil {
	// 		t.Log().Error().Msgf("File not found for contentID [%s]", lt.GetContentID())
	// 		continue
	// 	}
	//
	// 	bs, err := f.ReadAll()
	// 	if err != nil {
	// 		t.ReqNoErr(err)
	// 	}
	// 	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/file").WithBodyBytes(bs).WithQuery(
	// 		"fileID", lt.ID(),
	// 	).Call()
	// 	if err != nil {
	// 		t.ReqNoErr(err)
	// 	}
	//
	// 	ctx.Notify(ctx, notify.NewTaskNotification(t, websocket_mod.RestoreProgressEvent, *task.TaskResult{"filesTotal": len(lts), "filesRestored": i}))
	// }

	_, err = proxy.NewCoreRequest(meta.Core, "POST", "/restore/complete").Call()
	if err != nil {
		tsk.ReqNoErr(err)
	}

	meta.Core.SetReportedRole(tower_model.RoleCore)
	ctx.Notify(ctx, notify.NewTaskNotification(tsk, websocket_mod.RestoreCompleteEvent, nil))

	// Disconnect the core client to force a reconnection
	// coreClient := meta.Pack.ClientService.GetClientByServerID(meta.Core.TowerID)
	// meta.Pack.ClientService.ClientDisconnect(coreClient)

	tsk.Success()
}
