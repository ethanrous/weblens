# UserInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Activated** | **bool** |  | 
**FullName** | **string** |  | 
**HomeID** | **string** |  | 
**IsOnline** | Pointer to **bool** |  | [optional] 
**PermissionLevel** | **int32** |  | 
**Token** | Pointer to **string** |  | [optional] 
**TrashID** | **string** |  | 
**UpdatedAt** | **int64** |  | 
**Username** | **string** |  | 

## Methods

### NewUserInfo

`func NewUserInfo(activated bool, fullName string, homeID string, permissionLevel int32, trashID string, updatedAt int64, username string, ) *UserInfo`

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


### GetHomeID

`func (o *UserInfo) GetHomeID() string`

GetHomeID returns the HomeID field if non-nil, zero value otherwise.

### GetHomeIDOk

`func (o *UserInfo) GetHomeIDOk() (*string, bool)`

GetHomeIDOk returns a tuple with the HomeID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHomeID

`func (o *UserInfo) SetHomeID(v string)`

SetHomeID sets HomeID field to given value.


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

### GetTrashID

`func (o *UserInfo) GetTrashID() string`

GetTrashID returns the TrashID field if non-nil, zero value otherwise.

### GetTrashIDOk

`func (o *UserInfo) GetTrashIDOk() (*string, bool)`

GetTrashIDOk returns a tuple with the TrashID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrashID

`func (o *UserInfo) SetTrashID(v string)`

SetTrashID sets TrashID field to given value.


### GetUpdatedAt

`func (o *UserInfo) GetUpdatedAt() int64`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *UserInfo) GetUpdatedAtOk() (*int64, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *UserInfo) SetUpdatedAt(v int64)`

SetUpdatedAt sets UpdatedAt field to given value.


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


