# PasswordUpdateParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NewPassword** | **string** |  | 
**OldPassword** | Pointer to **string** |  | [optional] 

## Methods

### NewPasswordUpdateParams

`func NewPasswordUpdateParams(newPassword string, ) *PasswordUpdateParams`

NewPasswordUpdateParams instantiates a new PasswordUpdateParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordUpdateParamsWithDefaults

`func NewPasswordUpdateParamsWithDefaults() *PasswordUpdateParams`

NewPasswordUpdateParamsWithDefaults instantiates a new PasswordUpdateParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNewPassword

`func (o *PasswordUpdateParams) GetNewPassword() string`

GetNewPassword returns the NewPassword field if non-nil, zero value otherwise.

### GetNewPasswordOk

`func (o *PasswordUpdateParams) GetNewPasswordOk() (*string, bool)`

GetNewPasswordOk returns a tuple with the NewPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewPassword

`func (o *PasswordUpdateParams) SetNewPassword(v string)`

SetNewPassword sets NewPassword field to given value.


### GetOldPassword

`func (o *PasswordUpdateParams) GetOldPassword() string`

GetOldPassword returns the OldPassword field if non-nil, zero value otherwise.

### GetOldPasswordOk

`func (o *PasswordUpdateParams) GetOldPasswordOk() (*string, bool)`

GetOldPasswordOk returns a tuple with the OldPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOldPassword

`func (o *PasswordUpdateParams) SetOldPassword(v string)`

SetOldPassword sets OldPassword field to given value.

### HasOldPassword

`func (o *PasswordUpdateParams) HasOldPassword() bool`

HasOldPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


