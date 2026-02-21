package tower

import (
	"context"
	"net/http"

	"github.com/ethanrous/weblens/models/auth"
	file_model "github.com/ethanrous/weblens/models/file"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
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
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return context_service.ErrNoContext
	}

	localTower, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	// If no init role is set, skip automatic initialization.
	if localTower.Role != tower_model.RoleUninitialized {
		return nil
	}

	// Check if initialization is needed based on the configured role. If the
	// config has an init role, and the tower is uninitialized, we set up the server automatically.
	initRole := tower_model.Role(cnf.InitRole)

	// Check for required files based on the init role. If they are not present, defer initialization.
	switch initRole {
	case tower_model.RoleCore:
		_, err := appCtx.FileService.GetFileByID(appCtx, file_model.UsersTreeKey)
		if err != nil {
			return startup.ErrDeferStartup
		}
	case tower_model.RoleBackup:
		_, err := appCtx.FileService.GetFileByID(appCtx, file_model.BackupTreeKey)
		if err != nil {
			return startup.ErrDeferStartup
		}
	}

	if localTower.Role == initRole {
		return nil
	}

	// Perform initialization based on the configured role.
	// No transaction wrapper - initialization calls RunStartups which includes
	// operations like listIndexes that cannot run inside a MongoDB transaction.
	// Initialization is idempotent and can be retried on failure.
	switch initRole {
	case tower_model.RoleCore:
		err = InitializeCoreServer(ctx, structs.InitServerParams{
			Name:     "Weblens Core",
			Username: "admin",
			Password: "adminadmin1",
			FullName: "Weblens Admin",
			Role:     string(tower_model.RoleCore),
		}, cnf)
		if err != nil {
			return err
		}
	case tower_model.RoleBackup:
		err = InitializeBackupServer(ctx, structs.InitServerParams{
			Name:     "Weblens Backup",
			Username: "admin",
			Password: "adminadmin1",
			FullName: "Weblens Backup Admin",
			Role:     string(tower_model.RoleBackup),
		}, cnf)
		if err != nil {
			return err
		}
	}

	return nil
}

func newOwner(ctx context.Context, initBody structs.InitServerParams, withAPIKey bool) (*user_model.User, error) {
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

	// Create user home directory if this is a core server
	if tower_model.Role(initBody.Role) == tower_model.RoleCore {
		err := appCtx.FileService.CreateUserHome(ctx, owner)
		if err != nil {
			return nil, err
		}

		if owner.HomeID == "" {
			return nil, wlerrors.New("failed to create user home directory")
		}
	}

	localTower, err := tower_model.GetLocal(ctx)
	if err != nil {
		return nil, err
	}

	owner.CreatedBy = localTower.TowerID

	err = user_model.SaveUser(ctx, owner)
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

	if withAPIKey {
		log.FromContext(ctx).Debug().Msgf("Generating initial admin API token for user [%s]", owner.GetUsername())

		_, err = auth.GenerateNewToken(ctx, "Initial Admin API Token", owner.GetUsername(), localTower.TowerID)
		if err != nil {
			return nil, err
		}
	}

	return owner, nil
}

// InitializeCoreServer initializes a tower as a core server with the provided configuration.
func InitializeCoreServer(ctx context.Context, initBody structs.InitServerParams, cnf config.Provider) error {
	if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
		return wlerrors.New("missing required fields for core server initialization")
	}

	log.FromContext(ctx).Info().Msgf("Initializing server as CORE with name [%s]", initBody.Name)

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

	cnf.InitRole = string(tower_model.RoleCore)

	err = startup.RunStartups(ctx, cnf)
	if err != nil {
		return err
	}

	_, err = newOwner(ctx, initBody, cnf.GenerateAdminAPIToken)
	if err != nil {
		return err
	}

	return nil
}

// InitializeBackupServer initializes a tower as a backup server connected to a core server.
func InitializeBackupServer(ctx context.Context, initBody structs.InitServerParams, cnf config.Provider) error {
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

	coreAddress := initBody.CoreAddress
	if cnf.CoreAddress != "" {
		coreAddress = cnf.CoreAddress
	} else if coreAddress == "" {
		return wlerrors.New("core address is required for backup server initialization")
	}

	coreKey := initBody.CoreKey
	if cnf.CoreToken != "" {
		coreKey = cnf.CoreToken
	} else if coreKey == "" {
		return wlerrors.New("core token is required for backup server initialization")
	}

	log.FromContext(ctx).Info().Msgf("Initializing server as BACKUP connecting to CORE at [%s]", coreAddress)

	core := tower_model.Instance{
		Role:        tower_model.RoleCore,
		Address:     coreAddress,
		OutgoingKey: coreKey,
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

	_, err = newOwner(ctx, initBody, false)
	if err != nil {
		return err
	}

	err = AttachToCore(ctx, core)
	if err != nil {
		return err
	}

	err = startup.RunStartups(ctx, cnf)
	if err != nil {
		return err
	}

	return nil
}
