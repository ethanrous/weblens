# WlstructsInitServerParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CoreAddress** | Pointer to **string** |  | [optional] 
**CoreKey** | Pointer to **string** |  | [optional] 
**FullName** | Pointer to **string** |  | [optional] 
**LocalID** | Pointer to **string** | For restoring a server, remoind the core of its serverID and api key the remote last used | [optional] 
**Name** | **string** |  | 
**Password** | **string** |  | 
**RemoteID** | Pointer to **string** |  | [optional] 
**Role** | **string** |  | 
**Username** | **string** |  | 
**UsingKeyInfo** | Pointer to **string** |  | [optional] 

## Methods

### NewWlstructsInitServerParams

`func NewWlstructsInitServerParams(name string, password string, role string, username string, ) *WlstructsInitServerParams`

NewWlstructsInitServerParams instantiates a new WlstructsInitServerParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWlstructsInitServerParamsWithDefaults

`func NewWlstructsInitServerParamsWithDefaults() *WlstructsInitServerParams`

NewWlstructsInitServerParamsWithDefaults instantiates a new WlstructsInitServerParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCoreAddress

`func (o *WlstructsInitServerParams) GetCoreAddress() string`

GetCoreAddress returns the CoreAddress field if non-nil, zero value otherwise.

### GetCoreAddressOk

`func (o *WlstructsInitServerParams) GetCoreAddressOk() (*string, bool)`

GetCoreAddressOk returns a tuple with the CoreAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCoreAddress

`func (o *WlstructsInitServerParams) SetCoreAddress(v string)`

SetCoreAddress sets CoreAddress field to given value.

### HasCoreAddress

`func (o *WlstructsInitServerParams) HasCoreAddress() bool`

HasCoreAddress returns a boolean if a field has been set.

### GetCoreKey

`func (o *WlstructsInitServerParams) GetCoreKey() string`

GetCoreKey returns the CoreKey field if non-nil, zero value otherwise.

### GetCoreKeyOk

`func (o *WlstructsInitServerParams) GetCoreKeyOk() (*string, bool)`

GetCoreKeyOk returns a tuple with the CoreKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCoreKey

`func (o *WlstructsInitServerParams) SetCoreKey(v string)`

SetCoreKey sets CoreKey field to given value.

### HasCoreKey

`func (o *WlstructsInitServerParams) HasCoreKey() bool`

HasCoreKey returns a boolean if a field has been set.

### GetFullName

`func (o *WlstructsInitServerParams) GetFullName() string`

GetFullName returns the FullName field if non-nil, zero value otherwise.

### GetFullNameOk

`func (o *WlstructsInitServerParams) GetFullNameOk() (*string, bool)`

GetFullNameOk returns a tuple with the FullName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullName

`func (o *WlstructsInitServerParams) SetFullName(v string)`

SetFullName sets FullName field to given value.

### HasFullName

`func (o *WlstructsInitServerParams) HasFullName() bool`

HasFullName returns a boolean if a field has been set.

### GetLocalID

`func (o *WlstructsInitServerParams) GetLocalID() string`

GetLocalID returns the LocalID field if non-nil, zero value otherwise.

### GetLocalIDOk

`func (o *WlstructsInitServerParams) GetLocalIDOk() (*string, bool)`

GetLocalIDOk returns a tuple with the LocalID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocalID

`func (o *WlstructsInitServerParams) SetLocalID(v string)`

SetLocalID sets LocalID field to given value.

### HasLocalID

`func (o *WlstructsInitServerParams) HasLocalID() bool`

HasLocalID returns a boolean if a field has been set.

### GetName

`func (o *WlstructsInitServerParams) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WlstructsInitServerParams) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WlstructsInitServerParams) SetName(v string)`

SetName sets Name field to given value.


### GetPassword

`func (o *WlstructsInitServerParams) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *WlstructsInitServerParams) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *WlstructsInitServerParams) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetRemoteID

`func (o *WlstructsInitServerParams) GetRemoteID() string`

GetRemoteID returns the RemoteID field if non-nil, zero value otherwise.

### GetRemoteIDOk

`func (o *WlstructsInitServerParams) GetRemoteIDOk() (*string, bool)`

GetRemoteIDOk returns a tuple with the RemoteID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRemoteID

`func (o *WlstructsInitServerParams) SetRemoteID(v string)`

SetRemoteID sets RemoteID field to given value.

### HasRemoteID

`func (o *WlstructsInitServerParams) HasRemoteID() bool`

HasRemoteID returns a boolean if a field has been set.

### GetRole

`func (o *WlstructsInitServerParams) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *WlstructsInitServerParams) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *WlstructsInitServerParams) SetRole(v string)`

SetRole sets Role field to given value.


### GetUsername

`func (o *WlstructsInitServerParams) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *WlstructsInitServerParams) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *WlstructsInitServerParams) SetUsername(v string)`

SetUsername sets Username field to given value.


### GetUsingKeyInfo

`func (o *WlstructsInitServerParams) GetUsingKeyInfo() string`

GetUsingKeyInfo returns the UsingKeyInfo field if non-nil, zero value otherwise.

### GetUsingKeyInfoOk

`func (o *WlstructsInitServerParams) GetUsingKeyInfoOk() (*string, bool)`

GetUsingKeyInfoOk returns a tuple with the UsingKeyInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsingKeyInfo

`func (o *WlstructsInitServerParams) SetUsingKeyInfo(v string)`

SetUsingKeyInfo sets UsingKeyInfo field to given value.

### HasUsingKeyInfo

`func (o *WlstructsInitServerParams) HasUsingKeyInfo() bool`

HasUsingKeyInfo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


