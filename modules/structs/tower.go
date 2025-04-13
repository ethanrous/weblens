package structs

type TowerInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey string `json:"-"`

	// Core or Backup
	Role string `json:"role"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress"`

	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	ReportedRole string `json:"reportedRole"`

	LastBackup int64 `json:"lastBackup"`

	BackupSize int64 `json:"backupSize"`

	UserCount int `json:"userCount"`

	// If this server info represents this local server
	IsThisServer bool `json:"-"`

	Online bool `json:"online"`

	Started bool `json:"started"`
} // @name TowerInfo
