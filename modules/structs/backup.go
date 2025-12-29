// Package structs provides data transfer objects and API request/response structures.
package structs

// BackupInfo represents the complete state of a backup archive.
type BackupInfo struct {
	FileHistory    []FileActionInfo
	Users          []UserInfoArchive
	Instances      []TowerInfo
	Tokens         []TokenInfo
	LifetimesCount int
} // @name BackupInfo
