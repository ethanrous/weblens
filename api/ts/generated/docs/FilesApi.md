# FilesApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**addFilesToUpload**](#addfilestoupload) | **POST** /upload/{uploadID} | Add a file to an upload task|
|[**autocompletePath**](#autocompletepath) | **GET** /files/autocomplete | Get path completion suggestions|
|[**createTakeout**](#createtakeout) | **POST** /takeout | Create a zip file|
|[**deleteFiles**](#deletefiles) | **DELETE** /files | Delete Files \&quot;permanently\&quot;|
|[**downloadFile**](#downloadfile) | **GET** /files/{fileID}/download | Download a file|
|[**getFile**](#getfile) | **GET** /files/{fileID} | Get information about a file|
|[**getFileStats**](#getfilestats) | **GET** /files/{fileID}/stats | Get the statistics of a file|
|[**getFileText**](#getfiletext) | **GET** /files/{fileID}/text | Get the text of a text file|
|[**getSharedFiles**](#getsharedfiles) | **GET** /files/shared | Get files shared with the logged in user|
|[**getUploadResult**](#getuploadresult) | **GET** /upload/{uploadID} | Get the result of an upload task. This will block until the upload is complete|
|[**moveFiles**](#movefiles) | **PATCH** /files | Move a list of files to a new parent folder|
|[**restoreFiles**](#restorefiles) | **POST** /files/structsore | structsore files from some time in the past|
|[**searchByFilename**](#searchbyfilename) | **GET** /files/search | Search for files by filename|
|[**startUpload**](#startupload) | **POST** /upload | Begin a new upload task|
|[**unTrashFiles**](#untrashfiles) | **PATCH** /files/untrash | Move a list of files out of the trash, structsoring them to where they were before|
|[**updateFile**](#updatefile) | **PATCH** /files/{fileID} | Update a File|
|[**uploadFileChunk**](#uploadfilechunk) | **PUT** /upload/{uploadID}/file/{fileID} | Add a chunk to a file upload|

# **addFilesToUpload**
> NewFilesInfo addFilesToUpload(request)


### Example

```typescript
import {
    FilesApi,
    Configuration,
    NewFilesParams
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let uploadID: string; //Upload ID (default to undefined)
let request: NewFilesParams; //New file params
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.addFilesToUpload(
    uploadID,
    request,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **NewFilesParams**| New file params | |
| **uploadID** | [**string**] | Upload ID | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


### Return type

**NewFilesInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | FileIds |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **autocompletePath**
> FolderInfo autocompletePath()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let searchPath: string; //Search path (default to undefined)

const { status, data } = await apiInstance.autocompletePath(
    searchPath
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **searchPath** | [**string**] | Search path | defaults to undefined|


### Return type

**FolderInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Path info |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTakeout**
> TakeoutInfo createTakeout(request)

Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists

### Example

```typescript
import {
    FilesApi,
    Configuration,
    FilesListParams
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let request: FilesListParams; //File Ids
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.createTakeout(
    request,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **FilesListParams**| File Ids | |
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


### Return type

**TakeoutInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Zip Takeout Info |  -  |
|**202** | Task Dispatch Info |  -  |
|**400** | Bad Request |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteFiles**
> deleteFiles(request)


### Example

```typescript
import {
    FilesApi,
    Configuration,
    FilesListParams
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let request: FilesListParams; //Delete files request body
let ignoreTrash: boolean; //Delete files even if they are not in the trash (optional) (default to undefined)
let preserveFolder: boolean; //Preserve parent folder if it is empty after deletion (optional) (default to undefined)

const { status, data } = await apiInstance.deleteFiles(
    request,
    ignoreTrash,
    preserveFolder
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **FilesListParams**| Delete files request body | |
| **ignoreTrash** | [**boolean**] | Delete files even if they are not in the trash | (optional) defaults to undefined|
| **preserveFolder** | [**boolean**] | Preserve parent folder if it is empty after deletion | (optional) defaults to undefined|


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
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **downloadFile**
> string downloadFile()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let fileID: string; //File ID (default to undefined)
let shareID: string; //Share ID (optional) (default to undefined)
let format: string; //File format conversion (optional) (default to undefined)
let isTakeout: boolean; //Is this a takeout file (optional) (default to false)

const { status, data } = await apiInstance.downloadFile(
    fileID,
    shareID,
    format,
    isTakeout
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fileID** | [**string**] | File ID | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|
| **format** | [**string**] | File format conversion | (optional) defaults to undefined|
| **isTakeout** | [**boolean**] | Is this a takeout file | (optional) defaults to false|


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File content |  -  |
|**404** | Error Info |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFile**
> FileInfo getFile()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let fileID: string; //File ID (default to undefined)
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.getFile(
    fileID,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fileID** | [**string**] | File ID | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


### Return type

**FileInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File Info |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFileStats**
> getFileStats()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let fileID: string; //File ID (default to undefined)

const { status, data } = await apiInstance.getFileStats(
    fileID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fileID** | [**string**] | File ID | defaults to undefined|


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
|**400** | Bad Request |  -  |
|**501** | Not Implemented |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFileText**
> string getFileText()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let fileID: string; //File ID (default to undefined)
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.getFileText(
    fileID,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fileID** | [**string**] | File ID | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File text |  -  |
|**400** | Bad Request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSharedFiles**
> FolderInfo getSharedFiles()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

const { status, data } = await apiInstance.getSharedFiles();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**FolderInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | All the top-level files shared with the user |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUploadResult**
> getUploadResult()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let uploadID: string; //Upload ID (default to undefined)

const { status, data } = await apiInstance.getUploadResult(
    uploadID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uploadID** | [**string**] | Upload ID | defaults to undefined|


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
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **moveFiles**
> moveFiles(request)


### Example

```typescript
import {
    FilesApi,
    Configuration,
    MoveFilesParams
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let request: MoveFilesParams; //Move files request body
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.moveFiles(
    request,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **MoveFilesParams**| Move files request body | |
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


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
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **restoreFiles**
> RestoreFilesInfo restoreFiles(request)


### Example

```typescript
import {
    FilesApi,
    Configuration,
    RestoreFilesBody
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let request: RestoreFilesBody; //RestoreFiles files request body

const { status, data } = await apiInstance.restoreFiles(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **RestoreFilesBody**| RestoreFiles files request body | |


### Return type

**RestoreFilesInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | structsore files info |  -  |
|**400** | Bad Request |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **searchByFilename**
> Array<FileInfo> searchByFilename()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let search: string; //Filename to search for (default to undefined)
let baseFolderID: string; //The folder to search in, defaults to the user\'s home folder (optional) (default to undefined)

const { status, data } = await apiInstance.searchByFilename(
    search,
    baseFolderID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **search** | [**string**] | Filename to search for | defaults to undefined|
| **baseFolderID** | [**string**] | The folder to search in, defaults to the user\&#39;s home folder | (optional) defaults to undefined|


### Return type

**Array<FileInfo>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File Info |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **startUpload**
> NewUploadInfo startUpload(request)


### Example

```typescript
import {
    FilesApi,
    Configuration,
    NewUploadParams
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let request: NewUploadParams; //New upload request body
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.startUpload(
    request,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **NewUploadParams**| New upload request body | |
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


### Return type

**NewUploadInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Upload Info |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **unTrashFiles**
> unTrashFiles(request)


### Example

```typescript
import {
    FilesApi,
    Configuration,
    FilesListParams
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let request: FilesListParams; //Un-trash files request body

const { status, data } = await apiInstance.unTrashFiles(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **FilesListParams**| Un-trash files request body | |


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
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateFile**
> updateFile(request)


### Example

```typescript
import {
    FilesApi,
    Configuration,
    UpdateFileParams
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let fileID: string; //File ID (default to undefined)
let request: UpdateFileParams; //Update file request body
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.updateFile(
    fileID,
    request,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **UpdateFileParams**| Update file request body | |
| **fileID** | [**string**] | File ID | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **uploadFileChunk**
> uploadFileChunk()


### Example

```typescript
import {
    FilesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FilesApi(configuration);

let uploadID: string; //Upload ID (default to undefined)
let fileID: string; //File ID (default to undefined)
let chunk: File; //File chunk (default to undefined)
let shareID: string; //Share ID (optional) (default to undefined)

const { status, data } = await apiInstance.uploadFileChunk(
    uploadID,
    fileID,
    chunk,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uploadID** | [**string**] | Upload ID | defaults to undefined|
| **fileID** | [**string**] | File ID | defaults to undefined|
| **chunk** | [**File**] | File chunk | defaults to undefined|
| **shareID** | [**string**] | Share ID | (optional) defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

