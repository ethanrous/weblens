# UserInfoArchive

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Activated** | **bool** |  | 
**FullName** | **string** |  | 
**HomeID** | **string** |  | 
**IsOnline** | Pointer to **bool** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**PermissionLevel** | **int32** |  | 
**Token** | Pointer to **string** |  | [optional] 
**TrashID** | **string** |  | 
**Username** | **string** |  | 

## Methods

### NewUserInfoArchive

`func NewUserInfoArchive(activated bool, fullName string, homeID string, permissionLevel int32, trashID string, username string, ) *UserInfoArchive`

NewUserInfoArchive instantiates a new UserInfoArchive object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserInfoArchiveWithDefaults

`func NewUserInfoArchiveWithDefaults() *UserInfoArchive`

NewUserInfoArchiveWithDefaults instantiates a new UserInfoArchive object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActivated

`func (o *UserInfoArchive) GetActivated() bool`

GetActivated returns the Activated field if non-nil, zero value otherwise.

### GetActivatedOk

`func (o *UserInfoArchive) GetActivatedOk() (*bool, bool)`

GetActivatedOk returns a tuple with the Activated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActivated

`func (o *UserInfoArchive) SetActivated(v bool)`

SetActivated sets Activated field to given value.


### GetFullName

`func (o *UserInfoArchive) GetFullName() string`

GetFullName returns the FullName field if non-nil, zero value otherwise.

### GetFullNameOk

`func (o *UserInfoArchive) GetFullNameOk() (*string, bool)`

GetFullNameOk returns a tuple with the FullName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullName

`func (o *UserInfoArchive) SetFullName(v string)`

SetFullName sets FullName field to given value.


### GetHomeID

`func (o *UserInfoArchive) GetHomeID() string`

GetHomeID returns the HomeID field if non-nil, zero value otherwise.

### GetHomeIDOk

`func (o *UserInfoArchive) GetHomeIDOk() (*string, bool)`

GetHomeIDOk returns a tuple with the HomeID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHomeID

`func (o *UserInfoArchive) SetHomeID(v string)`

SetHomeID sets HomeID field to given value.


### GetIsOnline

`func (o *UserInfoArchive) GetIsOnline() bool`

GetIsOnline returns the IsOnline field if non-nil, zero value otherwise.

### GetIsOnlineOk

`func (o *UserInfoArchive) GetIsOnlineOk() (*bool, bool)`

GetIsOnlineOk returns a tuple with the IsOnline field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsOnline

`func (o *UserInfoArchive) SetIsOnline(v bool)`

SetIsOnline sets IsOnline field to given value.

### HasIsOnline

`func (o *UserInfoArchive) HasIsOnline() bool`

HasIsOnline returns a boolean if a field has been set.

### GetPassword

`func (o *UserInfoArchive) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *UserInfoArchive) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *UserInfoArchive) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *UserInfoArchive) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetPermissionLevel

`func (o *UserInfoArchive) GetPermissionLevel() int32`

GetPermissionLevel returns the PermissionLevel field if non-nil, zero value otherwise.

### GetPermissionLevelOk

`func (o *UserInfoArchive) GetPermissionLevelOk() (*int32, bool)`

GetPermissionLevelOk returns a tuple with the PermissionLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissionLevel

`func (o *UserInfoArchive) SetPermissionLevel(v int32)`

SetPermissionLevel sets PermissionLevel field to given value.


### GetToken

`func (o *UserInfoArchive) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *UserInfoArchive) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *UserInfoArchive) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *UserInfoArchive) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetTrashID

`func (o *UserInfoArchive) GetTrashID() string`

GetTrashID returns the TrashID field if non-nil, zero value otherwise.

### GetTrashIDOk

`func (o *UserInfoArchive) GetTrashIDOk() (*string, bool)`

GetTrashIDOk returns a tuple with the TrashID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrashID

`func (o *UserInfoArchive) SetTrashID(v string)`

SetTrashID sets TrashID field to given value.


### GetUsername

`func (o *UserInfoArchive) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *UserInfoArchive) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *UserInfoArchive) SetUsername(v string)`

SetUsername sets Username field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


