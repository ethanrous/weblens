package e2e_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/require"
)

func TestServerStartup(t *testing.T) {
	// Setup a core server for the backup to connect to
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	for _, role := range []tower.Role{
		tower.RoleUninitialized,
		tower.RoleCore,
		tower.RoleBackup,
	} {
		t.Run(string(role), func(t *testing.T) {
			setup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(role), CoreAddress: coreSetup.address, CoreToken: coreSetup.token})
			require.NoError(t, err, "Failed to start test server")

			client := getAPIClientFromConfig(setup.cnf, "")

			apiTowerInfo, _, err := client.TowersAPI.GetServerInfo(t.Context()).Execute()
			require.NoError(t, err, "Failed to get server info")

			if apiTowerInfo.GetRole() != string(role) {
				t.Fatalf("Expected server role to be '%s', got '%s'", role, apiTowerInfo.GetRole())
			}
		})
	}
}
