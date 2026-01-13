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

func TestGetMediaBatch(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get media batch with empty request - should return empty result
	mediaBatch, resp, err := client.MediaAPI.GetMedia(t.Context()).Request(openapi.MediaBatchParams{
		Page:  openapi.PtrInt32(0),
		Limit: openapi.PtrInt32(10),
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, mediaBatch)
	// Fresh server has no media
	assert.Equal(t, int32(0), mediaBatch.GetTotalMediaCount())
}

func TestGetMediaTypes(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get media types - should return type dictionaries
	mediaTypes, resp, err := client.MediaAPI.GetMediaTypes(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, mediaTypes)
	// Should have some mime and extension mappings
	assert.NotNil(t, mediaTypes.GetMimeMap())
	assert.NotNil(t, mediaTypes.GetExtMap())
}

func TestGetRandomMedia(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get random media - should return empty on fresh server
	mediaBatch, resp, err := client.MediaAPI.GetRandomMedia(t.Context()).Count(5).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, mediaBatch)
	// Fresh server has no media
	assert.Equal(t, 0, len(mediaBatch.GetMedia()))
}
