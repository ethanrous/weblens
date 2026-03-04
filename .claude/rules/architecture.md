# Architecture

## Backend Layers

- **`cmd/weblens/`** - Entry point
- **`routers/`** - HTTP routing (chi), middleware (auth, CORS, logging), request context
- **`services/`** - Business logic (`file/`, `media/`, `jobs/`, `auth/`, `notify/`, `journal/`, `tower/`)
- **`models/`** - Data models, interfaces, DB operations (`file/`, `media/`, `usermodel/`, `task/`, `share/`, `history/`, `tower/`)
- **`modules/`** - Pure utilities (`config/`, `startup/`, `wlfs/`, `log/`)
- **`e2e/`** - Integration tests using full server stack with generated API client

## Key Patterns

**DI container**: `AppContext` (`services/ctxservice/`) holds all services. `RequestContext` extends it with per-request data (user, share, HTTP req/res).

**Startup hooks**: Services self-register via `init()` + `startup.RegisterHook()` in `services/*/init.go`. Hooks run sequentially; `ErrDeferStartup` retries later.

**Task system**: `models/task/WorkerPool` runs background jobs (scan, upload, zip, backup). Jobs registered in `services/jobs/`. Progress broadcast via WebSocket.

**Portable paths**: Files use alias-based paths (`root:relative/path`) via `modules/wlfs/Filepath` for cross-system backup/restore.

**Route guards**: `RequireSignIn`, `RequireAdmin`, `RequireOwner`, `RequireCoreTower` in `routers/api/v1/api.go`.

## Frontend (`weblens-vue/weblens-nuxt/`)

Nuxt 3 SPA (SSR disabled). Pinia stores in `stores/`. Atomic design components (`atom/`, `molecule/`, `organism/`). TypeScript API client at `api/ts/` (`@ethanrous/weblens-api`).

## Rust Image Library (`agno/`)

GPU-accelerated image processing (wgpu + SPIR-V, CPU fallback). Linked as `libagno.a` via CGO. See `agno/CLAUDE.md`.

## Generated Code

- **`api/`** - Go API client from Swagger (used by e2e tests)
- **`api/ts/`** - TypeScript API client
- **`docs/`** - Swagger docs via `swaggo/swag`

## Configuration

All env vars with `WEBLENS_` prefix (optional `.env`). Key: `WEBLENS_MONGODB_URI`, `WEBLENS_DATA_PATH`, `WEBLENS_CACHE_PATH`, `WEBLENS_PORT`, `WEBLENS_LOG_LEVEL`, `WEBLENS_INIT_ROLE`.
