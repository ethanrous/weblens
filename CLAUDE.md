# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Weblens — self-hosted file management and photo server. Go backend (chi, MongoDB), Nuxt 4 SPA frontend, Rust image processing library (`agno/`) linked via CGO. Two server roles: **core** and **backup**.

## Quick Reference

```bash
make dev                                         # Dev server (air + nuxt dev)
make lint                                        # golangci-lint + pnpm lint
make swag                                        # Regenerate Swagger docs
./scripts/test-weblens.bash                      # All Go tests
./scripts/test-weblens.bash ./path/to/package/... # Single package tests
make test-ui                                     # Playwright E2E
make cover                                       # Coverage report (text)
make agno                                        # Build Rust libagno.a
make gen-ui                                      # Build frontend
```

Dev servers: frontend `:3000` (proxies API), backend `:8080`.

## Detailed Specs (auto-loaded from `.claude/rules/`)

| File                    | Contents                                                                   |
| ----------------------- | -------------------------------------------------------------------------- |
| `architecture.md`       | Backend layers, DI pattern, startup hooks, task system, frontend structure |
| `testing.md`            | Test conventions, what to test, what NOT to test, coverage                 |
| `memory-profile.md`     | Facts about the user                                                       |
| `memory-preferences.md` | How the user likes things done                                             |
| `memory-decisions.md`   | Past decisions for consistency                                             |
| `memory-sessions.md`    | Rolling summary of recent work                                             |

## Auto-Update Memory (MANDATORY)

**Update memory files AS YOU GO, not at the end.** When you learn something new, update immediately.

| Trigger                     | Action                                 |
| --------------------------- | -------------------------------------- |
| User states a preference    | Update `memory-preferences.md`         |
| A decision is made          | Update `memory-decisions.md` with date |
| Completing substantive work | Add to `memory-sessions.md`            |

**Skip:** Quick factual questions, trivial tasks with no new info.

**DO NOT ASK. Just update the files when you learn something.**
