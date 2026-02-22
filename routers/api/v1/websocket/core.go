// Package websocket provides WebSocket connection handlers and utilities.
package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	client_model "github.com/ethanrous/weblens/models/client"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/jobs"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// retryInterval defines the duration to wait between websocket connection retry attempts.
const retryInterval = time.Second
const maxRetryInterval = time.Minute * 5

// timeout specifies the maximum duration for the websocket handshake to complete.
const timeout = time.Second * 10

func init() {
	startup.RegisterHook(connectToCores)
}

func connectToCores(c context.Context, _ config.Provider) error {
	ctx, ok := context_service.FromContext(c)
	if !ok {
		return wlerrors.New("Failed to get context from context")
	}

	context.AfterFunc(ctx, func() {
		err := ctx.ClientService.DisconnectAll(ctx)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("Failed to disconnect all clients")
		}
	})

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return wlerrors.Errorf("Failed to get local tower instance while connecting to core: %w", err)
	}

	if local.Role != tower_model.RoleBackup {
		return nil
	}

	remotes, err := tower_model.GetRemotes(ctx)
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		if remote.Role != tower_model.RoleCore {
			continue
		}

		err = ConnectCore(ctx, &remote)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msgf("Failed to connect to core %s [%s]", remote.Name, remote.TowerID)
		}
	}

	return nil
}

// ConnectCore establishes and maintains a websocket connection to a core tower instance.
func ConnectCore(c context.Context, core *tower_model.Instance) error {
	ctx, ok := context_service.FromContext(c)
	if !ok {
		return wlerrors.New("Failed to get context from context")
	}

	coreURL, err := url.Parse(core.Address)
	if err != nil {
		return wlerrors.WithStack(err)
	}

	if coreURL.Host == "" {
		return wlerrors.Errorf("Failed to parse core address: [%s]", core.Address)
	}

	if coreURL.Scheme == "https" {
		coreURL.Scheme = "wss"
	} else {
		coreURL.Scheme = "ws"
	}

	coreURL.Path = "/api/v1/ws"

	dialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: timeout}

	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+string(core.OutgoingKey))
	authHeader.Add(tower_service.TowerIDHeader, ctx.LocalTowerID)

	log.Debug().Msgf("Websocket connecting to core \"%s\" [%s]", core.Name, core.TowerID)

	var client *client_model.WsClient

	appCtx, ok := context_service.FromContext(c)
	if !ok {
		return wlerrors.New("not an app context")
	}

	dialWithRetry := func() {
		activeRetry := retryInterval

		for {
			client, err = dial(ctx, dialer, *coreURL, authHeader, core)
			if err != nil {
				ctx.Log().Error().Msgf(
					"Failed to connect to core server at %s: %s. Trying again in %s",
					coreURL.String(), err, activeRetry,
				)

				select {
				case <-c.Done():
					return
				case <-time.After(activeRetry):
				case <-appCtx.ClientService.ListenForTowerUpdate(*core):
					ctx.Log().Info().Msgf("Core [%s] updated, attempting to reconnect...", core.Name)

					continue
				}

				activeRetry *= 2
				if activeRetry > maxRetryInterval {
					activeRetry = maxRetryInterval
				}

				continue
			}

			activeRetry = retryInterval

			ctx.Log().Debug().Msgf("Connection to core [%s] at [%s] successfully established", core.Name, coreURL.String())

			err = coreWsHandler(ctx, client)

			select {
			case <-c.Done():
				return
			default:
			}

			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msg("Error in core websocket handler")
			}

			ctx.Log().Warn().Msgf("Websocket connection to core [%s] closed, reconnecting...", core.Name)
		}
	}

	go dialWithRetry()

	return nil
}

func dial(ctx context_service.AppContext, dialer *websocket.Dialer, host url.URL, authHeader http.Header, core *tower_model.Instance) (*client_model.WsClient, error) {
	ctx.Log().Trace().Msgf("Dialing %s", host.String())

	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	client := ctx.ClientService.RemoteConnect(ctx, conn, core)

	return client, nil
}

func coreWsHandler(ctx context_service.AppContext, c *client_model.WsClient) error {
	defer func() {
		err := ctx.ClientService.ClientDisconnect(ctx, c)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("Failed to disconnect client")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, msgBuf, err := c.ReadOne()
		if err != nil {
			return wlerrors.WithStack(err)
		}

		wsCoreClientSwitchboard(ctx, msgBuf, c)
	}
}

func wsCoreClientSwitchboard(ctx context_service.AppContext, msgBuf []byte, c *client_model.WsClient) {
	defer wsRecover(ctx, c)

	var msg websocket_mod.WsResponseInfo

	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)

		return
	}

	c.Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Got wsmsg from %s [%s]: %v", c.GetInstance().Name, c.GetInstance().TowerID, msg)
	})

	switch msg.EventTag {
	case "do_backup":
		coreIDI, ok := msg.Content["coreID"]
		if !ok {
			c.Error(wlerrors.Errorf("Missing coreID in do_backup message"))

			return
		}

		coreID, ok := coreIDI.(string)
		if !ok {
			c.Error(wlerrors.Errorf("Invalid coreID in do_backup message: %v", coreIDI))

			return
		}

		coreTower, err := tower_model.GetTowerByID(ctx, coreID)
		if err != nil {
			c.Error(wlerrors.Wrapf(err, "Invalid coreID in do_backup message: %s", coreID))

			return
		}

		log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Backup requested for %s", coreTower.Name) })

		_, err = jobs.BackupOne(ctx, *c.GetInstance())
		if err != nil {
			c.Error(err)
		}

	case websocket_mod.WeblensLoadedEvent:
		roleI, ok := msg.Content["role"]
		if !ok {
			c.Error(wlerrors.Errorf("Missing role in weblens_loaded message"))

			return
		}

		roleStr, ok := roleI.(string)
		if !ok {
			c.Error(wlerrors.Errorf("Invalid role in weblens_loaded message: %v", roleI))

			return
		}

		var role tower_model.Role

		switch tower_model.Role(roleStr) {
		case tower_model.RoleCore, tower_model.RoleBackup, tower_model.RoleUninitialized:
			role = tower_model.Role(roleStr)
		default:
			c.Error(wlerrors.Errorf("Invalid role in weblens_loaded message: %v", roleI))
		}

		ctx.Log().Debug().Msgf("Setting server [%s] reported role to [%s]", c.GetInstance().TowerID, role)
		c.GetInstance().SetReportedRole(role)

		// Launch backup task whenever we reconnect to the core server
		_, err = jobs.BackupOne(ctx, *c.GetInstance())
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("")
		}
	case websocket_mod.StartupProgressEvent, websocket_mod.RemoteConnectionChangedEvent: // Do nothing
	case "error":
		c.Log().Error().Interface("websocket_msg_content", msg)
	default:
		c.Log().Error().Msgf("Unknown ws event %s", msg.EventTag)
	}
}
