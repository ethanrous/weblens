# TowerInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BackupSize** | **int64** |  | 
**CoreAddress** | **string** | Address of the remote server, only if the instance is a core. Not set for any remotes/backups on core server, as it IS the core | 
**Id** | **string** |  | 
**LastBackup** | **int64** |  | 
**LogLevel** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**Online** | **bool** |  | 
**ReportedRole** | **string** | Role the server is currently reporting. This is used to determine if the server is online (and functional) or not | 
**Role** | **string** | Core or Backup | 
**Started** | **bool** |  | 
**UserCount** | **int32** |  | 

## Methods

### NewTowerInfo

`func NewTowerInfo(backupSize int64, coreAddress string, id string, lastBackup int64, name string, online bool, reportedRole string, role string, started bool, userCount int32, ) *TowerInfo`

NewTowerInfo instantiates a new TowerInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTowerInfoWithDefaults

`func NewTowerInfoWithDefaults() *TowerInfo`

NewTowerInfoWithDefaults instantiates a new TowerInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBackupSize

`func (o *TowerInfo) GetBackupSize() int64`

GetBackupSize returns the BackupSize field if non-nil, zero value otherwise.

### GetBackupSizeOk

`func (o *TowerInfo) GetBackupSizeOk() (*int64, bool)`

GetBackupSizeOk returns a tuple with the BackupSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackupSize

`func (o *TowerInfo) SetBackupSize(v int64)`

SetBackupSize sets BackupSize field to given value.


### GetCoreAddress

`func (o *TowerInfo) GetCoreAddress() string`

GetCoreAddress returns the CoreAddress field if non-nil, zero value otherwise.

### GetCoreAddressOk

`func (o *TowerInfo) GetCoreAddressOk() (*string, bool)`

GetCoreAddressOk returns a tuple with the CoreAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCoreAddress

`func (o *TowerInfo) SetCoreAddress(v string)`

SetCoreAddress sets CoreAddress field to given value.


### GetId

`func (o *TowerInfo) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TowerInfo) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TowerInfo) SetId(v string)`

SetId sets Id field to given value.


### GetLastBackup

`func (o *TowerInfo) GetLastBackup() int64`

GetLastBackup returns the LastBackup field if non-nil, zero value otherwise.

### GetLastBackupOk

`func (o *TowerInfo) GetLastBackupOk() (*int64, bool)`

GetLastBackupOk returns a tuple with the LastBackup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastBackup

`func (o *TowerInfo) SetLastBackup(v int64)`

SetLastBackup sets LastBackup field to given value.


### GetLogLevel

`func (o *TowerInfo) GetLogLevel() string`

GetLogLevel returns the LogLevel field if non-nil, zero value otherwise.

### GetLogLevelOk

`func (o *TowerInfo) GetLogLevelOk() (*string, bool)`

GetLogLevelOk returns a tuple with the LogLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogLevel

`func (o *TowerInfo) SetLogLevel(v string)`

SetLogLevel sets LogLevel field to given value.

### HasLogLevel

`func (o *TowerInfo) HasLogLevel() bool`

HasLogLevel returns a boolean if a field has been set.

### GetName

`func (o *TowerInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TowerInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TowerInfo) SetName(v string)`

SetName sets Name field to given value.


### GetOnline

`func (o *TowerInfo) GetOnline() bool`

GetOnline returns the Online field if non-nil, zero value otherwise.

### GetOnlineOk

`func (o *TowerInfo) GetOnlineOk() (*bool, bool)`

GetOnlineOk returns a tuple with the Online field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnline

`func (o *TowerInfo) SetOnline(v bool)`

SetOnline sets Online field to given value.


### GetReportedRole

`func (o *TowerInfo) GetReportedRole() string`

GetReportedRole returns the ReportedRole field if non-nil, zero value otherwise.

### GetReportedRoleOk

`func (o *TowerInfo) GetReportedRoleOk() (*string, bool)`

GetReportedRoleOk returns a tuple with the ReportedRole field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReportedRole

`func (o *TowerInfo) SetReportedRole(v string)`

SetReportedRole sets ReportedRole field to given value.


### GetRole

`func (o *TowerInfo) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *TowerInfo) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *TowerInfo) SetRole(v string)`

SetRole sets Role field to given value.


### GetStarted

`func (o *TowerInfo) GetStarted() bool`

GetStarted returns the Started field if non-nil, zero value otherwise.

### GetStartedOk

`func (o *TowerInfo) GetStartedOk() (*bool, bool)`

GetStartedOk returns a tuple with the Started field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStarted

`func (o *TowerInfo) SetStarted(v bool)`

SetStarted sets Started field to given value.


### GetUserCount

`func (o *TowerInfo) GetUserCount() int32`

GetUserCount returns the UserCount field if non-nil, zero value otherwise.

### GetUserCountOk

`func (o *TowerInfo) GetUserCountOk() (*int32, bool)`

GetUserCountOk returns a tuple with the UserCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCount

`func (o *TowerInfo) SetUserCount(v int32)`

SetUserCount sets UserCount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


