package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service/proxy"
	"github.com/gorilla/websocket"
)

const RetryInterval = time.Second * 10

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
		return werror.Errorf("Failed to parse core address: %s", addrStr)
	}

	if coreUrl.Scheme == "https" {
		coreUrl.Scheme = "wss"
	} else {
		coreUrl.Scheme = "ws"
	}

	coreUrl.Path = "/api/ws"

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
					coreUrl.String(), err, RetryInterval,
				)
				time.Sleep(RetryInterval)
				continue
			}
			coreWsHandler(conn, pack)
			log.Warning.Printf("Connection to core websocket closed, reconnecting...")
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
	log.Debug.Println("Dialing", host.String())
	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	client := clientService.RemoteConnect(conn, core)

	// err = client.Raw(WsAuthorize{Auth: authHeader.Get("Authorization")})
	// if err != nil {
	// 	return nil, werror.WithStack(err)
	// }

	log.Debug.Printf("Connection to core server at %s successfully established", host.String())
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

	switch msg.EventTag {
	case "do_backup":
		core := pack.InstanceService.GetCore()
		meta := models.BackupMeta{
			RemoteId:            core.ServerId(),
			FileService:         pack.FileService,
			UserService:         pack.UserService,
			WebsocketService:    pack.ClientService,
			InstanceService:     pack.InstanceService,
			TaskService:         pack.TaskService,
			Caster:              pack.Caster,
			ProxyFileService:    &proxy.ProxyFileService{Core: core},
			ProxyJournalService: &proxy.ProxyJournalService{Core: core},
			ProxyUserService:    proxy.NewProxyUserService(core),
			ProxyMediaService:   &proxy.ProxyMediaService{Core: core},
		}
		_, err = pack.TaskService.DispatchJob(models.BackupTask, meta, nil)
		if err != nil {
			c.Error(err)
		}
	default:
		log.Error.Printf("Unknown ws message from core: %s", msg.EventTag)
	}
}