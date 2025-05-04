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
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/jobs"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const retryInterval = time.Second * 10
const timeout = time.Second * 10

func init() {
	startup.RegisterStartup(connectToCores)
}

func connectToCores(c context.Context, cnf config.ConfigProvider) error {
	ctx, ok := context_service.FromContext(c)
	if !ok {
		return errors.New("Failed to get context from context")
	}

	context.AfterFunc(ctx, func() {
		ctx.ClientService.DisconnectAll(ctx)
	})

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
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
			ctx.Log().Error().Stack().Err(err).Msgf("Failed to connect to core %s [%s]", remote.Name, remote.TowerId)
		}
	}

	return nil
}

func ConnectCore(c context.Context, core *tower_model.Instance) error {
	ctx, ok := context_service.FromContext(c)
	if !ok {
		return errors.New("Failed to get context from context")
	}

	coreUrl, err := url.Parse(core.Address)
	if err != nil {
		return errors.WithStack(err)
	}

	if coreUrl.Host == "" {
		return errors.Errorf("Failed to parse core address: [%s]", core.Address)
	}

	if coreUrl.Scheme == "https" {
		coreUrl.Scheme = "wss"
	} else {
		coreUrl.Scheme = "ws"
	}

	coreUrl.Path = "/api/v1/ws"

	dialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: timeout}

	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+string(core.OutgoingKey))
	authHeader.Add(tower_service.TowerIdHeader, ctx.LocalTowerId)

	log.Debug().Msgf("Websocket connecting to core \"%s\" [%s] using %s", core.Name, core.TowerId, core.OutgoingKey)

	var client *client_model.WsClient

	dialWithRetry := func() {
		for {
			client, err = dial(ctx, dialer, *coreUrl, authHeader, core)
			if err != nil {
				ctx.Log().Error().Msgf(
					"Failed to connect to core server at %s: %s Trying again in %s",
					coreUrl.String(), err, retryInterval,
				)

				select {
				case <-c.Done():
					return
				case <-time.After(retryInterval):
				}

				continue
			}

			ctx.Log().Debug().Func(func(e *zerolog.Event) {
				e.Msgf("Connection to core [%s] at [%s] successfully established", core.Name, coreUrl.String())
			})

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
	log.Trace().Msgf("Dialing %s", host.String())

	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client := ctx.ClientService.RemoteConnect(ctx, conn, core)

	return client, nil
}

func coreWsHandler(ctx context_service.AppContext, c *client_model.WsClient) error {
	defer func() { ctx.ClientService.ClientDisconnect(ctx, c) }()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, msgBuf, err := c.ReadOne()
		if err != nil {
			return errors.WithStack(err)
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
		e.Msgf("Got wsmsg from %s [%s]: %v", c.GetInstance().Name, c.GetInstance().TowerId, msg)
	})

	switch msg.EventTag {
	case "do_backup":
		coreIdI, ok := msg.Content["coreId"]
		if !ok {
			c.Error(errors.Errorf("Missing coreId in do_backup message"))

			return
		}

		coreId, ok := coreIdI.(string)
		if !ok {
			c.Error(errors.Errorf("Invalid coreId in do_backup message: %v", coreIdI))

			return
		}

		coreTower, err := tower_model.GetTowerById(ctx, coreId)
		if err != nil {
			c.Error(errors.Wrapf(err, "Invalid coreId in do_backup message: %s", coreId))

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
			c.Error(errors.Errorf("Missing role in weblens_loaded message"))

			return
		}

		roleStr, ok := roleI.(string)
		if !ok {
			c.Error(errors.Errorf("Invalid role in weblens_loaded message: %v", roleI))

			return
		}

		var role tower_model.Role

		switch tower_model.Role(roleStr) {
		case tower_model.RoleCore, tower_model.RoleBackup, tower_model.RoleInit:
			role = tower_model.Role(roleStr)
		default:
			c.Error(errors.Errorf("Invalid role in weblens_loaded message: %v", roleI))
		}

		ctx.Log().Debug().Msgf("Setting server [%s] reported role to [%s]", c.GetInstance().TowerId, role)
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
