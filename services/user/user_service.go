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

func loadUserService(c context.Context, cnf config.ConfigProvider) error {
	ctx := c.(context_service.AppContext)

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	if local.Role != tower_model.RoleCore {
		ctx.Log().Debug().Msg("Not a core tower, skipping user service load")

		return nil
	}

	userRoot, err := ctx.FileService.GetFileById(file_model.UsersTreeKey)
	if err != nil {
		ctx.Log().Debug().Msg("Deferring user service load")

		return startup.ErrDeferStartup
	}

	ctx.Log().Debug().Msg("Loading user service")

	users, err := user.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	for _, u := range users {
		homeFolder, err := ctx.FileService.GetFileById(u.HomeId)
		if errors.Is(err, file_model.ErrFileNotFound) {
			homePath := file_model.UsersRootPath.Child(u.Username, true)

			homeFolder, err = ctx.FileService.GetFileByFilepath(ctx, homePath)
			if err != nil {
				ctx.Log().Debug().Msgf("Home folder [%s][%s] not found, creating for user %s", u.HomeId, homePath, u.Username)

				homeFolder, err = ctx.FileService.CreateFolder(ctx, userRoot, u.Username)
				if err != nil {
					return err
				}
			}

			err = u.UpdateHomeId(ctx, homeFolder.ID())
			if err != nil {
				return err
			}
		}

		_, err = ctx.FileService.GetFileById(u.TrashId)
		if errors.Is(err, file_model.ErrFileNotFound) {
			trashPath := file_model.UsersRootPath.Child(u.Username, true).Child(file_model.UserTrashDirName, true)

			trashFolder, err := ctx.FileService.GetFileByFilepath(ctx, trashPath)
			if err != nil {
				ctx.Log().Debug().Msgf("Trash folder not found, creating for user %s", u.Username)

				trashFolder, err = ctx.FileService.CreateFolder(ctx, homeFolder, file_model.UserTrashDirName)
				if err != nil {
					return err
				}
			}

			err = u.UpdateTrashId(ctx, trashFolder.ID())
			if err != nil {
				return err
			}
		}
	}

	// Load user service
	// userService := user.NewUserService()
	// userService.Load()
	// log.Info().Msg("User service loaded")

	return nil
}
