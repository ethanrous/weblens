package client_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/client"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWsHandler struct {
	upgrader         websocket.Upgrader
	Done             chan struct{}
	m                sync.Mutex
	messagesReceived []websocket_mod.WsResponseInfo
}

func (s *mockWsHandler) MessagesReceived() []websocket_mod.WsResponseInfo {
	s.m.Lock()
	defer s.m.Unlock()

	return s.messagesReceived
}

func (s *mockWsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Done = make(chan struct{})
	defer close(s.Done)

	conn, err := s.upgrader.Upgrade(w, r, nil)
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
		s.messagesReceived = append(s.messagesReceived, msg)
		s.m.Unlock()
	}
}

func createMockConnection() (*websocket.Conn, *mockWsHandler) {
	h := &mockWsHandler{}
	d := wstest.NewDialer(h)

	conn, _, err := d.Dial("ws://whatever/ws", nil)
	if err != nil {
		panic(err)
	}

	return conn, h
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("creates client for user", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)

		assert.NotNil(t, c)
		assert.True(t, c.IsOpen())
		assert.NotEmpty(t, c.GetClientID())
		assert.Equal(t, user, c.GetUser())
		assert.Nil(t, c.GetInstance())
		assert.Equal(t, websocket_mod.WebClient, c.ClientType())

		_ = c.Disconnect()
	})

	t.Run("creates client for tower instance", func(t *testing.T) {
		conn, _ := createMockConnection()
		tower := &tower_model.Instance{Name: "test-tower", TowerID: "tower-123"}

		c := client.NewClient(context.Background(), conn, tower)

		assert.NotNil(t, c)
		assert.True(t, c.IsOpen())
		assert.NotEmpty(t, c.GetClientID())
		assert.Equal(t, tower, c.GetInstance())
		assert.Nil(t, c.GetUser())
		assert.Equal(t, websocket_mod.TowerClient, c.ClientType())

		_ = c.Disconnect()
	})
}

func TestClientType(t *testing.T) {
	t.Parallel()

	t.Run("returns WebClient for user connection", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		assert.Equal(t, websocket_mod.WebClient, c.ClientType())
	})

	t.Run("returns TowerClient for tower connection", func(t *testing.T) {
		conn, _ := createMockConnection()
		tower := &tower_model.Instance{Name: "test-tower"}

		c := client.NewClient(context.Background(), conn, tower)
		defer c.Disconnect() //nolint:errcheck

		assert.Equal(t, websocket_mod.TowerClient, c.ClientType())
	})
}

func TestGetShortID(t *testing.T) {
	t.Parallel()

	t.Run("returns last 8 characters of UUID", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		fullID := c.GetClientID()
		shortID := c.GetShortID()

		// UUID format is xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
		// ShortID returns the last 8 characters (starting at index 28)
		assert.Len(t, shortID, 8)
		assert.Equal(t, fullID[28:], shortID)
	})
}

func TestSubscriptionManagement(t *testing.T) {
	t.Parallel()

	t.Run("add and retrieve subscriptions", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		sub1 := websocket_mod.Subscription{
			SubscriptionID: "sub-1",
			Type:           websocket_mod.FolderSubscribe,
			When:           time.Now(),
		}
		sub2 := websocket_mod.Subscription{
			SubscriptionID: "sub-2",
			Type:           websocket_mod.TaskSubscribe,
			When:           time.Now(),
		}

		c.AddSubscription(sub1)
		c.AddSubscription(sub2)

		subs := c.GetSubscriptions()
		assert.Len(t, subs, 2)

		ids := make([]string, len(subs))
		for i, s := range subs {
			ids[i] = s.SubscriptionID
		}

		assert.Contains(t, ids, "sub-1")
		assert.Contains(t, ids, "sub-2")
	})

	t.Run("remove subscription by key", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		sub1 := websocket_mod.Subscription{SubscriptionID: "sub-1", Type: websocket_mod.FolderSubscribe}
		sub2 := websocket_mod.Subscription{SubscriptionID: "sub-2", Type: websocket_mod.TaskSubscribe}
		sub3 := websocket_mod.Subscription{SubscriptionID: "sub-3", Type: websocket_mod.FolderSubscribe}

		c.AddSubscription(sub1)
		c.AddSubscription(sub2)
		c.AddSubscription(sub3)

		c.RemoveSubscription("sub-2")

		subs := (c.GetSubscriptions())
		assert.Len(t, subs, 2)

		ids := make([]string, len(subs))
		for i, s := range subs {
			ids[i] = s.SubscriptionID
		}

		assert.Contains(t, ids, "sub-1")
		assert.Contains(t, ids, "sub-3")
		assert.NotContains(t, ids, "sub-2")
	})

	t.Run("remove non-existent subscription is no-op", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		sub := websocket_mod.Subscription{SubscriptionID: "sub-1", Type: websocket_mod.FolderSubscribe}
		c.AddSubscription(sub)

		// Should not panic or error
		c.RemoveSubscription("non-existent")

		subs := (c.GetSubscriptions())
		assert.Len(t, subs, 1)
	})
}

func TestSubscriptionConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("concurrent add operations are thread-safe", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		var wg sync.WaitGroup

		numGoroutines := 10
		subsPerGoroutine := 5

		// Concurrently add subscriptions
		for i := range numGoroutines {
			wg.Add(1)

			go func(idx int) {
				defer wg.Done()

				for j := range subsPerGoroutine {
					sub := websocket_mod.Subscription{
						SubscriptionID: string(rune('a'+idx)) + string(rune('0'+j)),
						Type:           websocket_mod.FolderSubscribe,
					}
					c.AddSubscription(sub)
				}
			}(i)
		}

		wg.Wait()

		// Verify all subscriptions were added
		subs := (c.GetSubscriptions())
		assert.Len(t, subs, numGoroutines*subsPerGoroutine)
	})

	t.Run("concurrent remove operations are thread-safe", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		// Add subscriptions first
		for i := range 20 {
			sub := websocket_mod.Subscription{
				SubscriptionID: string(rune('a' + i)),
				Type:           websocket_mod.FolderSubscribe,
			}
			c.AddSubscription(sub)
		}

		var wg sync.WaitGroup

		// Concurrently remove subscriptions
		for i := range 10 {
			wg.Add(1)

			go func(idx int) {
				defer wg.Done()

				c.RemoveSubscription(string(rune('a' + idx)))
			}(i)
		}

		wg.Wait()

		// Verify the remaining subscriptions
		subs := (c.GetSubscriptions())
		assert.Len(t, subs, 10)
	})
}

func TestSend(t *testing.T) {
	t.Parallel()

	t.Run("send message to connected client", func(t *testing.T) {
		conn, handler := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		msg := websocket_mod.WsResponseInfo{
			EventTag:     websocket_mod.FileUpdatedEvent,
			SubscribeKey: "test-key",
		}

		err := c.Send(msg)
		require.NoError(t, err)

		// Wait for message to be received
		assert.Eventually(t, func() bool {
			return len(handler.MessagesReceived()) > 0
		}, 2*time.Second, 50*time.Millisecond)

		received := handler.MessagesReceived()[0]
		assert.Equal(t, websocket_mod.FileUpdatedEvent, received.EventTag)
		assert.Equal(t, "test-key", received.SubscribeKey)
		assert.NotZero(t, received.SentTime)
	})

	t.Run("send fails on closed client", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		err := c.Disconnect()
		require.NoError(t, err)

		msg := websocket_mod.WsResponseInfo{
			EventTag: websocket_mod.FileUpdatedEvent,
		}

		err = c.Send(msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed client")
	})

	t.Run("send preserves existing SentTime", func(t *testing.T) {
		conn, handler := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		customTime := int64(1234567890)
		msg := websocket_mod.WsResponseInfo{
			EventTag: websocket_mod.FileUpdatedEvent,
			SentTime: customTime,
		}

		err := c.Send(msg)
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			return len(handler.MessagesReceived()) > 0
		}, 2*time.Second, 50*time.Millisecond)

		received := handler.MessagesReceived()[0]
		assert.Equal(t, customTime, received.SentTime)
	})
}

func TestReadOne(t *testing.T) {
	t.Parallel()

	t.Run("read fails on closed client", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		err := c.Disconnect()
		require.NoError(t, err)

		_, _, err = c.ReadOne()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client is closed")
	})
}

func TestDisconnect(t *testing.T) {
	t.Parallel()

	t.Run("disconnect closes connection and sets inactive", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		assert.True(t, c.IsOpen())

		err := c.Disconnect()
		require.NoError(t, err)

		assert.False(t, c.IsOpen())
	})

	t.Run("disconnect is idempotent", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)

		err := c.Disconnect()
		require.NoError(t, err)

		// Second disconnect should not error
		err = c.Disconnect()
		assert.NoError(t, err)
		assert.False(t, c.IsOpen())
	})
}

func TestLog(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil logger", func(t *testing.T) {
		conn, _ := createMockConnection()
		user := &user_model.User{Username: "testuser"}

		c := client.NewClient(context.Background(), conn, user)
		defer c.Disconnect() //nolint:errcheck

		logger := c.Log()
		assert.NotNil(t, logger)
	})
}
