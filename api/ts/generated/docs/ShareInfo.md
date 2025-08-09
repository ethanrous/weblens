# ShareInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**accessors** | [**Array&lt;UserInfo&gt;**](UserInfo.md) |  | [optional] [default to undefined]
**enabled** | **boolean** |  | [optional] [default to undefined]
**expires** | **number** |  | [optional] [default to undefined]
**fileId** | **string** |  | [optional] [default to undefined]
**owner** | **string** |  | [optional] [default to undefined]
**permissions** | [**{ [key: string]: PermissionsInfo; }**](PermissionsInfo.md) |  | [optional] [default to undefined]
**_public** | **boolean** |  | [optional] [default to undefined]
**shareId** | **string** |  | [optional] [default to undefined]
**shareName** | **string** |  | [optional] [default to undefined]
**shareType** | **string** |  | [optional] [default to undefined]
**timelineOnly** | **boolean** |  | [optional] [default to undefined]
**updated** | **number** |  | [optional] [default to undefined]
**wormhole** | **boolean** |  | [optional] [default to undefined]

## Example

```typescript
import { ShareInfo } from './api';

const instance: ShareInfo = {
    accessors,
    enabled,
    expires,
    fileId,
    owner,
    permissions,
    _public,
    shareId,
    shareName,
    shareType,
    timelineOnly,
    updated,
    wormhole,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
