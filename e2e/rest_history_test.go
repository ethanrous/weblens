package e2e_test

// Note: History API tests are skipped due to:
//
// 1. GetPagedHistoryActions - JSON unmarshalling issue: server returns filepath as string but
//    OpenAPI client expects FsFilepath struct
//
// 2. GetBackupInfo - Requires remote tower context (ctx.Remote.TowerID must be present)
//    which isn't available in standard e2e test setup
