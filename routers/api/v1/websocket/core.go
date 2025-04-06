package websocket

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	client_model "github.com/ethanrous/weblens/models/client"
	tower_model "github.com/ethanrous/weblens/models/tower"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const retryInterval = time.Second * 10

func WebsocketToCore(ctx *context.RequestContext, core *tower_model.Instance) error {
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

	coreUrl.Path = "/api/ws"
	q := coreUrl.Query()
	q.Add("server", "true")
	coreUrl.RawQuery = q.Encode()

	dialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 10 * time.Second}

	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+string(core.OutgoingKey))
	authHeader.Add("Wl-Server-Id", ctx.LocalTowerId)

	log.Debug().Msgf("Connecting to core server using %s", core.OutgoingKey)
	var client *client_model.WsClient
	go func() {
		for {
			client, err = dial(ctx, dialer, *coreUrl, authHeader, core)
			if err != nil {
				ctx.Log().Error().Msgf(
					"Failed to connect to core server at %s: %s Trying again in %s",
					coreUrl.String(), err, retryInterval,
				)
				time.Sleep(retryInterval)
				continue
			}
			ctx.Log().Debug().Func(func(e *zerolog.Event) {
				e.Msgf("Connection to core [%s] at [%s] successfully established", core.Name, coreUrl.String())
			})

			notif := client_model.NewSystemNotification(
				websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{
					"serverId": core.TowerId,
					"online":   true,
				},
			)
			ctx.Notify(notif)

			coreWsHandler(&ctx.AppContext, client)
			notif = client_model.NewSystemNotification(
				websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{
					"serverId": core.TowerId,
					"online":   false,
				},
			)
			ctx.Notify(notif)

			// if pack.Closing.Load() {
			// 	return
			// }
			ctx.Log().Warn().Msgf("Websocket connection to core [%s] closed, reconnecting...", core.Name)
		}
	}()

	return nil
}

func dial(ctx *context.RequestContext, dialer *websocket.Dialer, host url.URL, authHeader http.Header, core *tower_model.Instance) (*client_model.WsClient, error) {
	log.Trace().Msgf("Dialing %s", host.String())
	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client := ctx.ClientService.RemoteConnect(ctx, conn, core)
	return client, nil
}

func coreWsHandler(ctx *context.AppContext, c *client_model.WsClient) {
	defer func() { c.Disconnect() }()

	for {
		_, msgBuf, err := c.ReadOne()
		if err != nil {
			ctx.Log().Error().Stack().Err(errors.WithStack(err)).Msg("")
			break
		}

		wsCoreClientSwitchboard(ctx, msgBuf, c)
	}
}

func wsCoreClientSwitchboard(ctx *context.AppContext, msgBuf []byte, c *client_model.WsClient) {
	defer wsRecover(ctx, c)

	var msg websocket_mod.WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

	c.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Got wsmsg from R[%s]: %v", c.GetInstance().Name, msg) })

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
		_, err = jobs.BackupOne(ctx, c.GetInstance())
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

		var role tower_model.TowerRole
		switch tower_model.TowerRole(roleStr) {
		case tower_model.CoreServerRole, tower_model.BackupServerRole, tower_model.InitServerRole:
			role = tower_model.TowerRole(roleStr)
		default:
			c.Error(errors.Errorf("Invalid role in weblens_loaded message: %v", roleI))
		}

		ctx.Log().Debug().Msgf("Setting server [%s] reported role to [%s]", c.GetInstance().TowerId, role)
		c.GetInstance().SetReportedRole(role)

		// Launch backup task whenever we reconnect to the core server
		_, err = jobs.BackupOne(ctx, c.GetInstance())
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
