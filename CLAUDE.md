## Project

Weblens — self-hosted file management and photo server.

- Go backend (chi, MongoDB)
- Vue 3/Nuxt 4 + Tailwind SPA frontend at `./weblens-vue/weblens-nuxt/`
- Rust image processing library (`agno/`) linked via CGO.

## Quick Reference

```bash
make dev                                                # Dev server (air + nuxt dev)
make lint                                        		# golangci-lint + pnpm lint
make swag                                        		# Regenerate Swagger docs and typescript sdk
make test-server                                 		# All Go (backend) tests
./scripts/test-weblens.bash ./path/to/package/... 		# Single package tests
make test-ui                                     		# Playwright E2E
./scripts/test-playwright.bash --grep 'test name here' 	# Single playwright E2E test
make cover                                       		# Coverage report (text)
make agno                                        		# Build Rust libagno.a
make gen-ui                                      		# Build frontend
```

Dev servers: frontend `:3000` (proxies API), backend `:8080`.

## Debugging workflow:

ALWAYS use the `debug-fix` skill for any bug, test failure, or unexpected behavior. It automatically routes to the correct agents and orchestrates the full diagnose-then-fix pipeline.

**Debug agents** (opus — diagnosis only, no implementation):

- `debug-backend` — Go backend issues. Uses MongoDB MCP, structured logging, pprof. Writes a failing test and reports root cause.
- `debug-frontend` — Vue/Nuxt/Tailwind issues. Uses Playwright MCP to interact with live dev server. Writes a failing test and reports root cause.

**Fix agents** (sonnet — implementation only, requires diagnosed root cause):

- `fix-backend` — Implements Go backend fixes following project conventions (layers, DI, error handling, Swagger).
- `fix-frontend` — Implements Nuxt frontend fixes following project conventions (Pinia, atomic design, Tailwind, Playwright tests).

The `debug-fix` skill launches the appropriate debug agent first, then passes the root cause to the matching fix agent. Do not use fix agents without a diagnosis — they expect a root cause and failing test as input.

## Development Workflow: Test-Driven Development (MANDATORY)

Every bug fix and feature MUST follow this sequence. Do not skip steps.

1. **Understand** — Read the relevant code. Use plan mode for non-trivial work. Identify the root cause (bug) or the exact behavior change (feature).
2. **Design the solution** — Decide what to change, but **do NOT write implementation code yet**.
3. **Write the test first** — Add a test that captures the expected behavior. Extend an existing test file whenever possible (see `.claude/skills/write-backend-test.md` and `.claude/skills/write-playwright-test.md`). Do not create new test files unless the domain is genuinely new.
4. **Run the test — watch it fail** — Confirm the test fails for the right reason (not a syntax error or import issue). This validates the test actually tests something.
5. **Implement the fix/feature** — Write the minimum code to make the test pass.
6. **Run the test — watch it pass** — If it fails, fix the implementation, not the test (unless the test itself was wrong).
7. **Run the full relevant test suite** — `./scripts/test-weblens.bash ./path/to/package/...` for backend, `make test-ui` or `./scripts/test-playwright.bash --grep '...'` for frontend.

**Why this order matters:** Writing the test first forces you to define the expected behavior precisely before touching implementation code. It catches regressions, prevents over-engineering, and proves the fix actually works. Skipping to implementation and writing tests after is not TDD — it's rationalization.

## Detailed Specs (auto-loaded from `.claude/rules/`)

| File              | Contents                                                                   |
| ----------------- | -------------------------------------------------------------------------- |
| `architecture.md` | Backend layers, DI pattern, startup hooks, task system, frontend structure |
| `testing.md`      | Test conventions, what to test, what NOT to test, coverage                 |
