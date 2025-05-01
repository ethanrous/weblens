# FileInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChildrenIds** | Pointer to **[]string** |  | [optional] 
**ContentId** | Pointer to **string** |  | [optional] 
**CurrentId** | Pointer to **string** |  | [optional] 
**HasRestoreMedia** | Pointer to **bool** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**IsDir** | Pointer to **bool** |  | [optional] 
**Modifiable** | Pointer to **bool** |  | [optional] 
**ModifyTimestamp** | Pointer to **int32** |  | [optional] 
**Owner** | Pointer to **string** |  | [optional] 
**ParentId** | Pointer to **string** |  | [optional] 
**PastFile** | Pointer to **bool** |  | [optional] 
**PortablePath** | Pointer to **string** |  | [optional] 
**ShareId** | Pointer to **string** |  | [optional] 
**Size** | Pointer to **int32** |  | [optional] 

## Methods

### NewFileInfo

`func NewFileInfo() *FileInfo`

NewFileInfo instantiates a new FileInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFileInfoWithDefaults

`func NewFileInfoWithDefaults() *FileInfo`

NewFileInfoWithDefaults instantiates a new FileInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChildrenIds

`func (o *FileInfo) GetChildrenIds() []string`

GetChildrenIds returns the ChildrenIds field if non-nil, zero value otherwise.

### GetChildrenIdsOk

`func (o *FileInfo) GetChildrenIdsOk() (*[]string, bool)`

GetChildrenIdsOk returns a tuple with the ChildrenIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildrenIds

`func (o *FileInfo) SetChildrenIds(v []string)`

SetChildrenIds sets ChildrenIds field to given value.

### HasChildrenIds

`func (o *FileInfo) HasChildrenIds() bool`

HasChildrenIds returns a boolean if a field has been set.

### GetContentId

`func (o *FileInfo) GetContentId() string`

GetContentId returns the ContentId field if non-nil, zero value otherwise.

### GetContentIdOk

`func (o *FileInfo) GetContentIdOk() (*string, bool)`

GetContentIdOk returns a tuple with the ContentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentId

`func (o *FileInfo) SetContentId(v string)`

SetContentId sets ContentId field to given value.

### HasContentId

`func (o *FileInfo) HasContentId() bool`

HasContentId returns a boolean if a field has been set.

### GetCurrentId

`func (o *FileInfo) GetCurrentId() string`

GetCurrentId returns the CurrentId field if non-nil, zero value otherwise.

### GetCurrentIdOk

`func (o *FileInfo) GetCurrentIdOk() (*string, bool)`

GetCurrentIdOk returns a tuple with the CurrentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentId

`func (o *FileInfo) SetCurrentId(v string)`

SetCurrentId sets CurrentId field to given value.

### HasCurrentId

`func (o *FileInfo) HasCurrentId() bool`

HasCurrentId returns a boolean if a field has been set.

### GetHasRestoreMedia

`func (o *FileInfo) GetHasRestoreMedia() bool`

GetHasRestoreMedia returns the HasRestoreMedia field if non-nil, zero value otherwise.

### GetHasRestoreMediaOk

`func (o *FileInfo) GetHasRestoreMediaOk() (*bool, bool)`

GetHasRestoreMediaOk returns a tuple with the HasRestoreMedia field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasRestoreMedia

`func (o *FileInfo) SetHasRestoreMedia(v bool)`

SetHasRestoreMedia sets HasRestoreMedia field to given value.

### HasHasRestoreMedia

`func (o *FileInfo) HasHasRestoreMedia() bool`

HasHasRestoreMedia returns a boolean if a field has been set.

### GetId

`func (o *FileInfo) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *FileInfo) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *FileInfo) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *FileInfo) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIsDir

`func (o *FileInfo) GetIsDir() bool`

GetIsDir returns the IsDir field if non-nil, zero value otherwise.

### GetIsDirOk

`func (o *FileInfo) GetIsDirOk() (*bool, bool)`

GetIsDirOk returns a tuple with the IsDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDir

`func (o *FileInfo) SetIsDir(v bool)`

SetIsDir sets IsDir field to given value.

### HasIsDir

`func (o *FileInfo) HasIsDir() bool`

HasIsDir returns a boolean if a field has been set.

### GetModifiable

`func (o *FileInfo) GetModifiable() bool`

GetModifiable returns the Modifiable field if non-nil, zero value otherwise.

### GetModifiableOk

`func (o *FileInfo) GetModifiableOk() (*bool, bool)`

GetModifiableOk returns a tuple with the Modifiable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModifiable

`func (o *FileInfo) SetModifiable(v bool)`

SetModifiable sets Modifiable field to given value.

### HasModifiable

`func (o *FileInfo) HasModifiable() bool`

HasModifiable returns a boolean if a field has been set.

### GetModifyTimestamp

`func (o *FileInfo) GetModifyTimestamp() int32`

GetModifyTimestamp returns the ModifyTimestamp field if non-nil, zero value otherwise.

### GetModifyTimestampOk

`func (o *FileInfo) GetModifyTimestampOk() (*int32, bool)`

GetModifyTimestampOk returns a tuple with the ModifyTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModifyTimestamp

`func (o *FileInfo) SetModifyTimestamp(v int32)`

SetModifyTimestamp sets ModifyTimestamp field to given value.

### HasModifyTimestamp

`func (o *FileInfo) HasModifyTimestamp() bool`

HasModifyTimestamp returns a boolean if a field has been set.

### GetOwner

`func (o *FileInfo) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *FileInfo) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *FileInfo) SetOwner(v string)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *FileInfo) HasOwner() bool`

HasOwner returns a boolean if a field has been set.

### GetParentId

`func (o *FileInfo) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *FileInfo) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *FileInfo) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *FileInfo) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### GetPastFile

`func (o *FileInfo) GetPastFile() bool`

GetPastFile returns the PastFile field if non-nil, zero value otherwise.

### GetPastFileOk

`func (o *FileInfo) GetPastFileOk() (*bool, bool)`

GetPastFileOk returns a tuple with the PastFile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPastFile

`func (o *FileInfo) SetPastFile(v bool)`

SetPastFile sets PastFile field to given value.

### HasPastFile

`func (o *FileInfo) HasPastFile() bool`

HasPastFile returns a boolean if a field has been set.

### GetPortablePath

`func (o *FileInfo) GetPortablePath() string`

GetPortablePath returns the PortablePath field if non-nil, zero value otherwise.

### GetPortablePathOk

`func (o *FileInfo) GetPortablePathOk() (*string, bool)`

GetPortablePathOk returns a tuple with the PortablePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPortablePath

`func (o *FileInfo) SetPortablePath(v string)`

SetPortablePath sets PortablePath field to given value.

### HasPortablePath

`func (o *FileInfo) HasPortablePath() bool`

HasPortablePath returns a boolean if a field has been set.

### GetShareId

`func (o *FileInfo) GetShareId() string`

GetShareId returns the ShareId field if non-nil, zero value otherwise.

### GetShareIdOk

`func (o *FileInfo) GetShareIdOk() (*string, bool)`

GetShareIdOk returns a tuple with the ShareId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShareId

`func (o *FileInfo) SetShareId(v string)`

SetShareId sets ShareId field to given value.

### HasShareId

`func (o *FileInfo) HasShareId() bool`

HasShareId returns a boolean if a field has been set.

### GetSize

`func (o *FileInfo) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *FileInfo) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *FileInfo) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *FileInfo) HasSize() bool`

HasSize returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


