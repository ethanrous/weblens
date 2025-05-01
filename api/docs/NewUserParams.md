# NewUserParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Admin** | Pointer to **bool** |  | [optional] 
**AutoActivate** | Pointer to **bool** |  | [optional] 
**FullName** | **string** |  | 
**Password** | **string** |  | 
**Username** | **string** |  | 

## Methods

### NewNewUserParams

`func NewNewUserParams(fullName string, password string, username string, ) *NewUserParams`

NewNewUserParams instantiates a new NewUserParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNewUserParamsWithDefaults

`func NewNewUserParamsWithDefaults() *NewUserParams`

NewNewUserParamsWithDefaults instantiates a new NewUserParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAdmin

`func (o *NewUserParams) GetAdmin() bool`

GetAdmin returns the Admin field if non-nil, zero value otherwise.

### GetAdminOk

`func (o *NewUserParams) GetAdminOk() (*bool, bool)`

GetAdminOk returns a tuple with the Admin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdmin

`func (o *NewUserParams) SetAdmin(v bool)`

SetAdmin sets Admin field to given value.

### HasAdmin

`func (o *NewUserParams) HasAdmin() bool`

HasAdmin returns a boolean if a field has been set.

### GetAutoActivate

`func (o *NewUserParams) GetAutoActivate() bool`

GetAutoActivate returns the AutoActivate field if non-nil, zero value otherwise.

### GetAutoActivateOk

`func (o *NewUserParams) GetAutoActivateOk() (*bool, bool)`

GetAutoActivateOk returns a tuple with the AutoActivate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoActivate

`func (o *NewUserParams) SetAutoActivate(v bool)`

SetAutoActivate sets AutoActivate field to given value.

### HasAutoActivate

`func (o *NewUserParams) HasAutoActivate() bool`

HasAutoActivate returns a boolean if a field has been set.

### GetFullName

`func (o *NewUserParams) GetFullName() string`

GetFullName returns the FullName field if non-nil, zero value otherwise.

### GetFullNameOk

`func (o *NewUserParams) GetFullNameOk() (*string, bool)`

GetFullNameOk returns a tuple with the FullName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullName

`func (o *NewUserParams) SetFullName(v string)`

SetFullName sets FullName field to given value.


### GetPassword

`func (o *NewUserParams) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *NewUserParams) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *NewUserParams) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetUsername

`func (o *NewUserParams) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *NewUserParams) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *NewUserParams) SetUsername(v string)`

SetUsername sets Username field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


