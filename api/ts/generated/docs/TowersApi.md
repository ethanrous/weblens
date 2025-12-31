# TowersApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createRemote**](#createremote) | **POST** /tower/remote | Create a new remote|
|[**deleteRemote**](#deleteremote) | **DELETE** /tower/{serverID} | Delete a remote|
|[**flushCache**](#flushcache) | **DELETE** /tower/cache | Flush Cache|
|[**getBackupInfo**](#getbackupinfo) | **GET** /tower/backup | Get information about a file|
|[**getPagedHistoryActions**](#getpagedhistoryactions) | **GET** /tower/history | Get a page of file actions|
|[**getRemotes**](#getremotes) | **GET** /tower | Get all remotes|
|[**getRunningTasks**](#getrunningtasks) | **GET** /tower/tasks | Get Running Tasks|
|[**getServerInfo**](#getserverinfo) | **GET** /info | Get server info|
|[**initializeTower**](#initializetower) | **POST** /tower/init | Initialize the target server|
|[**launchBackup**](#launchbackup) | **POST** /tower/{serverID}/backup | Launch backup on a tower|
|[**resetTower**](#resettower) | **POST** /tower/reset | Reset tower|

# **createRemote**
> TowerInfo createRemote(request)


### Example

```typescript
import {
    TowersApi,
    Configuration,
    NewServerParams
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

let request: NewServerParams; //New Server Params

const { status, data } = await apiInstance.createRemote(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **NewServerParams**| New Server Params | |


### Return type

**TowerInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | New Server Info |  -  |
|**400** | Bad Request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteRemote**
> deleteRemote()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

let serverID: string; //Server ID to delete (default to undefined)

const { status, data } = await apiInstance.deleteRemote(
    serverID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **serverID** | [**string**] | Server ID to delete | defaults to undefined|


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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **flushCache**
> WLResponseInfo flushCache()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

const { status, data } = await apiInstance.flushCache();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**WLResponseInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Cache flushed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getBackupInfo**
> BackupInfo getBackupInfo()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

let timestamp: string; //Timestamp in milliseconds since epoch (default to undefined)

const { status, data } = await apiInstance.getBackupInfo(
    timestamp
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **timestamp** | [**string**] | Timestamp in milliseconds since epoch | defaults to undefined|


### Return type

**BackupInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Backup Info |  -  |
|**400** | Bad Request |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getPagedHistoryActions**
> Array<HistoryFileAction> getPagedHistoryActions()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

let page: number; //Page number (optional) (default to undefined)
let pageSize: number; //Number of items per page (optional) (default to undefined)

const { status, data } = await apiInstance.getPagedHistoryActions(
    page,
    pageSize
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to undefined|
| **pageSize** | [**number**] | Number of items per page | (optional) defaults to undefined|


### Return type

**Array<HistoryFileAction>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | File Actions |  -  |
|**400** | Bad Request |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getRemotes**
> Array<TowerInfo> getRemotes()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

const { status, data } = await apiInstance.getRemotes();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<TowerInfo>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tower Info |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getRunningTasks**
> Array<TaskInfo> getRunningTasks()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

const { status, data } = await apiInstance.getRunningTasks();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<TaskInfo>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Task Infos |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getServerInfo**
> TowerInfo getServerInfo()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

const { status, data } = await apiInstance.getServerInfo();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**TowerInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Server info |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **initializeTower**
> Array<TowerInfo> initializeTower(request)


### Example

```typescript
import {
    TowersApi,
    Configuration,
    StructsInitServerParams
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

let request: StructsInitServerParams; //Server initialization body

const { status, data } = await apiInstance.initializeTower(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **StructsInitServerParams**| Server initialization body | |


### Return type

**Array<TowerInfo>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | New server info |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **launchBackup**
> launchBackup()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

let serverID: string; //Server ID (default to undefined)

const { status, data } = await apiInstance.launchBackup(
    serverID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **serverID** | [**string**] | Server ID | defaults to undefined|


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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **resetTower**
> resetTower()


### Example

```typescript
import {
    TowersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TowersApi(configuration);

const { status, data } = await apiInstance.resetTower();
```

### Parameters
This endpoint does not have any parameters.


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
|**202** | Accepted |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

