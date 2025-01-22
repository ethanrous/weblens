package werror

import "errors"

var ErrServerNotInitialized = ClientSafeErr{
	realError:  errors.New("server not initialized"),
	statusCode: 503,
}

var ErrServerIsBackup = errors.New("this operation cannot be performed on a backup server")
var ErrNoServerId = errors.New("server must have an id")
var ErrNoCreator = errors.New("instances must have a creator id")
var ErrNoServerName = errors.New("no server name specified")
var ErrNoServerKey = errors.New("server has no key specified")
var ErrDuplicateLocalServer = errors.New("cannot add local server to instance service, it must already exist")
var ErrDuplicateCoreServer = errors.New("cannot add more than one core server to instance service")
var ErrNoCoreAddress = errors.New("core server cannot be added with no address")
var ErrNoLocal = errors.New("could not get local server")

var ErrNoInstance = errors.New("instance not found")
