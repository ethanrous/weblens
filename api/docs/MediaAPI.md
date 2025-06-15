# \MediaAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CleanupMedia**](MediaAPI.md#CleanupMedia) | **Post** /media/cleanup | Make sure all media is correctly synced with the file system
[**DropMedia**](MediaAPI.md#DropMedia) | **Post** /media/drop | DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.
[**GetMedia**](MediaAPI.md#GetMedia) | **Get** /media | Get paginated media
[**GetMediaFile**](MediaAPI.md#GetMediaFile) | **Get** /media/{mediaId}/file | Get file of media by id
[**GetMediaImage**](MediaAPI.md#GetMediaImage) | **Get** /media/{mediaId}.{extension} | Get a media image bytes
[**GetMediaInfo**](MediaAPI.md#GetMediaInfo) | **Get** /media/{mediaId}/info | Get media info
[**GetMediaTypes**](MediaAPI.md#GetMediaTypes) | **Get** /media/types | Get media type dictionary
[**GetRandomMedia**](MediaAPI.md#GetRandomMedia) | **Get** /media/random | Get random media
[**SetMediaLiked**](MediaAPI.md#SetMediaLiked) | **Patch** /media/{mediaId}/liked | Like a media
[**SetMediaVisibility**](MediaAPI.md#SetMediaVisibility) | **Patch** /media/visibility | Set media visibility
[**StreamVideo**](MediaAPI.md#StreamVideo) | **Get** /media/{mediaId}/video | Stream a video



## CleanupMedia

> CleanupMedia(ctx).Execute()

Make sure all media is correctly synced with the file system

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.MediaAPI.CleanupMedia(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.CleanupMedia``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiCleanupMediaRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DropMedia

> DropMedia(ctx).Execute()

DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.MediaAPI.DropMedia(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.DropMedia``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDropMediaRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMedia

> MediaBatchInfo GetMedia(ctx).Raw(raw).Hidden(hidden).Sort(sort).Search(search).Page(page).Limit(limit).FolderIds(folderIds).MediaIds(mediaIds).Execute()

Get paginated media

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	raw := true // bool | Include raw files (optional) (default to false)
	hidden := true // bool | Include hidden media (optional) (default to false)
	sort := "sort_example" // string | Sort by field (optional) (default to "createDate")
	search := "search_example" // string | Search string (optional)
	page := int32(56) // int32 | Page of medias to get (optional)
	limit := int32(56) // int32 | Number of medias to get (optional)
	folderIds := "folderIds_example" // string | Search only in given folders (optional)
	mediaIds := "mediaIds_example" // string | Get only media with the provided ids (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MediaAPI.GetMedia(context.Background()).Raw(raw).Hidden(hidden).Sort(sort).Search(search).Page(page).Limit(limit).FolderIds(folderIds).MediaIds(mediaIds).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.GetMedia``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMedia`: MediaBatchInfo
	fmt.Fprintf(os.Stdout, "Response from `MediaAPI.GetMedia`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetMediaRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **raw** | **bool** | Include raw files | [default to false]
 **hidden** | **bool** | Include hidden media | [default to false]
 **sort** | **string** | Sort by field | [default to &quot;createDate&quot;]
 **search** | **string** | Search string | 
 **page** | **int32** | Page of medias to get | 
 **limit** | **int32** | Number of medias to get | 
 **folderIds** | **string** | Search only in given folders | 
 **mediaIds** | **string** | Get only media with the provided ids | 

### Return type

[**MediaBatchInfo**](MediaBatchInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMediaFile

> FileInfo GetMediaFile(ctx, mediaId).Execute()

Get file of media by id

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	mediaId := "mediaId_example" // string | Id of media

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MediaAPI.GetMediaFile(context.Background(), mediaId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.GetMediaFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMediaFile`: FileInfo
	fmt.Fprintf(os.Stdout, "Response from `MediaAPI.GetMediaFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**mediaId** | **string** | Id of media | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetMediaFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**FileInfo**](FileInfo.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMediaImage

> string GetMediaImage(ctx, mediaId, extension).Quality(quality).Page(page).Execute()

Get a media image bytes

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	mediaId := "mediaId_example" // string | Media Id
	extension := "extension_example" // string | Extension
	quality := "quality_example" // string | Image Quality
	page := int32(56) // int32 | Page number (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MediaAPI.GetMediaImage(context.Background(), mediaId, extension).Quality(quality).Page(page).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.GetMediaImage``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMediaImage`: string
	fmt.Fprintf(os.Stdout, "Response from `MediaAPI.GetMediaImage`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**mediaId** | **string** | Media Id | 
**extension** | **string** | Extension | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetMediaImageRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **quality** | **string** | Image Quality | 
 **page** | **int32** | Page number | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: image/*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMediaInfo

> MediaInfo GetMediaInfo(ctx, mediaId).Execute()

Get media info

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	mediaId := "mediaId_example" // string | Media Id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MediaAPI.GetMediaInfo(context.Background(), mediaId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.GetMediaInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMediaInfo`: MediaInfo
	fmt.Fprintf(os.Stdout, "Response from `MediaAPI.GetMediaInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**mediaId** | **string** | Media Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetMediaInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**MediaInfo**](MediaInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMediaTypes

> MediaTypesInfo GetMediaTypes(ctx).Execute()

Get media type dictionary

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MediaAPI.GetMediaTypes(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.GetMediaTypes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMediaTypes`: MediaTypesInfo
	fmt.Fprintf(os.Stdout, "Response from `MediaAPI.GetMediaTypes`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetMediaTypesRequest struct via the builder pattern


### Return type

[**MediaTypesInfo**](MediaTypesInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRandomMedia

> MediaBatchInfo GetRandomMedia(ctx).Count(count).Execute()

Get random media

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	count := float32(8.14) // float32 | Number of random medias to get

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MediaAPI.GetRandomMedia(context.Background()).Count(count).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.GetRandomMedia``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRandomMedia`: MediaBatchInfo
	fmt.Fprintf(os.Stdout, "Response from `MediaAPI.GetRandomMedia`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetRandomMediaRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **count** | **float32** | Number of random medias to get | 

### Return type

[**MediaBatchInfo**](MediaBatchInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetMediaLiked

> SetMediaLiked(ctx, mediaId).Liked(liked).ShareId(shareId).Execute()

Like a media

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	mediaId := "mediaId_example" // string | Id of media
	liked := true // bool | Liked status to set
	shareId := "shareId_example" // string | ShareId (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.MediaAPI.SetMediaLiked(context.Background(), mediaId).Liked(liked).ShareId(shareId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.SetMediaLiked``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**mediaId** | **string** | Id of media | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetMediaLikedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **liked** | **bool** | Liked status to set | 
 **shareId** | **string** | ShareId | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetMediaVisibility

> SetMediaVisibility(ctx).Hidden(hidden).MediaIds(mediaIds).Execute()

Set media visibility

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	hidden := true // bool | Set the media visibility
	mediaIds := *openapiclient.NewMediaIdsParams() // MediaIdsParams | MediaIds to change visibility of

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.MediaAPI.SetMediaVisibility(context.Background()).Hidden(hidden).MediaIds(mediaIds).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.SetMediaVisibility``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetMediaVisibilityRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **hidden** | **bool** | Set the media visibility | 
 **mediaIds** | [**MediaIdsParams**](MediaIdsParams.md) | MediaIds to change visibility of | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StreamVideo

> StreamVideo(ctx, mediaId).Execute()

Stream a video

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/ethanrous/weblens/api"
)

func main() {
	mediaId := "mediaId_example" // string | Id of media

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.MediaAPI.StreamVideo(context.Background(), mediaId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MediaAPI.StreamVideo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**mediaId** | **string** | Id of media | 

### Other Parameters

Other parameters are passed through a pointer to a apiStreamVideoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

