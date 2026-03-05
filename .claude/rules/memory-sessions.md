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
