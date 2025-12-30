# MoveFilesParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FileIDs** | Pointer to **[]string** |  | [optional] 
**NewParentID** | Pointer to **string** |  | [optional] 

## Methods

### NewMoveFilesParams

`func NewMoveFilesParams() *MoveFilesParams`

NewMoveFilesParams instantiates a new MoveFilesParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMoveFilesParamsWithDefaults

`func NewMoveFilesParamsWithDefaults() *MoveFilesParams`

NewMoveFilesParamsWithDefaults instantiates a new MoveFilesParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFileIDs

`func (o *MoveFilesParams) GetFileIDs() []string`

GetFileIDs returns the FileIDs field if non-nil, zero value otherwise.

### GetFileIDsOk

`func (o *MoveFilesParams) GetFileIDsOk() (*[]string, bool)`

GetFileIDsOk returns a tuple with the FileIDs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileIDs

`func (o *MoveFilesParams) SetFileIDs(v []string)`

SetFileIDs sets FileIDs field to given value.

### HasFileIDs

`func (o *MoveFilesParams) HasFileIDs() bool`

HasFileIDs returns a boolean if a field has been set.

### GetNewParentID

`func (o *MoveFilesParams) GetNewParentID() string`

GetNewParentID returns the NewParentID field if non-nil, zero value otherwise.

### GetNewParentIDOk

`func (o *MoveFilesParams) GetNewParentIDOk() (*string, bool)`

GetNewParentIDOk returns a tuple with the NewParentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewParentID

`func (o *MoveFilesParams) SetNewParentID(v string)`

SetNewParentID sets NewParentID field to given value.

### HasNewParentID

`func (o *MoveFilesParams) HasNewParentID() bool`

HasNewParentID returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


