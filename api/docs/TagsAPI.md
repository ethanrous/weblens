# \TagsAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddFilesToTag**](TagsAPI.md#AddFilesToTag) | **Post** /tags/{tagID}/files | Add files to a tag
[**CreateTag**](TagsAPI.md#CreateTag) | **Post** /tags | Create a new tag
[**DeleteTag**](TagsAPI.md#DeleteTag) | **Delete** /tags/{tagID} | Delete a tag
[**GetFilesByTag**](TagsAPI.md#GetFilesByTag) | **Get** /tags/{tagID}/files | Get all files in a tag
[**GetTag**](TagsAPI.md#GetTag) | **Get** /tags/{tagID} | Get a tag by ID
[**GetTagsForFile**](TagsAPI.md#GetTagsForFile) | **Get** /tags/file/{fileID} | Get tags for a file
[**GetUserTags**](TagsAPI.md#GetUserTags) | **Get** /tags | Get all tags for the authenticated user
[**RemoveFilesFromTag**](TagsAPI.md#RemoveFilesFromTag) | **Delete** /tags/{tagID}/files | Remove files from a tag
[**UpdateTag**](TagsAPI.md#UpdateTag) | **Patch** /tags/{tagID} | Update a tag&#39;s name and/or color



## AddFilesToTag

> AddFilesToTag(ctx, tagID).Request(request).Execute()

Add files to a tag

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
	tagID := "tagID_example" // string | Tag ID
	request := *openapiclient.NewTagFileIDsParams() // TagFileIDsParams | File IDs to add

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TagsAPI.AddFilesToTag(context.Background(), tagID).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.AddFilesToTag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tagID** | **string** | Tag ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddFilesToTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**TagFileIDsParams**](TagFileIDsParams.md) | File IDs to add | 

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


## CreateTag

> GithubComEthanrousWeblensModelsTagTag CreateTag(ctx).Request(request).Execute()

Create a new tag

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
	request := *openapiclient.NewTagCreateTagParams() // TagCreateTagParams | Create tag request body

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TagsAPI.CreateTag(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.CreateTag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateTag`: GithubComEthanrousWeblensModelsTagTag
	fmt.Fprintf(os.Stdout, "Response from `TagsAPI.CreateTag`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**TagCreateTagParams**](TagCreateTagParams.md) | Create tag request body | 

### Return type

[**GithubComEthanrousWeblensModelsTagTag**](GithubComEthanrousWeblensModelsTagTag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteTag

> DeleteTag(ctx, tagID).Execute()

Delete a tag

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
	tagID := "tagID_example" // string | Tag ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TagsAPI.DeleteTag(context.Background(), tagID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.DeleteTag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tagID** | **string** | Tag ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTagRequest struct via the builder pattern


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


## GetFilesByTag

> []FileInfo GetFilesByTag(ctx, tagID).Execute()

Get all files in a tag

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
	tagID := "tagID_example" // string | Tag ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TagsAPI.GetFilesByTag(context.Background(), tagID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.GetFilesByTag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFilesByTag`: []FileInfo
	fmt.Fprintf(os.Stdout, "Response from `TagsAPI.GetFilesByTag`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tagID** | **string** | Tag ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFilesByTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]FileInfo**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTag

> GithubComEthanrousWeblensModelsTagTag GetTag(ctx, tagID).Execute()

Get a tag by ID

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
	tagID := "tagID_example" // string | Tag ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TagsAPI.GetTag(context.Background(), tagID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.GetTag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTag`: GithubComEthanrousWeblensModelsTagTag
	fmt.Fprintf(os.Stdout, "Response from `TagsAPI.GetTag`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tagID** | **string** | Tag ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GithubComEthanrousWeblensModelsTagTag**](GithubComEthanrousWeblensModelsTagTag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTagsForFile

> []GithubComEthanrousWeblensModelsTagTag GetTagsForFile(ctx, fileID).Execute()

Get tags for a file

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
	resp, r, err := apiClient.TagsAPI.GetTagsForFile(context.Background(), fileID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.GetTagsForFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTagsForFile`: []GithubComEthanrousWeblensModelsTagTag
	fmt.Fprintf(os.Stdout, "Response from `TagsAPI.GetTagsForFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fileID** | **string** | File ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTagsForFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]GithubComEthanrousWeblensModelsTagTag**](GithubComEthanrousWeblensModelsTagTag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUserTags

> []GithubComEthanrousWeblensModelsTagTag GetUserTags(ctx).Execute()

Get all tags for the authenticated user

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
	resp, r, err := apiClient.TagsAPI.GetUserTags(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.GetUserTags``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUserTags`: []GithubComEthanrousWeblensModelsTagTag
	fmt.Fprintf(os.Stdout, "Response from `TagsAPI.GetUserTags`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserTagsRequest struct via the builder pattern


### Return type

[**[]GithubComEthanrousWeblensModelsTagTag**](GithubComEthanrousWeblensModelsTagTag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveFilesFromTag

> RemoveFilesFromTag(ctx, tagID).Request(request).Execute()

Remove files from a tag

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
	tagID := "tagID_example" // string | Tag ID
	request := *openapiclient.NewTagFileIDsParams() // TagFileIDsParams | File IDs to remove

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TagsAPI.RemoveFilesFromTag(context.Background(), tagID).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.RemoveFilesFromTag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tagID** | **string** | Tag ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveFilesFromTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**TagFileIDsParams**](TagFileIDsParams.md) | File IDs to remove | 

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


## UpdateTag

> UpdateTag(ctx, tagID).Request(request).Execute()

Update a tag's name and/or color

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
	tagID := "tagID_example" // string | Tag ID
	request := *openapiclient.NewTagUpdateTagParams() // TagUpdateTagParams | Update tag request body

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TagsAPI.UpdateTag(context.Background(), tagID).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TagsAPI.UpdateTag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tagID** | **string** | Tag ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTagRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**TagUpdateTagParams**](TagUpdateTagParams.md) | Update tag request body | 

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

