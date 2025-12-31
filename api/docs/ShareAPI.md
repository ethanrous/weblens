# \ShareAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddUserToShare**](ShareAPI.md#AddUserToShare) | **Post** /share/{shareID}/accessors | Add a user to a file share
[**CreateFileShare**](ShareAPI.md#CreateFileShare) | **Post** /share/file | Share a file
[**DeleteFileShare**](ShareAPI.md#DeleteFileShare) | **Delete** /share/{shareID} | Delete a file share
[**GetFileShare**](ShareAPI.md#GetFileShare) | **Get** /share/{shareID} | Get a file share
[**RemoveUserFromShare**](ShareAPI.md#RemoveUserFromShare) | **Delete** /share/{shareID}/accessors/{username} | Remove a user from a file share
[**SetSharePublic**](ShareAPI.md#SetSharePublic) | **Patch** /share/{shareID}/public | Update a share&#39;s \&quot;public\&quot; status
[**UpdateShareAccessorPermissions**](ShareAPI.md#UpdateShareAccessorPermissions) | **Patch** /share/{shareID}/accessors/{username} | Update a share&#39;s user permissions



## AddUserToShare

> ShareInfo AddUserToShare(ctx, shareID).Request(request).Execute()

Add a user to a file share

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
	shareID := "shareID_example" // string | Share ID
	request := *openapiclient.NewAddUserParams("Username_example") // AddUserParams | Share Accessors

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ShareAPI.AddUserToShare(context.Background(), shareID).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ShareAPI.AddUserToShare``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AddUserToShare`: ShareInfo
	fmt.Fprintf(os.Stdout, "Response from `ShareAPI.AddUserToShare`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**shareID** | **string** | Share ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddUserToShareRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**AddUserParams**](AddUserParams.md) | Share Accessors | 

### Return type

[**ShareInfo**](ShareInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateFileShare

> ShareInfo CreateFileShare(ctx).Request(request).Execute()

Share a file

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
	request := *openapiclient.NewFileShareParams() // FileShareParams | New File Share Params

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ShareAPI.CreateFileShare(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ShareAPI.CreateFileShare``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateFileShare`: ShareInfo
	fmt.Fprintf(os.Stdout, "Response from `ShareAPI.CreateFileShare`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateFileShareRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**FileShareParams**](FileShareParams.md) | New File Share Params | 

### Return type

[**ShareInfo**](ShareInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteFileShare

> DeleteFileShare(ctx, shareID).Execute()

Delete a file share

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
	shareID := "shareID_example" // string | Share ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ShareAPI.DeleteFileShare(context.Background(), shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ShareAPI.DeleteFileShare``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**shareID** | **string** | Share ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFileShareRequest struct via the builder pattern


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


## GetFileShare

> ShareInfo GetFileShare(ctx, shareID).Execute()

Get a file share

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
	shareID := "shareID_example" // string | Share ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ShareAPI.GetFileShare(context.Background(), shareID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ShareAPI.GetFileShare``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFileShare`: ShareInfo
	fmt.Fprintf(os.Stdout, "Response from `ShareAPI.GetFileShare`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**shareID** | **string** | Share ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFileShareRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ShareInfo**](ShareInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveUserFromShare

> ShareInfo RemoveUserFromShare(ctx, shareID, username).Execute()

Remove a user from a file share

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
	shareID := "shareID_example" // string | Share ID
	username := "username_example" // string | Username

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ShareAPI.RemoveUserFromShare(context.Background(), shareID, username).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ShareAPI.RemoveUserFromShare``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RemoveUserFromShare`: ShareInfo
	fmt.Fprintf(os.Stdout, "Response from `ShareAPI.RemoveUserFromShare`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**shareID** | **string** | Share ID | 
**username** | **string** | Username | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveUserFromShareRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ShareInfo**](ShareInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetSharePublic

> SetSharePublic(ctx, shareID).Public(public).Execute()

Update a share's \"public\" status

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
	shareID := "shareID_example" // string | Share ID
	public := true // bool | Share Public Status

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ShareAPI.SetSharePublic(context.Background(), shareID).Public(public).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ShareAPI.SetSharePublic``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**shareID** | **string** | Share ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetSharePublicRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **public** | **bool** | Share Public Status | 

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


## UpdateShareAccessorPermissions

> ShareInfo UpdateShareAccessorPermissions(ctx, shareID, username).Request(request).Execute()

Update a share's user permissions

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
	shareID := "shareID_example" // string | Share ID
	username := "username_example" // string | Username
	request := *openapiclient.NewPermissionsParams() // PermissionsParams | Share Permissions Params

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ShareAPI.UpdateShareAccessorPermissions(context.Background(), shareID, username).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ShareAPI.UpdateShareAccessorPermissions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateShareAccessorPermissions`: ShareInfo
	fmt.Fprintf(os.Stdout, "Response from `ShareAPI.UpdateShareAccessorPermissions`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**shareID** | **string** | Share ID | 
**username** | **string** | Username | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateShareAccessorPermissionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **request** | [**PermissionsParams**](PermissionsParams.md) | Share Permissions Params | 

### Return type

[**ShareInfo**](ShareInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

