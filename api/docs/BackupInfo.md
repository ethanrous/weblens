# BackupInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FileHistory** | Pointer to [**[]FileActionInfo**](FileActionInfo.md) |  | [optional] 
**Instances** | Pointer to [**[]TowerInfo**](TowerInfo.md) |  | [optional] 
**LifetimesCount** | Pointer to **int32** |  | [optional] 
**Tokens** | Pointer to [**[]TokenInfo**](TokenInfo.md) |  | [optional] 
**Users** | Pointer to [**[]UserInfoArchive**](UserInfoArchive.md) |  | [optional] 

## Methods

### NewBackupInfo

`func NewBackupInfo() *BackupInfo`

NewBackupInfo instantiates a new BackupInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBackupInfoWithDefaults

`func NewBackupInfoWithDefaults() *BackupInfo`

NewBackupInfoWithDefaults instantiates a new BackupInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFileHistory

`func (o *BackupInfo) GetFileHistory() []FileActionInfo`

GetFileHistory returns the FileHistory field if non-nil, zero value otherwise.

### GetFileHistoryOk

`func (o *BackupInfo) GetFileHistoryOk() (*[]FileActionInfo, bool)`

GetFileHistoryOk returns a tuple with the FileHistory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileHistory

`func (o *BackupInfo) SetFileHistory(v []FileActionInfo)`

SetFileHistory sets FileHistory field to given value.

### HasFileHistory

`func (o *BackupInfo) HasFileHistory() bool`

HasFileHistory returns a boolean if a field has been set.

### GetInstances

`func (o *BackupInfo) GetInstances() []TowerInfo`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *BackupInfo) GetInstancesOk() (*[]TowerInfo, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *BackupInfo) SetInstances(v []TowerInfo)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *BackupInfo) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### GetLifetimesCount

`func (o *BackupInfo) GetLifetimesCount() int32`

GetLifetimesCount returns the LifetimesCount field if non-nil, zero value otherwise.

### GetLifetimesCountOk

`func (o *BackupInfo) GetLifetimesCountOk() (*int32, bool)`

GetLifetimesCountOk returns a tuple with the LifetimesCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLifetimesCount

`func (o *BackupInfo) SetLifetimesCount(v int32)`

SetLifetimesCount sets LifetimesCount field to given value.

### HasLifetimesCount

`func (o *BackupInfo) HasLifetimesCount() bool`

HasLifetimesCount returns a boolean if a field has been set.

### GetTokens

`func (o *BackupInfo) GetTokens() []TokenInfo`

GetTokens returns the Tokens field if non-nil, zero value otherwise.

### GetTokensOk

`func (o *BackupInfo) GetTokensOk() (*[]TokenInfo, bool)`

GetTokensOk returns a tuple with the Tokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokens

`func (o *BackupInfo) SetTokens(v []TokenInfo)`

SetTokens sets Tokens field to given value.

### HasTokens

`func (o *BackupInfo) HasTokens() bool`

HasTokens returns a boolean if a field has been set.

### GetUsers

`func (o *BackupInfo) GetUsers() []UserInfoArchive`

GetUsers returns the Users field if non-nil, zero value otherwise.

### GetUsersOk

`func (o *BackupInfo) GetUsersOk() (*[]UserInfoArchive, bool)`

GetUsersOk returns a tuple with the Users field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsers

`func (o *BackupInfo) SetUsers(v []UserInfoArchive)`

SetUsers sets Users field to given value.

### HasUsers

`func (o *BackupInfo) HasUsers() bool`

HasUsers returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


