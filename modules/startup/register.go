package startup

import (
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/context"
)

type StartupFunc func(context.ContextZ, config.ConfigProvider) error

var startups []StartupFunc

func RegisterStartup(f StartupFunc) {
	startups = append(startups, f)
}

func RunStartups(ctx context.ContextZ, cnf config.ConfigProvider) error {
	for _, startup := range startups {
		if err := startup(ctx, cnf); err != nil {
			return err
		}
	}
	return nil
}
