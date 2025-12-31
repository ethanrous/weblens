// Package test provides testing utilities for Weblens.
package test

import (
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/rs/zerolog"
)

// NewWeblensTestInstance creates a new Weblens test instance with the specified configuration.
func NewWeblensTestInstance(_ string, _ config.Provider, _ zerolog.Logger) (ctxservice.AppContext, error) {
	return ctxservice.AppContext{}, wlerrors.New("not implemented")
}

