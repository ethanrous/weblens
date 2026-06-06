# Architecture

## Backend Layers

- **`cmd/weblens/`** - Entry point
- **`routers/`** - HTTP routing (chi), middleware (auth, CORS, logging), request context
- **`services/`** - Business logic (`file/`, `media/`, `jobs/`, `auth/`, `notify/`, `journal/`, `tower/`, `embed/`)
- **`models/`** - Data models, interfaces, DB operations (`file/`, `media/`, `usermodel/`, `task/`, `share/`, `history/`, `tower/`, `embedding/`)
- **`modules/`** - Pure utilities (`config/`, `startup/`, `wlfs/`, `wlog/`)
- **`embed/`** - Python multimodal-embedding sidecar (Flask + Jina CLIP v2). See "Embedding pipeline" below.
- **`e2e/`** - Integration tests using full server stack with generated API client

Each layer depends only on the one below it. `routers` depend on `services`, which depend on `models` and `modules`. No circular dependencies.

### Layer rules (REQUIRED)

**All MongoDB access lives in `models/`.** `db.GetCollection`, raw `bson.M` filters, `Aggregate`/`Find`/`UpdateMany`/`DeleteMany`/`ReplaceOne`, and Atlas search calls belong in a model package — never in `services/` or `routers/`. The model package exports CRUD/query helpers like `GetByID`, `Save`, `Delete*`, `Count*`, `Search`, etc. that hide the driver entirely. Services orchestrate by calling these helpers; routers call services.

Anti-pattern (DO NOT do this in services or routers):

```go
// in services/jobs/something.go
col, err := db.GetCollection[embedding.Embedding](ctx, embedding.CollectionKey)
n, _ := col.CountDocuments(ctx, bson.M{"sourceId": id, "kind": "image"})
```

Correct pattern:

```go
// in models/embedding/store.go
func CountForSource(ctx context.Context, kind Kind, sourceID string) (int64, error) { ... }

// in services/jobs/something.go
n, err := embedding.CountForSource(ctx, embedding.KindImage, sourceID)
```

If a new query is needed, add the helper to the model package and use it from the caller. This rule keeps DB schema knowledge out of the business-logic layer and makes the model package the single place that owns its collection.

**Test helpers** in `*_test.go` may use `db.GetCollection` for setup/teardown when no suitable model helper exists, but prefer adding the helper to the model package and using it from tests too.

## Key Patterns

**DI container**: `AppContext` (`services/ctxservice/`) holds all services. `RequestContext` extends it with per-request data (user, share, HTTP req/res).

**Startup hooks**: Services self-register via `init()` + `startup.RegisterHook()` in `services/*/init.go`. Hooks run sequentially; `ErrDeferStartup` retries later.

**Task system**: `models/task/WorkerPool` runs background jobs (scan, upload, zip, backup). Jobs registered in `services/jobs/`. Progress broadcast via WebSocket.

**Portable paths**: Files use alias-based paths (`root:relative/path`) via `modules/wlfs/Filepath` for cross-system backup/restore.

**Route guards**: `RequireSignIn`, `RequireAdmin`, `RequireOwner`, `RequireCoreTower` in `routers/api/v1/api.go`.

## Embedding pipeline

Semantic search over images and document text is powered by a separate Python sidecar.

**Container (`embed/`)**: Flask app loading a unified multimodal model (Jina CLIP v2, 1024-dim). Endpoints:
- `GET /encode?img-path=...` — image → vector
- `POST /encode-text {text}` — text query → vector
- `POST /extract-and-embed {path, mimeHint}` — extract text from a file (`extract.py` handles PDF / DOCX / XLSX / PPTX / OCR / plaintext), chunk it, return per-chunk vectors
- `GET /health`

The container image is `ethrous/weblens-embed`, built from `docker/embed.Dockerfile`. In dev mode `scripts/lib/embed.bash` runs `uv sync && uv run main.py` directly on the host (no container). `make dev` launches it via `scripts/envup.bash`.

**Go side (`services/embed/`)**: HTTP-only — no DB.
- `client.go` — `NewClient`, `EncodeImage`, `EncodeQueryText` (cached), `ExtractAndEmbedFile`, circuit-breaker via `ServiceUnavailable()`.
- `init.go` — process-wide `Default()` client, 30 s health-check ticker that clears the circuit breaker when `/health` comes back.

**Storage (`models/embedding/`)**: one collection, `embeddings`. Each row is `{kind, sourceId, chunkIndex, vector, snippet, model, contentHash, createdAt}`. `kind="image"` rows use `media.contentID` as `sourceId` with one row per page (chunkIndex = pageNum). `kind="file_chunk"` rows use `fileID` as `sourceId` with one row per text chunk. The startup hook in `models/embedding/init.go` creates the collection, registers a B-tree index on `sourceId`, a unique compound on `(kind, sourceId, chunkIndex)`, and an Atlas vector index `embeddings_vector` on `vector` (1024-dim, cosine).

**Background job**: `job.ExtractAndEmbedTask` (registered in `services/jobs/jobs.go`, handler `services/jobs/embed.go`). Dispatched from `services/jobs/upload.go` after a successful upload-finalize. Handler checks extension/size/availability gates, hashes the file, skips on hash match (idempotency), calls `/extract-and-embed`, upserts each chunk via `embedding.Upsert`, then prunes trailing rows via `embedding.PruneTrailingChunks`.

**Image embeddings**: written inline during media scan (`services/jobs/file_parser.go:writeImageEmbedding`). Iterates `media.PageCount` so multi-page documents get one vector per page (HighRes cache file); single-page media use the LowRes thumbnail. Per-page existence is checked, so partial embeddings backfill on rerun.

**Search merge**: `/files/search` (`routers/api/v1/file/rest_files_search.go:SearchFiles`) runs filename substring/regex + semantic vector search in parallel via `errgroup`. Semantic path encodes the query, calls `embedding.Search`, demuxes image hits via `media.contentID → []*file`, blends scores. Response: `[]SearchResult{file, matchKind ([]string; may contain "filename" and/or "content"), matchSnippet, matchPage, score}`. Semantic path is non-fatal — it's skipped silently when the embed service is unavailable, when `regex=true`, or when `includeContent=false`.

**Health reporting**: `/info` (`GetServerInfo`) includes `embedAvailable: bool` for the local server, sourced from `embed.Default().ServiceUnavailable()`. Up to 30 s stale (driven by the health ticker).

**Feature flag**: `embed.processing_enabled` in `feature_flags`. When false, both `writeImageEmbedding` and `ExtractAndEmbedTask` are skipped.

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

All env vars with `WEBLENS_` prefix (optional `.env`). Key: `WEBLENS_MONGODB_URI`, `WEBLENS_DATA_PATH`, `WEBLENS_CACHE_PATH`, `WEBLENS_PORT`, `WEBLENS_LOG_LEVEL`, `WEBLENS_INIT_ROLE`, `WEBLENS_EMBED_URI` (embed sidecar base URL, default `http://weblens-embed:5500`), `WEBLENS_EMBED_MAX_FILE_SIZE` (bytes; files larger are skipped by `ExtractAndEmbedTask`, default 50 MiB). `WEBLENS_HDIR_URI` is honored as a deprecated alias for `WEBLENS_EMBED_URI` for one release.
