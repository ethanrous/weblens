package types

import (
	"time"

	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccessService interface {
	WeblensService[WeblensApiKey, ApiKeyInfo, AccessStore]
	GenerateApiKey(AccessMeta) (ApiKeyInfo, error)
	GetAllKeys(AccessMeta) ([]ApiKeyInfo, error)
}

type Instance interface {
	ServerId() InstanceId
	SetServerId(InstanceId)
	GetName() string
	GetUsingKey() WeblensApiKey
	SetUsingKey(WeblensApiKey)
	ServerRole() ServerRole
	GetAddress() (string, error)
	SetAddress(address string) error
	Info() Instance
	IsLocal() bool
	IsCore() bool
}

var ErrAlreadyCore = error2.NewWeblensError("core server cannot have a remote core")
