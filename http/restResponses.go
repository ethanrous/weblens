package http

import "github.com/ethrousseau/weblens/api/types"

type serverInfo struct {
	Id   types.InstanceId `json:"id"`
	Name string           `json:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey types.WeblensApiKey `json:"-"`

	// Core or Backup
	Role types.ServerRole `json:"role"`

	// If this server info represents this local server
	IsThisServer bool `json:"-"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress"`

	Online bool `json:"online"`
}
