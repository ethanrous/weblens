package tower

import (
	"context"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/errors"
	context_service "github.com/ethanrous/weblens/services/context"
)

func ResetTower(ctx context.Context) error {
	_, err := tower_model.ResetLocal(ctx)
	if err != nil {
		return err
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return errors.New("failed to get app context")
	}

	appCtx.ClearCache()

	return nil
}
