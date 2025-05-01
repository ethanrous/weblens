package structs

type TowerInfo struct {
	Id   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey string `json:"-" validate:"required"`

	// Core or Backup
	Role string `json:"role" validate:"required"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress" validate:"required"`

	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	ReportedRole string `json:"reportedRole" validate:"required"`

	LastBackup int64 `json:"lastBackup" validate:"required" format:"int64"`

	BackupSize int64 `json:"backupSize" validate:"required" format:"int64"`

	UserCount int `json:"userCount" validate:"required"`

	// If this server info represents this local server
	IsThisServer bool `json:"-" validate:"required"`

	Online bool `json:"online" validate:"required"`

	Started bool `json:"started" validate:"required"`
} // @name TowerInfo
