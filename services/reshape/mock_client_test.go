package reshape_test

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/client"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/gorilla/websocket"
)

// mockClientManager implements client.Manager for testing purposes.
// It allows configuring which users appear online.
type mockClientManager struct {
	onlineUsers  map[string]bool
	onlineTowers map[string]bool
}

func newMockClientManager() *mockClientManager {
	return &mockClientManager{
		onlineUsers:  make(map[string]bool),
		onlineTowers: make(map[string]bool),
	}
}

func (m *mockClientManager) setUserOnline(username string, online bool) {
	m.onlineUsers[username] = online
}

func (m *mockClientManager) setTowerOnline(towerID string, online bool) {
	m.onlineTowers[towerID] = online
}

// GetClientByUsername returns a non-nil client if the user is marked as online.
func (m *mockClientManager) GetClientByUsername(username string) *client.WsClient {
	if m.onlineUsers[username] {
		return &client.WsClient{}
	}
	return nil
}

// GetClientByTowerID returns a non-nil client if the tower is marked as online.
func (m *mockClientManager) GetClientByTowerID(towerID string) *client.WsClient {
	if m.onlineTowers[towerID] {
		return &client.WsClient{}
	}
	return nil
}

// Stub implementations for remaining interface methods

func (m *mockClientManager) Notify(ctx context.Context, msg ...websocket_mod.WsResponseInfo) {}

func (m *mockClientManager) ClientConnect(ctx context_mod.LoggerContext, conn *websocket.Conn, user *user_model.User) (*client.WsClient, error) {
	return nil, nil
}

func (m *mockClientManager) RemoteConnect(ctx context_mod.LoggerContext, conn *websocket.Conn, remote *tower_model.Instance) *client.WsClient {
	return nil
}

func (m *mockClientManager) ClientDisconnect(ctx context.Context, c *client.WsClient) {}

func (m *mockClientManager) DisconnectAll(ctx context.Context) {}

func (m *mockClientManager) GetAllClients() []*client.WsClient {
	return nil
}

func (m *mockClientManager) GetConnectedAdmins() []*client.WsClient {
	return nil
}

func (m *mockClientManager) GetSubscribers(ctx context_mod.LoggerContext, st websocket_mod.SubscriptionType, key string) []*client.WsClient {
	return nil
}

func (m *mockClientManager) SubscribeToFile(ctx context_mod.Z, c *client.WsClient, file *file_model.WeblensFileImpl, share *share_model.FileShare, subTime time.Time) error {
	return nil
}

func (m *mockClientManager) SubscribeToTask(ctx context_mod.LoggerContext, c *client.WsClient, t *task_model.Task, subTime time.Time) error {
	return nil
}

func (m *mockClientManager) Unsubscribe(ctx context_mod.LoggerContext, c *client.WsClient, key string, unSubTime time.Time) error {
	return nil
}

func (m *mockClientManager) FolderSubToTask(ctx context_mod.LoggerContext, folderID string, t task.Task) {
}

func (m *mockClientManager) UnsubTask(ctx context.Context, taskID string) {}

func (m *mockClientManager) Send(ctx context.Context, msg websocket_mod.WsResponseInfo) {}

func (m *mockClientManager) Flush(ctx context.Context) {}

func (m *mockClientManager) Relay(msg websocket_mod.WsResponseInfo) {}

// Verify interface compliance at compile time
var _ client.Manager = (*mockClientManager)(nil)

// mockFileService implements file.Service for testing purposes.
type mockFileService struct {
	files map[string]*file_model.WeblensFileImpl
}

func newMockFileService() *mockFileService {
	return &mockFileService{
		files: make(map[string]*file_model.WeblensFileImpl),
	}
}

func (m *mockFileService) addFile(id string, f *file_model.WeblensFileImpl) {
	m.files[id] = f
}

func (m *mockFileService) GetFileByID(ctx context.Context, fileID string) (*file_model.WeblensFileImpl, error) {
	if f, ok := m.files[fileID]; ok {
		return f, nil
	}
	return nil, file_model.ErrFileNotFound
}

// Stub implementations for remaining interface methods

func (m *mockFileService) AddFile(ctx context.Context, file ...*file_model.WeblensFileImpl) error {
	return nil
}

func (m *mockFileService) Size(treeAlias string) int64 {
	return 0
}

func (m *mockFileService) GetFileByFilepath(ctx context.Context, path file_system.Filepath, dontLoadNew ...bool) (*file_model.WeblensFileImpl, error) {
	return nil, file_model.ErrFileNotFound
}

func (m *mockFileService) CreateFile(ctx context.Context, parent *file_model.WeblensFileImpl, filename string, data ...[]byte) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) CreateFolder(ctx context.Context, parent *file_model.WeblensFileImpl, folderName string) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) GetChildren(ctx context.Context, folder *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) RecursiveEnsureChildrenLoaded(ctx context.Context, folder *file_model.WeblensFileImpl) error {
	return nil
}

func (m *mockFileService) CreateUserHome(ctx context.Context, user *user_model.User) error {
	return nil
}

func (m *mockFileService) NewBackupRestoreFile(ctx context.Context, contentID, remoteTowerID string) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) InitBackupDirectory(ctx context.Context, tower tower_model.Instance) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) MoveFiles(ctx context.Context, files []*file_model.WeblensFileImpl, destFolder *file_model.WeblensFileImpl) error {
	return nil
}

func (m *mockFileService) RenameFile(ctx context.Context, file *file_model.WeblensFileImpl, newName string) error {
	return nil
}

func (m *mockFileService) ReturnFilesFromTrash(ctx context.Context, trashFiles []*file_model.WeblensFileImpl) error {
	return nil
}

func (m *mockFileService) DeleteFiles(ctx context.Context, files ...*file_model.WeblensFileImpl) error {
	return nil
}

func (m *mockFileService) RestoreFiles(ctx context.Context, ids []string, newParent *file_model.WeblensFileImpl, restoreTime time.Time) error {
	return nil
}

func (m *mockFileService) GetMediaCacheByFilename(ctx context.Context, filename string) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) GetFileByContentID(ctx context.Context, contentID string) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) NewCacheFile(mediaID string, quality string, pageNum int) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) DeleteCacheFile(file *file_model.WeblensFileImpl) error {
	return nil
}

func (m *mockFileService) NewZip(ctx context.Context, zipName string, owner *user_model.User) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

func (m *mockFileService) GetZip(ctx context.Context, id string) (*file_model.WeblensFileImpl, error) {
	return nil, nil
}

// Verify interface compliance at compile time
var _ file_model.Service = (*mockFileService)(nil)
