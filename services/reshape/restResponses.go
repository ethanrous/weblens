package reshape

// FIXME: This file is a mess, clean it up, move all of these structs to the structs package and all of the functions
// to their own files.

// func ApiKeyToApiKeyInfo(k auth.Token) structs.ApiKeyInfo {
// 	return structs.ApiKeyInfo{
// 		ID:           k.ID.Hex(),
// 		Name:         k.Nickname,
// 		Key:          string(k.Token[:]),
// 		Owner:        k.Owner,
// 		CreatedTime:  k.CreatedTime.UnixMilli(),
// 		RemoteUsing:  k.RemoteUsing,
// 		CreatedBy:    k.CreatedBy,
// 		LastUsedTime: k.LastUsed.UnixMilli(),
// 	}
// }
