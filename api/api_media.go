/*
Weblens API

Programmatic access to the Weblens server

API version: 1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)


// MediaAPIService MediaAPI service
type MediaAPIService service

type ApiCleanupMediaRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
}

func (r ApiCleanupMediaRequest) Execute() (*http.Response, error) {
	return r.ApiService.CleanupMediaExecute(r)
}

/*
CleanupMedia Make sure all media is correctly synced with the file system

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiCleanupMediaRequest
*/
func (a *MediaAPIService) CleanupMedia(ctx context.Context) ApiCleanupMediaRequest {
	return ApiCleanupMediaRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
func (a *MediaAPIService) CleanupMediaExecute(r ApiCleanupMediaRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.CleanupMedia")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/cleanup"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiDropMediaRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
}

func (r ApiDropMediaRequest) Execute() (*http.Response, error) {
	return r.ApiService.DropMediaExecute(r)
}

/*
DropMedia DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiDropMediaRequest
*/
func (a *MediaAPIService) DropMedia(ctx context.Context) ApiDropMediaRequest {
	return ApiDropMediaRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
func (a *MediaAPIService) DropMediaExecute(r ApiDropMediaRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.DropMedia")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/drop"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiGetMediaRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	raw *bool
	hidden *bool
	sort *string
	search *string
	page *int32
	limit *int32
	folderIds *string
	mediaIds *string
}

// Include raw files
func (r ApiGetMediaRequest) Raw(raw bool) ApiGetMediaRequest {
	r.raw = &raw
	return r
}

// Include hidden media
func (r ApiGetMediaRequest) Hidden(hidden bool) ApiGetMediaRequest {
	r.hidden = &hidden
	return r
}

// Sort by field
func (r ApiGetMediaRequest) Sort(sort string) ApiGetMediaRequest {
	r.sort = &sort
	return r
}

// Search string
func (r ApiGetMediaRequest) Search(search string) ApiGetMediaRequest {
	r.search = &search
	return r
}

// Page of medias to get
func (r ApiGetMediaRequest) Page(page int32) ApiGetMediaRequest {
	r.page = &page
	return r
}

// Number of medias to get
func (r ApiGetMediaRequest) Limit(limit int32) ApiGetMediaRequest {
	r.limit = &limit
	return r
}

// Search only in given folders
func (r ApiGetMediaRequest) FolderIds(folderIds string) ApiGetMediaRequest {
	r.folderIds = &folderIds
	return r
}

// Get only media with the provided ids
func (r ApiGetMediaRequest) MediaIds(mediaIds string) ApiGetMediaRequest {
	r.mediaIds = &mediaIds
	return r
}

func (r ApiGetMediaRequest) Execute() (*MediaBatchInfo, *http.Response, error) {
	return r.ApiService.GetMediaExecute(r)
}

/*
GetMedia Get paginated media

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiGetMediaRequest
*/
func (a *MediaAPIService) GetMedia(ctx context.Context) ApiGetMediaRequest {
	return ApiGetMediaRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return MediaBatchInfo
func (a *MediaAPIService) GetMediaExecute(r ApiGetMediaRequest) (*MediaBatchInfo, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *MediaBatchInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.GetMedia")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.raw != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "raw", r.raw, "", "")
	} else {
		var defaultValue bool = false
		r.raw = &defaultValue
	}
	if r.hidden != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "hidden", r.hidden, "", "")
	} else {
		var defaultValue bool = false
		r.hidden = &defaultValue
	}
	if r.sort != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "sort", r.sort, "", "")
	} else {
		var defaultValue string = "createDate"
		r.sort = &defaultValue
	}
	if r.search != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "search", r.search, "", "")
	}
	if r.page != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "page", r.page, "", "")
	}
	if r.limit != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "limit", r.limit, "", "")
	}
	if r.folderIds != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "folderIds", r.folderIds, "", "")
	}
	if r.mediaIds != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "mediaIds", r.mediaIds, "", "")
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetMediaFileRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	mediaId string
}

func (r ApiGetMediaFileRequest) Execute() (*FileInfo, *http.Response, error) {
	return r.ApiService.GetMediaFileExecute(r)
}

/*
GetMediaFile Get file of media by id

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param mediaId Id of media
 @return ApiGetMediaFileRequest
*/
func (a *MediaAPIService) GetMediaFile(ctx context.Context, mediaId string) ApiGetMediaFileRequest {
	return ApiGetMediaFileRequest{
		ApiService: a,
		ctx: ctx,
		mediaId: mediaId,
	}
}

// Execute executes the request
//  @return FileInfo
func (a *MediaAPIService) GetMediaFileExecute(r ApiGetMediaFileRequest) (*FileInfo, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FileInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.GetMediaFile")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/{mediaId}/file"
	localVarPath = strings.Replace(localVarPath, "{"+"mediaId"+"}", url.PathEscape(parameterValueToString(r.mediaId, "mediaId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetMediaImageRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	mediaId string
	extension string
	quality *string
	page *int32
}

// Image Quality
func (r ApiGetMediaImageRequest) Quality(quality string) ApiGetMediaImageRequest {
	r.quality = &quality
	return r
}

// Page number
func (r ApiGetMediaImageRequest) Page(page int32) ApiGetMediaImageRequest {
	r.page = &page
	return r
}

func (r ApiGetMediaImageRequest) Execute() (string, *http.Response, error) {
	return r.ApiService.GetMediaImageExecute(r)
}

/*
GetMediaImage Get a media image bytes

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param mediaId Media Id
 @param extension Extension
 @return ApiGetMediaImageRequest
*/
func (a *MediaAPIService) GetMediaImage(ctx context.Context, mediaId string, extension string) ApiGetMediaImageRequest {
	return ApiGetMediaImageRequest{
		ApiService: a,
		ctx: ctx,
		mediaId: mediaId,
		extension: extension,
	}
}

// Execute executes the request
//  @return string
func (a *MediaAPIService) GetMediaImageExecute(r ApiGetMediaImageRequest) (string, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  string
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.GetMediaImage")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/{mediaId}.{extension}"
	localVarPath = strings.Replace(localVarPath, "{"+"mediaId"+"}", url.PathEscape(parameterValueToString(r.mediaId, "mediaId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"extension"+"}", url.PathEscape(parameterValueToString(r.extension, "extension")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.quality == nil {
		return localVarReturnValue, nil, reportError("quality is required and must be specified")
	}

	parameterAddToHeaderOrQuery(localVarQueryParams, "quality", r.quality, "", "")
	if r.page != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "page", r.page, "", "")
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"image/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetMediaInfoRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	mediaId string
}

func (r ApiGetMediaInfoRequest) Execute() (*MediaInfo, *http.Response, error) {
	return r.ApiService.GetMediaInfoExecute(r)
}

/*
GetMediaInfo Get media info

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param mediaId Media Id
 @return ApiGetMediaInfoRequest
*/
func (a *MediaAPIService) GetMediaInfo(ctx context.Context, mediaId string) ApiGetMediaInfoRequest {
	return ApiGetMediaInfoRequest{
		ApiService: a,
		ctx: ctx,
		mediaId: mediaId,
	}
}

// Execute executes the request
//  @return MediaInfo
func (a *MediaAPIService) GetMediaInfoExecute(r ApiGetMediaInfoRequest) (*MediaInfo, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *MediaInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.GetMediaInfo")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/{mediaId}/info"
	localVarPath = strings.Replace(localVarPath, "{"+"mediaId"+"}", url.PathEscape(parameterValueToString(r.mediaId, "mediaId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetMediaTypesRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
}

func (r ApiGetMediaTypesRequest) Execute() (*MediaTypesInfo, *http.Response, error) {
	return r.ApiService.GetMediaTypesExecute(r)
}

/*
GetMediaTypes Get media type dictionary

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiGetMediaTypesRequest
*/
func (a *MediaAPIService) GetMediaTypes(ctx context.Context) ApiGetMediaTypesRequest {
	return ApiGetMediaTypesRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return MediaTypesInfo
func (a *MediaAPIService) GetMediaTypesExecute(r ApiGetMediaTypesRequest) (*MediaTypesInfo, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *MediaTypesInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.GetMediaTypes")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/types"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetRandomMediaRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	count *float32
}

// Number of random medias to get
func (r ApiGetRandomMediaRequest) Count(count float32) ApiGetRandomMediaRequest {
	r.count = &count
	return r
}

func (r ApiGetRandomMediaRequest) Execute() (*MediaBatchInfo, *http.Response, error) {
	return r.ApiService.GetRandomMediaExecute(r)
}

/*
GetRandomMedia Get random media

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiGetRandomMediaRequest
*/
func (a *MediaAPIService) GetRandomMedia(ctx context.Context) ApiGetRandomMediaRequest {
	return ApiGetRandomMediaRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return MediaBatchInfo
func (a *MediaAPIService) GetRandomMediaExecute(r ApiGetRandomMediaRequest) (*MediaBatchInfo, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *MediaBatchInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.GetRandomMedia")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/random"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.count == nil {
		return localVarReturnValue, nil, reportError("count is required and must be specified")
	}

	parameterAddToHeaderOrQuery(localVarQueryParams, "count", r.count, "", "")
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSetMediaLikedRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	mediaId string
	liked *bool
	shareId *string
}

// Liked status to set
func (r ApiSetMediaLikedRequest) Liked(liked bool) ApiSetMediaLikedRequest {
	r.liked = &liked
	return r
}

// ShareId
func (r ApiSetMediaLikedRequest) ShareId(shareId string) ApiSetMediaLikedRequest {
	r.shareId = &shareId
	return r
}

func (r ApiSetMediaLikedRequest) Execute() (*http.Response, error) {
	return r.ApiService.SetMediaLikedExecute(r)
}

/*
SetMediaLiked Like a media

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param mediaId Id of media
 @return ApiSetMediaLikedRequest
*/
func (a *MediaAPIService) SetMediaLiked(ctx context.Context, mediaId string) ApiSetMediaLikedRequest {
	return ApiSetMediaLikedRequest{
		ApiService: a,
		ctx: ctx,
		mediaId: mediaId,
	}
}

// Execute executes the request
func (a *MediaAPIService) SetMediaLikedExecute(r ApiSetMediaLikedRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPatch
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.SetMediaLiked")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/{mediaId}/liked"
	localVarPath = strings.Replace(localVarPath, "{"+"mediaId"+"}", url.PathEscape(parameterValueToString(r.mediaId, "mediaId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.liked == nil {
		return nil, reportError("liked is required and must be specified")
	}

	if r.shareId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "shareId", r.shareId, "", "")
	}
	parameterAddToHeaderOrQuery(localVarQueryParams, "liked", r.liked, "", "")
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiSetMediaVisibilityRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	hidden *bool
	mediaIds *MediaIdsParams
}

// Set the media visibility
func (r ApiSetMediaVisibilityRequest) Hidden(hidden bool) ApiSetMediaVisibilityRequest {
	r.hidden = &hidden
	return r
}

// MediaIds to change visibility of
func (r ApiSetMediaVisibilityRequest) MediaIds(mediaIds MediaIdsParams) ApiSetMediaVisibilityRequest {
	r.mediaIds = &mediaIds
	return r
}

func (r ApiSetMediaVisibilityRequest) Execute() (*http.Response, error) {
	return r.ApiService.SetMediaVisibilityExecute(r)
}

/*
SetMediaVisibility Set media visibility

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiSetMediaVisibilityRequest
*/
func (a *MediaAPIService) SetMediaVisibility(ctx context.Context) ApiSetMediaVisibilityRequest {
	return ApiSetMediaVisibilityRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
func (a *MediaAPIService) SetMediaVisibilityExecute(r ApiSetMediaVisibilityRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPatch
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.SetMediaVisibility")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/visibility"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.hidden == nil {
		return nil, reportError("hidden is required and must be specified")
	}
	if r.mediaIds == nil {
		return nil, reportError("mediaIds is required and must be specified")
	}

	parameterAddToHeaderOrQuery(localVarQueryParams, "hidden", r.hidden, "", "")
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.mediaIds
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiStreamVideoRequest struct {
	ctx context.Context
	ApiService *MediaAPIService
	mediaId string
}

func (r ApiStreamVideoRequest) Execute() (*http.Response, error) {
	return r.ApiService.StreamVideoExecute(r)
}

/*
StreamVideo Stream a video

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param mediaId Id of media
 @return ApiStreamVideoRequest
*/
func (a *MediaAPIService) StreamVideo(ctx context.Context, mediaId string) ApiStreamVideoRequest {
	return ApiStreamVideoRequest{
		ApiService: a,
		ctx: ctx,
		mediaId: mediaId,
	}
}

// Execute executes the request
func (a *MediaAPIService) StreamVideoExecute(r ApiStreamVideoRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "MediaAPIService.StreamVideo")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/media/{mediaId}/video"
	localVarPath = strings.Replace(localVarPath, "{"+"mediaId"+"}", url.PathEscape(parameterValueToString(r.mediaId, "mediaId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}
