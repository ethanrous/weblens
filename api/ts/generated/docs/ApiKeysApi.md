# ApiKeysApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createApiKey**](#createapikey) | **POST** /keys | Create a new api key|
|[**deleteApiKey**](#deleteapikey) | **DELETE** /keys/{tokenID} | Delete an api key|
|[**getApiKeys**](#getapikeys) | **GET** /keys | Get all api keys|

# **createApiKey**
> TokenInfo createApiKey(params)


### Example

```typescript
import {
    ApiKeysApi,
    Configuration,
    ApiKeyParams
} from './api';

const configuration = new Configuration();
const apiInstance = new ApiKeysApi(configuration);

let params: ApiKeyParams; //The new token params

const { status, data } = await apiInstance.createApiKey(
    params
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **params** | **ApiKeyParams**| The new token params | |


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

# **deleteApiKey**
> deleteApiKey()


### Example

```typescript
import {
    ApiKeysApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ApiKeysApi(configuration);

let tokenID: string; //Api key id (default to undefined)

const { status, data } = await apiInstance.deleteApiKey(
    tokenID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tokenID** | [**string**] | Api key id | defaults to undefined|


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

# **getApiKeys**
> Array<TokenInfo> getApiKeys()


### Example

```typescript
import {
    ApiKeysApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ApiKeysApi(configuration);

const { status, data } = await apiInstance.getApiKeys();
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

