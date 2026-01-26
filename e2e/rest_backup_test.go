package e2e_test

// Note: Backup API tests are skipped due to:
//
// LaunchBackup (POST /tower/{serverID}/backup) requires either:
//   - A backup server connected to the core server (multi-server setup)
//   - Or testing only the error case (server ID not found)
//
// The multi-server setup is complex and requires:
//   1. Starting a core server
//   2. Starting a backup server
//   3. Connecting them via AttachRemote
//   4. Then testing the backup launch
//
// This infrastructure is not available in the standard e2e test setup.
// See e2e/backup_test.go for integration tests that may cover this functionality.
