package e2e_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
)

func TestServerStartup(t *testing.T) {
	// Setup a core server for the backup to connect to
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	for _, role := range []tower.Role{
		tower.RoleUninitialized,
		tower.RoleCore,
		tower.RoleBackup,
	} {
		t.Run(string(role), func(t *testing.T) {
			setup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(role), CoreAddress: coreSetup.address, CoreToken: coreSetup.token})
			if err != nil {
				log.GlobalLogger().Error().Err(err).Msg("Failed to start test server")
				t.FailNow()
			}

			client := getAPIClientFromConfig(setup.cnf, "")

			apiTowerInfo, _, err := client.TowersAPI.GetServerInfo(t.Context()).Execute()
			if err != nil {
				log.GlobalLogger().Error().Err(err).Msg("Failed to get server info")
				t.FailNow()
			}

			if apiTowerInfo.GetRole() != string(role) {
				t.Fatalf("Expected server role to be '%s', got '%s'", role, apiTowerInfo.GetRole())
			}
		})
	}
}
