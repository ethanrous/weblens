package startup

import (
	"context"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
)

type StartupFunc func(context.Context, config.ConfigProvider) error

var startups []StartupFunc

func RegisterStartup(f StartupFunc) {
	startups = append(startups, f)
}

var ErrDeferStartup = errors.New("defer startup")

func RunStartups(ctx context.Context, cnf config.ConfigProvider) error {
	toRun := startups
	for len(toRun) != 0 {
		var startup StartupFunc

		startup, toRun = toRun[0], toRun[1:]
		if err := startup(ctx, cnf); err != nil {
			if errors.Is(err, ErrDeferStartup) {
				if len(toRun) == 0 {
					return errors.New("startup requested to be defered, but there are no more startups to run")
				}

				// Defer the startup
				toRun = append(toRun, startup)

				continue
			}

			return err
		}
	}

	return nil
}
