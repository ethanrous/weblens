# FileInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChildrenIds** | Pointer to **[]string** |  | [optional] 
**ContentID** | Pointer to **string** |  | [optional] 
**HasRestoreMedia** | Pointer to **bool** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**IsDir** | Pointer to **bool** |  | [optional] 
**Modifiable** | Pointer to **bool** |  | [optional] 
**ModifyTimestamp** | Pointer to **int64** |  | [optional] 
**Owner** | Pointer to **string** |  | [optional] 
**ParentID** | Pointer to **string** |  | [optional] 
**PastFile** | Pointer to **bool** |  | [optional] 
**PortablePath** | Pointer to **string** |  | [optional] 
**ShareID** | Pointer to **string** |  | [optional] 
**Size** | Pointer to **int64** |  | [optional] 

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

### GetContentID

`func (o *FileInfo) GetContentID() string`

GetContentID returns the ContentID field if non-nil, zero value otherwise.

### GetContentIDOk

`func (o *FileInfo) GetContentIDOk() (*string, bool)`

GetContentIDOk returns a tuple with the ContentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentID

`func (o *FileInfo) SetContentID(v string)`

SetContentID sets ContentID field to given value.

### HasContentID

`func (o *FileInfo) HasContentID() bool`

HasContentID returns a boolean if a field has been set.

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

`func (o *FileInfo) GetModifyTimestamp() int64`

GetModifyTimestamp returns the ModifyTimestamp field if non-nil, zero value otherwise.

### GetModifyTimestampOk

`func (o *FileInfo) GetModifyTimestampOk() (*int64, bool)`

GetModifyTimestampOk returns a tuple with the ModifyTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModifyTimestamp

`func (o *FileInfo) SetModifyTimestamp(v int64)`

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

### GetParentID

`func (o *FileInfo) GetParentID() string`

GetParentID returns the ParentID field if non-nil, zero value otherwise.

### GetParentIDOk

`func (o *FileInfo) GetParentIDOk() (*string, bool)`

GetParentIDOk returns a tuple with the ParentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentID

`func (o *FileInfo) SetParentID(v string)`

SetParentID sets ParentID field to given value.

### HasParentID

`func (o *FileInfo) HasParentID() bool`

HasParentID returns a boolean if a field has been set.

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

### GetShareID

`func (o *FileInfo) GetShareID() string`

GetShareID returns the ShareID field if non-nil, zero value otherwise.

### GetShareIDOk

`func (o *FileInfo) GetShareIDOk() (*string, bool)`

GetShareIDOk returns a tuple with the ShareID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShareID

`func (o *FileInfo) SetShareID(v string)`

SetShareID sets ShareID field to given value.

### HasShareID

`func (o *FileInfo) HasShareID() bool`

HasShareID returns a boolean if a field has been set.

### GetSize

`func (o *FileInfo) GetSize() int64`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *FileInfo) GetSizeOk() (*int64, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *FileInfo) SetSize(v int64)`

SetSize sets Size field to given value.

### HasSize

`func (o *FileInfo) HasSize() bool`

HasSize returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


