# Testing Rules (Weblens-specific)

## Go Tests

- Use separate `_test` package (e.g., `package file_test`), not the main package
- Focus on behavior and logic, not constant definitions or struct assignments
- Always run via `./scripts/test-weblens.bash` (sets up MongoDB, env vars, build tags)
- Single package: `./scripts/test-weblens.bash ./path/to/package/...`
- Tests use `-tags=test`, `-race`, and `-cover`
- Check coverage: `make cover` (text) or `make cover-view` (HTML)

## E2E Tests

- Go integration tests in `/e2e/` spin up real server instances with isolated DBs and dynamic ports
- Frontend E2E: `make test-ui` (Playwright via `./scripts/test-playwright.bash`)

## What NOT to Test

- Constant definitions or their values
- Default struct zero values
- Simple struct field assignments
- Error variable existence or message strings
- Private/unexported functions
