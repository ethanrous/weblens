package user

import (
	"context"
	"errors"

	file_model "github.com/ethanrous/weblens/models/file"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	context_service "github.com/ethanrous/weblens/services/context"
)

func init() {
	startup.RegisterStartup(loadUserService)
}

func loadUserService(ctx context.Context, cnf config.ConfigProvider) error {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return context_service.ErrNoContext
	}

	local, err := tower_model.GetLocal(appCtx)
	if err != nil {
		return err
	}

	if local.Role != tower_model.RoleCore {
		appCtx.Log().Debug().Msg("Not a core tower, skipping user service load")

		return nil
	}

	userRoot, err := appCtx.FileService.GetFileById(appCtx, file_model.UsersTreeKey)
	if err != nil {
		appCtx.Log().Debug().Msg("Deferring user service load")

		return startup.ErrDeferStartup
	}

	appCtx.Log().Debug().Msg("Loading user service")

	users, err := user.GetAllUsers(appCtx)
	if err != nil {
		return err
	}

	for _, u := range users {
		homeFolder, err := appCtx.FileService.GetFileById(appCtx, u.HomeId)
		if errors.Is(err, file_model.ErrFileNotFound) {
			homePath := file_model.UsersRootPath.Child(u.Username, true)

			homeFolder, err = appCtx.FileService.GetFileByFilepath(appCtx, homePath)
			if err != nil {
				appCtx.Log().Debug().Msgf("Home folder [%s][%s] not found, creating for user %s", u.HomeId, homePath, u.Username)

				homeFolder, err = appCtx.FileService.CreateFolder(appCtx, userRoot, u.Username)
				if err != nil {
					return err
				}
			}

			err = u.UpdateHomeId(appCtx, homeFolder.ID())
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		_, err = appCtx.FileService.GetFileById(appCtx, u.TrashId)
		if errors.Is(err, file_model.ErrFileNotFound) {
			trashPath := file_model.UsersRootPath.Child(u.Username, true).Child(file_model.UserTrashDirName, true)

			trashFolder, err := appCtx.FileService.GetFileByFilepath(appCtx, trashPath)
			if err != nil {
				appCtx.Log().Debug().Msgf("Trash folder not found, creating for user %s", u.Username)

				trashFolder, err = appCtx.FileService.CreateFolder(appCtx, homeFolder, file_model.UserTrashDirName)
				if err != nil {
					return err
				}
			}

			err = u.UpdateTrashId(appCtx, trashFolder.ID())
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}
