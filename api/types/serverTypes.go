package types

type ServerInfo interface {
	ServerId() string
	ServerRole() ServerRole
	GetCoreAddress() (string, error)
}

type ServerRole string

const (
	CoreMode   ServerRole = "core"
	BackupMode ServerRole = "backup"
)
