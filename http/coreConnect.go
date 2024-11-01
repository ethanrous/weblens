package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/gorilla/websocket"
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
	var conn *models.WsClient
	go func() {
		for {
			conn, err = dial(dialer, *coreUrl, authHeader, core, pack.ClientService)
			if err != nil {
				log.Error.Printf(
					"Failed to connect to core server at %s: %s Trying again in %s",
					coreUrl.String(), err, retryInterval,
				)
				time.Sleep(retryInterval)
				continue
			}

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
			log.Warning.Printf("Websocket connection to core [%s] closed, reconnecting...", core.GetName())
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
	log.Trace.Println("Dialing", host.String())
	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	client := clientService.RemoteConnect(conn, core)

	log.Debug.Printf("Connection to core [%s] at [%s] successfully established", core.GetName(), host.String())
	return client, nil
}

func coreWsHandler(c *models.WsClient, pack *models.ServicePack) {
	defer func() { c.Disconnect() }()

	for {
		_, msgBuf, err := c.ReadOne()
		if err != nil {
			log.ShowErr(werror.WithStack(err))
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

	log.Trace.Printf("Got wsmsg from R[%s]: %v", c.GetRemote().GetName(), msg)

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

		log.Trace.Printf("Backup requested by %s", core.GetName())
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

		c.GetRemote().SetReportedRole(roleI.(string))

		// Launch backup task whenever we reconnect to the core server
		_, err = jobs.BackupOne(c.GetRemote(), pack)
		if err != nil {
			log.ErrTrace(err)
		}
	case models.StartupProgressEvent, models.RemoteConnectionChangedEvent: // Do nothing
	case "error":
		log.Trace.Println(msg)
	default:
		log.Error.Printf("Unknown ws event from %s: %s", c.GetRemote().GetName(), msg.EventTag)
	}
}
