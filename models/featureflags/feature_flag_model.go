// Package featureflags provides feature flag management for Weblens.
package featureflags

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FeatureFlagCollectionKey is the database collection name for feature flag document(s).
const FeatureFlagCollectionKey = "featureFlags"

// FeatureFlagDocumentID is the ID of the main feature flag document.
const FeatureFlagDocumentID = "ff"

// FlagKey represents a feature flag key used to identify specific flags.
type FlagKey = string

// Feature Flag keys for Weblens settings.
const (
	// AllowRegistrations controls whether new user registration is allowed.
	AllowRegistrations FlagKey = "auth.allow_registrations"
	// EnableHDIR controls whether HDIR (high-dynamic-range image rendering) is enabled.
	EnableHDIR FlagKey = "media.hdir_processing_enabled"
	// EnableWebDAV controls whether WebDAV file access is enabled.
	EnableWebDAV FlagKey = "webdav.enabled"
)

// Bundle represents the application feature flag document.
type Bundle struct {
	AllowRegistrations bool `bson:"auth.allow_registrations" json:"auth.allow_registrations"`
	EnableHDIR         bool `bson:"media.hdir_processing_enabled" json:"media.hdir_processing_enabled"`
	EnableWebDAV       bool `bson:"webdav.enabled" json:"webdav.enabled"`
} // @name Bundle

// Default returns the default flags
func Default() Bundle {
	return Bundle{
		AllowRegistrations: true,
		EnableHDIR:         false,
	}
}

// GetFlags retrieves the current flag set from cache or database.
func GetFlags(ctx context.Context) (Bundle, error) {
	flags, ok := cache.GetOneAs[Bundle](ctx, FeatureFlagCollectionKey, FeatureFlagDocumentID)
	if ok {
		return flags, nil
	}

	col, err := db.GetCollection[Bundle](ctx, FeatureFlagCollectionKey)
	if err != nil {
		return Bundle{}, err
	}

	flags, err = col.FindOneAs(ctx, bson.M{"_id": FeatureFlagDocumentID})
	if err != nil {
		if db.IsNotFound(err) {
			defaultFlags := Default()

			err = SaveFlags(ctx, defaultFlags)
			if err != nil {
				return Bundle{}, err
			}

			return defaultFlags, nil
		}

		return Bundle{}, err
	}

	err = cache.SetOne(ctx, FeatureFlagCollectionKey, FeatureFlagDocumentID, flags)
	if err != nil {
		return Bundle{}, err
	}

	return flags, nil
}

// SaveFlags saves the flags to the database and updates the cache.
func SaveFlags(ctx context.Context, ffs Bundle) error {
	col, err := db.GetCollection[Bundle](ctx, FeatureFlagCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.ReplaceOne(ctx, bson.M{"_id": FeatureFlagDocumentID}, ffs, options.Replace().SetUpsert(true))
	if err != nil {
		return db.WrapError(err, "update feature flag config")
	}

	err = cache.SetOne(ctx, FeatureFlagCollectionKey, FeatureFlagDocumentID, ffs)
	if err != nil {
		return err
	}

	return nil
}
