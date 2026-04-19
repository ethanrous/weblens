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

### Avoid:

- Writing new E2E tests for backend logic that can be tested with Go unit tests. E2E tests are slower and more brittle, so reserve them for critical user flows or frontend interactions.
- Creating new tests or test files for every bug. Instead, add to existing test files that cover the relevant domain. Only create new test files if the domain is genuinely new and doesn't fit existing tests.
- Adding custom timeouts to every assertion in Playwright tests. Use the default timeout unless you have a specific reason to change it for a particular assertion. i.e. `expect(...).toBeVisible({ timeout: 10000 })` should only be used when absolutely necessary, not as a blanket approach.
- Using raw `net/http` requests in e2e tests. Always use the generated API client (`api/` package). If the client is missing a parameter (e.g., a query param or header), add it to the Swagger annotation on the endpoint in `routers/api/v1/` and run `make swag` to regenerate the client before writing the test.

## Writing tests

- Follow TDD: write the test BEFORE any fix. Do not skip to implementation.
- Add to existing test files whenever possible (e.g., `file_test.go` for file-related bugs). Only create new test files if the domain is genuinely new.
- Reuse setup code and test helpers from existing tests. Don't write new boilerplate if you can extend an existing test.
- If you uncover a bug while writing a test, stop and write the test first. Don't fix the bug before you have a failing test that captures it.
- NEVER write a test that passes when you know the code is currently broken. If you find a bug unrelated to your current task, ask the user how to proceed — you may need to write a separate test for the new bug, or it may be out of scope.
- Do not write comments in the test code explaining the bug you are trying to fix. Comments explaining the test logic are fine, but comments like "CRITICAL: " or "BUG: " that describe the issue are not — the test itself should capture the behavior, and instead the commit message will describe the bug and fix.

## What NOT to Test

- Constant definitions or their values
- Default struct zero values
- Simple struct field assignments
- Error variable existence or message strings
- Private/unexported functions

## Running Tests

- When running tests, ALWAYS redirect the output to a tmp file for easier debugging. For example:

```bash
./scripts/test-weblens.bash ./path/to/package/... > /tmp/test_output.txt 2>&1
```

This allows you to easily search or grep for specific error messages, stack traces, or test names in the output, without having to re-run the tests multiple times.
