# FileActionInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionType** | **string** |  | 
**ContentID** | Pointer to **string** |  | [optional] 
**DestinationPath** | Pointer to **string** |  | [optional] 
**EventID** | **string** |  | 
**FileID** | **string** |  | 
**Filepath** | Pointer to **string** |  | [optional] 
**OriginPath** | Pointer to **string** |  | [optional] 
**ParentID** | **string** |  | 
**Size** | **int64** |  | 
**Timestamp** | **int64** |  | 
**TowerID** | **string** |  | 

## Methods

### NewFileActionInfo

`func NewFileActionInfo(actionType string, eventID string, fileID string, parentID string, size int64, timestamp int64, towerID string, ) *FileActionInfo`

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


### GetContentID

`func (o *FileActionInfo) GetContentID() string`

GetContentID returns the ContentID field if non-nil, zero value otherwise.

### GetContentIDOk

`func (o *FileActionInfo) GetContentIDOk() (*string, bool)`

GetContentIDOk returns a tuple with the ContentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentID

`func (o *FileActionInfo) SetContentID(v string)`

SetContentID sets ContentID field to given value.

### HasContentID

`func (o *FileActionInfo) HasContentID() bool`

HasContentID returns a boolean if a field has been set.

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

### GetEventID

`func (o *FileActionInfo) GetEventID() string`

GetEventID returns the EventID field if non-nil, zero value otherwise.

### GetEventIDOk

`func (o *FileActionInfo) GetEventIDOk() (*string, bool)`

GetEventIDOk returns a tuple with the EventID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventID

`func (o *FileActionInfo) SetEventID(v string)`

SetEventID sets EventID field to given value.


### GetFileID

`func (o *FileActionInfo) GetFileID() string`

GetFileID returns the FileID field if non-nil, zero value otherwise.

### GetFileIDOk

`func (o *FileActionInfo) GetFileIDOk() (*string, bool)`

GetFileIDOk returns a tuple with the FileID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileID

`func (o *FileActionInfo) SetFileID(v string)`

SetFileID sets FileID field to given value.


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

### GetParentID

`func (o *FileActionInfo) GetParentID() string`

GetParentID returns the ParentID field if non-nil, zero value otherwise.

### GetParentIDOk

`func (o *FileActionInfo) GetParentIDOk() (*string, bool)`

GetParentIDOk returns a tuple with the ParentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentID

`func (o *FileActionInfo) SetParentID(v string)`

SetParentID sets ParentID field to given value.


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


### GetTowerID

`func (o *FileActionInfo) GetTowerID() string`

GetTowerID returns the TowerID field if non-nil, zero value otherwise.

### GetTowerIDOk

`func (o *FileActionInfo) GetTowerIDOk() (*string, bool)`

GetTowerIDOk returns a tuple with the TowerID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTowerID

`func (o *FileActionInfo) SetTowerID(v string)`

SetTowerID sets TowerID field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


