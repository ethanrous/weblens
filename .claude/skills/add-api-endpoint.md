---
name: add-api-endpoint
description: Add a new REST API endpoint to the weblens Go backend, including route registration, swagger annotations, and generated clients
---

# Add API Endpoint

Follow this checklist when adding a new API endpoint to the weblens backend.

## 1. Determine the route group

Look at `routers/api/v1/api.go` to find the correct route group (media, files, tags, etc.) or create a new sub-package under `routers/api/v1/`.

Route guards available in `routers/api/v1/api.go`:
- `router.RequireSignIn` — authenticated user
- `router.RequireAdmin` — admin user
- `router.RequireOwner` — file/resource owner
- `router.RequireCoreTower` — core server only

## 2. Write the handler

Create or edit a handler file in the appropriate `routers/api/v1/<domain>/` package.

Handler signature pattern — handlers receive `context_service.RequestContext`:

```go
func HandlerName(ctx context_service.RequestContext) {
    // Access user: ctx.Requester
    // Access share: ctx.Share
    // Services via ctx: ctx.FileService(), ctx.MediaService(), etc.
    // Write response: ctx.Ok(data), ctx.Error(status, err)
}
```

Request body parsing:
```go
var params myParams
if err := ctx.ReadBody(&params); err != nil {
    ctx.Error(http.StatusBadRequest, err)
    return
}
```

URL parameters:
```go
fileID := ctx.Param("fileID")
```

## 3. Add Swagger annotations (MANDATORY)

Every handler MUST have complete Swagger annotations. Example:

```go
// HandlerName godoc
//
//  @ID         HandlerName
//  @Security   SessionAuth
//  @Summary    One-line description
//  @Tags       DomainTag
//  @Accept     json
//  @Produce    json
//  @Param      paramName  path    string          true  "Description"
//  @Param      body       body    requestStruct   true  "Description"
//  @Success    200        {object} responseType    "Description"
//  @Failure    400
//  @Failure    401
//  @Failure    500
//  @Router     /route/path [method]
```

## 4. Register the route

In `routers/api/v1/api.go`, add the route within the appropriate group:

```go
r.Get("/path", router.RequireSignIn, handler_api.HandlerName)
```

## 5. Regenerate Swagger and clients

```bash
make swag
```

This runs `scripts/swaggo.bash` which:
1. Generates `docs/swagger.json` and `docs/swagger.yaml`
2. Generates TypeScript client at `api/ts/generated/`
3. Generates Go client at `api/` (used by e2e tests)

## 6. Add e2e test

Add a test in the appropriate `e2e/rest_*_test.go` file using the generated Go API client:

```go
func TestHandlerName(t *testing.T) {
    setup, err := setupTestServer(t.Context(), t.Name(), config.Provider{...})
    require.NoError(t, err)
    client := getAPIClientFromConfig(setup.cnf, setup.token)

    // Use client.DomainAPI.HandlerName(setup.ctx).Execute()
}
```

## Files typically changed

- `routers/api/v1/<domain>/rest_<domain>.go` — handler
- `routers/api/v1/api.go` — route registration
- `docs/` — auto-generated swagger (via `make swag`)
- `api/` — auto-generated Go client (via `make swag`)
- `api/ts/generated/` — auto-generated TS client (via `make swag`)
- `e2e/rest_<domain>_test.go` — integration test
