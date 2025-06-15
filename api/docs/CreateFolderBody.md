# CreateFolderBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Children** | Pointer to **[]string** |  | [optional] 
**NewFolderName** | **string** |  | 
**ParentFolderId** | **string** |  | 

## Methods

### NewCreateFolderBody

`func NewCreateFolderBody(newFolderName string, parentFolderId string, ) *CreateFolderBody`

NewCreateFolderBody instantiates a new CreateFolderBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateFolderBodyWithDefaults

`func NewCreateFolderBodyWithDefaults() *CreateFolderBody`

NewCreateFolderBodyWithDefaults instantiates a new CreateFolderBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChildren

`func (o *CreateFolderBody) GetChildren() []string`

GetChildren returns the Children field if non-nil, zero value otherwise.

### GetChildrenOk

`func (o *CreateFolderBody) GetChildrenOk() (*[]string, bool)`

GetChildrenOk returns a tuple with the Children field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildren

`func (o *CreateFolderBody) SetChildren(v []string)`

SetChildren sets Children field to given value.

### HasChildren

`func (o *CreateFolderBody) HasChildren() bool`

HasChildren returns a boolean if a field has been set.

### GetNewFolderName

`func (o *CreateFolderBody) GetNewFolderName() string`

GetNewFolderName returns the NewFolderName field if non-nil, zero value otherwise.

### GetNewFolderNameOk

`func (o *CreateFolderBody) GetNewFolderNameOk() (*string, bool)`

GetNewFolderNameOk returns a tuple with the NewFolderName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewFolderName

`func (o *CreateFolderBody) SetNewFolderName(v string)`

SetNewFolderName sets NewFolderName field to given value.


### GetParentFolderId

`func (o *CreateFolderBody) GetParentFolderId() string`

GetParentFolderId returns the ParentFolderId field if non-nil, zero value otherwise.

### GetParentFolderIdOk

`func (o *CreateFolderBody) GetParentFolderIdOk() (*string, bool)`

GetParentFolderIdOk returns a tuple with the ParentFolderId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentFolderId

`func (o *CreateFolderBody) SetParentFolderId(v string)`

SetParentFolderId sets ParentFolderId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


