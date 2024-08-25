package dataStore

import (
	"errors"
)

// Errors

var ErrNoCache = func() error {
	return errors.New("media references cache file that does not exist")
}
var ErrNoUser = errors.New("user does not exist")
var ErrNoFileAccess = errors.New("user does not have access to file")
var ErrPageOutOfRange = errors.New("page number does not exist on media")
var ErrNoImage = errors.New("media is missing required image")
var ErrKeyInUse = errors.New("api key is already being used to identify another remote server")

var ErrBadRequestMode = errors.New("access struct does not have correct request mode set for the given function")
var ErrNoShare = errors.New("no share found")
var ErrBadShareType = errors.New("expected share type does not match given share type")
var ErrUnsupportedImgType = errors.New("image type is not supported by weblens")
var ErrNoKey = errors.New("api key is does not exist")
var ErrNotCore = errors.New("tried to perform core only action on non-core server")
var ErrNotBackup = errors.New("tried to perform backup only action on non-backup server")
var ErrAlreadyInit = errors.New("server is already initialized, cannot re-initialize server")
var ErrNoBackup = errors.New("no prior backups exist")
var ErrBadJournalAction = errors.New("unknown journal action type")
