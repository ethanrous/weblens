# TagsApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**addFilesToTag**](#addfilestotag) | **POST** /tags/{tagID}/files | Add files to a tag|
|[**createTag**](#createtag) | **POST** /tags | Create a new tag|
|[**deleteTag**](#deletetag) | **DELETE** /tags/{tagID} | Delete a tag|
|[**getFilesByTag**](#getfilesbytag) | **GET** /tags/{tagID}/files | Get all files in a tag|
|[**getTag**](#gettag) | **GET** /tags/{tagID} | Get a tag by ID|
|[**getTagsForFile**](#gettagsforfile) | **GET** /tags/file/{fileID} | Get tags for a file|
|[**getUserTags**](#getusertags) | **GET** /tags | Get all tags for the authenticated user|
|[**removeFilesFromTag**](#removefilesfromtag) | **DELETE** /tags/{tagID}/files | Remove files from a tag|
|[**updateTag**](#updatetag) | **PATCH** /tags/{tagID} | Update a tag\&#39;s name and/or color|

# **addFilesToTag**
> addFilesToTag(request)


### Example

```typescript
import {
    TagsApi,
    Configuration,
    TagFileIDsParams
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let tagID: string; //Tag ID (default to undefined)
let request: TagFileIDsParams; //File IDs to add

const { status, data } = await apiInstance.addFilesToTag(
    tagID,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **TagFileIDsParams**| File IDs to add | |
| **tagID** | [**string**] | Tag ID | defaults to undefined|


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
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTag**
> GithubComEthanrousWeblensModelsTagTag createTag(request)


### Example

```typescript
import {
    TagsApi,
    Configuration,
    TagCreateTagParams
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let request: TagCreateTagParams; //Create tag request body

const { status, data } = await apiInstance.createTag(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **TagCreateTagParams**| Create tag request body | |


### Return type

**GithubComEthanrousWeblensModelsTagTag**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Created tag |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**409** | Conflict |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteTag**
> deleteTag()


### Example

```typescript
import {
    TagsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let tagID: string; //Tag ID (default to undefined)

const { status, data } = await apiInstance.deleteTag(
    tagID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tagID** | [**string**] | Tag ID | defaults to undefined|


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
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFilesByTag**
> FilesInfo getFilesByTag()


### Example

```typescript
import {
    TagsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let tagID: string; //Tag ID (default to undefined)

const { status, data } = await apiInstance.getFilesByTag(
    tagID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tagID** | [**string**] | Tag ID | defaults to undefined|


### Return type

**FilesInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Files in the tag |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTag**
> GithubComEthanrousWeblensModelsTagTag getTag()


### Example

```typescript
import {
    TagsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let tagID: string; //Tag ID (default to undefined)

const { status, data } = await apiInstance.getTag(
    tagID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tagID** | [**string**] | Tag ID | defaults to undefined|


### Return type

**GithubComEthanrousWeblensModelsTagTag**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tag |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTagsForFile**
> Array<GithubComEthanrousWeblensModelsTagTag> getTagsForFile()


### Example

```typescript
import {
    TagsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let fileID: string; //File ID (default to undefined)

const { status, data } = await apiInstance.getTagsForFile(
    fileID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **fileID** | [**string**] | File ID | defaults to undefined|


### Return type

**Array<GithubComEthanrousWeblensModelsTagTag>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tags containing the file |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUserTags**
> Array<GithubComEthanrousWeblensModelsTagTag> getUserTags()


### Example

```typescript
import {
    TagsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

const { status, data } = await apiInstance.getUserTags();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<GithubComEthanrousWeblensModelsTagTag>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | User\&#39;s tags |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **removeFilesFromTag**
> removeFilesFromTag(request)


### Example

```typescript
import {
    TagsApi,
    Configuration,
    TagFileIDsParams
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let tagID: string; //Tag ID (default to undefined)
let request: TagFileIDsParams; //File IDs to remove

const { status, data } = await apiInstance.removeFilesFromTag(
    tagID,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **TagFileIDsParams**| File IDs to remove | |
| **tagID** | [**string**] | Tag ID | defaults to undefined|


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
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateTag**
> updateTag(request)


### Example

```typescript
import {
    TagsApi,
    Configuration,
    TagUpdateTagParams
} from './api';

const configuration = new Configuration();
const apiInstance = new TagsApi(configuration);

let tagID: string; //Tag ID (default to undefined)
let request: TagUpdateTagParams; //Update tag request body

const { status, data } = await apiInstance.updateTag(
    tagID,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **TagUpdateTagParams**| Update tag request body | |
| **tagID** | [**string**] | Tag ID | defaults to undefined|


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
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

