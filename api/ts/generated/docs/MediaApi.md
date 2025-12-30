# MediaApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**cleanupMedia**](#cleanupmedia) | **POST** /media/cleanup | Make sure all media is correctly synced with the file system|
|[**dropHDIRs**](#drophdirs) | **POST** /media/drop/hdirs | Drop all computed media HDIR data. Must be server owner.|
|[**dropMedia**](#dropmedia) | **POST** /media/drop | DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.|
|[**getMedia**](#getmedia) | **POST** /media | Get paginated media|
|[**getMediaFile**](#getmediafile) | **GET** /media/{mediaID}/file | Get file of media by id|
|[**getMediaImage**](#getmediaimage) | **GET** /media/{mediaID}.{extension} | Get a media image bytes|
|[**getMediaInfo**](#getmediainfo) | **GET** /media/{mediaID}/info | Get media info|
|[**getMediaTypes**](#getmediatypes) | **GET** /media/types | Get media type dictionary|
|[**getRandomMedia**](#getrandommedia) | **GET** /media/random | Get random media|
|[**setMediaLiked**](#setmedialiked) | **PATCH** /media/{mediaID}/liked | Like a media|
|[**setMediaVisibility**](#setmediavisibility) | **PATCH** /media/visibility | Set media visibility|
|[**streamVideo**](#streamvideo) | **GET** /media/{mediaID}/video | Stream a video|

# **cleanupMedia**
> cleanupMedia()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

const { status, data } = await apiInstance.cleanupMedia();
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
|**200** | OK |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **dropHDIRs**
> dropHDIRs()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

const { status, data } = await apiInstance.dropHDIRs();
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
|**200** | OK |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **dropMedia**
> dropMedia()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

const { status, data } = await apiInstance.dropMedia();
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
|**200** | OK |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMedia**
> MediaBatchInfo getMedia(request)


### Example

```typescript
import {
    MediaApi,
    Configuration,
    MediaBatchParams
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let request: MediaBatchParams; //Media Batch Params
let shareID: string; //File ShareID (optional) (default to undefined)

const { status, data } = await apiInstance.getMedia(
    request,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **MediaBatchParams**| Media Batch Params | |
| **shareID** | [**string**] | File ShareID | (optional) defaults to undefined|


### Return type

**MediaBatchInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Media Batch |  -  |
|**400** | Bad Request |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMediaFile**
> FileInfo getMediaFile()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let mediaID: string; //ID of media (default to undefined)

const { status, data } = await apiInstance.getMediaFile(
    mediaID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mediaID** | [**string**] | ID of media | defaults to undefined|


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
|**200** | File info of file media was created from |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMediaImage**
> string getMediaImage()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let mediaID: string; //Media ID (default to undefined)
let extension: string; //Extension (default to undefined)
let quality: 'thumbnail' | 'fullres'; //Image Quality (default to undefined)
let page: number; //Page number (optional) (default to undefined)

const { status, data } = await apiInstance.getMediaImage(
    mediaID,
    extension,
    quality,
    page
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mediaID** | [**string**] | Media ID | defaults to undefined|
| **extension** | [**string**] | Extension | defaults to undefined|
| **quality** | [**&#39;thumbnail&#39; | &#39;fullres&#39;**]**Array<&#39;thumbnail&#39; &#124; &#39;fullres&#39;>** | Image Quality | defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to undefined|


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: image/*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | image bytes |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMediaInfo**
> MediaInfo getMediaInfo()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let mediaID: string; //Media ID (default to undefined)

const { status, data } = await apiInstance.getMediaInfo(
    mediaID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mediaID** | [**string**] | Media ID | defaults to undefined|


### Return type

**MediaInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Media Info |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMediaTypes**
> MediaTypesInfo getMediaTypes()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

const { status, data } = await apiInstance.getMediaTypes();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**MediaTypesInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Media types |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getRandomMedia**
> MediaBatchInfo getRandomMedia()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let count: number; //Number of random medias to get (default to undefined)

const { status, data } = await apiInstance.getRandomMedia(
    count
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **count** | [**number**] | Number of random medias to get | defaults to undefined|


### Return type

**MediaBatchInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Media Batch |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setMediaLiked**
> setMediaLiked()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let mediaID: string; //ID of media (default to undefined)
let liked: boolean; //Liked status to set (default to undefined)
let shareID: string; //ShareID (optional) (default to undefined)

const { status, data } = await apiInstance.setMediaLiked(
    mediaID,
    liked,
    shareID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mediaID** | [**string**] | ID of media | defaults to undefined|
| **liked** | [**boolean**] | Liked status to set | defaults to undefined|
| **shareID** | [**string**] | ShareID | (optional) defaults to undefined|


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

# **setMediaVisibility**
> setMediaVisibility(mediaIDs)


### Example

```typescript
import {
    MediaApi,
    Configuration,
    MediaIDsParams
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let hidden: boolean; //Set the media visibility (default to undefined)
let mediaIDs: MediaIDsParams; //MediaIDs to change visibility of

const { status, data } = await apiInstance.setMediaVisibility(
    hidden,
    mediaIDs
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mediaIDs** | **MediaIDsParams**| MediaIDs to change visibility of | |
| **hidden** | [**boolean**] | Set the media visibility | defaults to undefined|


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

# **streamVideo**
> streamVideo()


### Example

```typescript
import {
    MediaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let mediaID: string; //ID of media (default to undefined)

const { status, data } = await apiInstance.streamVideo(
    mediaID
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mediaID** | [**string**] | ID of media | defaults to undefined|


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

