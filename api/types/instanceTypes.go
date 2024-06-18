package types

type WeblensApiKey string
type ServerRole string
type InstanceId string

const (
	Initialization ServerRole = "init"

	Core   ServerRole = "core"
	Backup ServerRole = "backup"
)

type Instance interface {
	ServerId() InstanceId
	GetName() string
	GetUsingKey() WeblensApiKey
	ServerRole() ServerRole
	GetCoreAddress() (string, error)
	Info() Instance
	IsLocal() bool
}

type InstanceService interface {
	BaseService[InstanceId, Instance]
	GetLocal() Instance
	GetRemotes() []Instance
	GenerateNewId(name string) InstanceId
	InitCore(Instance) error
	InitBackup(Instance) error
}

var ErrAlreadyCore = NewWeblensError("core server cannot have a remote core")
