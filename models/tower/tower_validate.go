package tower

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrNilTower     = errors.New("tower is nil")
	ErrEmptyID      = errors.New("tower ID is empty")
	ErrEmptyName    = errors.New("tower name is empty")
	ErrEmptyRole    = errors.New("tower role is empty")
	ErrEmptyAddress = errors.New("tower address is empty")
	ErrEmptyPort    = errors.New("tower port is empty")
)

func validateNewTower(ctx context.Context, t *Instance) error {
	if t == nil {
		return errors.WithStack(ErrNilTower)
	}

	if t.TowerId == "" {
		return errors.WithStack(ErrEmptyID)
	}

	if t.Name == "" {
		return errors.WithStack(ErrEmptyName)
	}

	if t.Role == "" {
		return errors.WithStack(ErrEmptyRole)
	}

	local, err := GetLocal(ctx)
	if err != nil {
		return err
	}

	if local.Role == RoleBackup && t.Role == RoleCore {
		if t.Address == "" {
			return errors.WithStack(ErrEmptyAddress)
		}
	}

	return nil
}
