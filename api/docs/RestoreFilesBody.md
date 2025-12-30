# RestoreFilesBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FileIDs** | Pointer to **[]string** |  | [optional] 
**NewParentID** | Pointer to **string** |  | [optional] 
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

### GetFileIDs

`func (o *RestoreFilesBody) GetFileIDs() []string`

GetFileIDs returns the FileIDs field if non-nil, zero value otherwise.

### GetFileIDsOk

`func (o *RestoreFilesBody) GetFileIDsOk() (*[]string, bool)`

GetFileIDsOk returns a tuple with the FileIDs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileIDs

`func (o *RestoreFilesBody) SetFileIDs(v []string)`

SetFileIDs sets FileIDs field to given value.

### HasFileIDs

`func (o *RestoreFilesBody) HasFileIDs() bool`

HasFileIDs returns a boolean if a field has been set.

### GetNewParentID

`func (o *RestoreFilesBody) GetNewParentID() string`

GetNewParentID returns the NewParentID field if non-nil, zero value otherwise.

### GetNewParentIDOk

`func (o *RestoreFilesBody) GetNewParentIDOk() (*string, bool)`

GetNewParentIDOk returns a tuple with the NewParentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewParentID

`func (o *RestoreFilesBody) SetNewParentID(v string)`

SetNewParentID sets NewParentID field to given value.

### HasNewParentID

`func (o *RestoreFilesBody) HasNewParentID() bool`

HasNewParentID returns a boolean if a field has been set.

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


