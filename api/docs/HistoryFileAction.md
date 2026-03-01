# HistoryFileAction

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionType** | Pointer to **string** |  | [optional] 
**ContentID** | Pointer to **string** |  | [optional] 
**DestinationPath** | Pointer to [**WlfsFilepath**](WlfsFilepath.md) |  | [optional] 
**Doer** | Pointer to **string** | The user or system that performed the action | [optional] 
**EventID** | Pointer to **string** |  | [optional] 
**FileID** | Pointer to **string** |  | [optional] 
**Filepath** | Pointer to [**WlfsFilepath**](WlfsFilepath.md) |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**OriginPath** | Pointer to [**WlfsFilepath**](WlfsFilepath.md) |  | [optional] 
**Size** | Pointer to **int32** |  | [optional] 
**Timestamp** | Pointer to **string** |  | [optional] 
**TowerID** | Pointer to **string** |  | [optional] 

## Methods

### NewHistoryFileAction

`func NewHistoryFileAction() *HistoryFileAction`

NewHistoryFileAction instantiates a new HistoryFileAction object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHistoryFileActionWithDefaults

`func NewHistoryFileActionWithDefaults() *HistoryFileAction`

NewHistoryFileActionWithDefaults instantiates a new HistoryFileAction object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActionType

`func (o *HistoryFileAction) GetActionType() string`

GetActionType returns the ActionType field if non-nil, zero value otherwise.

### GetActionTypeOk

`func (o *HistoryFileAction) GetActionTypeOk() (*string, bool)`

GetActionTypeOk returns a tuple with the ActionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionType

`func (o *HistoryFileAction) SetActionType(v string)`

SetActionType sets ActionType field to given value.

### HasActionType

`func (o *HistoryFileAction) HasActionType() bool`

HasActionType returns a boolean if a field has been set.

### GetContentID

`func (o *HistoryFileAction) GetContentID() string`

GetContentID returns the ContentID field if non-nil, zero value otherwise.

### GetContentIDOk

`func (o *HistoryFileAction) GetContentIDOk() (*string, bool)`

GetContentIDOk returns a tuple with the ContentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentID

`func (o *HistoryFileAction) SetContentID(v string)`

SetContentID sets ContentID field to given value.

### HasContentID

`func (o *HistoryFileAction) HasContentID() bool`

HasContentID returns a boolean if a field has been set.

### GetDestinationPath

`func (o *HistoryFileAction) GetDestinationPath() WlfsFilepath`

GetDestinationPath returns the DestinationPath field if non-nil, zero value otherwise.

### GetDestinationPathOk

`func (o *HistoryFileAction) GetDestinationPathOk() (*WlfsFilepath, bool)`

GetDestinationPathOk returns a tuple with the DestinationPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDestinationPath

`func (o *HistoryFileAction) SetDestinationPath(v WlfsFilepath)`

SetDestinationPath sets DestinationPath field to given value.

### HasDestinationPath

`func (o *HistoryFileAction) HasDestinationPath() bool`

HasDestinationPath returns a boolean if a field has been set.

### GetDoer

`func (o *HistoryFileAction) GetDoer() string`

GetDoer returns the Doer field if non-nil, zero value otherwise.

### GetDoerOk

`func (o *HistoryFileAction) GetDoerOk() (*string, bool)`

GetDoerOk returns a tuple with the Doer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDoer

`func (o *HistoryFileAction) SetDoer(v string)`

SetDoer sets Doer field to given value.

### HasDoer

`func (o *HistoryFileAction) HasDoer() bool`

HasDoer returns a boolean if a field has been set.

### GetEventID

`func (o *HistoryFileAction) GetEventID() string`

GetEventID returns the EventID field if non-nil, zero value otherwise.

### GetEventIDOk

`func (o *HistoryFileAction) GetEventIDOk() (*string, bool)`

GetEventIDOk returns a tuple with the EventID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventID

`func (o *HistoryFileAction) SetEventID(v string)`

SetEventID sets EventID field to given value.

### HasEventID

`func (o *HistoryFileAction) HasEventID() bool`

HasEventID returns a boolean if a field has been set.

### GetFileID

`func (o *HistoryFileAction) GetFileID() string`

GetFileID returns the FileID field if non-nil, zero value otherwise.

### GetFileIDOk

`func (o *HistoryFileAction) GetFileIDOk() (*string, bool)`

GetFileIDOk returns a tuple with the FileID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileID

`func (o *HistoryFileAction) SetFileID(v string)`

SetFileID sets FileID field to given value.

### HasFileID

`func (o *HistoryFileAction) HasFileID() bool`

HasFileID returns a boolean if a field has been set.

### GetFilepath

`func (o *HistoryFileAction) GetFilepath() WlfsFilepath`

GetFilepath returns the Filepath field if non-nil, zero value otherwise.

### GetFilepathOk

`func (o *HistoryFileAction) GetFilepathOk() (*WlfsFilepath, bool)`

GetFilepathOk returns a tuple with the Filepath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilepath

`func (o *HistoryFileAction) SetFilepath(v WlfsFilepath)`

SetFilepath sets Filepath field to given value.

### HasFilepath

`func (o *HistoryFileAction) HasFilepath() bool`

HasFilepath returns a boolean if a field has been set.

### GetId

`func (o *HistoryFileAction) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *HistoryFileAction) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *HistoryFileAction) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *HistoryFileAction) HasId() bool`

HasId returns a boolean if a field has been set.

### GetOriginPath

`func (o *HistoryFileAction) GetOriginPath() WlfsFilepath`

GetOriginPath returns the OriginPath field if non-nil, zero value otherwise.

### GetOriginPathOk

`func (o *HistoryFileAction) GetOriginPathOk() (*WlfsFilepath, bool)`

GetOriginPathOk returns a tuple with the OriginPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPath

`func (o *HistoryFileAction) SetOriginPath(v WlfsFilepath)`

SetOriginPath sets OriginPath field to given value.

### HasOriginPath

`func (o *HistoryFileAction) HasOriginPath() bool`

HasOriginPath returns a boolean if a field has been set.

### GetSize

`func (o *HistoryFileAction) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *HistoryFileAction) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *HistoryFileAction) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *HistoryFileAction) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetTimestamp

`func (o *HistoryFileAction) GetTimestamp() string`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *HistoryFileAction) GetTimestampOk() (*string, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *HistoryFileAction) SetTimestamp(v string)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *HistoryFileAction) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.

### GetTowerID

`func (o *HistoryFileAction) GetTowerID() string`

GetTowerID returns the TowerID field if non-nil, zero value otherwise.

### GetTowerIDOk

`func (o *HistoryFileAction) GetTowerIDOk() (*string, bool)`

GetTowerIDOk returns a tuple with the TowerID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTowerID

`func (o *HistoryFileAction) SetTowerID(v string)`

SetTowerID sets TowerID field to given value.

### HasTowerID

`func (o *HistoryFileAction) HasTowerID() bool`

HasTowerID returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


