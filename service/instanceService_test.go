package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestInstanceServiceImpl_Add(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		i *models.Instance
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
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				tt.wantErr(t, is.Add(tt.args.i), fmt.Sprintf("Add(%v)", tt.args.i))
			},
		)
	}
}

func TestInstanceServiceImpl_AddLoading(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		loadingKey string
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
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				is.AddLoading(tt.args.loadingKey)
			},
		)
	}
}

func TestInstanceServiceImpl_Del(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		iId models.InstanceId
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
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				tt.wantErr(t, is.Del(tt.args.iId), fmt.Sprintf("Del(%v)", tt.args.iId))
			},
		)
	}
}

func TestInstanceServiceImpl_GenerateNewId(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   models.InstanceId
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(t, tt.want, is.GenerateNewId(tt.args.name), "GenerateNewId(%v)", tt.args.name)
			},
		)
	}
}

func TestInstanceServiceImpl_Get(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		iId models.InstanceId
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *models.Instance
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(t, tt.want, is.Get(tt.args.iId), "Get(%v)", tt.args.iId)
			},
		)
	}
}

func TestInstanceServiceImpl_GetCore(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   *models.Instance
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(t, tt.want, is.GetCore(), "GetCore()")
			},
		)
	}
}

func TestInstanceServiceImpl_GetLocal(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   *models.Instance
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(t, tt.want, is.GetLocal(), "GetLocal()")
			},
		)
	}
}

func TestInstanceServiceImpl_GetRemotes(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   []*models.Instance
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(t, tt.want, is.GetRemotes(), "GetRemotes()")
			},
		)
	}
}

func TestInstanceServiceImpl_Init(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
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
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				tt.wantErr(t, is.Init(), fmt.Sprintf("Init()"))
			},
		)
	}
}

func TestInstanceServiceImpl_InitBackup(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		name     string
		coreAddr string
		key      models.WeblensApiKey
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
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				tt.wantErr(
					t, is.InitBackup(tt.args.name, tt.args.coreAddr, tt.args.key),
					fmt.Sprintf("InitBackup(%v, %v, %v)", tt.args.name, tt.args.coreAddr, tt.args.key),
				)
			},
		)
	}
}

func TestInstanceServiceImpl_InitCore(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		instance *models.Instance
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
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				tt.wantErr(t, is.InitCore(tt.args.instance), fmt.Sprintf("InitCore(%v)", tt.args.instance))
			},
		)
	}
}

func TestInstanceServiceImpl_IsLocalLoaded(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(t, tt.want, is.IsLocalLoaded(), "IsLocalLoaded()")
			},
		)
	}
}

func TestInstanceServiceImpl_RemoveLoading(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
	}
	type args struct {
		loadingKey string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantDoneLoading bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(
					t, tt.wantDoneLoading, is.RemoveLoading(tt.args.loadingKey), "RemoveLoading(%v)",
					tt.args.loadingKey,
				)
			},
		)
	}
}

func TestInstanceServiceImpl_Size(t *testing.T) {
	type fields struct {
		instanceMap     map[models.InstanceId]*models.Instance
		instanceMapLock sync.RWMutex
		local           *models.Instance
		core            *models.Instance
		localLoading    map[string]bool
		col             *mongo.Collection
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
				is := &InstanceServiceImpl{
					instanceMap:     tt.fields.instanceMap,
					instanceMapLock: tt.fields.instanceMapLock,
					local:           tt.fields.local,
					core:            tt.fields.core,
					localLoading:    tt.fields.localLoading,
					col:             tt.fields.col,
				}
				assert.Equalf(t, tt.want, is.Size(), "Size()")
			},
		)
	}
}

func TestMakeUniqueChildName(t *testing.T) {
	type args struct {
		parent *fileTree.WeblensFileImpl
		childName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t, tt.want, MakeUniqueChildName(tt.args.parent, tt.args.childName), "MakeUniqueChildName(%v, %v)",
					tt.args.parent, tt.args.childName,
				)
			},
		)
	}
}
