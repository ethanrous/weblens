package db

import (
	"context"

	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/pkg/errors"
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

	l := context_mod.ToZ(ctx)

	if hasTransaction(ctx) {
		l.Log().Debug().Msg("Already in a transaction, skipping straight to callback function")

		return fn(ctx)
	}

	session, err := db.Client().StartSession()
	if err != nil {
		return errors.WithStack(err)
	}

	defer session.EndSession(ctx)

	l.Log().Debug().Msg("Starting transaction")

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		var ctx context.Context = sessCtx

		for _, hook := range transactionHooks {
			ctx = hook(ctx)
		}

		err := fn(ctx)
		if err != nil {
			l.Log().Debug().Msg("Transaction func returned error, aborting")

			return nil, err
		}

		l.Log().Debug().Msg("Transaction complete, committing")

		err = sessCtx.CommitTransaction(sessCtx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return nil, nil
	})

	return err
}
