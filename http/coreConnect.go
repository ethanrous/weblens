package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const retryInterval = time.Second * 10

func WebsocketToCore(core *models.Instance, pack *models.ServicePack) error {
	addrStr, err := core.GetAddress()
	if err != nil {
		return err
	}

	coreUrl, err := url.Parse(addrStr)
	if err != nil {
		return werror.WithStack(err)
	}

	if coreUrl.Host == "" {
		return werror.Errorf("Failed to parse core address: [%s]", addrStr)
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
	authHeader.Add("Authorization", "Bearer "+string(core.GetUsingKey()))
	authHeader.Add("Wl-Server-Id", pack.InstanceService.GetLocal().ServerId())

	log.Debug().Msgf("Connecting to core server using %s", core.GetUsingKey())
	var conn *models.WsClient
	go func() {
		for {
			conn, err = dial(dialer, *coreUrl, authHeader, core, pack.ClientService)
			if err != nil {
				pack.Log.Error().Msgf(
					"Failed to connect to core server at %s: %s Trying again in %s",
					coreUrl.String(), err, retryInterval,
				)
				time.Sleep(retryInterval)
				continue
			}
			pack.Log.Debug().Func(func(e *zerolog.Event) {
				e.Msgf("Connection to core [%s] at [%s] successfully established", core.GetName(), coreUrl.String())
			})

			pack.Caster.PushWeblensEvent(
				models.RemoteConnectionChangedEvent, models.WsC{
					"serverId": core.ServerId(),
					"online":   true,
				},
			)

			coreWsHandler(conn, pack)
			pack.Caster.PushWeblensEvent(
				models.RemoteConnectionChangedEvent, models.WsC{
					"serverId": core.ServerId(),
					"online":   false,
				},
			)

			if pack.Closing.Load() {
				return
			}
			pack.Log.Warn().Msgf("Websocket connection to core [%s] closed, reconnecting...", core.GetName())
		}
	}()

	return nil
}

func dial(
	dialer *websocket.Dialer, host url.URL, authHeader http.Header, core *models.Instance,
	clientService models.ClientManager,
) (
	*models.WsClient, error,
) {
	log.Trace().Msgf("Dialing %s", host.String())
	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	client := clientService.RemoteConnect(conn, core)

	return client, nil
}

func coreWsHandler(c *models.WsClient, pack *models.ServicePack) {
	defer func() { c.Disconnect() }()

	for {
		_, msgBuf, err := c.ReadOne()
		if err != nil {
			pack.Log.Error().Stack().Err(werror.WithStack(err)).Msg("")
			break
		}

		wsCoreClientSwitchboard(msgBuf, c, pack)
	}
}

func wsCoreClientSwitchboard(msgBuf []byte, c *models.WsClient, pack *models.ServicePack) {
	defer wsRecover(c)

	var msg models.WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

	c.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Got wsmsg from R[%s]: %v", c.GetInstance().GetName(), msg) })

	switch msg.EventTag {
	case "do_backup":
		coreIdI, ok := msg.Content["coreId"]
		if !ok {
			c.Error(werror.Errorf("Missing coreId in do_backup message"))
			return
		}
		coreId, ok := coreIdI.(string)
		if !ok {
			c.Error(werror.Errorf("Invalid coreId in do_backup message: %v", coreIdI))
			return
		}
		core := pack.InstanceService.GetByInstanceId(coreId)
		if core == nil {
			c.Error(werror.Errorf("Core server not found: %s", msg.Content["coreId"]))
			return
		}

		log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Backup requested by %s", core.GetName()) })
		_, err = jobs.BackupOne(core, pack)
		if err != nil {
			c.Error(err)
		}

	case models.WeblensLoadedEvent:
		roleI, ok := msg.Content["role"]
		if !ok {
			c.Error(werror.Errorf("Missing role in weblens_loaded message"))
			return
		}

		roleStr, ok := roleI.(string)
		if !ok {
			c.Error(werror.Errorf("Invalid role in weblens_loaded message: %v", roleI))
			return
		}

		var role models.ServerRole
		switch models.ServerRole(roleStr) {
		case models.CoreServerRole, models.BackupServerRole, models.InitServerRole:
			role = models.ServerRole(roleStr)
		default:
			c.Error(werror.Errorf("Invalid role in weblens_loaded message: %v", roleI))
		}

		pack.Log.Debug().Msgf("Setting server [%s] reported role to [%s]", c.GetInstance().ServerId(), role)
		c.GetInstance().SetReportedRole(role)

		// Launch backup task whenever we reconnect to the core server
		_, err = jobs.BackupOne(c.GetInstance(), pack)
		if err != nil {
			pack.Log.Error().Stack().Err(err).Msg("")
		}
	case models.StartupProgressEvent, models.RemoteConnectionChangedEvent: // Do nothing
	case "error":
		c.Log().Error().Interface("websocket_msg_content", msg)
	default:
		c.Log().Error().Msgf("Unknown ws event %s", msg.EventTag)
	}
}
