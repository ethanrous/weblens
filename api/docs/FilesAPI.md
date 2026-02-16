# \FilesAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddFilesToUpload**](FilesAPI.md#AddFilesToUpload) | **Post** /upload/{uploadID} | Add a file to an upload task
[**AutocompletePath**](FilesAPI.md#AutocompletePath) | **Get** /files/autocomplete | Get path completion suggestions
[**ClearZipCache**](FilesAPI.md#ClearZipCache) | **Delete** /takeout | Clear all cached zip files
[**CreateTakeout**](FilesAPI.md#CreateTakeout) | **Post** /takeout | Create a zip file
[**DeleteFiles**](FilesAPI.md#DeleteFiles) | **Delete** /files | Delete Files \&quot;permanently\&quot;
[**DownloadFile**](FilesAPI.md#DownloadFile) | **Get** /files/{fileID}/download | Download a file
[**GetFile**](FilesAPI.md#GetFile) | **Get** /files/{fileID} | Get information about a file
[**GetFileStats**](FilesAPI.md#GetFileStats) | **Get** /files/{fileID}/stats | Get the statistics of a file
[**GetFileText**](FilesAPI.md#GetFileText) | **Get** /files/{fileID}/text | Get the text of a text file
[**GetSharedFiles**](FilesAPI.md#GetSharedFiles) | **Get** /files/shared | Get files shared with the logged in user
[**GetUploadResult**](FilesAPI.md#GetUploadResult) | **Get** /upload/{uploadID} | Get the result of an upload task. This will block until the upload is complete
[**MoveFiles**](FilesAPI.md#MoveFiles) | **Patch** /files | Move a list of files to a new parent folder
[**RestoreFiles**](FilesAPI.md#RestoreFiles) | **Post** /files/structsore | structsore files from some time in the past
[**SearchByFilename**](FilesAPI.md#SearchByFilename) | **Get** /files/search | Search for files by filename
[**StartUpload**](FilesAPI.md#StartUpload) | **Post** /upload | Begin a new upload task
[**UnTrashFiles**](FilesAPI.md#UnTrashFiles) | **Patch** /files/untrash | Move a list of files out of the trash, structsoring them to where they were before
[**UpdateFile**](FilesAPI.md#UpdateFile) | **Patch** /files/{fileID} | Update a File
[**UploadFileChunk**](FilesAPI.md#UploadFileChunk) | **Put** /upload/{uploadID}/file/{fileID} | Add a chunk to a file upload



## AddFilesToUpload

> NewFilesInfo AddFilesToUpload(ctx, uploadID).Request(request).ShareID(shareID).Execute()

Add a file to an upload task

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
	uploadID := "uploadID_example" // string | Upload ID
	request := *openapiclient.NewNewFilesParams() // NewFilesParams | New file params
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.AddFilesToUpload(context.Background(), uploadID).Request(request).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.AddFilesToUpload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AddFilesToUpload`: NewFilesInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.AddFilesToUpload`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uploadID** | **string** | Upload ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddFilesToUploadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**NewFilesParams**](NewFilesParams.md) | New file params | 
 **shareID** | **string** | Share ID | 

### Return type

[**NewFilesInfo**](NewFilesInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AutocompletePath

> FolderInfo AutocompletePath(ctx).SearchPath(searchPath).Execute()

Get path completion suggestions

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
	searchPath := "searchPath_example" // string | Search path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.AutocompletePath(context.Background()).SearchPath(searchPath).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.AutocompletePath``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AutocompletePath`: FolderInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.AutocompletePath`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAutocompletePathRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **searchPath** | **string** | Search path | 

### Return type

[**FolderInfo**](FolderInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ClearZipCache

> ClearZipCache(ctx).Execute()

Clear all cached zip files

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
	r, err := apiClient.FilesAPI.ClearZipCache(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.ClearZipCache``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiClearZipCacheRequest struct via the builder pattern


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


## CreateTakeout

> TakeoutInfo CreateTakeout(ctx).Request(request).ShareID(shareID).Execute()

Create a zip file



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
	request := *openapiclient.NewFilesListParams() // FilesListParams | File Ids
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.CreateTakeout(context.Background()).Request(request).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.CreateTakeout``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateTakeout`: TakeoutInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.CreateTakeout`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTakeoutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**FilesListParams**](FilesListParams.md) | File Ids | 
 **shareID** | **string** | Share ID | 

### Return type

[**TakeoutInfo**](TakeoutInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteFiles

> DeleteFiles(ctx).Request(request).IgnoreTrash(ignoreTrash).PreserveFolder(preserveFolder).Execute()

Delete Files \"permanently\"

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
	request := *openapiclient.NewFilesListParams() // FilesListParams | Delete files request body
	ignoreTrash := true // bool | Delete files even if they are not in the trash (optional)
	preserveFolder := true // bool | Preserve parent folder if it is empty after deletion (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FilesAPI.DeleteFiles(context.Background()).Request(request).IgnoreTrash(ignoreTrash).PreserveFolder(preserveFolder).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.DeleteFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**FilesListParams**](FilesListParams.md) | Delete files request body | 
 **ignoreTrash** | **bool** | Delete files even if they are not in the trash | 
 **preserveFolder** | **bool** | Preserve parent folder if it is empty after deletion | 

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


## DownloadFile

> string DownloadFile(ctx, fileID).ShareID(shareID).Format(format).Quality(quality).IsTakeout(isTakeout).Execute()

Download a file

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
	shareID := "shareID_example" // string | Share ID (optional)
	format := "format_example" // string | File format conversion (optional)
	quality := int32(56) // int32 | JPEG quality (1-100) (optional) (default to 85)
	isTakeout := true // bool | Is this a takeout file (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.DownloadFile(context.Background(), fileID).ShareID(shareID).Format(format).Quality(quality).IsTakeout(isTakeout).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.DownloadFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DownloadFile`: string
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.DownloadFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDownloadFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **shareID** | **string** | Share ID | 
 **format** | **string** | File format conversion | 
 **quality** | **int32** | JPEG quality (1-100) | [default to 85]
 **isTakeout** | **bool** | Is this a takeout file | [default to false]

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFile

> FileInfo GetFile(ctx, fileID).ShareID(shareID).Execute()

Get information about a file

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
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.GetFile(context.Background(), fileID).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.GetFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFile`: FileInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.GetFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **shareID** | **string** | Share ID | 

### Return type

[**FileInfo**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFileStats

> GetFileStats(ctx, fileID).Execute()

Get the statistics of a file

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
	r, err := apiClient.FilesAPI.GetFileStats(context.Background(), fileID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.GetFileStats``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFileStatsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## GetFileText

> string GetFileText(ctx, fileID).ShareID(shareID).Execute()

Get the text of a text file

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
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.GetFileText(context.Background(), fileID).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.GetFileText``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFileText`: string
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.GetFileText`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFileTextRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **shareID** | **string** | Share ID | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSharedFiles

> FolderInfo GetSharedFiles(ctx).Execute()

Get files shared with the logged in user

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
	resp, r, err := apiClient.FilesAPI.GetSharedFiles(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.GetSharedFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSharedFiles`: FolderInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.GetSharedFiles`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetSharedFilesRequest struct via the builder pattern


### Return type

[**FolderInfo**](FolderInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUploadResult

> GetUploadResult(ctx, uploadID).Execute()

Get the result of an upload task. This will block until the upload is complete

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
	uploadID := "uploadID_example" // string | Upload ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FilesAPI.GetUploadResult(context.Background(), uploadID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.GetUploadResult``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uploadID** | **string** | Upload ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUploadResultRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## MoveFiles

> MoveFiles(ctx).Request(request).ShareID(shareID).Execute()

Move a list of files to a new parent folder

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
	request := *openapiclient.NewMoveFilesParams() // MoveFilesParams | Move files request body
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FilesAPI.MoveFiles(context.Background()).Request(request).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.MoveFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMoveFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**MoveFilesParams**](MoveFilesParams.md) | Move files request body | 
 **shareID** | **string** | Share ID | 

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


## RestoreFiles

> RestoreFilesInfo RestoreFiles(ctx).Request(request).Execute()

structsore files from some time in the past

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
	request := *openapiclient.NewRestoreFilesBody() // RestoreFilesBody | RestoreFiles files request body

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.RestoreFiles(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.RestoreFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RestoreFiles`: RestoreFilesInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.RestoreFiles`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRestoreFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**RestoreFilesBody**](RestoreFilesBody.md) | RestoreFiles files request body | 

### Return type

[**RestoreFilesInfo**](RestoreFilesInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SearchByFilename

> []FileInfo SearchByFilename(ctx).Search(search).BaseFolderID(baseFolderID).SortProp(sortProp).SortOrder(sortOrder).Recursive(recursive).Regex(regex).Execute()

Search for files by filename

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
	search := "search_example" // string | Filename to search for
	baseFolderID := "baseFolderID_example" // string | The folder to search in, defaults to the user's home folder (optional)
	sortProp := "sortProp_example" // string | Property to sort by (optional) (default to "name")
	sortOrder := "sortOrder_example" // string | Sort order (optional) (default to "asc")
	recursive := true // bool | Search recursively (optional) (default to false)
	regex := true // bool | Whether to treat the search term as a regex pattern (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.SearchByFilename(context.Background()).Search(search).BaseFolderID(baseFolderID).SortProp(sortProp).SortOrder(sortOrder).Recursive(recursive).Regex(regex).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.SearchByFilename``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchByFilename`: []FileInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.SearchByFilename`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSearchByFilenameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **search** | **string** | Filename to search for | 
 **baseFolderID** | **string** | The folder to search in, defaults to the user&#39;s home folder | 
 **sortProp** | **string** | Property to sort by | [default to &quot;name&quot;]
 **sortOrder** | **string** | Sort order | [default to &quot;asc&quot;]
 **recursive** | **bool** | Search recursively | [default to false]
 **regex** | **bool** | Whether to treat the search term as a regex pattern | [default to false]

### Return type

[**[]FileInfo**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StartUpload

> NewUploadInfo StartUpload(ctx).Request(request).ShareID(shareID).Execute()

Begin a new upload task

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
	request := *openapiclient.NewNewUploadParams() // NewUploadParams | New upload request body
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FilesAPI.StartUpload(context.Background()).Request(request).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.StartUpload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `StartUpload`: NewUploadInfo
	fmt.Fprintf(os.Stdout, "Response from `FilesAPI.StartUpload`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiStartUploadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**NewUploadParams**](NewUploadParams.md) | New upload request body | 
 **shareID** | **string** | Share ID | 

### Return type

[**NewUploadInfo**](NewUploadInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UnTrashFiles

> UnTrashFiles(ctx).Request(request).Execute()

Move a list of files out of the trash, structsoring them to where they were before

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
	request := *openapiclient.NewFilesListParams() // FilesListParams | Un-trash files request body

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FilesAPI.UnTrashFiles(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.UnTrashFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUnTrashFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**FilesListParams**](FilesListParams.md) | Un-trash files request body | 

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


## UpdateFile

> UpdateFile(ctx, fileID).Request(request).ShareID(shareID).Execute()

Update a File

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
	request := *openapiclient.NewUpdateFileParams() // UpdateFileParams | Update file request body
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FilesAPI.UpdateFile(context.Background(), fileID).Request(request).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.UpdateFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**UpdateFileParams**](UpdateFileParams.md) | Update file request body | 
 **shareID** | **string** | Share ID | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadFileChunk

> UploadFileChunk(ctx, uploadID, fileID).Chunk(chunk).ShareID(shareID).Execute()

Add a chunk to a file upload

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
	uploadID := "uploadID_example" // string | Upload ID
	fileID := "fileID_example" // string | File ID
	chunk := os.NewFile(1234, "some_file") // *os.File | File chunk
	shareID := "shareID_example" // string | Share ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FilesAPI.UploadFileChunk(context.Background(), uploadID, fileID).Chunk(chunk).ShareID(shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesAPI.UploadFileChunk``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uploadID** | **string** | Upload ID | 
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadFileChunkRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **chunk** | ***os.File** | File chunk | 
 **shareID** | **string** | Share ID | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

