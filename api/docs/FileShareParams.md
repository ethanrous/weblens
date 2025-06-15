# FileShareParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FileId** | Pointer to **string** |  | [optional] 
**Public** | Pointer to **bool** |  | [optional] 
**Users** | Pointer to **[]string** |  | [optional] 
**Wormhole** | Pointer to **bool** |  | [optional] 

## Methods

### NewFileShareParams

`func NewFileShareParams() *FileShareParams`

NewFileShareParams instantiates a new FileShareParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFileShareParamsWithDefaults

`func NewFileShareParamsWithDefaults() *FileShareParams`

NewFileShareParamsWithDefaults instantiates a new FileShareParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFileId

`func (o *FileShareParams) GetFileId() string`

GetFileId returns the FileId field if non-nil, zero value otherwise.

### GetFileIdOk

`func (o *FileShareParams) GetFileIdOk() (*string, bool)`

GetFileIdOk returns a tuple with the FileId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileId

`func (o *FileShareParams) SetFileId(v string)`

SetFileId sets FileId field to given value.

### HasFileId

`func (o *FileShareParams) HasFileId() bool`

HasFileId returns a boolean if a field has been set.

### GetPublic

`func (o *FileShareParams) GetPublic() bool`

GetPublic returns the Public field if non-nil, zero value otherwise.

### GetPublicOk

`func (o *FileShareParams) GetPublicOk() (*bool, bool)`

GetPublicOk returns a tuple with the Public field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublic

`func (o *FileShareParams) SetPublic(v bool)`

SetPublic sets Public field to given value.

### HasPublic

`func (o *FileShareParams) HasPublic() bool`

HasPublic returns a boolean if a field has been set.

### GetUsers

`func (o *FileShareParams) GetUsers() []string`

GetUsers returns the Users field if non-nil, zero value otherwise.

### GetUsersOk

`func (o *FileShareParams) GetUsersOk() (*[]string, bool)`

GetUsersOk returns a tuple with the Users field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsers

`func (o *FileShareParams) SetUsers(v []string)`

SetUsers sets Users field to given value.

### HasUsers

`func (o *FileShareParams) HasUsers() bool`

HasUsers returns a boolean if a field has been set.

### GetWormhole

`func (o *FileShareParams) GetWormhole() bool`

GetWormhole returns the Wormhole field if non-nil, zero value otherwise.

### GetWormholeOk

`func (o *FileShareParams) GetWormholeOk() (*bool, bool)`

GetWormholeOk returns a tuple with the Wormhole field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWormhole

`func (o *FileShareParams) SetWormhole(v bool)`

SetWormhole sets Wormhole field to given value.

### HasWormhole

`func (o *FileShareParams) HasWormhole() bool`

HasWormhole returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


