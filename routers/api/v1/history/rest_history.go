// Package history provides functionalities for managing and retrieving file history.
package history

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/ethanrous/weblens/services/reshape"
)

// DoFullBackup godoc
//
//	@ID			GetBackupInfo
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Get information about a file
//	@Tags		Towers
//	@Produce	json
//	@Param		timestamp	query		string				true	"Timestamp in milliseconds since epoch"
//	@Success	200			{object}	structs.BackupInfo	"Backup Info"
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/tower/backup [get]
func DoFullBackup(ctx ctxservice.RequestContext) {
	if ctx.Remote.TowerID == "" {
		ctx.Error(http.StatusUnauthorized, wlerrors.New("missing tower in request context"))

		return
	}

	millisString := ctx.Query("timestamp")
	if millisString == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("missing timestamp"))

		return
	}

	millis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil || millis < 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("invalid timestamp"))

		return
	}

	since := time.UnixMilli(millis)

	ctx.Log().Trace().Msgf("Getting backup info since %s", since.String())

	fileActions, err := history.GetActionsAtPathAfter(ctx, wlfs.Filepath{}, since, false)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "failed to get actions"))

		return
	}

	users, err := usermodel.GetAllUsers(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "failed to get users"))

		return
	}

	towers, err := tower.GetAllTowersByTowerID(ctx, ctx.LocalTowerID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "failed to get towers"))

		return
	}

	// tokenID, err := primitive.ObjectIDFromHex(ctx.Remote.IncomingKey)
	// if err != nil {
	// 	ctx.Error(http.StatusBadRequest, errors.Wrap(err, "invalid token id"))
	// 	return
	// }
	// auth.GetTokenByID(ctx, tokenID)
	// TODO: Get tokens from the database
	tokens := make([]*auth.Token, 0)

	res := reshape.NewBackupInfo(
		ctx,
		fileActions,
		users,
		towers,
		tokens,
	)

	ctx.JSON(http.StatusOK, res)
}

// GetPagedHistoryActions godoc
//
//	@ID			GetPagedHistoryActions
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Get a page of file actions
//	@Tags		Towers
//	@Produce	json
//	@Param		page		query		int					false	"Page number"
//	@Param		pageSize	query		int					false	"Number of items per page"
//	@Success	200			{array}	history.FileAction	"File Actions"
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/tower/history [get]
func GetPagedHistoryActions(ctx ctxservice.RequestContext) {
	page, err := ctx.QueryInt("page")
	if err != nil {
		ctx.Error(http.StatusBadRequest, wlerrors.New("invalid page number"))

		return
	}

	pageSize, err := ctx.QueryIntDefault("pageSize", 50)
	if err != nil {
		ctx.Error(http.StatusBadRequest, wlerrors.New("invalid page size"))

		return
	}

	actions, err := journal.GetActionsPage(ctx, int(pageSize), int(page))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "failed to get actions"))

		return
	}

	actionInfos := make([]wlstructs.FileActionInfo, 0, len(actions))
	for _, action := range actions {
		actionInfos = append(actionInfos, reshape.FileActionToFileActionInfo(action))
	}

	ctx.JSON(http.StatusOK, actionInfos)
}
