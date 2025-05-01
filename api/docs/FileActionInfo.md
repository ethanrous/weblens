# FileActionInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionType** | **string** |  | 
**DestinationPath** | Pointer to **string** |  | [optional] 
**EventId** | **string** |  | 
**FileId** | **string** |  | 
**Filepath** | Pointer to **string** |  | [optional] 
**OriginPath** | Pointer to **string** |  | [optional] 
**ParentId** | **string** |  | 
**Size** | **int64** |  | 
**Timestamp** | **int64** |  | 
**TowerId** | **string** |  | 

## Methods

### NewFileActionInfo

`func NewFileActionInfo(actionType string, eventId string, fileId string, parentId string, size int64, timestamp int64, towerId string, ) *FileActionInfo`

NewFileActionInfo instantiates a new FileActionInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFileActionInfoWithDefaults

`func NewFileActionInfoWithDefaults() *FileActionInfo`

NewFileActionInfoWithDefaults instantiates a new FileActionInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActionType

`func (o *FileActionInfo) GetActionType() string`

GetActionType returns the ActionType field if non-nil, zero value otherwise.

### GetActionTypeOk

`func (o *FileActionInfo) GetActionTypeOk() (*string, bool)`

GetActionTypeOk returns a tuple with the ActionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionType

`func (o *FileActionInfo) SetActionType(v string)`

SetActionType sets ActionType field to given value.


### GetDestinationPath

`func (o *FileActionInfo) GetDestinationPath() string`

GetDestinationPath returns the DestinationPath field if non-nil, zero value otherwise.

### GetDestinationPathOk

`func (o *FileActionInfo) GetDestinationPathOk() (*string, bool)`

GetDestinationPathOk returns a tuple with the DestinationPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDestinationPath

`func (o *FileActionInfo) SetDestinationPath(v string)`

SetDestinationPath sets DestinationPath field to given value.

### HasDestinationPath

`func (o *FileActionInfo) HasDestinationPath() bool`

HasDestinationPath returns a boolean if a field has been set.

### GetEventId

`func (o *FileActionInfo) GetEventId() string`

GetEventId returns the EventId field if non-nil, zero value otherwise.

### GetEventIdOk

`func (o *FileActionInfo) GetEventIdOk() (*string, bool)`

GetEventIdOk returns a tuple with the EventId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventId

`func (o *FileActionInfo) SetEventId(v string)`

SetEventId sets EventId field to given value.


### GetFileId

`func (o *FileActionInfo) GetFileId() string`

GetFileId returns the FileId field if non-nil, zero value otherwise.

### GetFileIdOk

`func (o *FileActionInfo) GetFileIdOk() (*string, bool)`

GetFileIdOk returns a tuple with the FileId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileId

`func (o *FileActionInfo) SetFileId(v string)`

SetFileId sets FileId field to given value.


### GetFilepath

`func (o *FileActionInfo) GetFilepath() string`

GetFilepath returns the Filepath field if non-nil, zero value otherwise.

### GetFilepathOk

`func (o *FileActionInfo) GetFilepathOk() (*string, bool)`

GetFilepathOk returns a tuple with the Filepath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilepath

`func (o *FileActionInfo) SetFilepath(v string)`

SetFilepath sets Filepath field to given value.

### HasFilepath

`func (o *FileActionInfo) HasFilepath() bool`

HasFilepath returns a boolean if a field has been set.

### GetOriginPath

`func (o *FileActionInfo) GetOriginPath() string`

GetOriginPath returns the OriginPath field if non-nil, zero value otherwise.

### GetOriginPathOk

`func (o *FileActionInfo) GetOriginPathOk() (*string, bool)`

GetOriginPathOk returns a tuple with the OriginPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPath

`func (o *FileActionInfo) SetOriginPath(v string)`

SetOriginPath sets OriginPath field to given value.

### HasOriginPath

`func (o *FileActionInfo) HasOriginPath() bool`

HasOriginPath returns a boolean if a field has been set.

### GetParentId

`func (o *FileActionInfo) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *FileActionInfo) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *FileActionInfo) SetParentId(v string)`

SetParentId sets ParentId field to given value.


### GetSize

`func (o *FileActionInfo) GetSize() int64`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *FileActionInfo) GetSizeOk() (*int64, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *FileActionInfo) SetSize(v int64)`

SetSize sets Size field to given value.


### GetTimestamp

`func (o *FileActionInfo) GetTimestamp() int64`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *FileActionInfo) GetTimestampOk() (*int64, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *FileActionInfo) SetTimestamp(v int64)`

SetTimestamp sets Timestamp field to given value.


### GetTowerId

`func (o *FileActionInfo) GetTowerId() string`

GetTowerId returns the TowerId field if non-nil, zero value otherwise.

### GetTowerIdOk

`func (o *FileActionInfo) GetTowerIdOk() (*string, bool)`

GetTowerIdOk returns a tuple with the TowerId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTowerId

`func (o *FileActionInfo) SetTowerId(v string)`

SetTowerId sets TowerId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


