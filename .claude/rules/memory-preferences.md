# Memory: Preferences

How the user likes things done.

- Prefers concise, direct communication
- Wants tests run via `./scripts/test-weblens.bash`, not raw `go test`
- Uses Playwright MCP for frontend debugging
- Frontend dev server at :3000, API at :8080
- Prefers fixing root causes over workarounds
- Dev login credentials: admin / adminadmin1
- **NEVER create hand-written API clients (e.g. TagApi.ts).** All frontend API calls MUST use the generated `@ethanrous/weblens-api` package. The workflow is: add swagger annotations to Go handlers → `make swag` → use `useWeblensAPI().XxxAPI.method()` in frontend code. Types must come from the generated package, not local interfaces.
- Always run `make lint` after changes to validate code standards (golangci-lint + pnpm lint)
