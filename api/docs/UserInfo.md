# UserInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Activated** | **bool** |  | 
**FullName** | **string** |  | 
**HomeId** | **string** |  | 
**IsOnline** | Pointer to **bool** |  | [optional] 
**PermissionLevel** | **int32** |  | 
**Token** | Pointer to **string** |  | [optional] 
**TrashId** | **string** |  | 
**Username** | **string** |  | 

## Methods

### NewUserInfo

`func NewUserInfo(activated bool, fullName string, homeId string, permissionLevel int32, trashId string, username string, ) *UserInfo`

NewUserInfo instantiates a new UserInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserInfoWithDefaults

`func NewUserInfoWithDefaults() *UserInfo`

NewUserInfoWithDefaults instantiates a new UserInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActivated

`func (o *UserInfo) GetActivated() bool`

GetActivated returns the Activated field if non-nil, zero value otherwise.

### GetActivatedOk

`func (o *UserInfo) GetActivatedOk() (*bool, bool)`

GetActivatedOk returns a tuple with the Activated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActivated

`func (o *UserInfo) SetActivated(v bool)`

SetActivated sets Activated field to given value.


### GetFullName

`func (o *UserInfo) GetFullName() string`

GetFullName returns the FullName field if non-nil, zero value otherwise.

### GetFullNameOk

`func (o *UserInfo) GetFullNameOk() (*string, bool)`

GetFullNameOk returns a tuple with the FullName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullName

`func (o *UserInfo) SetFullName(v string)`

SetFullName sets FullName field to given value.


### GetHomeId

`func (o *UserInfo) GetHomeId() string`

GetHomeId returns the HomeId field if non-nil, zero value otherwise.

### GetHomeIdOk

`func (o *UserInfo) GetHomeIdOk() (*string, bool)`

GetHomeIdOk returns a tuple with the HomeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHomeId

`func (o *UserInfo) SetHomeId(v string)`

SetHomeId sets HomeId field to given value.


### GetIsOnline

`func (o *UserInfo) GetIsOnline() bool`

GetIsOnline returns the IsOnline field if non-nil, zero value otherwise.

### GetIsOnlineOk

`func (o *UserInfo) GetIsOnlineOk() (*bool, bool)`

GetIsOnlineOk returns a tuple with the IsOnline field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsOnline

`func (o *UserInfo) SetIsOnline(v bool)`

SetIsOnline sets IsOnline field to given value.

### HasIsOnline

`func (o *UserInfo) HasIsOnline() bool`

HasIsOnline returns a boolean if a field has been set.

### GetPermissionLevel

`func (o *UserInfo) GetPermissionLevel() int32`

GetPermissionLevel returns the PermissionLevel field if non-nil, zero value otherwise.

### GetPermissionLevelOk

`func (o *UserInfo) GetPermissionLevelOk() (*int32, bool)`

GetPermissionLevelOk returns a tuple with the PermissionLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissionLevel

`func (o *UserInfo) SetPermissionLevel(v int32)`

SetPermissionLevel sets PermissionLevel field to given value.


### GetToken

`func (o *UserInfo) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *UserInfo) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *UserInfo) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *UserInfo) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetTrashId

`func (o *UserInfo) GetTrashId() string`

GetTrashId returns the TrashId field if non-nil, zero value otherwise.

### GetTrashIdOk

`func (o *UserInfo) GetTrashIdOk() (*string, bool)`

GetTrashIdOk returns a tuple with the TrashId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrashId

`func (o *UserInfo) SetTrashId(v string)`

SetTrashId sets TrashId field to given value.


### GetUsername

`func (o *UserInfo) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *UserInfo) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *UserInfo) SetUsername(v string)`

SetUsername sets Username field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


