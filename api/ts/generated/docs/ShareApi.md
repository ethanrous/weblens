# ShareApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**addUserToShare**](#addusertoshare) | **POST** /share/{shareId}/accessors | Add a user to a file share|
|[**createFileShare**](#createfileshare) | **POST** /share/file | Share a file|
|[**deleteFileShare**](#deletefileshare) | **DELETE** /share/{shareId} | Delete a file share|
|[**getFileShare**](#getfileshare) | **GET** /share/{shareId} | Get a file share|
|[**removeUserFromShare**](#removeuserfromshare) | **DELETE** /share/{shareId}/accessors/{username} | Remove a user from a file share|
|[**setSharePublic**](#setsharepublic) | **PATCH** /share/{shareId}/public | Update a share\&#39;s \&quot;public\&quot; status|
|[**updateShareAccessorPermissions**](#updateshareaccessorpermissions) | **PATCH** /share/{shareId}/accessors/{username} | Update a share\&#39;s user permissions|

# **addUserToShare**
> ShareInfo addUserToShare(request)


### Example

```typescript
import {
    ShareApi,
    Configuration,
    AddUserParams
} from './api';

const configuration = new Configuration();
const apiInstance = new ShareApi(configuration);

let shareId: string; //Share Id (default to undefined)
let request: AddUserParams; //Share Accessors

const { status, data } = await apiInstance.addUserToShare(
    shareId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **AddUserParams**| Share Accessors | |
| **shareId** | [**string**] | Share Id | defaults to undefined|


### Return type

**ShareInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createFileShare**
> ShareInfo createFileShare(request)


### Example

```typescript
import {
    ShareApi,
    Configuration,
    FileShareParams
} from './api';

const configuration = new Configuration();
const apiInstance = new ShareApi(configuration);

let request: FileShareParams; //New File Share Params

const { status, data } = await apiInstance.createFileShare(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **FileShareParams**| New File Share Params | |


### Return type

**ShareInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | New File Share |  -  |
|**409** | Conflict |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteFileShare**
> deleteFileShare()


### Example

```typescript
import {
    ShareApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ShareApi(configuration);

let shareId: string; //Share Id (default to undefined)

const { status, data } = await apiInstance.deleteFileShare(
    shareId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **shareId** | [**string**] | Share Id | defaults to undefined|


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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFileShare**
> ShareInfo getFileShare()


### Example

```typescript
import {
    ShareApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ShareApi(configuration);

let shareId: string; //Share Id (default to undefined)

const { status, data } = await apiInstance.getFileShare(
    shareId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **shareId** | [**string**] | Share Id | defaults to undefined|


### Return type

**ShareInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File Share |  -  |
|**404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **removeUserFromShare**
> ShareInfo removeUserFromShare()


### Example

```typescript
import {
    ShareApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ShareApi(configuration);

let shareId: string; //Share Id (default to undefined)
let username: string; //Username (default to undefined)

const { status, data } = await apiInstance.removeUserFromShare(
    shareId,
    username
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **shareId** | [**string**] | Share Id | defaults to undefined|
| **username** | [**string**] | Username | defaults to undefined|


### Return type

**ShareInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setSharePublic**
> setSharePublic()


### Example

```typescript
import {
    ShareApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ShareApi(configuration);

let shareId: string; //Share Id (default to undefined)
let _public: boolean; //Share Public Status (default to undefined)

const { status, data } = await apiInstance.setSharePublic(
    shareId,
    _public
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **shareId** | [**string**] | Share Id | defaults to undefined|
| **_public** | [**boolean**] | Share Public Status | defaults to undefined|


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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateShareAccessorPermissions**
> ShareInfo updateShareAccessorPermissions(request)


### Example

```typescript
import {
    ShareApi,
    Configuration,
    PermissionsParams
} from './api';

const configuration = new Configuration();
const apiInstance = new ShareApi(configuration);

let shareId: string; //Share Id (default to undefined)
let username: string; //Username (default to undefined)
let request: PermissionsParams; //Share Permissions Params

const { status, data } = await apiInstance.updateShareAccessorPermissions(
    shareId,
    username,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **PermissionsParams**| Share Permissions Params | |
| **shareId** | [**string**] | Share Id | defaults to undefined|
| **username** | [**string**] | Username | defaults to undefined|


### Return type

**ShareInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

