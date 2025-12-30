package reshape

import (
	"context"

	openapi "github.com/ethanrous/weblens/api"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/structs"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// TowerToTowerInfo converts a tower Instance to a TowerInfo structure suitable for API responses.
func TowerToTowerInfo(ctx context.Context, tower tower_model.Instance) structs.TowerInfo {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		panic("not an app context")
	}

	online := false
	backupSize := int64(0)

	if tower.IsThisTower {
		online = true
	} else {
		client := appCtx.ClientService.GetClientByTowerID(tower.TowerID)
		online = client != nil

		towerBackupDir, err := appCtx.FileService.GetFileByID(ctx, tower.TowerID)
		if err == nil {
			backupSize = towerBackupDir.Size()
		}
	}

	return structs.TowerInfo{
		ID:           tower.TowerID,
		Name:         tower.Name,
		Role:         string(tower.Role),
		Address:      tower.Address,
		LastBackup:   tower.LastBackup,
		IsThisServer: tower.IsThisTower,
		Started:      true,
		BackupSize:   backupSize,

		// TODO: Get real reported role
		ReportedRole: string(tower.Role),
		Online:       online,
	}
}

// TowerInfoToTower converts a TowerInfo from the API to a tower Instance.
func TowerInfoToTower(t structs.TowerInfo) *tower_model.Instance {
	return &tower_model.Instance{
		TowerID:     t.ID,
		Name:        t.Name,
		Role:        tower_model.Role(t.Role),
		IsThisTower: t.IsThisServer,
		Address:     t.Address,
		LastBackup:  t.LastBackup,
	}
}

// APITowerInfoToTower converts an OpenAPI TowerInfo to a tower Instance.
func APITowerInfoToTower(t openapi.TowerInfo) tower_model.Instance {
	return tower_model.Instance{
		TowerID:     t.Id,
		Name:        t.Name,
		Role:        tower_model.Role(t.Role),
		IsThisTower: false,
		Address:     t.CoreAddress,
		LastBackup:  t.LastBackup,
	}
}
