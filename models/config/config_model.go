// Package config provides configuration management for Weblens.
package config

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConfigCollectionKey is the database collection name for configuration documents.
const ConfigCollectionKey = "config"

// OptionKey represents a configuration key used to identify specific settings.
type OptionKey = string

// Configuration keys for Weblens settings.
const (
	// AllowRegistrations controls whether new user registration is allowed.
	AllowRegistrations OptionKey = "allowRegistrations"
	// EnableHDIR controls whether HDIR (high-dynamic-range image rendering) is enabled.
	EnableHDIR OptionKey = "enableHDIR"
)

// Config represents the application configuration settings.
type Config struct {
	AllowRegistrations bool `bson:"allowRegistrations" json:"allowRegistrations"`
	EnableHDIR         bool `bson:"enableHDIR" json:"enableHDIR"`
} // @name Config

// DefaultConfig returns the default configuration with all settings enabled.
func DefaultConfig() Config {
	return Config{
		AllowRegistrations: true,
		EnableHDIR:         true,
	}
}

// GetConfig retrieves the current configuration from cache or database.
func GetConfig(ctx context.Context) (Config, error) {
	config, ok := cache.GetOneAs[Config](ctx, ConfigCollectionKey, "config")
	if ok {
		return config, nil
	}

	col, err := db.GetCollection[Config](ctx, ConfigCollectionKey)
	if err != nil {
		return Config{}, err
	}

	config, err = col.FindOneAs(ctx, bson.M{"_id": "config"})
	if err != nil {
		if db.IsNotFound(err) {
			defaultConfig := DefaultConfig()

			err = SaveConfig(ctx, defaultConfig)
			if err != nil {
				return Config{}, err
			}

			return defaultConfig, nil
		}

		return Config{}, err
	}

	err = cache.SetOne(ctx, ConfigCollectionKey, "config", config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

// SaveConfig saves the configuration to the database and updates the cache.
func SaveConfig(ctx context.Context, config Config) error {
	col, err := db.GetCollection[Config](ctx, ConfigCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.ReplaceOne(ctx, bson.M{"_id": "config"}, config, options.Replace().SetUpsert(true))
	if err != nil {
		return db.WrapError(err, "update config")
	}

	err = cache.SetOne(ctx, ConfigCollectionKey, "config", config)
	if err != nil {
		return err
	}

	return nil
}
