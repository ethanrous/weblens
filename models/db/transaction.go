package db

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionHook func(context.Context) context.Context

var transactionHooks = []TransactionHook{}

func InstallTransactionHook(hook TransactionHook) {
	transactionHooks = append(transactionHooks, hook)
}

func hasTransaction(ctx context.Context) bool {
	return mongo.SessionFromContext(ctx) != nil
}

func WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	db, err := getDbFromContext(ctx)
	if err != nil {
		return err
	}

	l := log.FromContext(ctx)

	if hasTransaction(ctx) {
		// return errors.Errorf("Already in a transaction")
		l.Debug().Msg("Already in a transaction, skipping straight to callback function")

		return fn(ctx)
	}

	session, err := db.Client().StartSession()
	if err != nil {
		return errors.WithStack(err)
	}

	defer session.EndSession(ctx)

	l.Debug().CallerSkipFrame(1).Msg("Starting transaction")

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		var ctx context.Context = sessCtx

		for _, hook := range transactionHooks {
			ctx = hook(ctx)
		}

		start := time.Now()

		err := fn(ctx)
		if err != nil {
			l.Debug().Msg("Transaction func returned error, aborting")

			return nil, err
		}

		l.Debug().Msgf("Transaction complete in %s, committing", time.Since(start))

		err = sessCtx.CommitTransaction(sessCtx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return nil, nil
	})

	return err
}
