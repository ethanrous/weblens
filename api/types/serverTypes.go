package types

type ServerInfo interface {
	ServerId() string
	ServerRole() ServerRole
	// Role() serv
	GetCoreAddress() (string, error)
	GetUsingKey() WeblensApiKey
}

type ServerRole string

const (
	Core           ServerRole = "core"
	Backup         ServerRole = "backup"
	Initialization ServerRole = "init"
)
