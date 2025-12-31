package tower

import (
	"context"
	"net/http"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/modules/wlerrors"
	access_service "github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

func init() {
	startup.RegisterHook(initTower)
}

func initTower(ctx context.Context, cnf config.Provider) error {
	initRole := tower_model.Role(cnf.InitRole)

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return context_service.ErrNoContext
	}

	if initRole != tower_model.RoleInit {
		_, err := appCtx.FileService.GetFileByID(appCtx, file_model.UsersTreeKey)
		if err != nil {
			return startup.ErrDeferStartup
		}
	}

	localTower, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	if localTower.Role != tower_model.RoleInit {
		return nil
	}

	if localTower.Role != initRole && initRole == tower_model.RoleCore {
		err = db.WithTransaction(ctx, func(sessionCtx context.Context) error {
			return InitializeCoreServer(sessionCtx, structs.InitServerParams{
				Name:     "Weblens Core",
				Username: "admin",
				Password: "adminadmin1",
				FullName: "Weblens Admin",
				Role:     string(tower_model.RoleCore),
			})
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func newOwner(ctx context.Context, initBody structs.InitServerParams) (*user_model.User, error) {
	owner := &user_model.User{
		Username:    initBody.Username,
		Password:    initBody.Password,
		DisplayName: initBody.FullName,
		UserPerms:   user_model.UserPermissionOwner,
		Activated:   true,
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, wlerrors.New("failed to get request context")
	}

	if exists, err := user_model.GetUserByUsername(ctx, owner.Username); err == nil {
		err = exists.Delete(ctx)
		if err != nil {
			return nil, err
		}
	}

	if tower_model.Role(initBody.Role) == tower_model.RoleCore {
		// Create user home directory
		err := appCtx.FileService.CreateUserHome(ctx, owner)
		if err != nil {
			return nil, err
		}

		if owner.HomeID == "" {
			return nil, wlerrors.New("failed to create user home directory")
		}
	}

	err := user_model.SaveUser(ctx, owner)
	if err != nil {
		return nil, err
	}

	reqCtx, ok := context_service.ReqFromContext(ctx)
	if ok {
		reqCtx.Requester = owner

		err = access_service.SetSessionToken(reqCtx)
		if err != nil {
			reqCtx.Error(http.StatusInternalServerError, err)
		}
	}

	return owner, nil
}

// InitializeCoreServer initializes a tower as a core server with the provided configuration.
func InitializeCoreServer(ctx context.Context, initBody structs.InitServerParams) error {
	if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
		return wlerrors.New("missing required fields for core server initialization")
	}

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	local.Role = tower_model.RoleCore
	local.Name = initBody.Name
	local.Address = initBody.CoreAddress

	err = tower_model.UpdateTower(ctx, &local)
	if err != nil {
		return err
	}

	cnf := config.GetConfig()
	cnf.InitRole = string(tower_model.RoleCore)

	err = startup.RunStartups(ctx, cnf)
	if err != nil {
		return err
	}

	_, err = newOwner(ctx, initBody)
	if err != nil {
		return err
	}

	return nil
}

// InitializeBackupServer initializes a tower as a backup server connected to a core server.
func InitializeBackupServer(ctx context.Context, initBody structs.InitServerParams) error {
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	local.Role = tower_model.RoleBackup
	local.Name = initBody.Name

	err = tower_model.UpdateTower(ctx, &local)
	if err != nil {
		return err
	}

	core := tower_model.Instance{
		Role:        tower_model.RoleCore,
		Address:     initBody.CoreAddress,
		OutgoingKey: initBody.CoreKey,
	}

	coreInfo, err := Ping(ctx, core)
	if err != nil {
		return err
	}

	if tower_model.Role(coreInfo.GetRole()) != tower_model.RoleCore {
		return tower_model.ErrNotCore
	}

	core.TowerID = coreInfo.GetId()
	core.Name = coreInfo.GetName()

	err = tower_model.SaveTower(ctx, &core)
	if err != nil {
		return err
	}

	_, err = newOwner(ctx, initBody)
	if err != nil {
		return err
	}

	return nil
}
