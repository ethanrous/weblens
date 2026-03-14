# Architecture

## Backend Layers

- **`cmd/weblens/`** - Entry point
- **`routers/`** - HTTP routing (chi), middleware (auth, CORS, logging), request context
- **`services/`** - Business logic (`file/`, `media/`, `jobs/`, `auth/`, `notify/`, `journal/`, `tower/`)
- **`models/`** - Data models, interfaces, DB operations (`file/`, `media/`, `usermodel/`, `task/`, `share/`, `history/`, `tower/`)
- **`modules/`** - Pure utilities (`config/`, `startup/`, `wlfs/`, `log/`)
- **`e2e/`** - Integration tests using full server stack with generated API client

Each layer depends only on the one below it. `routers` depend on `services`, which depend on `models` and `modules`. No circular dependencies.

## Key Patterns

**DI container**: `AppContext` (`services/ctxservice/`) holds all services. `RequestContext` extends it with per-request data (user, share, HTTP req/res).

**Startup hooks**: Services self-register via `init()` + `startup.RegisterHook()` in `services/*/init.go`. Hooks run sequentially; `ErrDeferStartup` retries later.

**Task system**: `models/task/WorkerPool` runs background jobs (scan, upload, zip, backup). Jobs registered in `services/jobs/`. Progress broadcast via WebSocket.

**Portable paths**: Files use alias-based paths (`root:relative/path`) via `modules/wlfs/Filepath` for cross-system backup/restore.

**Route guards**: `RequireSignIn`, `RequireAdmin`, `RequireOwner`, `RequireCoreTower` in `routers/api/v1/api.go`.

## Frontend (`weblens-vue/weblens-nuxt/`)

Nuxt 4 SPA (SSR disabled). Pinia stores in `stores/`. Atomic design components (`atom/`, `molecule/`, `organism/`). TypeScript API client at `api/ts/` generated from openapi spec (`@ethanrous/weblens-api`).

## Rust Image Library (`agno/`)

GPU-accelerated image processing (wgpu + SPIR-V, CPU fallback). Linked as `libagno.a` via CGO. See `agno/CLAUDE.md`.
Imported as a git submodule at `agno/` with its own CLAUDE rules.

## Generated Code

- **`api/`** - Go API client from Swagger (used by e2e tests)
- **`api/ts/`** - TypeScript API client
- **`docs/`** - Swagger docs via `swaggo/swag`

All API route handlers must have complete Swagger annotations for parameters, request body, and responses to ensure accurate documentation and client generation.
When making changes to API routes, update Swagger annotations in `routers/api/v1/` and run `make swag` to regenerate docs and clients.

## Configuration

All env vars with `WEBLENS_` prefix (optional `.env`). Key: `WEBLENS_MONGODB_URI`, `WEBLENS_DATA_PATH`, `WEBLENS_CACHE_PATH`, `WEBLENS_PORT`, `WEBLENS_LOG_LEVEL`, `WEBLENS_INIT_ROLE`.
