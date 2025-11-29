# FolderApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createFolder**](#createfolder) | **POST** /folder | Create a new folder|
|[**getFolder**](#getfolder) | **GET** /folder/{folderId} | Get a folder|
|[**getFolderHistory**](#getfolderhistory) | **GET** /files/{fileId}/history | Get actions of a folder at a given time|
|[**scanFolder**](#scanfolder) | **POST** /folder/{folderId}/scan | Dispatch a folder scan|
|[**setFolderCover**](#setfoldercover) | **PATCH** /folder/{folderId}/cover | Set the cover image of a folder|

# **createFolder**
> FileInfo createFolder(request)


### Example

```typescript
import {
    FolderApi,
    Configuration,
    CreateFolderBody
} from './api';

const configuration = new Configuration();
const apiInstance = new FolderApi(configuration);

let request: CreateFolderBody; //New folder body
let shareId: string; //Share Id (optional) (default to undefined)

const { status, data } = await apiInstance.createFolder(
    request,
    shareId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateFolderBody**| New folder body | |
| **shareId** | [**string**] | Share Id | (optional) defaults to undefined|


### Return type

**FileInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File Info |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFolder**
> FolderInfo getFolder()


### Example

```typescript
import {
    FolderApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FolderApi(configuration);

let folderId: string; //Folder Id (default to undefined)
let shareId: string; //Share Id (optional) (default to undefined)
let timestamp: number; //Past timestamp to view the folder at, in ms since epoch (optional) (default to undefined)

const { status, data } = await apiInstance.getFolder(
    folderId,
    shareId,
    timestamp
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **folderId** | [**string**] | Folder Id | defaults to undefined|
| **shareId** | [**string**] | Share Id | (optional) defaults to undefined|
| **timestamp** | [**number**] | Past timestamp to view the folder at, in ms since epoch | (optional) defaults to undefined|


### Return type

**FolderInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Folder Info |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFolderHistory**
> Array<FileActionInfo> getFolderHistory()


### Example

```typescript
import {
    FolderApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FolderApi(configuration);

let fileId: string; //File Id (default to undefined)
let timestamp: number; //Past timestamp to view the folder at, in ms since epoch (default to undefined)

const { status, data } = await apiInstance.getFolderHistory(
    fileId,
    timestamp
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fileId** | [**string**] | File Id | defaults to undefined|
| **timestamp** | [**number**] | Past timestamp to view the folder at, in ms since epoch | defaults to undefined|


### Return type

**Array<FileActionInfo>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File actions |  -  |
|**400** | Bad Request |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **scanFolder**
> TaskInfo scanFolder()


### Example

```typescript
import {
    FolderApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FolderApi(configuration);

let folderId: string; //Folder Id (default to undefined)
let shareId: string; //Share Id (optional) (default to undefined)

const { status, data } = await apiInstance.scanFolder(
    folderId,
    shareId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **folderId** | [**string**] | Folder Id | defaults to undefined|
| **shareId** | [**string**] | Share Id | (optional) defaults to undefined|


### Return type

**TaskInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Task Info |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setFolderCover**
> setFolderCover()


### Example

```typescript
import {
    FolderApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FolderApi(configuration);

let folderId: string; //Folder Id (default to undefined)
let mediaId: string; //Media Id (default to undefined)

const { status, data } = await apiInstance.setFolderCover(
    folderId,
    mediaId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **folderId** | [**string**] | Folder Id | defaults to undefined|
| **mediaId** | [**string**] | Media Id | defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

