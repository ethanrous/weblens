package structs

// InitServerParams contains parameters for initializing a server.
type InitServerParams struct {
	Name string `json:"name"`
	Role string `json:"role"`

	Username    string `json:"username"`
	Password    string `json:"password"`
	FullName    string `json:"fullName"`
	CoreAddress string `json:"coreAddress"`
	CoreKey     string `json:"coreKey"`
	RemoteID    string `json:"remoteID"`

	// For restoring a server, remoind the core of its serverID and api key the remote last used
	LocalID      string `json:"localID"`
	UsingKeyInfo string `json:"usingKeyInfo"`
}

// NewServerParams contains parameters for creating a new server.
type NewServerParams struct {
	ID          string `json:"serverID"`
	Role        string `json:"role"`
	Name        string `json:"name"`
	CoreAddress string `json:"coreAddress"`
	UsingKey    string `json:"usingKey"`
} // @name NewServerParams

// UpdateServerParams contains parameters for updating a server.
type UpdateServerParams struct {
	Name        string `json:"name"`
	CoreAddress string `json:"coreAddress"`
	UsingKey    string `json:"usingKey"`
} // @name UpdateServerParams

// NewCoreBody contains parameters for registering with a core server.
type NewCoreBody struct {
	CoreAddress string `json:"coreAddress"`
	UsingKey    string `json:"usingKey"`
}

// SetConfigParam represents a single configuration key-value pair.
type SetConfigParam struct {
	ConfigKey   string `json:"configKey"`
	ConfigValue any    `json:"configValue"`
}

// SetConfigParams represents a list of configuration parameters to set.
type SetConfigParams []SetConfigParam // @name SetConfigParams

// type SetConfigParams struct {
// 	AllowRegistrations bool `json:"allowRegistrations"`
// 	EnableHDIR         bool `json:"enableHDIR"`
// } // @name SetConfigParams
