# \FolderAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateFolder**](FolderAPI.md#CreateFolder) | **Post** /folder | Create a new folder
[**GetFolder**](FolderAPI.md#GetFolder) | **Get** /folder/{folderId} | Get a folder
[**GetFolderHistory**](FolderAPI.md#GetFolderHistory) | **Get** /files/{fileId}/history | Get actions of a folder at a given time
[**ScanFolder**](FolderAPI.md#ScanFolder) | **Post** /folder/{folderId}/scan | Dispatch a folder scan
[**SetFolderCover**](FolderAPI.md#SetFolderCover) | **Patch** /folder/{folderId}/cover | Set the cover image of a folder



## CreateFolder

> FileInfo CreateFolder(ctx).Request(request).ShareId(shareId).Execute()

Create a new folder

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
	request := *openapiclient.NewCreateFolderBody("NewFolderName_example", "ParentFolderId_example") // CreateFolderBody | New folder body
	shareId := "shareId_example" // string | Share Id (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.CreateFolder(context.Background()).Request(request).ShareId(shareId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FolderAPI.CreateFolder``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateFolder`: FileInfo
	fmt.Fprintf(os.Stdout, "Response from `FolderAPI.CreateFolder`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateFolderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**CreateFolderBody**](CreateFolderBody.md) | New folder body | 
 **shareId** | **string** | Share Id | 

### Return type

[**FileInfo**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFolder

> FolderInfo GetFolder(ctx, folderId).ShareId(shareId).Timestamp(timestamp).Execute()

Get a folder

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
	folderId := "folderId_example" // string | Folder Id
	shareId := "shareId_example" // string | Share Id (optional)
	timestamp := int32(56) // int32 | Past timestamp to view the folder at, in ms since epoch (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.GetFolder(context.Background(), folderId).ShareId(shareId).Timestamp(timestamp).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FolderAPI.GetFolder``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFolder`: FolderInfo
	fmt.Fprintf(os.Stdout, "Response from `FolderAPI.GetFolder`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**folderId** | **string** | Folder Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFolderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **shareId** | **string** | Share Id | 
 **timestamp** | **int32** | Past timestamp to view the folder at, in ms since epoch | 

### Return type

[**FolderInfo**](FolderInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFolderHistory

> []FileActionInfo GetFolderHistory(ctx, fileId).Timestamp(timestamp).Execute()

Get actions of a folder at a given time

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
	fileId := "fileId_example" // string | File Id
	timestamp := int32(56) // int32 | Past timestamp to view the folder at, in ms since epoch

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.GetFolderHistory(context.Background(), fileId).Timestamp(timestamp).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FolderAPI.GetFolderHistory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFolderHistory`: []FileActionInfo
	fmt.Fprintf(os.Stdout, "Response from `FolderAPI.GetFolderHistory`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fileId** | **string** | File Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFolderHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **timestamp** | **int32** | Past timestamp to view the folder at, in ms since epoch | 

### Return type

[**[]FileActionInfo**](FileActionInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ScanFolder

> TaskInfo ScanFolder(ctx, folderId).ShareId(shareId).Execute()

Dispatch a folder scan

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
	folderId := "folderId_example" // string | Folder Id
	shareId := "shareId_example" // string | Share Id (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.ScanFolder(context.Background(), folderId).ShareId(shareId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FolderAPI.ScanFolder``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ScanFolder`: TaskInfo
	fmt.Fprintf(os.Stdout, "Response from `FolderAPI.ScanFolder`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**folderId** | **string** | Folder Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiScanFolderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **shareId** | **string** | Share Id | 

### Return type

[**TaskInfo**](TaskInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetFolderCover

> SetFolderCover(ctx, folderId).MediaId(mediaId).Execute()

Set the cover image of a folder

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
	folderId := "folderId_example" // string | Folder Id
	mediaId := "mediaId_example" // string | Media Id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FolderAPI.SetFolderCover(context.Background(), folderId).MediaId(mediaId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FolderAPI.SetFolderCover``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**folderId** | **string** | Folder Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetFolderCoverRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **mediaId** | **string** | Media Id | 

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

