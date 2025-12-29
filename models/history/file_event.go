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

// fileEventContextKey is a context key for storing FileEvent values.
type fileEventContextKey struct{}

// FileEvent represents a file operation event with timing and user information.
type FileEvent struct {
	EventID   string
	StartTime time.Time
	Doer      string
}

// WithFileEvent creates a new FileEvent and adds it to the context.
// It extracts the doer from the context or defaults to UnknownUserName if not found.
func WithFileEvent(ctx context.Context) context.Context {
	doer, ok := ctx.Value(context_mod.RequestDoerKey).(string)
	if !ok || doer == "" {
		doer = user_model.UnknownUserName
	}

	e := FileEvent{
		EventID:   primitive.NewObjectID().Hex(),
		StartTime: time.Now(),
		Doer:      doer,
	}

	return context.WithValue(ctx, fileEventContextKey{}, e)
}

// FileEventFromContext retrieves the FileEvent from the context if it exists.
// Returns the FileEvent and true if found, or an empty FileEvent and false otherwise.
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
