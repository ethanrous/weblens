# APIKeysApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createAPIKey**](#createapikey) | **POST** /keys | Create a new api key|
|[**deleteAPIKey**](#deleteapikey) | **DELETE** /keys/{tokenID} | Delete an api key|
|[**getAPIKeys**](#getapikeys) | **GET** /keys | Get all api keys|

# **createAPIKey**
> TokenInfo createAPIKey(params)


### Example

```typescript
import {
    APIKeysApi,
    Configuration,
    APIKeyParams
} from './api';

const configuration = new Configuration();
const apiInstance = new APIKeysApi(configuration);

let params: APIKeyParams; //The new token params

const { status, data } = await apiInstance.createAPIKey(
    params
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **params** | **APIKeyParams**| The new token params | |


### Return type

**TokenInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | The new token |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteAPIKey**
> deleteAPIKey()


### Example

```typescript
import {
    APIKeysApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new APIKeysApi(configuration);

let tokenID: string; //API key id (default to undefined)

const { status, data } = await apiInstance.deleteAPIKey(
    tokenID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tokenID** | [**string**] | API key id | defaults to undefined|


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
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAPIKeys**
> Array<TokenInfo> getAPIKeys()


### Example

```typescript
import {
    APIKeysApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new APIKeysApi(configuration);

const { status, data } = await apiInstance.getAPIKeys();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<TokenInfo>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tokens |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

