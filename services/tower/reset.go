package tower

import (
	"context"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// ResetTower resets the local tower to its initial state.
func ResetTower(ctx context.Context) error {
	_, err := tower_model.ResetLocal(ctx)
	if err != nil {
		return err
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return wlerrors.New("failed to get app context")
	}

	appCtx.ClearCache()

	return nil
}
