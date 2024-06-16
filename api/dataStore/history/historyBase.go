package history

import "github.com/ethrousseau/weblens/api/types"

var hc historyControllers

type historyControllers struct {
	fileTree types.FileTree
	dbServer types.DatabaseService
}

func SetHistoryControllers(fileTree types.FileTree, dbServer types.DatabaseService) {
	hc = historyControllers{
		fileTree: fileTree,
		dbServer: dbServer,
	}
}
