# \ServersAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LaunchBackup**](ServersAPI.md#LaunchBackup) | **Post** /servers/{serverID}/backup | Launch backup on a server



## LaunchBackup

> LaunchBackup(ctx, serverID).Execute()

Launch backup on a server

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
	serverID := "serverID_example" // string | Server ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ServersAPI.LaunchBackup(context.Background(), serverID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServersAPI.LaunchBackup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**serverID** | **string** | Server ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiLaunchBackupRequest struct via the builder pattern


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

