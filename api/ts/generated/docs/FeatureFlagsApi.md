# FeatureFlagsApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getFlags**](#getflags) | **GET** /flags | Get Feature Flags|
|[**setFlags**](#setflags) | **POST** /flags | Set Feature Flags|

# **getFlags**
> Bundle getFlags()


### Example

```typescript
import {
    FeatureFlagsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FeatureFlagsApi(configuration);

const { status, data } = await apiInstance.getFlags();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Bundle**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Feature Flags |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setFlags**
> setFlags(request)


### Example

```typescript
import {
    FeatureFlagsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new FeatureFlagsApi(configuration);

let request: Array<StructsSetConfigParam>; //Feature Flag Params

const { status, data } = await apiInstance.setFlags(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **Array<StructsSetConfigParam>**| Feature Flag Params | |


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

