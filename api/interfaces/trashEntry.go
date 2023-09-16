package interfaces

type TrashEntry struct {
	OriginalPath	string				`bson:"originalPath"`
	TrashPath 		string 				`bson:"trashPath"`
}