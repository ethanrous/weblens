package reshape_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/file"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/viccon/sturdyc"
	"go.mongodb.org/mongo-driver/mongo"
)

// testAppContextOptions configures the test AppContext.
type testAppContextOptions struct {
	onlineUsers  map[string]bool
	onlineTowers map[string]bool
	fileService  file.Service
}

// testAppContextOption is a functional option for configuring test AppContext.
type testAppContextOption func(*testAppContextOptions)

// withOnlineUser marks a user as online in the test context.
func withOnlineUser(username string) testAppContextOption {
	return func(opts *testAppContextOptions) {
		opts.onlineUsers[username] = true
	}
}

// withOnlineTower marks a tower as online in the test context.
func withOnlineTower(towerID string) testAppContextOption {
	return func(opts *testAppContextOptions) {
		opts.onlineTowers[towerID] = true
	}
}

// withFileService sets a custom file service for testing.
func withFileService(fs file.Service) testAppContextOption {
	return func(opts *testAppContextOptions) {
		opts.fileService = fs
	}
}

// newTestAppContext creates an AppContext suitable for testing.
// It uses a mock ClientManager and optionally connects to the test database.
func newTestAppContext(t *testing.T, opts ...testAppContextOption) context.Context {
	t.Helper()

	// Apply options
	options := &testAppContextOptions{
		onlineUsers:  make(map[string]bool),
		onlineTowers: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(options)
	}

	// Create a logger for the test
	logger := wlog.NewZeroLogger()

	// Create basic context with logger
	basicCtx := ctxservice.NewBasicContext(context.Background(), logger)

	// Create AppContext with mock services
	appCtx := ctxservice.AppContext{
		BasicContext:  basicCtx,
		ClientService: nil,
		FileService:   options.fileService,
		Cache:         make(map[string]*sturdyc.Client[any]),
		WG:            &sync.WaitGroup{},
	}

	// Return as a context that can be extracted via FromContext
	return appCtx.WithContext(context.Background())
}

// newTestAppContextWithDB creates an AppContext with a real test database connection.
// Useful for integration tests that need database access.
func newTestAppContextWithDB(t *testing.T, opts ...testAppContextOption) context.Context {
	t.Helper()

	// Set up test database
	dbCtx := db.SetupTestDB(t, user_model.UserCollectionKey)

	// Apply options
	options := &testAppContextOptions{
		onlineUsers:  make(map[string]bool),
		onlineTowers: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(options)
	}

	// Get the database from the context
	dbAny := dbCtx.Value(db.DatabaseContextKey)
	database, _ := dbAny.(*mongo.Database)

	// Create a logger for the test
	logger := wlog.NewZeroLogger()

	// Create basic context with logger
	basicCtx := ctxservice.NewBasicContext(dbCtx, logger)

	// Create AppContext with mock services and real DB
	appCtx := ctxservice.AppContext{
		BasicContext:  basicCtx,
		ClientService: nil,
		FileService:   options.fileService,
		DB:            database,
		Cache:         make(map[string]*sturdyc.Client[any]),
		WG:            &sync.WaitGroup{},
	}

	// Return as a context
	return appCtx.WithContext(dbCtx)
}

// createTestUser creates a user in the test database.
// The password meets validation requirements (6+ chars with a digit).
func createTestUser(ctx context.Context, t *testing.T, username string) *user_model.User {
	t.Helper()

	u := &user_model.User{
		Username:  username,
		Password:  "testpass1", // Meets: 6+ chars with digit
		Activated: true,
	}

	err := user_model.SaveUser(ctx, u)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return u
}

// newTestRequestContextWithDB creates a RequestContext with a real test database connection.
// This is useful for testing functions that take a RequestContext parameter.
func newTestRequestContextWithDB(t *testing.T, opts ...testAppContextOption) ctxservice.RequestContext {
	t.Helper()

	// Set up test database
	dbCtx := db.SetupTestDB(t, user_model.UserCollectionKey)

	// Apply options
	options := &testAppContextOptions{
		onlineUsers:  make(map[string]bool),
		onlineTowers: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(options)
	}

	// Get the database from the context
	dbAny := dbCtx.Value(db.DatabaseContextKey)
	database, _ := dbAny.(*mongo.Database)

	// Create a logger for the test
	logger := wlog.NewZeroLogger()

	// Create basic context with logger
	basicCtx := ctxservice.NewBasicContext(dbCtx, logger)

	// Create AppContext with mock services and real DB
	appCtx := ctxservice.AppContext{
		BasicContext:  basicCtx,
		ClientService: nil,
		FileService:   options.fileService,
		DB:            database,
		Cache:         make(map[string]*sturdyc.Client[any]),
		WG:            &sync.WaitGroup{},
	}

	// Create RequestContext embedding the AppContext
	reqCtx := ctxservice.RequestContext{
		AppContext: appCtx,
		ReqCtx:     dbCtx,
	}

	return reqCtx
}
