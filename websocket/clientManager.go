package websocket

import (
	"flag"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type clientManager struct {
	webClientMap    map[weblens.Username]*WsClient
	remoteClientMap map[weblens.InstanceId]*WsClient
	clientMu        *sync.RWMutex

	core *WsClient

	// Key: subscription identifier, value: clientConn instance
	// Use map to take advantage of O(1) lookup time when finding or removing clients
	// by subscription identifier
	// {
	// 	"fileId": [
	// 		*client1,
	// 		*client2,
	//     ]
	// }
	folderSubs   map[types.SubId][]*WsClient
	taskSubs     map[types.SubId][]*WsClient
	taskTypeSubs map[types.SubId][]*WsClient
	folderMu     sync.Mutex
	taskMu       sync.Mutex
	taskTypeMu   sync.Mutex
}

func NewClientManager() types.ClientManager {
	return &clientManager{
		webClientMap:    map[weblens.Username]*WsClient{},
		remoteClientMap: map[weblens.InstanceId]*WsClient{},
		clientMu:        &sync.RWMutex{},

		folderSubs:   map[types.SubId][]*WsClient{},
		taskSubs:     map[types.SubId][]*WsClient{},
		taskTypeSubs: map[types.SubId][]*WsClient{},
	}
}

func (cm *clientManager) ClientConnect(conn *websocket.Conn, user types.User) *WsClient {
	connectionId := ClientId(uuid.New().String())
	newClient := clientConn{Active: true, connId: connectionId, conn: conn, user: user}

	cm.clientMu.Lock()
	cm.webClientMap[user.GetUsername()] = &newClient
	cm.clientMu.Unlock()

	newClient.debug("Connected")
	return &newClient
}

func (cm *clientManager) RemoteConnect(conn *websocket.Conn, remote *weblens.WeblensInstance) *WsClient {
	connectionId := ClientId(uuid.New().String())
	newClient := &clientConn{Active: true, conn: conn, remote: remote, connId: connectionId}

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.ServerId()] = newClient
	cm.clientMu.Unlock()

	if remote.IsCore() {
		cm.core = newClient
	}

	return newClient
}

func (cm *clientManager) ClientDisconnect(c *WsClient) {
	for _, s := range c.GetSubscriptions() {
		cm.RemoveSubscription(s, c, true)
	}

	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	if c.GetUser() != nil {
		delete(cm.webClientMap, c.GetUser().GetUsername())
	} else if c.GetRemote() != nil {
		delete(cm.remoteClientMap, c.GetRemote().ServerId())
	}
}

func (cm *clientManager) GetClientByUsername(username weblens.Username) *WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.webClientMap[username]
}

func (cm *clientManager) GetClientByInstanceId(instanceId weblens.InstanceId) *WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.remoteClientMap[instanceId]
}

func (cm *clientManager) GetAllClients() []*WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return internal.MapToValues(cm.webClientMap)
}

func (cm *clientManager) GetConnectedAdmins() []*WsClient {
	clients := cm.GetAllClients()
	admins := internal.Filter(
		clients, func(c *WsClient) bool {
			return c.GetUser().IsAdmin()
		},
	)
	return admins
}

func (cm *clientManager) GetSubscribers(st types.WsAction, key types.SubId) (clients []*WsClient) {
	switch st {
	case types.FolderSubscribe:
		{
			cm.folderMu.Lock()
			clients = cm.folderSubs[key]
			cm.folderMu.Unlock()
		}
	case types.TaskSubscribe:
		{
			cm.taskMu.Lock()
			clients = cm.taskSubs[key]
			cm.taskMu.Unlock()
		}
	case types.UserSubscribe:
		{
			cm.clientMu.Lock()
			allClients := internal.MapToValues(cm.webClientMap)
			cm.clientMu.Unlock()
			clients = internal.Filter(
				allClients, func(c *WsClient) bool {
					return types.SubId(c.GetUser().GetUsername()) == key
				},
			)
		}
	case types.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			clients = cm.taskTypeSubs[key]
			cm.taskTypeMu.Unlock()
		}
	default:
		wlog.Error.Printf("Unknown subscriber type: [%s]", st)
	}

	// Copy slice to not modify reference to mapped slice
	return clients[:]
}

func (cm *clientManager) AddSubscription(subInfo types.Subscription, client *WsClient) {
	switch subInfo.Type {
	case types.FolderSubscribe:
		{
			cm.folderMu.Lock()
			addSub(cm.folderSubs, subInfo, client)
			cm.folderMu.Unlock()
		}
	case types.TaskSubscribe:
		{
			cm.taskMu.Lock()
			addSub(cm.taskSubs, subInfo, client)
			cm.taskMu.Unlock()
		}
	case types.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			addSub(cm.taskTypeSubs, subInfo, client)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			wlog.Error.Println("Unknown subType", subInfo.Type)
			return
		}
	}
}

func (cm *clientManager) RemoveSubscription(subInfo types.Subscription, client *WsClient, removeAll bool) {
	switch subInfo.Type {
	case types.FolderSubscribe:
		{
			cm.folderMu.Lock()
			removeSubs(cm.folderSubs, subInfo, client, removeAll)
			cm.folderMu.Unlock()
		}
	case types.TaskSubscribe:
		{
			cm.taskMu.Lock()
			removeSubs(cm.taskSubs, subInfo, client, removeAll)
			cm.taskMu.Unlock()
		}
	case types.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			removeSubs(cm.taskTypeSubs, subInfo, client, removeAll)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			wlog.Error.Println("Unknown subType", string(subInfo.Type))
			return
		}
	}
}

func addSub(subMap map[types.SubId][]*WsClient, subInfo types.Subscription, client *WsClient) {
	subs, ok := subMap[subInfo.Key]

	if !ok {
		subs = []*WsClient{}
	}

	subMap[subInfo.Key] = append(subs, client)
}

func removeSubs(
	subMap map[types.SubId][]*WsClient, subInfo types.Subscription, client *WsClient, removeAll bool,
) {
	subs, ok := subMap[subInfo.Key]
	if !ok {
		wlog.Warning.Println("Tried to unsubscribe from non-existent key", string(subInfo.Key))
		return
	}
	if removeAll {
		subs = internal.Filter(subs, func(c *WsClient) bool { return c.GetClientId() != client.GetClientId() })
	} else {
		index := slices.IndexFunc(subs, func(c *WsClient) bool { return c.GetClientId() == client.GetClientId() })
		if index != -1 {
			subs = internal.Banish(subs, index)
		}
	}
	subMap[subInfo.Key] = subs
}

type WsAuthorize struct {
	Auth string `json:"auth"`
}

const retryInterval = time.Second * 10

func WebsocketToCore(core *weblens.WeblensInstance, clientService *clientManager) error {
	addrStr, err := core.GetAddress()
	if err != nil {
		return err
	}

	if addrStr == "" {
		return errors.New("Core server address is empty")
	}

	re, err := regexp.Compile(`http(s)?://([^/]*)`)
	if err != nil {
		return errors.WithStack(err)
	}

	parts := re.FindStringSubmatch(addrStr)

	addr := flag.String("addr", parts[2], "http service address")
	host := url.URL{Scheme: "ws" + parts[1], Host: *addr, Path: "/api/core/ws"}
	dialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 10 * time.Second}

	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+string(core.GetUsingKey()))
	var conn *WsClient
	go func() {
		for {
			conn, err = dial(dialer, host, authHeader, core, clientService)
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

func dial(
	dialer *websocket.Dialer, host url.URL, authHeader http.Header, core *weblens.WeblensInstance,
	clientService *clientManager,
) (
	*WsClient, error,
) {
	wlog.Debug.Println("Dialing", host.String())
	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, werror.Wrap(err)
	}

	client := clientService.RemoteConnect(conn, core)

	err = client.Raw(WsAuthorize{Auth: authHeader.Get("Authorization")})
	if err != nil {
		return nil, werror.Wrap(err)
	}

	wlog.Info.Printf("Connection to core server at %s successfully established", host.String())
	return realC, nil
}

func coreWsHandler(c *websocket2.clientConn) {
	defer func() { c.Disconnect() }()
	defer func() { recover() }()

	for {
		mt, message, err := c.ReadOne()
		if err != nil {
			wlog.ShowErr(werror.Wrap(err))
			break
		}
		wlog.Debug.Println(mt, string(message))
	}
}
