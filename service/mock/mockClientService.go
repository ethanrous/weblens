package mock

import (
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
	"github.com/gorilla/websocket"
)

var _ models.ClientManager = (*MockClientService)(nil)

type MockClientService struct{}

func (m *MockClientService) ClientConnect(conn *websocket.Conn, user *models.User) *models.WsClient {
	
	panic("implement me")
}

func (m *MockClientService) RemoteConnect(conn *websocket.Conn, remote *models.Instance) *models.WsClient {
	
	panic("implement me")
}

func (m *MockClientService) GetSubscribers(st models.WsAction, key models.SubId) (clients []*models.WsClient) {
	
	panic("implement me")
}

func (m *MockClientService) GetClientByUsername(username models.Username) *models.WsClient {
	return nil
}

func (m *MockClientService) GetClientByServerId(id models.InstanceId) *models.WsClient {
	return nil
}

func (m *MockClientService) GetAllClients() []*models.WsClient {
	return []*models.WsClient{}
}

func (m *MockClientService) GetConnectedAdmins() []*models.WsClient {
	
	panic("implement me")
}

func (m *MockClientService) FolderSubToPool(folderId fileTree.FileId, poolId task.Id) {}

func (m *MockClientService) TaskSubToPool(taskId task.Id, poolId task.Id) {}

func (m *MockClientService) Subscribe(
	c *models.WsClient, key models.SubId, action models.WsAction, subTime time.Time, share models.Share,
) (complete bool, results map[string]any, err error) {
	
	panic("implement me")
}

func (m *MockClientService) Unsubscribe(c *models.WsClient, key models.SubId, unSubTime time.Time) error {
	
	panic("implement me")
}

func (m *MockClientService) Send(msg models.WsResponseInfo) {
	
	panic("implement me")
}

func (m *MockClientService) ClientDisconnect(c *models.WsClient) {
	
	panic("implement me")
}
