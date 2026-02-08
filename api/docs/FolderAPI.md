# \FolderAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateFolder**](FolderAPI.md#CreateFolder) | **Post** /folder | Create a new folder
[**GetFolder**](FolderAPI.md#GetFolder) | **Get** /folder/{folderID} | Get a folder
[**GetFolderHistory**](FolderAPI.md#GetFolderHistory) | **Get** /files/{fileID}/history | Get actions of a folder at a given time
[**ScanFolder**](FolderAPI.md#ScanFolder) | **Post** /folder/{folderID}/scan | Dispatch a folder scan
[**SetFolderCover**](FolderAPI.md#SetFolderCover) | **Patch** /folder/{folderID}/cover | Set the cover image of a folder



## CreateFolder

> FileInfo CreateFolder(ctx).Request(request).ShareID(shareID).Execute()

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
	request := *openapiclient.NewCreateFolderBody("NewFolderName_example", "ParentFolderID_example") // CreateFolderBody | New folder body
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.CreateFolder(context.Background()).Request(request).ShareID(shareID).Execute()
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
 **shareID** | **string** | Share ID | 

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

> FolderInfo GetFolder(ctx, folderID).ShareID(shareID).Timestamp(timestamp).SortProp(sortProp).SortOrder(sortOrder).Execute()

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
	folderID := "folderID_example" // string | Folder ID
	shareID := "shareID_example" // string | Share ID (optional)
	timestamp := int32(56) // int32 | Past timestamp to view the folder at, in ms since epoch (optional)
	sortProp := "sortProp_example" // string | Property to sort by (optional) (default to "name")
	sortOrder := "sortOrder_example" // string | Sort order (optional) (default to "asc")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.GetFolder(context.Background(), folderID).ShareID(shareID).Timestamp(timestamp).SortProp(sortProp).SortOrder(sortOrder).Execute()
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
**folderID** | **string** | Folder ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFolderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **shareID** | **string** | Share ID | 
 **timestamp** | **int32** | Past timestamp to view the folder at, in ms since epoch | 
 **sortProp** | **string** | Property to sort by | [default to &quot;name&quot;]
 **sortOrder** | **string** | Sort order | [default to &quot;asc&quot;]

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

> []FileActionInfo GetFolderHistory(ctx, fileID).Execute()

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
	fileID := "fileID_example" // string | File ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.GetFolderHistory(context.Background(), fileID).Execute()
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
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFolderHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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

> TaskInfo ScanFolder(ctx, folderID).ShareID(shareID).Execute()

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
	folderID := "folderID_example" // string | Folder ID
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FolderAPI.ScanFolder(context.Background(), folderID).ShareID(shareID).Execute()
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
**folderID** | **string** | Folder ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiScanFolderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **shareID** | **string** | Share ID | 

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

> SetFolderCover(ctx, folderID).MediaID(mediaID).Execute()

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
	folderID := "folderID_example" // string | Folder ID
	mediaID := "mediaID_example" // string | Media ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FolderAPI.SetFolderCover(context.Background(), folderID).MediaID(mediaID).Execute()
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
**folderID** | **string** | Folder ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetFolderCoverRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **mediaID** | **string** | Media ID | 

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

