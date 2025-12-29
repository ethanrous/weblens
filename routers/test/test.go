// Package test provides testing utilities for Weblens.
package test

import (
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/services/context"
	"github.com/rs/zerolog"
)

// NewWeblensTestInstance creates a new Weblens test instance with the specified configuration.
func NewWeblensTestInstance(_ string, _ config.Provider, _ zerolog.Logger) (context.AppContext, error) {
	return context.AppContext{}, errors.New("not implemented")
}