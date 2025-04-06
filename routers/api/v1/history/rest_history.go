package history

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/journal"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
)

func GetLifetimesSince(ctx context.RequestContext) {

	millisString := ctx.Query("timestamp")
	if millisString == "" {
		ctx.Error(http.StatusBadRequest, errors.New("missing timestamp"))
	}

	millis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil || millis < 0 {
		ctx.Error(http.StatusBadRequest, errors.New("invalid timestamp"))
		return
	}

	date := time.UnixMilli(millis)

	actions, err := journal.GetActionsSince(ctx, date)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "failed to get lifetimes"))
		return
	}

	ctx.JSON(http.StatusNotImplemented, actions)
}

func DoFullBackup(ctx context.RequestContext) {
	if ctx.Remote == nil {
		ctx.Error(http.StatusUnauthorized, errors.New("missing tower in request context"))
		return
	}

	millisString := ctx.Query("timestamp")
	if millisString == "" {
		ctx.Error(http.StatusBadRequest, errors.New("missing timestamp"))
		return
	}

	millis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil || millis < 0 {
		ctx.Error(http.StatusBadRequest, errors.New("invalid timestamp"))
		return
	}

	since := time.UnixMilli(millis)

	fileActions, err := journal.GetActionsSince(ctx, since)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "failed to get actions"))
		return
	}

	users, err := user.GetAllUsers(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "failed to get users"))
		return
	}

	towers, err := tower.GetAllTowersByTowerId(ctx, ctx.LocalTowerId)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "failed to get towers"))
		return
	}

	// tokenId, err := primitive.ObjectIDFromHex(ctx.Remote.IncomingKey)
	// if err != nil {
	// 	ctx.Error(http.StatusBadRequest, errors.Wrap(err, "invalid token id"))
	// 	return
	// }
	// auth.GetTokenById(ctx, tokenId)
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
