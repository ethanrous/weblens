# \FeatureFlagsAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetFlags**](FeatureFlagsAPI.md#GetFlags) | **Get** /flags | Get Feature Flags
[**SetFlags**](FeatureFlagsAPI.md#SetFlags) | **Post** /flags | Set Feature Flags



## GetFlags

> Bundle GetFlags(ctx).Execute()

Get Feature Flags

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
	resp, r, err := apiClient.FeatureFlagsAPI.GetFlags(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FeatureFlagsAPI.GetFlags``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFlags`: Bundle
	fmt.Fprintf(os.Stdout, "Response from `FeatureFlagsAPI.GetFlags`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetFlagsRequest struct via the builder pattern


### Return type

[**Bundle**](Bundle.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetFlags

> SetFlags(ctx).Request(request).Execute()

Set Feature Flags

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
	request := []openapiclient.StructsSetConfigParam{*openapiclient.NewStructsSetConfigParam()} // []StructsSetConfigParam | Feature Flag Params

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FeatureFlagsAPI.SetFlags(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FeatureFlagsAPI.SetFlags``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetFlagsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**[]StructsSetConfigParam**](StructsSetConfigParam.md) | Feature Flag Params | 

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

