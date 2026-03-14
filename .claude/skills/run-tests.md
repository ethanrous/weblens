---
name: run-tests
description: Run Go backend tests, frontend Playwright tests, or coverage for the weblens project
---

# Run Tests

## Go backend tests

Run all:
```bash
make test-server
```

Run a single package:
```bash
./scripts/test-weblens.bash ./path/to/package/...
```

Examples:
```bash
./scripts/test-weblens.bash ./services/file/...
./scripts/test-weblens.bash ./models/tag/...
./scripts/test-weblens.bash ./e2e/...
```

What `test-weblens.bash` does:
1. Ensures MongoDB test instance is running (port 27019, replica set)
2. Builds `libagno.a` if needed (lazy mode skips if present)
3. Runs `gotestsum` with `-race`, `-cover`, `-tags=test`, `-timeout=1m`
4. Coverage output goes to `_build/cover/coverage.out`

### Flags

- `--no-lazy` — rebuild mongo and agno even if already running
- `-c` / `--containerize` — run tests inside Docker container

### Environment

Key env vars set by the script:
- `WEBLENS_MONGODB_URI` — defaults to `mongodb://127.0.0.1:27019/?replicaSet=rs0&directConnection=true`
- `WEBLENS_DO_CACHE=false`
- `WEBLENS_LOG_LEVEL=debug`

### Prerequisites

- MongoDB running (script handles this automatically)
- `libagno.a` built (script handles this in non-lazy mode)
- `gotestsum` installed (script installs via `go install`)

## Frontend Playwright tests

Run all:
```bash
make test-ui
```

Run specific test:
```bash
./scripts/test-playwright.bash --grep 'test name here'
```

## Coverage

Text summary:
```bash
make cover
```

HTML report:
```bash
make cover-view
```

## Important notes

- Go tests use `-tags=test` build tag — use `//go:build test` for test-only code
- Go tests use separate `_test` package (e.g., `package file_test`)
- E2E tests in `e2e/` spin up real server instances with isolated databases
- Frontend tests need a running dev server (`make dev`)
