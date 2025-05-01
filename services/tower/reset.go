package tower

import (
	"context"

	tower_model "github.com/ethanrous/weblens/models/tower"
	context_mod "github.com/ethanrous/weblens/modules/context"
)

func ResetTower(ctx context.Context) error {
	_, err := tower_model.ResetLocal(ctx)
	if err != nil {
		return err
	}

	context_mod.ToZ(ctx).ClearCache()

	return nil
}
