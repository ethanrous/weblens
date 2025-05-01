package history

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	db.InstallTransactionHook(WithFileEvent)
}

type FileEventContextKey struct{}

type FileEvent struct {
	EventId   string
	StartTime time.Time
}

func WithFileEvent(ctx context.Context) context.Context {
	e := FileEvent{
		EventId:   primitive.NewObjectID().Hex(),
		StartTime: time.Now(),
	}

	return context.WithValue(ctx, FileEventContextKey{}, e)
}

func FileEventFromContext(ctx context.Context) (FileEvent, bool) {
	if ctx == nil {
		return FileEvent{}, false
	}

	fe, ok := ctx.Value(FileEventContextKey{}).(FileEvent)
	if !ok {
		return FileEvent{}, false
	}

	return fe, true
}
