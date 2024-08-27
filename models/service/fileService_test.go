package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestFileServiceImpl_AddTask(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		f *fileTree.WeblensFile
		t *task.Task
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(t, fs.AddTask(tt.args.f, tt.args.t), fmt.Sprintf("AddTask(%v, %v)", tt.args.f, tt.args.t))
			},
		)
	}
}

func TestFileServiceImpl_CreateFile(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		parent   *fileTree.WeblensFile
		fileName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.CreateFile(tt.args.parent, tt.args.fileName)
				if !tt.wantErr(t, err, fmt.Sprintf("CreateFile(%v, %v)", tt.args.parent, tt.args.fileName)) {
					return
				}
				assert.Equalf(t, tt.want, got, "CreateFile(%v, %v)", tt.args.parent, tt.args.fileName)
			},
		)
	}
}

func TestFileServiceImpl_CreateFolder(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		parent     *fileTree.WeblensFile
		folderName string
		caster     models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.CreateFolder(tt.args.parent, tt.args.folderName, tt.args.caster)
				if !tt.wantErr(
					t, err, fmt.Sprintf("CreateFolder(%v, %v, %v)", tt.args.parent, tt.args.folderName, tt.args.caster),
				) {
					return
				}
				assert.Equalf(
					t, tt.want, got, "CreateFolder(%v, %v, %v)", tt.args.parent, tt.args.folderName, tt.args.caster,
				)
			},
		)
	}
}

func TestFileServiceImpl_DeleteCacheFile(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		f *fileTree.WeblensFile
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(t, fs.DeleteCacheFile(tt.args.f), fmt.Sprintf("DeleteCacheFile(%v)", tt.args.f))
			},
		)
	}
}

func TestFileServiceImpl_GetFileOwner(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		file *fileTree.WeblensFile
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *models.User
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				assert.Equalf(t, tt.want, fs.GetFileOwner(tt.args.file), "GetFileOwner(%v)", tt.args.file)
			},
		)
	}
}

func TestFileServiceImpl_GetFileSafe(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		id    fileTree.FileId
		user  *models.User
		share *models.FileShare
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.GetFileSafe(tt.args.id, tt.args.user, tt.args.share)
				if !tt.wantErr(
					t, err, fmt.Sprintf("GetFileSafe(%v, %v, %v)", tt.args.id, tt.args.user, tt.args.share),
				) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetFileSafe(%v, %v, %v)", tt.args.id, tt.args.user, tt.args.share)
			},
		)
	}
}

func TestFileServiceImpl_GetMediaJournal(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   fileTree.JournalService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				assert.Equalf(t, tt.want, fs.GetMediaJournal(), "GetMediaJournal()")
			},
		)
	}
}

func TestFileServiceImpl_GetMediaRoot(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   *fileTree.WeblensFile
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				assert.Equalf(t, tt.want, fs.GetMediaRoot(), "GetMediaRoot()")
			},
		)
	}
}

func TestFileServiceImpl_GetTasks(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		f *fileTree.WeblensFile
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*task.Task
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				assert.Equalf(t, tt.want, fs.GetTasks(tt.args.f), "GetTasks(%v)", tt.args.f)
			},
		)
	}
}

func TestFileServiceImpl_GetThumbFileId(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		id fileTree.FileId
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.GetThumbFileId(tt.args.id)
				if !tt.wantErr(t, err, fmt.Sprintf("GetThumbFileId(%v)", tt.args.id)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetThumbFileId(%v)", tt.args.id)
			},
		)
	}
}

func TestFileServiceImpl_GetThumbFileName(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		thumbFileName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.GetThumbFileName(tt.args.thumbFileName)
				if !tt.wantErr(t, err, fmt.Sprintf("GetThumbFileName(%v)", tt.args.thumbFileName)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetThumbFileName(%v)", tt.args.thumbFileName)
			},
		)
	}
}

func TestFileServiceImpl_IsFileInTrash(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		f *fileTree.WeblensFile
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				assert.Equalf(t, tt.want, fs.IsFileInTrash(tt.args.f), "IsFileInTrash(%v)", tt.args.f)
			},
		)
	}
}

func TestFileServiceImpl_MoveFile(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		file        *fileTree.WeblensFile
		destParent  *fileTree.WeblensFile
		newFilename string
		caster      models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.MoveFile(tt.args.file, tt.args.destParent, tt.args.newFilename, tt.args.caster), fmt.Sprintf(
						"MoveFile(%v, %v, %v, %v)", tt.args.file, tt.args.destParent, tt.args.newFilename,
						tt.args.caster,
					),
				)
			},
		)
	}
}

func TestFileServiceImpl_MoveFileToTrash(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		file   *fileTree.WeblensFile
		user   *models.User
		share  *models.FileShare
		caster models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.MoveFileToTrash(tt.args.file, tt.args.user, tt.args.share, tt.args.caster), fmt.Sprintf(
						"MoveFileToTrash(%v, %v, %v, %v)", tt.args.file, tt.args.user, tt.args.share, tt.args.caster,
					),
				)
			},
		)
	}
}

func TestFileServiceImpl_NewCacheFile(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		contentId string
		quality   models.MediaQuality
		pageNum   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.NewCacheFile(tt.args.contentId, tt.args.quality, tt.args.pageNum)
				if !tt.wantErr(
					t, err,
					fmt.Sprintf("NewCacheFile(%v, %v, %v)", tt.args.contentId, tt.args.quality, tt.args.pageNum),
				) {
					return
				}
				assert.Equalf(
					t, tt.want, got, "NewCacheFile(%v, %v, %v)", tt.args.contentId, tt.args.quality, tt.args.pageNum,
				)
			},
		)
	}
}

func TestFileServiceImpl_NewZip(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		zipName string
		owner   *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.NewZip(tt.args.zipName, tt.args.owner)
				if !tt.wantErr(t, err, fmt.Sprintf("NewZip(%v, %v)", tt.args.zipName, tt.args.owner)) {
					return
				}
				assert.Equalf(t, tt.want, got, "NewZip(%v, %v)", tt.args.zipName, tt.args.owner)
			},
		)
	}
}

func TestFileServiceImpl_PathToFile(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		searchPath string
		u          *models.User
		share      *models.FileShare
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		want1   []*fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, got1, err := fs.PathToFile(tt.args.searchPath, tt.args.u, tt.args.share)
				if !tt.wantErr(
					t, err, fmt.Sprintf("PathToFile(%v, %v, %v)", tt.args.searchPath, tt.args.u, tt.args.share),
				) {
					return
				}
				assert.Equalf(t, tt.want, got, "PathToFile(%v, %v, %v)", tt.args.searchPath, tt.args.u, tt.args.share)
				assert.Equalf(t, tt.want1, got1, "PathToFile(%v, %v, %v)", tt.args.searchPath, tt.args.u, tt.args.share)
			},
		)
	}
}

func TestFileServiceImpl_PermanentlyDeleteFiles(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		files  []*fileTree.WeblensFile
		caster models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.PermanentlyDeleteFiles(tt.args.files, tt.args.caster),
					fmt.Sprintf("PermanentlyDeleteFiles(%v, %v)", tt.args.files, tt.args.caster),
				)
			},
		)
	}
}

func TestFileServiceImpl_RemoveTask(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		f *fileTree.WeblensFile
		t *task.Task
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.RemoveTask(tt.args.f, tt.args.t), fmt.Sprintf("RemoveTask(%v, %v)", tt.args.f, tt.args.t),
				)
			},
		)
	}
}

func TestFileServiceImpl_ResizeDown(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		f      *fileTree.WeblensFile
		caster models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.ResizeDown(tt.args.f, tt.args.caster),
					fmt.Sprintf("ResizeDown(%v, %v)", tt.args.f, tt.args.caster),
				)
			},
		)
	}
}

func TestFileServiceImpl_ResizeUp(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		f      *fileTree.WeblensFile
		caster models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.ResizeUp(tt.args.f, tt.args.caster),
					fmt.Sprintf("ResizeUp(%v, %v)", tt.args.f, tt.args.caster),
				)
			},
		)
	}
}

func TestFileServiceImpl_ReturnFilesFromTrash(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		trashFiles []*fileTree.WeblensFile
		c          models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.ReturnFilesFromTrash(tt.args.trashFiles, tt.args.c),
					fmt.Sprintf("ReturnFilesFromTrash(%v, %v)", tt.args.trashFiles, tt.args.c),
				)
			},
		)
	}
}

func TestFileServiceImpl_SetAccessService(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		accessService *AccessServiceImpl
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				fs.SetAccessService(tt.args.accessService)
			},
		)
	}
}

func TestFileServiceImpl_SetMediaService(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		mediaService *MediaServiceImpl
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				fs.SetMediaService(tt.args.mediaService)
			},
		)
	}
}

func TestFileServiceImpl_Size(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				assert.Equalf(t, tt.want, fs.Size(), "Size()")
			},
		)
	}
}

func TestFileServiceImpl_clearTakeoutDir(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(t, fs.clearTakeoutDir(), fmt.Sprintf("clearTakeoutDir()"))
			},
		)
	}
}

func TestFileServiceImpl_clearTempDir(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(t, fs.clearTempDir(), fmt.Sprintf("clearTempDir()"))
			},
		)
	}
}

func TestFileServiceImpl_clearThumbsDir(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(t, fs.clearThumbsDir(), fmt.Sprintf("clearThumbsDir()"))
			},
		)
	}
}

func TestFileServiceImpl_getFileByIdAndRoot(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		id        fileTree.FileId
		rootAlias string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *fileTree.WeblensFile
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				got, err := fs.getFileByIdAndRoot(tt.args.id, tt.args.rootAlias)
				if !tt.wantErr(t, err, fmt.Sprintf("getFileByIdAndRoot(%v, %v)", tt.args.id, tt.args.rootAlias)) {
					return
				}
				assert.Equalf(t, tt.want, got, "getFileByIdAndRoot(%v, %v)", tt.args.id, tt.args.rootAlias)
			},
		)
	}
}

func TestFileServiceImpl_resizeMultiple(t *testing.T) {
	type fields struct {
		mediaTree       fileTree.FileTree
		cachesTree      fileTree.FileTree
		userService     models.UserService
		accessService   models.AccessService
		mediaService    models.MediaService
		instanceService models.InstanceService
		fileTaskLink    map[fileTree.FileId][]*task.Task
		fileTaskLock    sync.RWMutex
		trashCol        *mongo.Collection
	}
	type args struct {
		old    *fileTree.WeblensFile
		new    *fileTree.WeblensFile
		caster models.FileCaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fs := &FileServiceImpl{
					mediaTree:       tt.fields.mediaTree,
					cachesTree:      tt.fields.cachesTree,
					userService:     tt.fields.userService,
					accessService:   tt.fields.accessService,
					mediaService:    tt.fields.mediaService,
					instanceService: tt.fields.instanceService,
					fileTaskLink:    tt.fields.fileTaskLink,
					fileTaskLock:    tt.fields.fileTaskLock,
					trashCol:        tt.fields.trashCol,
				}
				tt.wantErr(
					t, fs.resizeMultiple(tt.args.old, tt.args.new, tt.args.caster),
					fmt.Sprintf("resizeMultiple(%v, %v, %v)", tt.args.old, tt.args.new, tt.args.caster),
				)
			},
		)
	}
}