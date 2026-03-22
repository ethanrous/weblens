# FilesInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Files** | [**[]FileInfo**](FileInfo.md) |  | 
**Medias** | Pointer to [**[]MediaInfo**](MediaInfo.md) |  | [optional] 

## Methods

### NewFilesInfo

`func NewFilesInfo(files []FileInfo, ) *FilesInfo`

NewFilesInfo instantiates a new FilesInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFilesInfoWithDefaults

`func NewFilesInfoWithDefaults() *FilesInfo`

NewFilesInfoWithDefaults instantiates a new FilesInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFiles

`func (o *FilesInfo) GetFiles() []FileInfo`

GetFiles returns the Files field if non-nil, zero value otherwise.

### GetFilesOk

`func (o *FilesInfo) GetFilesOk() (*[]FileInfo, bool)`

GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFiles

`func (o *FilesInfo) SetFiles(v []FileInfo)`

SetFiles sets Files field to given value.


### GetMedias

`func (o *FilesInfo) GetMedias() []MediaInfo`

GetMedias returns the Medias field if non-nil, zero value otherwise.

### GetMediasOk

`func (o *FilesInfo) GetMediasOk() (*[]MediaInfo, bool)`

GetMediasOk returns a tuple with the Medias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMedias

`func (o *FilesInfo) SetMedias(v []MediaInfo)`

SetMedias sets Medias field to given value.

### HasMedias

`func (o *FilesInfo) HasMedias() bool`

HasMedias returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


