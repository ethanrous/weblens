package reshape

import (
	"context"

	openapi "github.com/ethanrous/weblens/api"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/structs"
	context_service "github.com/ethanrous/weblens/services/context"
)

func TowerToTowerInfo(ctx context.Context, tower tower_model.Instance) structs.TowerInfo {
	online := false
	if tower.IsThisTower {
		online = true
	} else {
		appCtx, ok := context_service.FromContext(ctx)
		if ok {
			client := appCtx.ClientService.GetClientByTowerId(tower.TowerId)
			online = client != nil
		}
	}

	return structs.TowerInfo{
		Id:           tower.TowerId,
		Name:         tower.Name,
		Role:         string(tower.Role),
		Address:      tower.Address,
		LastBackup:   tower.LastBackup,
		IsThisServer: tower.IsThisTower,
		Started:      true,

		// TODO: Get real reported role
		ReportedRole: string(tower.Role),
		Online:       online,
	}
}

func TowerInfoToTower(t structs.TowerInfo) *tower_model.Instance {
	return &tower_model.Instance{
		TowerId:     t.Id,
		Name:        t.Name,
		Role:        tower_model.Role(t.Role),
		IsThisTower: t.IsThisServer,
		Address:     t.Address,
		LastBackup:  t.LastBackup,
	}
}

func ApiTowerInfoToTower(t openapi.TowerInfo) tower_model.Instance {
	return tower_model.Instance{
		TowerId:     t.Id,
		Name:        t.Name,
		Role:        tower_model.Role(t.Role),
		IsThisTower: false,
		Address:     t.CoreAddress,
		LastBackup:  t.LastBackup,
	}
}
