package config_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("returns config with defaults enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		assert.True(t, cfg.AllowRegistrations)
		assert.True(t, cfg.EnableHDIR)
	})
}

func TestOptionKeyConstants(t *testing.T) {
	t.Run("AllowRegistrations key", func(t *testing.T) {
		assert.Equal(t, config.OptionKey("allowRegistrations"), config.AllowRegistrations)
	})

	t.Run("EnableHDIR key", func(t *testing.T) {
		assert.Equal(t, config.OptionKey("enableHDIR"), config.EnableHDIR)
	})
}

func TestConfigCollectionKey(t *testing.T) {
	t.Run("returns correct collection name", func(t *testing.T) {
		assert.Equal(t, "config", config.ConfigCollectionKey)
	})
}
