# FolderInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Children** | Pointer to [**[]FileInfo**](FileInfo.md) |  | [optional] 
**Medias** | Pointer to [**[]MediaInfo**](MediaInfo.md) |  | [optional] 
**Parents** | Pointer to [**[]FileInfo**](FileInfo.md) |  | [optional] 
**Self** | Pointer to [**FileInfo**](FileInfo.md) |  | [optional] 

## Methods

### NewFolderInfo

`func NewFolderInfo() *FolderInfo`

NewFolderInfo instantiates a new FolderInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFolderInfoWithDefaults

`func NewFolderInfoWithDefaults() *FolderInfo`

NewFolderInfoWithDefaults instantiates a new FolderInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChildren

`func (o *FolderInfo) GetChildren() []FileInfo`

GetChildren returns the Children field if non-nil, zero value otherwise.

### GetChildrenOk

`func (o *FolderInfo) GetChildrenOk() (*[]FileInfo, bool)`

GetChildrenOk returns a tuple with the Children field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildren

`func (o *FolderInfo) SetChildren(v []FileInfo)`

SetChildren sets Children field to given value.

### HasChildren

`func (o *FolderInfo) HasChildren() bool`

HasChildren returns a boolean if a field has been set.

### GetMedias

`func (o *FolderInfo) GetMedias() []MediaInfo`

GetMedias returns the Medias field if non-nil, zero value otherwise.

### GetMediasOk

`func (o *FolderInfo) GetMediasOk() (*[]MediaInfo, bool)`

GetMediasOk returns a tuple with the Medias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMedias

`func (o *FolderInfo) SetMedias(v []MediaInfo)`

SetMedias sets Medias field to given value.

### HasMedias

`func (o *FolderInfo) HasMedias() bool`

HasMedias returns a boolean if a field has been set.

### GetParents

`func (o *FolderInfo) GetParents() []FileInfo`

GetParents returns the Parents field if non-nil, zero value otherwise.

### GetParentsOk

`func (o *FolderInfo) GetParentsOk() (*[]FileInfo, bool)`

GetParentsOk returns a tuple with the Parents field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParents

`func (o *FolderInfo) SetParents(v []FileInfo)`

SetParents sets Parents field to given value.

### HasParents

`func (o *FolderInfo) HasParents() bool`

HasParents returns a boolean if a field has been set.

### GetSelf

`func (o *FolderInfo) GetSelf() FileInfo`

GetSelf returns the Self field if non-nil, zero value otherwise.

### GetSelfOk

`func (o *FolderInfo) GetSelfOk() (*FileInfo, bool)`

GetSelfOk returns a tuple with the Self field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelf

`func (o *FolderInfo) SetSelf(v FileInfo)`

SetSelf sets Self field to given value.

### HasSelf

`func (o *FolderInfo) HasSelf() bool`

HasSelf returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


