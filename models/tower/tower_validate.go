package tower

import (
	"context"

	"github.com/ethanrous/weblens/modules/wlerrors"
)

var (
	// ErrNilTower indicates a nil tower instance was provided.
	ErrNilTower = wlerrors.New("tower is nil")
	// ErrEmptyID indicates a tower has an empty ID.
	ErrEmptyID = wlerrors.New("tower ID is empty")
	// ErrEmptyName indicates a tower has an empty name.
	ErrEmptyName = wlerrors.New("tower name is empty")
	// ErrEmptyRole indicates a tower has an empty role.
	ErrEmptyRole = wlerrors.New("tower role is empty")
	// ErrEmptyAddress indicates a tower has an empty address.
	ErrEmptyAddress = wlerrors.New("tower address is empty")
	// ErrEmptyPort indicates a tower has an empty port.
	ErrEmptyPort = wlerrors.New("tower port is empty")
)

func validateNewTower(_ context.Context, t *Instance) error {
	if t == nil {
		return wlerrors.WithStack(ErrNilTower)
	}

	if t.TowerID == "" {
		return wlerrors.WithStack(ErrEmptyID)
	}

	if t.Name == "" {
		return wlerrors.WithStack(ErrEmptyName)
	}

	if t.Role == "" {
		return wlerrors.WithStack(ErrEmptyRole)
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
