---
name: add-backend-feature
description: Add a new backend feature spanning models, services, routes, and tests in the weblens Go backend
---

# Add Backend Feature

Use this when adding a new domain feature (like tags, history, backup) that touches multiple backend layers.

## Layer order (build bottom-up)

1. **Models** (`models/<domain>/`)
2. **Services** (`services/<domain>/`)
3. **Routes** (`routers/api/v1/<domain>/`)
4. **E2E tests** (`e2e/`)

Each layer depends only on the one below. Never import from `routers` into `services` or from `services` into `models`.

## Step 1: Models (`models/<domain>/`)

Define data structures and DB operations.

- MongoDB collection registration in `models/db/collection.go`
- Model structs with BSON and JSON tags
- DB accessor functions that take `context_service.AppContext` or a mongo collection

Pattern for DB operations:
```go
func GetThingByID(ctx context_service.AppContext, id primitive.ObjectID) (*Thing, error) {
    col := db.GetCollection[Thing](ctx, db.ThingsCol)
    return col.FindOne(ctx, bson.M{"_id": id})
}
```

## Step 2: Services (`services/<domain>/`)

Business logic layer.

- Services are injected via `AppContext` (the DI container in `services/ctxservice/`)
- If the feature needs a persistent service, add it to `AppContext`
- Service methods take `AppContext` or `RequestContext` as first argument

## Step 3: Routes (`routers/api/v1/<domain>/`)

See the `add-api-endpoint` skill for the full handler pattern.

- Create a new package under `routers/api/v1/` for the domain
- Register routes in `routers/api/v1/api.go`
- Add complete Swagger annotations on every handler

## Step 4: E2E tests (`e2e/`)

- Create `e2e/rest_<domain>_test.go`
- Use `setupTestServer()` from `e2e/e2e_common_test.go`
- Tests spin up isolated server instances with unique DB names and ports
- Use the generated Go API client from `api/`

## Step 5: Regenerate clients

```bash
make swag
```

## Checklist

- [ ] Model structs with BSON/JSON tags
- [ ] DB collection registered in `models/db/collection.go`
- [ ] Service methods on `AppContext`
- [ ] Route handlers with Swagger annotations
- [ ] Routes registered in `api.go` with correct guards
- [ ] `make swag` run to regenerate clients
- [ ] E2E test covering happy path
- [ ] `make lint` passes
