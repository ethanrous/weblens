// Package tower provides services for connecting and communicating with tower instances.
package tower

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	api "github.com/ethanrous/weblens/api"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/structs"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// TowerIDHeader is the HTTP header name used to identify the source tower in requests.
const TowerIDHeader = "Weblens-TowerID"

type clientOpts struct {
	// noTowerIDHeader indicates whether to omit the TowerID header in requests.
	// This is useful for initial connections where the tower local tower is not yet known by the remote.
	noTowerIDHeader bool
}

func getAPIClient(ctx context.Context, tower tower_model.Instance, o clientOpts) (*api.APIClient, error) {
	if tower.Address == "" {
		return nil, fmt.Errorf("tower address is empty")
	}

	towerURL, err := url.Parse(tower.Address)
	if err != nil {
		return nil, err
	}

	context_mod.ToZ(ctx).Log().Trace().Msgf("Building API client for tower [%s]", tower.Address)

	cnf := api.NewConfiguration()
	cnf.Scheme = towerURL.Scheme
	cnf.Host = towerURL.Host
	cnf.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", tower.OutgoingKey)

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get appContext from context")
	}

	if !o.noTowerIDHeader {
		cnf.DefaultHeader[TowerIDHeader] = appCtx.LocalTowerID
	}

	return api.NewAPIClient(cnf), nil
}

// Ping checks if a tower is reachable and returns its information.
func Ping(ctx context.Context, tower tower_model.Instance) (*api.TowerInfo, error) {
	client, err := getAPIClient(ctx, tower, clientOpts{noTowerIDHeader: true})
	if err != nil {
		return nil, wlerrors.Wrapf(err, "failed to create API client for tower at [%s]", tower.Address)
	}

	towerInfo, _, err := client.TowersAPI.GetServerInfo(ctx).Execute()
	if err != nil {
		return nil, wlerrors.Wrapf(err, "failed to ping tower at [%s]", tower.Address)
	}

	return towerInfo, nil
}

// GetBackup retrieves backup information from a tower for all changes since the specified time.
func GetBackup(ctx context.Context, tower tower_model.Instance, since time.Time) (*structs.BackupInfo, error) {
	client, err := getAPIClient(ctx, tower, clientOpts{})
	if err != nil {
		return nil, err
	}

	apiBackupInfo, resp, err := client.TowersAPI.GetBackupInfo(ctx).Timestamp(strconv.FormatInt(since.UnixMilli(), 10)).Execute()
	if err != nil {
		return nil, netwrk.ReadError(ctx, resp, wlerrors.WithStack(err))
	}

	// Convert API backup info to internal BackupInfo struct
	backupInfo := &structs.BackupInfo{}

	// Convert FileHistory
	backupInfo.FileHistory = make([]structs.FileActionInfo, len(apiBackupInfo.FileHistory))
	for i, apiAction := range apiBackupInfo.FileHistory {
		backupInfo.FileHistory[i] = structs.FileActionInfo{
			FileID:     apiAction.FileID,
			ActionType: apiAction.ActionType,
			EventID:    apiAction.EventID,
			ParentID:   apiAction.ParentID,
			TowerID:    apiAction.TowerID,
			Timestamp:  apiAction.Timestamp,
			Size:       apiAction.Size,
		}
		// Handle optional pointer fields
		if apiAction.Filepath != nil {
			backupInfo.FileHistory[i].Filepath = *apiAction.Filepath
		}

		if apiAction.OriginPath != nil {
			backupInfo.FileHistory[i].OriginPath = *apiAction.OriginPath
		}

		if apiAction.DestinationPath != nil {
			backupInfo.FileHistory[i].DestinationPath = *apiAction.DestinationPath
		}

		if apiAction.ContentID != nil {
			backupInfo.FileHistory[i].ContentID = *apiAction.ContentID
		}
	}

	// Convert Users
	backupInfo.Users = make([]structs.UserInfoArchive, len(apiBackupInfo.Users))

	for i, apiUser := range apiBackupInfo.Users {
		isOnline := false
		if apiUser.IsOnline != nil {
			isOnline = *apiUser.IsOnline
		}

		token := ""
		if apiUser.Token != nil {
			token = *apiUser.Token
		}

		password := ""
		if apiUser.Password != nil {
			password = *apiUser.Password
		}

		backupInfo.Users[i] = structs.UserInfoArchive{
			UserInfo: structs.UserInfo{
				Activated:       apiUser.Activated,
				FullName:        apiUser.FullName,
				HomeID:          apiUser.HomeID,
				PermissionLevel: int(apiUser.PermissionLevel),
				Token:           token,
				TrashID:         apiUser.TrashID,
				Username:        apiUser.Username,
				IsOnline:        isOnline,
			},
			Password: password,
		}
	}

	// Convert Instances (towers)
	backupInfo.Instances = make([]structs.TowerInfo, len(apiBackupInfo.Instances))
	for i, apiTower := range apiBackupInfo.Instances {
		backupInfo.Instances[i] = structs.TowerInfo{
			ID:      apiTower.Id,
			Name:    apiTower.Name,
			Role:    apiTower.Role,
			Address: apiTower.CoreAddress,
			// UsingKey is not available from the API response, will need to be set separately if needed
		}
	}

	// Convert Tokens
	backupInfo.Tokens = make([]structs.TokenInfo, len(apiBackupInfo.Tokens))
	for i, apiToken := range apiBackupInfo.Tokens {
		backupInfo.Tokens[i] = structs.TokenInfo{
			ID:          apiToken.Id,
			CreatedTime: apiToken.CreatedTime,
			LastUsed:    apiToken.LastUsed,
			Nickname:    apiToken.Nickname,
			Owner:       apiToken.Owner,
			RemoteUsing: apiToken.RemoteUsing,
			CreatedBy:   apiToken.CreatedBy,
			Token:       apiToken.Token,
		}
	}

	// Convert LifetimesCount
	if apiBackupInfo.LifetimesCount != nil {
		backupInfo.LifetimesCount = int(*apiBackupInfo.LifetimesCount)
	}

	return backupInfo, nil
}

// AttachToCore registers this tower as a remote with a core tower.
func AttachToCore(ctx context.Context, core tower_model.Instance) error {
	client, err := getAPIClient(ctx, core, clientOpts{noTowerIDHeader: true})
	if err != nil {
		return err
	}

	delete(client.GetConfig().DefaultHeader, TowerIDHeader)

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	role := string(local.Role)
	newParams := api.NewServerParams{
		ServerID: &local.TowerID,
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
