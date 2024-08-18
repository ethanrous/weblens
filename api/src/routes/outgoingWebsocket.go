package routes

import (
	"flag"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util/wlog"
	"github.com/gorilla/websocket"
)

const retryInterval = time.Second * 10

func WebsocketToCore(core types.Instance) error {
	addrStr, err := core.GetAddress()
	if err != nil {
		return err
	}

	if addrStr == "" {
		return types.WeblensErrorMsg("Core server address is empty")
	}

	re, err := regexp.Compile(`http(s)?://([^/]*)`)
	if err != nil {
		return types.WeblensErrorFromError(err)
	}

	parts := re.FindStringSubmatch(addrStr)

	addr := flag.String("addr", parts[2], "http service address")
	host := url.URL{Scheme: "ws" + parts[1], Host: *addr, Path: "/api/core/ws"}
	dialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 10 * time.Second}

	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+string(core.GetUsingKey()))
	var conn *clientConn
	go func() {
		for {
			conn, err = dial(dialer, host, authHeader, core)
			if err != nil {
				wlog.Warning.Printf(
					"Failed to connect to core server at %s, trying again in %s",
					host.String(), retryInterval,
				)
				wlog.Debug.Println("Error was", err)
				time.Sleep(retryInterval)
				continue
			}
			coreWsHandler(conn)
			wlog.Warning.Printf("Connection to core websocket closed, reconnecting...")
		}
	}()
	return nil
}

func dial(dialer *websocket.Dialer, host url.URL, authHeader http.Header, core types.Instance) (*clientConn, error) {
	wlog.Debug.Println("Dialing", host.String())
	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, types.WeblensErrorFromError(err)
	}

	client := types.SERV.ClientManager.RemoteConnect(conn, core)

	realC := client.(*clientConn)
	err = realC.conn.WriteJSON(wsAuthorize{Auth: authHeader.Get("Authorization")})
	if err != nil {
		return nil, types.WeblensErrorFromError(err)
	}

	wlog.Info.Printf("Connection to core server at %s successfully established", host.String())
	return realC, nil
}

func coreWsHandler(c *clientConn) {
	defer func() { c.Disconnect() }()
	defer func() { recover() }()

	for {
		mt, message, err := c.ReadOne()
		if err != nil {
			wlog.ShowErr(types.WeblensErrorFromError(err))
			break
		}
		wlog.Debug.Println(mt, string(message))
	}
}
