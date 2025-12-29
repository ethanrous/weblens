package tower

import (
	"context"

	"github.com/ethanrous/weblens/modules/errors"
)

var (
	// ErrNilTower indicates a nil tower instance was provided.
	ErrNilTower = errors.New("tower is nil")
	// ErrEmptyID indicates a tower has an empty ID.
	ErrEmptyID = errors.New("tower ID is empty")
	// ErrEmptyName indicates a tower has an empty name.
	ErrEmptyName = errors.New("tower name is empty")
	// ErrEmptyRole indicates a tower has an empty role.
	ErrEmptyRole = errors.New("tower role is empty")
	// ErrEmptyAddress indicates a tower has an empty address.
	ErrEmptyAddress = errors.New("tower address is empty")
	// ErrEmptyPort indicates a tower has an empty port.
	ErrEmptyPort = errors.New("tower port is empty")
)

func validateNewTower(_ context.Context, t *Instance) error {
	if t == nil {
		return errors.WithStack(ErrNilTower)
	}

	if t.TowerID == "" {
		return errors.WithStack(ErrEmptyID)
	}

	if t.Name == "" {
		return errors.WithStack(ErrEmptyName)
	}

	if t.Role == "" {
		return errors.WithStack(ErrEmptyRole)
	}

	// local, err := GetLocal(ctx)
	// if err != nil {
	// 	return err
	// }
	//
	// if local.Role == RoleBackup && t.Role == RoleCore {
	// 	if t.Address == "" {
	// 		return errors.WithStack(ErrEmptyAddress)
	// 	}
	// }

	return nil
}
