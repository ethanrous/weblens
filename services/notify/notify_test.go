package notify_test

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/usermodel"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
	"github.com/stretchr/testify/assert"
)

type mockWsHandler struct {
	upgrader         websocket.Upgrader
	Done             chan struct{}
	m                sync.Mutex
	messagesRecieved []websocket_mod.WsResponseInfo
}

func (s *mockWsHandler) MessagesReceived() []websocket_mod.WsResponseInfo {
	s.m.Lock()
	defer s.m.Unlock()

	return s.messagesRecieved
}

func (s *mockWsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	s.Done = make(chan struct{})
	defer close(s.Done)

	var conn *websocket.Conn

	conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	for {
		var msg websocket_mod.WsResponseInfo

		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		s.m.Lock()
		s.messagesRecieved = append(s.messagesRecieved, msg)
		s.m.Unlock()
	}

	wlog.GlobalLogger().Info().Msg("Mock WS handler connection closed")
}

func mockClientConnect(ctx context.Context, manager *notify.ClientManager) (*client.WsClient, *mockWsHandler, error) {
	h := &mockWsHandler{}
	d := wstest.NewDialer(h)

	conn, _, err := d.Dial("ws://whatever/ws", nil)
	if err != nil {
		return nil, nil, err
	}

	usr := &usermodel.User{Username: "testuser"}

	client, err := manager.ClientConnect(ctx, conn, usr)
	if err != nil {
		return nil, nil, err
	}

	if client == nil {
		return nil, nil, err
	}

	return client, h, nil
}

func setupManagerAndClient(ctx ctxservice.AppContext) (*notify.ClientManager, *client.WsClient, *mockWsHandler, error) {
	manager := notify.NewClientManager(ctx)

	client, handler, err := mockClientConnect(ctx, manager)
	if err != nil {
		return nil, nil, nil, err
	}

	return manager, client, handler, nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	var sb strings.Builder
	sb.Grow(length)

	for range length {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}

func TestClientConnect(t *testing.T) {
	t.Parallel()
	appCtx := ctxservice.NewTestContext(t.Context())

	t.Run("connects client successfully", func(t *testing.T) {
		_, c, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.True(t, c.IsOpen())
	})
}

func TestClientSubscribe(t *testing.T) {
	t.Parallel()
	appCtx := ctxservice.NewTestContext(t.Context())

	t.Run("subscribes successfully to file", func(t *testing.T) {
		m, c, h, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileID := randomString(8)

		err = m.SubscribeToFile(t.Context(), c, &mockIDer{id: fileID}, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		clients := m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID)
		assert.Len(t, clients, 1)
		assert.Equal(t, c, clients[0])

		// Attempting to retrieve subscribers for the wrong subscription type should return no clients
		noClients := m.GetSubscribers(appCtx, websocket_mod.TaskSubscribe, fileID)
		assert.Len(t, noClients, 0)

		notif := notify.NewFileNotification(appCtx, wlstructs.FileInfo{ID: fileID}, websocket_mod.FileUpdatedEvent)
		m.Notify(appCtx, notif...)
		m.Flush(appCtx)

		assert.Eventually(t, func() bool {
			return len(h.MessagesReceived()) > 0
		}, 2*time.Second, 100*time.Millisecond, "expected to receive file notification message")

		assert.Equal(t, fileID, h.MessagesReceived()[0].SubscribeKey)
	})

	t.Run("only subscribers receive notifications", func(t *testing.T) {
		m, _, h, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileID := randomString(8)

		notif := notify.NewFileNotification(appCtx, wlstructs.FileInfo{ID: fileID}, websocket_mod.FileUpdatedEvent)
		m.Notify(appCtx, notif...)
		m.Flush(appCtx)

		assert.Never(t, func() bool {
			return len(h.MessagesReceived()) > 0
		}, 1*time.Second, 100*time.Millisecond, "did not expect to receive file notification message")
	})

	t.Run("FolderSubToTask correctly adds subscriptions to a task", func(t *testing.T) {
		m, client1, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		client2, _, err := mockClientConnect(appCtx, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileID := randomString(8)

		// Subscribe client1 to the file
		err = m.SubscribeToFile(t.Context(), client1, &mockIDer{id: fileID}, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		// Create a task subscription from the folder subscription via FolderSubToTask
		taskID := randomString(8)
		task := &mockIDer{id: taskID}
		m.FolderSubToTask(appCtx, fileID, task)

		// Verify that client1 is subscribed to the task
		subscribedClients := m.GetSubscribers(appCtx, websocket_mod.TaskSubscribe, taskID)
		assert.Len(t, subscribedClients, 1)
		assert.Equal(t, client1, subscribedClients[0])

		// Verify that client2 is not subscribed to the task
		assert.NotContains(t, subscribedClients, client2)
	})
}

func TestClientUnsubscribe(t *testing.T) {
	t.Parallel()
	appCtx := ctxservice.NewTestContext(t.Context())

	t.Run("unsubscribes successfully from file", func(t *testing.T) {
		m, c, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileID1 := randomString(8)

		err = m.SubscribeToFile(t.Context(), c, &mockIDer{id: fileID1}, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		fileID2 := randomString(8)

		err = m.SubscribeToFile(t.Context(), c, &mockIDer{id: fileID2}, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		assert.Equal(t, len(c.GetSubscriptions()), 2)

		err = m.Unsubscribe(appCtx, c, fileID1, time.Now())
		if err != nil {
			t.Fatalf("unexpected error unsubscribing from file: %v", err)
		}

		// Verify unsubscribed from fileID1
		clients := m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID1)
		assert.Len(t, clients, 0)

		// Verify still subscribed to fileID2
		clients = m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID2)
		assert.Len(t, clients, 1)
	})

	t.Run("correctly unsubscribes all subscribers by ID", func(t *testing.T) {
		m, client1, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Attach another client to verify all subscribers are removed
		client2, _, err := mockClientConnect(appCtx, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileID1 := randomString(8)
		f1 := &mockIDer{id: fileID1}

		err = m.SubscribeToFile(t.Context(), client1, f1, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		err = m.SubscribeToFile(t.Context(), client2, f1, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		fileID2 := randomString(8)
		f2 := &mockIDer{id: fileID2}

		err = m.SubscribeToFile(t.Context(), client1, f2, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		err = m.SubscribeToFile(t.Context(), client2, f2, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		// Verify subscribed to fileID1
		clients := m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID1)
		assert.Len(t, clients, 2)

		// Verify subscribed to fileID2
		clients = m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID2)
		assert.Len(t, clients, 2)

		// Ubsubscribe all subscribers from fileID1
		err = m.UnsubscribeAllByID(appCtx, fileID1, websocket_mod.FolderSubscribe)
		if err != nil {
			t.Fatalf("unexpected error unsubscribing all from file: %v", err)
		}

		// Verify unsubscribed from fileID1
		clients = m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID1)
		assert.Len(t, clients, 0)

		// Verify still subscribed to fileID2
		clients = m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID2)
		assert.Len(t, clients, 2)
	})
}

type mockIDer struct {
	id string
}

func (i *mockIDer) ID() string {
	return i.id
}

func TestClientDisconnect(t *testing.T) {
	t.Parallel()
	appCtx := ctxservice.NewTestContext(t.Context())

	t.Run("client disconnects successfully", func(t *testing.T) {
		m, c, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = m.ClientDisconnect(appCtx, c)
		if err != nil {
			t.Fatalf("unexpected error disconnecting client: %v", err)
		}

		assert.False(t, c.IsOpen())
		assert.Empty(t, m.GetAllClients())
	})

	t.Run("client subscriptions are removed successfully", func(t *testing.T) {
		m, c, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileID := "file-123"

		err = m.SubscribeToFile(t.Context(), c, &mockIDer{id: fileID}, time.Now())
		if err != nil {
			t.Fatalf("unexpected error subscribing to file: %v", err)
		}

		err = m.ClientDisconnect(appCtx, c)
		if err != nil {
			t.Fatalf("unexpected error disconnecting client: %v", err)
		}

		emptyClients := m.GetSubscribers(appCtx, websocket_mod.FolderSubscribe, fileID)
		assert.Empty(t, emptyClients)
	})

	t.Run("all clients disconnect successfully", func(t *testing.T) {
		m, c1, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		c2, _, err := mockClientConnect(appCtx, m)
		if err != nil {
			t.Fatalf("unexpected error connecting second client: %v", err)
		}

		err = m.DisconnectAll(appCtx)
		if err != nil {
			t.Fatalf("unexpected error disconnecting all clients: %v", err)
		}

		assert.False(t, c1.IsOpen())
		assert.False(t, c2.IsOpen())
		assert.Empty(t, m.GetAllClients())
	})
}

func TestClientManager_GetClientByUsername(t *testing.T) {
	t.Parallel()
	appCtx := ctxservice.NewTestContext(t.Context())

	t.Run("retrieves client by username successfully", func(t *testing.T) {
		m, c, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrievedClient := m.GetClientByUsername("testuser")
		assert.NotNil(t, retrievedClient)
		assert.Equal(t, c, retrievedClient)
	})

	t.Run("returns nil for non-existent username", func(t *testing.T) {
		m, _, _, err := setupManagerAndClient(appCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrievedClient := m.GetClientByUsername("nonexistentuser")
		assert.Nil(t, retrievedClient)
	})
}
