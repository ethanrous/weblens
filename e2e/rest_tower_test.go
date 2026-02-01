package e2e_test

import (
	"net/http"
	"testing"

	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetServerInfo(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get server info - no auth required
	serverInfo, resp, err := client.TowersAPI.GetServerInfo(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, serverInfo.GetId())
	assert.Equal(t, "core", serverInfo.GetRole())
}

func TestGetRemotes(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get remotes (admin only) - should return empty list for fresh server
	remotes, resp, err := client.TowersAPI.GetRemotes(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Fresh server should have no remotes
	assert.Equal(t, 0, len(remotes))
}

func TestEnableTraceLogging(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Enable trace logging (admin only)
	resp, err := client.TowersAPI.EnableTraceLogging(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRunningTasks(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get running tasks (admin only)
	tasks, resp, err := client.TowersAPI.GetRunningTasks(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Tasks array should be non-nil (may be empty)
	assert.NotNil(t, tasks)
}

func TestFlushCache(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Flush cache (admin only)
	respBody, resp, err := client.TowersAPI.FlushCache(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Cache flushed successfully", respBody.GetMessage())
}

func TestGetConfig(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get config (admin only)
	cfg, resp, err := client.FeatureFlagsAPI.GetFlags(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, cfg)
}

// Note: TestSetConfig is skipped because the OpenAPI interface for ConfigValue
// doesn't match the server's expected format (map vs direct bool value)
