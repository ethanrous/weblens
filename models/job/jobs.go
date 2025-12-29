// Package job defines task identifiers for various background jobs in the Weblens system.
package job

const (
	// ScanDirectoryTask is the task identifier for scanning a directory.
	ScanDirectoryTask = "scan_directory"
	// ScanFileTask is the task identifier for scanning a file.
	ScanFileTask = "scan_file"
	// MoveFileTask is the task identifier for moving a file.
	MoveFileTask = "move_file"
	// UploadFilesTask is the task identifier for uploading files.
	UploadFilesTask = "write_file"
	// CreateZipTask is the task identifier for creating a zip archive.
	CreateZipTask = "create_zip"
	// GatherFsStatsTask is the task identifier for gathering filesystem statistics.
	GatherFsStatsTask = "gather_filesystem_stats"
	// BackupTask is the task identifier for performing a backup.
	BackupTask = "do_backup"
	// HashFileTask is the task identifier for hashing a file.
	HashFileTask = "hash_file"
	// LoadFilesystemTask is the task identifier for loading the filesystem.
	LoadFilesystemTask = "load_filesystem"
	// CopyFileFromCoreTask is the task identifier for copying a file from core.
	CopyFileFromCoreTask = "copy_file_from_core"
	// RestoreCoreTask is the task identifier for restoring a core instance.
	RestoreCoreTask = "restore_core"
)
