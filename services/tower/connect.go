package tower

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	api "github.com/ethanrous/weblens/api"
	tower_model "github.com/ethanrous/weblens/models/tower"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/net"
	context_service "github.com/ethanrous/weblens/services/context"
)

const TowerIdHeader = "Weblens-TowerId"

func getApiClient(ctx context.Context, tower tower_model.Instance) (*api.APIClient, error) {
	if tower.Address == "" {
		return nil, fmt.Errorf("tower address is empty")
	}

	towerUrl, err := url.Parse(tower.Address)
	if err != nil {
		return nil, err
	}

	context_mod.ToZ(ctx).Log().Trace().Msgf("Building API client for tower [%s]", tower.Address)

	cnf := api.NewConfiguration()
	cnf.Scheme = towerUrl.Scheme
	cnf.Host = towerUrl.Host
	cnf.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", tower.OutgoingKey)

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get appContext from context")
	}

	cnf.DefaultHeader[TowerIdHeader] = appCtx.LocalTowerId

	return api.NewAPIClient(cnf), nil
}

func Ping(ctx context.Context, tower tower_model.Instance) (*api.TowerInfo, error) {
	client, err := getApiClient(ctx, tower)
	if err != nil {
		return nil, err
	}

	towerInfo, _, err := client.TowersAPI.GetServerInfo(ctx).Execute()
	if err != nil {
		return nil, err
	}

	return towerInfo, nil
}

func GetBackup(ctx context.Context, tower tower_model.Instance, since time.Time) (*api.BackupInfo, error) {
	client, err := getApiClient(ctx, tower)
	if err != nil {
		return nil, err
	}

	backupInfo, resp, err := client.TowersAPI.GetBackupInfo(ctx).Timestamp(strconv.FormatInt(since.UnixMilli(), 10)).Execute()
	if err != nil {
		return nil, net.ReadError(ctx, resp, err)
	}

	return backupInfo, nil
}

func AttachToCore(ctx context.Context, core tower_model.Instance) error {
	client, err := getApiClient(ctx, core)
	if err != nil {
		return err
	}

	delete(client.GetConfig().DefaultHeader, TowerIdHeader)

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	role := string(local.Role)
	newParams := api.NewServerParams{
		ServerId: &local.TowerId,
		Name:     &local.Name,
		Role:     &role,
		UsingKey: &core.OutgoingKey,
	}

	_, _, err = client.TowersAPI.CreateRemote(ctx).Request(newParams).Execute()
	if err != nil {
		return err
	}

	return nil
}
