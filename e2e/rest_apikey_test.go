package e2e_test

import (
	"net/http"
	"testing"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIKey(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a new API key
	tokenInfo, resp, err := client.APIKeysAPI.CreateAPIKey(t.Context()).Params(openapi.APIKeyParams{
		Name: "test-api-key",
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "test-api-key", tokenInfo.GetNickname())
	assert.NotEmpty(t, tokenInfo.GetToken())
	assert.NotEmpty(t, tokenInfo.GetId())
}

func TestGetAPIKeys(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a couple of API keys
	_, _, err = client.APIKeysAPI.CreateAPIKey(t.Context()).Params(openapi.APIKeyParams{
		Name: "key-one",
	}).Execute()
	require.NoError(t, err)

	_, _, err = client.APIKeysAPI.CreateAPIKey(t.Context()).Params(openapi.APIKeyParams{
		Name: "key-two",
	}).Execute()
	require.NoError(t, err)

	// Get all API keys
	keys, resp, err := client.APIKeysAPI.GetAPIKeys(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Should have at least the 2 we just created (plus any existing admin key)
	assert.GreaterOrEqual(t, len(keys), 2)

	// Verify our keys are in the list
	foundKeyOne := false
	foundKeyTwo := false
	for _, k := range keys {
		if k.GetNickname() == "key-one" {
			foundKeyOne = true
		}
		if k.GetNickname() == "key-two" {
			foundKeyTwo = true
		}
	}
	assert.True(t, foundKeyOne, "key-one should be in the list")
	assert.True(t, foundKeyTwo, "key-two should be in the list")
}

func TestDeleteAPIKey(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create an API key to delete
	tokenInfo, _, err := client.APIKeysAPI.CreateAPIKey(t.Context()).Params(openapi.APIKeyParams{
		Name: "key-to-delete",
	}).Execute()
	require.NoError(t, err)

	tokenID := tokenInfo.GetId()

	// Verify the key exists
	keys, _, err := client.APIKeysAPI.GetAPIKeys(t.Context()).Execute()
	require.NoError(t, err)
	found := false
	for _, k := range keys {
		if k.GetId() == tokenID {
			found = true
			break
		}
	}
	require.True(t, found, "key should exist before deletion")

	// Delete the key
	resp, err := client.APIKeysAPI.DeleteAPIKey(t.Context(), tokenID).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify the key no longer exists
	keys, _, err = client.APIKeysAPI.GetAPIKeys(t.Context()).Execute()
	require.NoError(t, err)
	found = false
	for _, k := range keys {
		if k.GetId() == tokenID {
			found = true
			break
		}
	}
	assert.False(t, found, "key should not exist after deletion")
}
