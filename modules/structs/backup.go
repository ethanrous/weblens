package structs

type BackupInfo struct {
	FileHistory    []FileActionInfo
	Users          []UserInfoArchive
	Instances      []TowerInfo
	Tokens         []Token
	LifetimesCount int
}
