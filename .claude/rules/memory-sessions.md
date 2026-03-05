# Memory: Sessions

Rolling summary of recent work. Keep last ~5 sessions, remove older ones.

## 2026-03-04 (Security Audit)

- Comprehensive security audit of access control across the application
- Fixed CRITICAL: Hardcoded JWT signing key → loads from WEBLENS_JWT_SECRET env var
- Fixed CRITICAL: `isTakeout=true` query param bypassed all file auth → removed bypass, added CheckFileAccessByID to CreateTakeout
- Fixed HIGH: Added RequireSignIn to /files, /share, /media routes; RequireAdmin to GET /users; moved /tower/history inside admin group
- Fixed HIGH: Added CanUserModifyShare ownership check to all share mutation handlers
- Fixed HIGH: Added owner check to DeleteToken (IDOR), auth to UnTrashFiles, auth to AutocompletePath
- Fixed HIGH: Added RequireSignIn to media info/image/stream/batch/random endpoints
- Fixed MEDIUM: WebSocket CheckOrigin now validates against ProxyAddress
- Fixed MEDIUM: Added Secure and SameSite=Lax flags to session cookies
- Fixed MEDIUM: Added filename validation to CreateFile/CreateFolder/RenameFile (path traversal prevention)
- Fixed MEDIUM: GetRandomMedia now requires auth instead of cookie fallback
- Added tests: JWT key loading from env, filename validation, share ownership, cookie security flags

## 2026-03-04 (Session 2)

- Implemented file tagging feature (full stack):
  - Backend: `models/tag/` (Tag model, MongoDB CRUD, indexes), `routers/api/v1/tag/` (REST API)
  - Routes registered in `api.go`, file deletion cleanup in `rest_files.go`
  - Frontend: `stores/tags.ts` (Pinia store), `api/TagApi.ts` (API client)
  - UI: `TagSelector.vue`, `TagManager.vue`, integrated into context menu, file cards, detail sidebar, search filters, file header

## 2026-03-04

- Fixed dynamic port allocation for E2E test backend manager (`e2e/backend-manager.ts`, `e2e/fixtures.ts`)
- Fixed file download naming in `rest_media.go` and `FileBrowserApi.ts`
- Created initial CLAUDE.md for the project
