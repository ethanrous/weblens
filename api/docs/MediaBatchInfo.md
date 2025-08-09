# MediaBatchInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Media** | Pointer to [**[]MediaInfo**](MediaInfo.md) |  | [optional] 
**MediaCount** | Pointer to **int32** |  | [optional] 
**TotalMediaCount** | Pointer to **int32** |  | [optional] 

## Methods

### NewMediaBatchInfo

`func NewMediaBatchInfo() *MediaBatchInfo`

NewMediaBatchInfo instantiates a new MediaBatchInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMediaBatchInfoWithDefaults

`func NewMediaBatchInfoWithDefaults() *MediaBatchInfo`

NewMediaBatchInfoWithDefaults instantiates a new MediaBatchInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMedia

`func (o *MediaBatchInfo) GetMedia() []MediaInfo`

GetMedia returns the Media field if non-nil, zero value otherwise.

### GetMediaOk

`func (o *MediaBatchInfo) GetMediaOk() (*[]MediaInfo, bool)`

GetMediaOk returns a tuple with the Media field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMedia

`func (o *MediaBatchInfo) SetMedia(v []MediaInfo)`

SetMedia sets Media field to given value.

### HasMedia

`func (o *MediaBatchInfo) HasMedia() bool`

HasMedia returns a boolean if a field has been set.

### GetMediaCount

`func (o *MediaBatchInfo) GetMediaCount() int32`

GetMediaCount returns the MediaCount field if non-nil, zero value otherwise.

### GetMediaCountOk

`func (o *MediaBatchInfo) GetMediaCountOk() (*int32, bool)`

GetMediaCountOk returns a tuple with the MediaCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMediaCount

`func (o *MediaBatchInfo) SetMediaCount(v int32)`

SetMediaCount sets MediaCount field to given value.

### HasMediaCount

`func (o *MediaBatchInfo) HasMediaCount() bool`

HasMediaCount returns a boolean if a field has been set.

### GetTotalMediaCount

`func (o *MediaBatchInfo) GetTotalMediaCount() int32`

GetTotalMediaCount returns the TotalMediaCount field if non-nil, zero value otherwise.

### GetTotalMediaCountOk

`func (o *MediaBatchInfo) GetTotalMediaCountOk() (*int32, bool)`

GetTotalMediaCountOk returns a tuple with the TotalMediaCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalMediaCount

`func (o *MediaBatchInfo) SetTotalMediaCount(v int32)`

SetTotalMediaCount sets TotalMediaCount field to given value.

### HasTotalMediaCount

`func (o *MediaBatchInfo) HasTotalMediaCount() bool`

HasTotalMediaCount returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


