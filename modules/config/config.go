package config

// ConfigProvider provides configuration for Weblens. All values provided are external to the application, and are expected to be set
// prior to initial startup using environment variables, etc. For management of runtime server settings, those will be stored in the
// database at /models/settings/...
type ConfigProvider struct {
	Host string
	Port string

	MongoDBUri  string
	MongoDBName string
}

func GetConfig() ConfigProvider {
	return getDefaultConfig()
}

func getDefaultConfig() ConfigProvider {
	return ConfigProvider{
		Host:        "localhost",
		Port:        "8080",
		MongoDBUri:  "mongodb://localhost:27017",
		MongoDBName: "weblens",
	}
}
