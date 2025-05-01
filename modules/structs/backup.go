package structs

type BackupInfo struct {
	FileHistory    []FileActionInfo
	Users          []UserInfoArchive
	Instances      []TowerInfo
	Tokens         []TokenInfo
	LifetimesCount int
} // @name BackupInfo
