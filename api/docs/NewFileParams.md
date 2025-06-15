# NewFileParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FileSize** | Pointer to **int32** |  | [optional] 
**IsDir** | Pointer to **bool** |  | [optional] 
**NewFileName** | Pointer to **string** |  | [optional] 
**ParentFolderId** | Pointer to **string** |  | [optional] 

## Methods

### NewNewFileParams

`func NewNewFileParams() *NewFileParams`

NewNewFileParams instantiates a new NewFileParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNewFileParamsWithDefaults

`func NewNewFileParamsWithDefaults() *NewFileParams`

NewNewFileParamsWithDefaults instantiates a new NewFileParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFileSize

`func (o *NewFileParams) GetFileSize() int32`

GetFileSize returns the FileSize field if non-nil, zero value otherwise.

### GetFileSizeOk

`func (o *NewFileParams) GetFileSizeOk() (*int32, bool)`

GetFileSizeOk returns a tuple with the FileSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileSize

`func (o *NewFileParams) SetFileSize(v int32)`

SetFileSize sets FileSize field to given value.

### HasFileSize

`func (o *NewFileParams) HasFileSize() bool`

HasFileSize returns a boolean if a field has been set.

### GetIsDir

`func (o *NewFileParams) GetIsDir() bool`

GetIsDir returns the IsDir field if non-nil, zero value otherwise.

### GetIsDirOk

`func (o *NewFileParams) GetIsDirOk() (*bool, bool)`

GetIsDirOk returns a tuple with the IsDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDir

`func (o *NewFileParams) SetIsDir(v bool)`

SetIsDir sets IsDir field to given value.

### HasIsDir

`func (o *NewFileParams) HasIsDir() bool`

HasIsDir returns a boolean if a field has been set.

### GetNewFileName

`func (o *NewFileParams) GetNewFileName() string`

GetNewFileName returns the NewFileName field if non-nil, zero value otherwise.

### GetNewFileNameOk

`func (o *NewFileParams) GetNewFileNameOk() (*string, bool)`

GetNewFileNameOk returns a tuple with the NewFileName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewFileName

`func (o *NewFileParams) SetNewFileName(v string)`

SetNewFileName sets NewFileName field to given value.

### HasNewFileName

`func (o *NewFileParams) HasNewFileName() bool`

HasNewFileName returns a boolean if a field has been set.

### GetParentFolderId

`func (o *NewFileParams) GetParentFolderId() string`

GetParentFolderId returns the ParentFolderId field if non-nil, zero value otherwise.

### GetParentFolderIdOk

`func (o *NewFileParams) GetParentFolderIdOk() (*string, bool)`

GetParentFolderIdOk returns a tuple with the ParentFolderId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentFolderId

`func (o *NewFileParams) SetParentFolderId(v string)`

SetParentFolderId sets ParentFolderId field to given value.

### HasParentFolderId

`func (o *NewFileParams) HasParentFolderId() bool`

HasParentFolderId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


