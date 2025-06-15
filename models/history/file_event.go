package history

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	db.InstallTransactionHook(WithFileEvent)
}

type fileEventContextKey struct{}

type FileEvent struct {
	EventId   string
	StartTime time.Time
	Doer      string
}

func WithFileEvent(ctx context.Context) context.Context {
	doer, ok := ctx.Value(context_mod.RequestDoerKey).(string)
	if !ok || doer == "" {
		doer = user_model.UnknownUserName
	}

	e := FileEvent{
		EventId:   primitive.NewObjectID().Hex(),
		StartTime: time.Now(),
		Doer:      doer,
	}

	return context.WithValue(ctx, fileEventContextKey{}, e)
}

func FileEventFromContext(ctx context.Context) (FileEvent, bool) {
	if ctx == nil {
		return FileEvent{}, false
	}

	fe, ok := ctx.Value(fileEventContextKey{}).(FileEvent)
	if !ok {
		return FileEvent{}, false
	}

	return fe, true
}
