package mock

// import (
// 	"github.com/ethanrous/weblens/fileTree"
// 	"github.com/ethanrous/weblens/models"
// 	"github.com/ethanrous/weblens/task"
// )
//
// var _ models.Broadcaster = (*MockCaster)(nil)
//
// type MockCaster struct{}
//
// func (m *MockCaster) PushWeblensEvent(eventTag string, content ...models.WsC) {}
//
// func (m *MockCaster) PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *models.Media) {}
//
// func (m *MockCaster) PushFilesUpdate(updatedFiles []*fileTree.WeblensFileImpl, medias []*models.Media) {}
//
// func (m *MockCaster) PushTaskUpdate(task *task.Task, event string, result task.TaskResult) {}
//
// func (m *MockCaster) PushPoolUpdate(pool task.Pool, event string, result task.TaskResult) {}
//
// func (m *MockCaster) PushFileCreate(newFile *fileTree.WeblensFileImpl) {}
//
// func (m *MockCaster) PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl) {
// }
//
// func (m *MockCaster) PushFilesMove(preMoveParentId, postMoveParentId fileTree.FileId, files []*fileTree.WeblensFileImpl) {
// }
//
// func (m *MockCaster) PushFileDelete(deletedFile *fileTree.WeblensFileImpl) {}
//
// func (m *MockCaster) PushFilesDelete(deletedFiles []*fileTree.WeblensFileImpl) {}
//
// func (m *MockCaster) PushShareUpdate(username string, newShareInfo models.Share) {}
//
// func (m *MockCaster) Enable() {}
//
// func (m *MockCaster) Disable() {}
//
// func (m *MockCaster) IsEnabled() bool {
// 	return false
// }
//
// func (m *MockCaster) IsBuffered() bool {
// 	return false
// }
//
// func (m *MockCaster) FolderSubToTask(folder fileTree.FileId, taskId task.Id) {}
//
// func (m *MockCaster) DisableAutoFlush() {}
//
// func (m *MockCaster) AutoFlushEnable() {}
//
// func (m *MockCaster) Flush() {}
//
// func (m *MockCaster) Relay(msg models.WsResponseInfo) {}
//
// func (m *MockCaster) Close() {}
