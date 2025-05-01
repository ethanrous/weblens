package user

import (
	"context"
	"errors"

	file_model "github.com/ethanrous/weblens/models/file"
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
