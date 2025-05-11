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
	for len(startups) != 0 {
		var startup StartupFunc

		startup, startups = startups[0], startups[1:]
		if err := startup(ctx, cnf); err != nil {
			if errors.Is(err, ErrDeferStartup) {
				if len(startups) == 0 {
					return errors.New("startup requested to be defered, but there are no more startups to run")
				}

				// Defer the startup
				startups = append(startups, startup)

				continue
			}

			return err
		}
	}

	return nil
}
