package comm

import (
	"errors"
	"flag"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/gorilla/websocket"
)

const RetryInterval = time.Second * 10

func WebsocketToCore(core *models.Instance, clientService models.ClientManager) error {
	addrStr, err := core.GetAddress()
	if err != nil {
		return err
	}

	if addrStr == "" {
		return errors.New("Core server address is empty")
	}

	re, err := regexp.Compile(`http(s)?://([^/]*)`)
	if err != nil {
		return werror.WithStack(err)
	}

	parts := re.FindStringSubmatch(addrStr)

	addr := flag.String("addr", parts[2], "http service address")
	host := url.URL{Scheme: "ws" + parts[1], Host: *addr, Path: "/api/core/ws"}
	dialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 10 * time.Second}

	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+string(core.GetUsingKey()))
	var conn *models.WsClient
	go func() {
		for {
			conn, err = dial(dialer, host, authHeader, core, clientService)
			if err != nil {
				log.Warning.Printf(
					"Failed to connect to core server at %s, trying again in %s",
					host.String(), RetryInterval,
				)
				log.Debug.Println("Error was", err)
				time.Sleep(RetryInterval)
				continue
			}
			coreWsHandler(conn)
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

	err = client.Raw(WsAuthorize{Auth: authHeader.Get("Authorization")})
	if err != nil {
		return nil, werror.WithStack(err)
	}

	log.Debug.Printf("Connection to core server at %s successfully established", host.String())
	return client, nil
}

func coreWsHandler(c *models.WsClient) {
	defer func() { c.Disconnect() }()
	defer func() { recover() }()

	for {
		mt, message, err := c.ReadOne()
		if err != nil {
			log.ShowErr(werror.WithStack(err))
			break
		}
		log.Debug.Println(mt, string(message))
	}
}
