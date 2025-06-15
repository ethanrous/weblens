# RestoreFilesBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FileIds** | Pointer to **[]string** |  | [optional] 
**NewParentId** | Pointer to **string** |  | [optional] 
**Timestamp** | Pointer to **int32** |  | [optional] 

## Methods

### NewRestoreFilesBody

`func NewRestoreFilesBody() *RestoreFilesBody`

NewRestoreFilesBody instantiates a new RestoreFilesBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRestoreFilesBodyWithDefaults

`func NewRestoreFilesBodyWithDefaults() *RestoreFilesBody`

NewRestoreFilesBodyWithDefaults instantiates a new RestoreFilesBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFileIds

`func (o *RestoreFilesBody) GetFileIds() []string`

GetFileIds returns the FileIds field if non-nil, zero value otherwise.

### GetFileIdsOk

`func (o *RestoreFilesBody) GetFileIdsOk() (*[]string, bool)`

GetFileIdsOk returns a tuple with the FileIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileIds

`func (o *RestoreFilesBody) SetFileIds(v []string)`

SetFileIds sets FileIds field to given value.

### HasFileIds

`func (o *RestoreFilesBody) HasFileIds() bool`

HasFileIds returns a boolean if a field has been set.

### GetNewParentId

`func (o *RestoreFilesBody) GetNewParentId() string`

GetNewParentId returns the NewParentId field if non-nil, zero value otherwise.

### GetNewParentIdOk

`func (o *RestoreFilesBody) GetNewParentIdOk() (*string, bool)`

GetNewParentIdOk returns a tuple with the NewParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewParentId

`func (o *RestoreFilesBody) SetNewParentId(v string)`

SetNewParentId sets NewParentId field to given value.

### HasNewParentId

`func (o *RestoreFilesBody) HasNewParentId() bool`

HasNewParentId returns a boolean if a field has been set.

### GetTimestamp

`func (o *RestoreFilesBody) GetTimestamp() int32`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *RestoreFilesBody) GetTimestampOk() (*int32, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *RestoreFilesBody) SetTimestamp(v int32)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *RestoreFilesBody) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


