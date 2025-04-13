package reshape

import (
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/structs"
)

func TowerToTowerInfo(tower *tower_model.Instance) structs.TowerInfo {
	return structs.TowerInfo{
		Id:   tower.TowerId,
		Name: tower.Name,
		// UsingKey:     tower.,
		Role:    string(tower.Role),
		Address: tower.Address,
		// ReportedRole: tower.ReportedRole,
		LastBackup: tower.LastBackup,
		// BackupSize:   tower.back,
		// UserCount:    tower.UserCount,
		IsThisServer: tower.IsThisTower,
		Started:      true,
	}
}

func TowerInfoToTower(t structs.TowerInfo) *tower_model.Instance {
	return &tower_model.Instance{
		TowerId:     t.Id,
		Name:        t.Name,
		Role:        tower_model.TowerRole(t.Role),
		IsThisTower: t.IsThisServer,
		Address:     t.Address,
		LastBackup:  t.LastBackup,
	}
}
