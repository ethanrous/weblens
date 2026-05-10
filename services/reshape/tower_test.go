package reshape_test

import (
	"testing"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/services/reshape"
)

func TestTowerToTowerInfo_LocalIncludesBuildVersion(t *testing.T) {
	t.Setenv("WEBLENS_BUILD_VERSION", "test-version-1.2.3")

	ctx := newTestAppContext(t)

	tower := tower_model.Instance{
		TowerID:     "test-tower",
		Name:        "Test Tower",
		Role:        tower_model.RoleCore,
		IsThisTower: true,
	}

	info := reshape.TowerToTowerInfo(ctx, tower)

	if info.BuildVersion != "test-version-1.2.3" {
		t.Fatalf("expected BuildVersion %q, got %q", "test-version-1.2.3", info.BuildVersion)
	}
}
