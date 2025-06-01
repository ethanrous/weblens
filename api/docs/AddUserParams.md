# AddUserParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CanDelete** | Pointer to **bool** |  | [optional] 
**CanDownload** | Pointer to **bool** |  | [optional] 
**CanEdit** | Pointer to **bool** |  | [optional] 
**Username** | **string** |  | 

## Methods

### NewAddUserParams

`func NewAddUserParams(username string, ) *AddUserParams`

NewAddUserParams instantiates a new AddUserParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAddUserParamsWithDefaults

`func NewAddUserParamsWithDefaults() *AddUserParams`

NewAddUserParamsWithDefaults instantiates a new AddUserParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCanDelete

`func (o *AddUserParams) GetCanDelete() bool`

GetCanDelete returns the CanDelete field if non-nil, zero value otherwise.

### GetCanDeleteOk

`func (o *AddUserParams) GetCanDeleteOk() (*bool, bool)`

GetCanDeleteOk returns a tuple with the CanDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanDelete

`func (o *AddUserParams) SetCanDelete(v bool)`

SetCanDelete sets CanDelete field to given value.

### HasCanDelete

`func (o *AddUserParams) HasCanDelete() bool`

HasCanDelete returns a boolean if a field has been set.

### GetCanDownload

`func (o *AddUserParams) GetCanDownload() bool`

GetCanDownload returns the CanDownload field if non-nil, zero value otherwise.

### GetCanDownloadOk

`func (o *AddUserParams) GetCanDownloadOk() (*bool, bool)`

GetCanDownloadOk returns a tuple with the CanDownload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanDownload

`func (o *AddUserParams) SetCanDownload(v bool)`

SetCanDownload sets CanDownload field to given value.

### HasCanDownload

`func (o *AddUserParams) HasCanDownload() bool`

HasCanDownload returns a boolean if a field has been set.

### GetCanEdit

`func (o *AddUserParams) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *AddUserParams) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *AddUserParams) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *AddUserParams) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### GetUsername

`func (o *AddUserParams) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *AddUserParams) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *AddUserParams) SetUsername(v string)`

SetUsername sets Username field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


