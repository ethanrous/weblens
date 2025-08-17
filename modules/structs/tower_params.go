package structs

type InitServerParams struct {
	Name string `json:"name"`
	Role string `json:"role"`

	Username    string `json:"username"`
	Password    string `json:"password"`
	FullName    string `json:"fullName"`
	CoreAddress string `json:"coreAddress"`
	CoreKey     string `json:"coreKey"`
	RemoteId    string `json:"remoteId"`

	// For restoring a server, remoind the core of its serverId and api key the remote last used
	LocalId      string `json:"localId"`
	UsingKeyInfo string `json:"usingKeyInfo"`
}

type NewServerParams struct {
	Id          string `json:"serverId"`
	Role        string `json:"role"`
	Name        string `json:"name"`
	CoreAddress string `json:"coreAddress"`
	UsingKey    string `json:"usingKey"`
} // @name NewServerParams

type UpdateServerParams struct {
	Name        string `json:"name"`
	CoreAddress string `json:"coreAddress"`
	UsingKey    string `json:"usingKey"`
} // @name UpdateServerParams

type NewCoreBody struct {
	CoreAddress string `json:"coreAddress"`
	UsingKey    string `json:"usingKey"`
}

type SetConfigParam struct {
	ConfigKey   string `json:"configKey"`
	ConfigValue any    `json:"configValue"`
}

type SetConfigParams []SetConfigParam // @name SetConfigParams

// type SetConfigParams struct {
// 	AllowRegistrations bool `json:"allowRegistrations"`
// 	EnableHDIR         bool `json:"enableHDIR"`
// } // @name SetConfigParams
