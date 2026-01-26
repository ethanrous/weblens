// Package proxy provides functionality to communicate with remote tower instances.
package proxy

import (
	"net/url"

	openapi "github.com/ethanrous/weblens/api"
	tower_model "github.com/ethanrous/weblens/models/tower"
)

// APIClientFromTower creates an API client configured to communicate with the specified tower instance.
func APIClientFromTower(tower tower_model.Instance) (*openapi.APIClient, error) {
	address, err := url.Parse(tower.Address)
	if err != nil {
		return nil, err
	}

	apiConfig := openapi.NewConfiguration()
	apiConfig.Host = address.Host
	apiConfig.UserAgent = "Weblens-Tower-Client"

	if tower.OutgoingKey != "" {
		apiConfig.AddDefaultHeader("Authorization", "Bearer "+string(tower.OutgoingKey))
	}

	client := openapi.NewAPIClient(apiConfig)

	return client, nil
}
