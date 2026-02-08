# FolderApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createFolder**](#createfolder) | **POST** /folder | Create a new folder|
|[**getFolder**](#getfolder) | **GET** /folder/{folderID} | Get a folder|
|[**getFolderHistory**](#getfolderhistory) | **GET** /files/{fileID}/history | Get actions of a folder at a given time|
|[**scanFolder**](#scanfolder) | **POST** /folder/{folderID}/scan | Dispatch a folder scan|
|[**setFolderCover**](#setfoldercover) | **PATCH** /folder/{folderID}/cover | Set the cover image of a folder|

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
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.createFolder(
    request,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateFolderBody**| New folder body | |
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


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

let folderID: string; //Folder ID (default to undefined)
let shareID: string; //Share ID (optional) (default to undefined)
let timestamp: number; //Past timestamp to view the folder at, in ms since epoch (optional) (default to undefined)
let sortProp: 'name' | 'size' | 'updatedAt'; //Property to sort by (optional) (default to 'name')
let sortOrder: 'asc' | 'desc'; //Sort order (optional) (default to 'asc')

const { status, data } = await apiInstance.getFolder(
    folderID,
    shareID,
    timestamp,
    sortProp,
    sortOrder
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **folderID** | [**string**] | Folder ID | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|
| **timestamp** | [**number**] | Past timestamp to view the folder at, in ms since epoch | (optional) defaults to undefined|
| **sortProp** | [**&#39;name&#39; | &#39;size&#39; | &#39;updatedAt&#39;**]**Array<&#39;name&#39; &#124; &#39;size&#39; &#124; &#39;updatedAt&#39;>** | Property to sort by | (optional) defaults to 'name'|
| **sortOrder** | [**&#39;asc&#39; | &#39;desc&#39;**]**Array<&#39;asc&#39; &#124; &#39;desc&#39;>** | Sort order | (optional) defaults to 'asc'|


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

let fileID: string; //File ID (default to undefined)

const { status, data } = await apiInstance.getFolderHistory(
    fileID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fileID** | [**string**] | File ID | defaults to undefined|


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

let folderID: string; //Folder ID (default to undefined)
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.scanFolder(
    folderID,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **folderID** | [**string**] | Folder ID | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


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

let folderID: string; //Folder ID (default to undefined)
let mediaID: string; //Media ID (default to undefined)

const { status, data } = await apiInstance.setFolderCover(
    folderID,
    mediaID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **folderID** | [**string**] | Folder ID | defaults to undefined|
| **mediaID** | [**string**] | Media ID | defaults to undefined|


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

