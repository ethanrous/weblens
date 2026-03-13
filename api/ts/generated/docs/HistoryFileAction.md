# HistoryFileAction


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**actionType** | **string** |  | [optional] [default to undefined]
**contentID** | **string** |  | [optional] [default to undefined]
**destinationPath** | [**WlfsFilepath**](WlfsFilepath.md) |  | [optional] [default to undefined]
**doer** | **string** | The user or system that performed the action | [optional] [default to undefined]
**eventID** | **string** |  | [optional] [default to undefined]
**fileID** | **string** |  | [optional] [default to undefined]
**filepath** | [**WlfsFilepath**](WlfsFilepath.md) |  | [optional] [default to undefined]
**id** | **string** |  | [optional] [default to undefined]
**oldFileID** | **string** | Used for restore actions to reference the file being restored | [optional] [default to undefined]
**originPath** | [**WlfsFilepath**](WlfsFilepath.md) |  | [optional] [default to undefined]
**size** | **number** |  | [optional] [default to undefined]
**timestamp** | **string** |  | [optional] [default to undefined]
**towerID** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { HistoryFileAction } from './api';

const instance: HistoryFileAction = {
    actionType,
    contentID,
    destinationPath,
    doer,
    eventID,
    fileID,
    filepath,
    id,
    oldFileID,
    originPath,
    size,
    timestamp,
    towerID,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
