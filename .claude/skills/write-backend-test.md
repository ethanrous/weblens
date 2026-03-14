---
name: write-backend-test
description: Write Go unit tests and e2e integration tests for the weblens backend — emphasizes extending existing test files over creating new ones
---

# Write Backend Test

## CRITICAL: Extend existing tests, don't create new files

Before writing any test, **search for an existing test file** that already covers the domain you're testing. Almost every domain already has tests — add subtests or new test functions to those files rather than creating new ones. Test slop (lots of thin, duplicative test files) makes the suite harder to maintain.

- Unit tests: check `models/<domain>/*_test.go` and `services/<domain>/*_test.go`
- E2E tests: check `e2e/rest_<domain>_test.go`

Only create a new test file if the domain is genuinely new and has no existing test file.

---

## Unit Tests

### Conventions

- **Separate `_test` package**: `package auth_test`, not `package auth`
- **No mocks**: tests use real MongoDB via `db.SetupTestDB()` — no mock libraries
- **Assertions**: `testify/assert` (non-fatal) and `testify/require` (fatal)
- **Build tag**: tests run with `-tags=test`
- **Run via script**: always use `./scripts/test-weblens.bash`, never raw `go test`

### DB setup pattern

For tests that need MongoDB, use `db.SetupTestDB()` which creates an isolated collection with automatic cleanup:

```go
package auth_test

import (
    "testing"
    "github.com/ethanrous/weblens/models/db"
    "github.com/ethanrous/weblens/models/auth"
    "github.com/stretchr/testify/assert"
)

func TestGetTokenByID(t *testing.T) {
    ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

    t.Run("found", func(t *testing.T) {
        // create test data, then assert
    })

    t.Run("not found", func(t *testing.T) {
        token, err := auth.GetTokenByID(ctx, primitive.NewObjectID())
        assert.Nil(t, token)
        assert.Equal(t, auth.ErrTokenNotFound, err)
    })
}
```

### Table-driven tests

Use for functions with multiple input/output combinations:

```go
tests := []struct {
    name          string
    ext           string
    expectMime    string
    isDisplayable bool
}{
    {"jpeg extension", "jpg", "image/jpeg", true},
    {"unknown extension", "xyz", "", false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test body using tt fields
    })
}
```

### Pure unit tests (no DB)

For utility/pure functions, skip `db.SetupTestDB()` — just call the function directly with table-driven subtests.

### Test helpers

- `db.SetupTestDB(t, collectionKey, ...indexModels)` — isolated MongoDB collection with cleanup
- `tests.Setup(t)` — test context with logger (from `modules/tests/`)
- `tests.Recover(t)` — panic recovery with stack traces (defer at top of test)

### Running

```bash
# Single package
./scripts/test-weblens.bash ./models/auth/...
./scripts/test-weblens.bash ./services/file/...

# All
make test-server

# Coverage
make cover
```

---

## E2E (Integration) Tests

E2E tests live in `e2e/` and spin up real server instances with isolated databases. They test full API flows through the HTTP layer using the generated Go API client.

### CRITICAL: Extend existing test files

The e2e directory already has test files organized by domain:

| File | Domain |
|------|--------|
| `rest_file_test.go` | File CRUD, search, upload, download |
| `rest_folder_test.go` | Folder operations |
| `rest_share_test.go` | Share CRUD, permissions |
| `rest_user_test.go` | User management |
| `rest_apikey_test.go` | API key operations |
| `rest_media_test.go` | Media endpoints |
| `rest_history_test.go` | File history |
| `rest_tower_test.go` | Server info, health |
| `rest_backup_test.go` | Backup workflow |
| `backup_test.go` | Backup integration (core + backup servers) |
| `startup_test.go` | Server startup scenarios |

**Add new test functions or subtests to the matching file.** For example, if you added a new file endpoint, add a `TestNewEndpoint` function to `rest_file_test.go` — don't create `rest_file_new_test.go`.

### Server setup

Each top-level `Test*` function creates its own isolated server. Within one function, use `t.Run()` subtests to share that server:

```go
func TestSearchFiles(t *testing.T) {
    coreSetup, err := setupTestServer(t.Context(), t.Name(),
        config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
    require.NoError(t, err)

    client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

    // Shared setup: create test data once
    userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
    require.NoError(t, err)
    homeID := userInfo.GetHomeID()

    // ... create folders, files, tags ...

    t.Run("search by name finds matching files", func(t *testing.T) {
        results, resp, err := client.FilesAPI.SearchFiles(t.Context()).Search("searchable").Execute()
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        assert.Equal(t, 2, len(results))
    })

    t.Run("empty search returns 400", func(t *testing.T) {
        _, resp, err := client.FilesAPI.SearchFiles(t.Context()).Search("").Execute()
        assert.Error(t, err)
        assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    })
}
```

**If you're adding test cases for an endpoint that already has a `Test*` function**, add `t.Run()` subtests inside that function rather than creating a separate top-level function. Each `Test*` function spins up its own server, which is expensive.

### API client pattern

The generated Go client at `api/` mirrors the Swagger spec:

```go
// GET
result, resp, err := client.FilesAPI.GetFile(ctx, fileID).Execute()

// POST with body
result, resp, err := client.TagsAPI.CreateTag(ctx).Request(openapi.TagCreateTagParams{
    Name:  openapi.PtrString("mytag"),
    Color: openapi.PtrString("#ff0000"),
}).Execute()

// PATCH with query params
resp, err := client.ShareAPI.SetSharePublic(ctx, shareID).Public(true).Execute()
```

When the generated client can't express something (e.g., raw query params, multipart upload), fall back to raw `http.NewRequestWithContext`:

```go
reqURL := fmt.Sprintf("%s/api/v1/endpoint?param=value", coreSetup.address)
req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, reqURL, nil)
req.Header.Set("Authorization", "Bearer "+coreSetup.token)
resp, err := http.DefaultClient.Do(req)
```

### Table-driven e2e subtests

For testing multiple variations of the same endpoint (different params, error cases), use table-driven subtests within a single `Test*` function:

```go
tests := []struct {
    name           string
    fileID         string
    expectedStatus int
    expectError    bool
}{
    {"basic download", uploadedFileID, http.StatusOK, false},
    {"non-existent file", "non-existent-id", http.StatusNotFound, true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        _, resp, err := client.FilesAPI.DownloadFile(t.Context(), tt.fileID).Execute()
        if tt.expectError {
            assert.Error(t, err)
        } else {
            require.NoError(t, err)
        }
        assert.Equal(t, tt.expectedStatus, resp.StatusCode)
    })
}
```

### Multi-server tests

For backup/restore flows that need core + backup:

```go
func TestBackupFlow(t *testing.T) {
    coreSetup, backupSetup := newCoreAndBackup(t)
    // Each has independent servers, databases, filesystems
}
```

### Helpers available in `e2e_common_test.go`

- `setupTestServer(ctx, name, config.Provider{...})` — isolated server instance
- `getAPIClientFromConfig(cnf, token)` — typed API client for the server
- `waitForHealthy(appCtx, cnf, token)` — poll health endpoint
- `newCoreAndBackup(t)` — core + backup server pair
- `buildTestConfig(testName, ...overrides)` — config with isolated paths

### Running

```bash
./scripts/test-weblens.bash ./e2e/...
```

### Important

- Always use `t.Context()` for cancellation
- `require.NoError` for setup steps (stop on failure), `assert` for actual checks
- The generated API client must be current — run `make swag` if endpoints changed
- Each `Test*` function gets a fresh server+DB — subtests within share that server

---

## What NOT to test

Per project rules:
- Constant definitions or their values
- Default struct zero values
- Simple struct field assignments
- Error variable existence or message strings
- Private/unexported functions
