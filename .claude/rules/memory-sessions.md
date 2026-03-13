# Memory: Sessions

Rolling summary of recent work. Keep last ~5 sessions, remove older ones.

## 2026-03-11

- File history panel: grouped actions by `eventID` for compact display
  - Created `FileEventGroup.vue` — collapsed header with chevron expand for multi-action events, single actions render inline
  - Created `ActionIcon.vue` — renders correct icon per action type
  - Refactored `FileAction.vue` — compact single-line layout with `compact` prop, removed fixed `h-20` height
  - Updated `FileHistory.vue` — groups flat `FileActionInfo[]` by `eventID` into `EventGroup[]` via computed, renders `FileEventGroup` instead of individual actions

## 2026-03-10

- Implemented file restore from history (full stack):
  - Backend: `NewRestoreAction` in `models/history/file_action.go`, `RestoreFiles` service method in `services/file/file_service.go` (BFS queue, hard-links from RESTORE tree, handles dirs recursively), REST handler in `routers/api/v1/file/rest_files.go`
  - Fixed `journal.GetPastFileByID` to include `ContentID` and `Size` from history action, and bypass path-based lookup to avoid ambiguity when multiple files existed at same path
  - Fixed swagger annotations ("structsore" → "restore"), ran `make swag` to regenerate
  - Frontend: Added "Restore" button to `ContextMenuActions.vue` (visible when `isViewingPast`), calls generated API's `restoreFiles` method
  - Tests: 3 integration tests (single file, directory with children, name conflict)
  - All 115 file service tests pass

## 2026-03-07

- Migrated tag API from hand-written `TagApi.ts` to generated `@ethanrous/weblens-api` client
  - Added swagger annotations to all 9 tag handlers in `rest_tags.go`
  - Ran `make swag` to regenerate swagger.json, TypeScript client, Go client
  - Added `TagsAPI` to `api/ts/AllApi.ts` WLAPI type
  - Deleted `api/TagApi.ts`, rewrote `stores/tags.ts` to use `useWeblensAPI().TagsAPI.*`
  - Updated all consumers (`TagPill.vue`, `TagManager.vue`, `TagSelector.vue`, `FileSearchFilters.vue`, `PresentationFileInfo.vue`, `[tagID].vue`)
  - Fixed `stores/files.ts` renamed types from swag regeneration (`SearchByFilename*` → `SearchFiles*`)
  - Re-exported generated type as `TagInfo` from `stores/tags.ts` for convenience

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
