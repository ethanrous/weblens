# TowerInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**backupSize** | **number** |  | [default to undefined]
**coreAddress** | **string** | Address of the remote server, only if the instance is a core. Not set for any remotes/backups on core server, as it IS the core | [default to undefined]
**id** | **string** |  | [default to undefined]
**lastBackup** | **number** |  | [default to undefined]
**logLevel** | **string** |  | [optional] [default to undefined]
**name** | **string** |  | [default to undefined]
**online** | **boolean** |  | [default to undefined]
**reportedRole** | **string** | Role the server is currently reporting. This is used to determine if the server is online (and functional) or not | [default to undefined]
**role** | **string** | Core or Backup | [default to undefined]
**started** | **boolean** |  | [default to undefined]
**userCount** | **number** |  | [default to undefined]

## Example

```typescript
import { TowerInfo } from './api';

const instance: TowerInfo = {
    backupSize,
    coreAddress,
    id,
    lastBackup,
    logLevel,
    name,
    online,
    reportedRole,
    role,
    started,
    userCount,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
