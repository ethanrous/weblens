// Package user manages higher level user operations
package user

import (
	"context"
	"errors"

	file_model "github.com/ethanrous/weblens/models/file"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

func init() {
	startup.RegisterHook(loadUserService)
}

func loadUserService(ctx context.Context, _ config.Provider) error {
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

	userRoot, err := appCtx.FileService.GetFileByID(appCtx, file_model.UsersTreeKey)
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
		homeFolder, err := appCtx.FileService.GetFileByID(appCtx, u.HomeID)
		if errors.Is(err, file_model.ErrFileNotFound) {
			homePath := file_model.UsersRootPath.Child(u.Username, true)

			homeFolder, err = appCtx.FileService.GetFileByFilepath(appCtx, homePath)
			if err != nil {
				appCtx.Log().Debug().Msgf("Home folder [%s][%s] not found, creating for user %s", u.HomeID, homePath, u.Username)

				homeFolder, err = appCtx.FileService.CreateFolder(appCtx, userRoot, u.Username)
				if err != nil {
					return err
				}
			}

			err = u.UpdateHomeID(appCtx, homeFolder.ID())
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		trashFile, err := appCtx.FileService.GetFileByID(appCtx, u.TrashID)
		if errors.Is(err, file_model.ErrFileNotFound) || trashFile.ID() != u.TrashID {
			trashPath := file_model.UsersRootPath.Child(u.Username, true).Child(file_model.UserTrashDirName, true)

			trashFolder, err := appCtx.FileService.GetFileByFilepath(appCtx, trashPath)
			if err != nil {
				appCtx.Log().Debug().Msgf("Trash folder not found, creating for user %s", u.Username)

				trashFolder, err = appCtx.FileService.CreateFolder(appCtx, homeFolder, file_model.UserTrashDirName)
				if err != nil {
					return err
				}
			}

			err = u.UpdateTrashID(appCtx, trashFolder.ID())
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}
