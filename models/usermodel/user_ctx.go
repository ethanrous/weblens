package usermodel

import "context"

type userContextKey struct{}

// WithUser attaches a user to the context.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// FromContext retrieves the user from the context if present.
func FromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey{}).(*User)

	return user, ok
}
