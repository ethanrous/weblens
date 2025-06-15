# MediaTypeInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FileExtension** | Pointer to **[]string** |  | [optional] 
**FriendlyName** | Pointer to **string** |  | [optional] 
**IsDisplayable** | Pointer to **bool** |  | [optional] 
**IsRaw** | Pointer to **bool** |  | [optional] 
**IsVideo** | Pointer to **bool** |  | [optional] 
**MultiPage** | Pointer to **bool** |  | [optional] 
**RawThumbExifKey** | Pointer to **string** |  | [optional] 
**SupportsImgRecog** | Pointer to **bool** |  | [optional] 
**Mime** | Pointer to **string** |  | [optional] 

## Methods

### NewMediaTypeInfo

`func NewMediaTypeInfo() *MediaTypeInfo`

NewMediaTypeInfo instantiates a new MediaTypeInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMediaTypeInfoWithDefaults

`func NewMediaTypeInfoWithDefaults() *MediaTypeInfo`

NewMediaTypeInfoWithDefaults instantiates a new MediaTypeInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFileExtension

`func (o *MediaTypeInfo) GetFileExtension() []string`

GetFileExtension returns the FileExtension field if non-nil, zero value otherwise.

### GetFileExtensionOk

`func (o *MediaTypeInfo) GetFileExtensionOk() (*[]string, bool)`

GetFileExtensionOk returns a tuple with the FileExtension field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileExtension

`func (o *MediaTypeInfo) SetFileExtension(v []string)`

SetFileExtension sets FileExtension field to given value.

### HasFileExtension

`func (o *MediaTypeInfo) HasFileExtension() bool`

HasFileExtension returns a boolean if a field has been set.

### GetFriendlyName

`func (o *MediaTypeInfo) GetFriendlyName() string`

GetFriendlyName returns the FriendlyName field if non-nil, zero value otherwise.

### GetFriendlyNameOk

`func (o *MediaTypeInfo) GetFriendlyNameOk() (*string, bool)`

GetFriendlyNameOk returns a tuple with the FriendlyName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFriendlyName

`func (o *MediaTypeInfo) SetFriendlyName(v string)`

SetFriendlyName sets FriendlyName field to given value.

### HasFriendlyName

`func (o *MediaTypeInfo) HasFriendlyName() bool`

HasFriendlyName returns a boolean if a field has been set.

### GetIsDisplayable

`func (o *MediaTypeInfo) GetIsDisplayable() bool`

GetIsDisplayable returns the IsDisplayable field if non-nil, zero value otherwise.

### GetIsDisplayableOk

`func (o *MediaTypeInfo) GetIsDisplayableOk() (*bool, bool)`

GetIsDisplayableOk returns a tuple with the IsDisplayable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDisplayable

`func (o *MediaTypeInfo) SetIsDisplayable(v bool)`

SetIsDisplayable sets IsDisplayable field to given value.

### HasIsDisplayable

`func (o *MediaTypeInfo) HasIsDisplayable() bool`

HasIsDisplayable returns a boolean if a field has been set.

### GetIsRaw

`func (o *MediaTypeInfo) GetIsRaw() bool`

GetIsRaw returns the IsRaw field if non-nil, zero value otherwise.

### GetIsRawOk

`func (o *MediaTypeInfo) GetIsRawOk() (*bool, bool)`

GetIsRawOk returns a tuple with the IsRaw field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsRaw

`func (o *MediaTypeInfo) SetIsRaw(v bool)`

SetIsRaw sets IsRaw field to given value.

### HasIsRaw

`func (o *MediaTypeInfo) HasIsRaw() bool`

HasIsRaw returns a boolean if a field has been set.

### GetIsVideo

`func (o *MediaTypeInfo) GetIsVideo() bool`

GetIsVideo returns the IsVideo field if non-nil, zero value otherwise.

### GetIsVideoOk

`func (o *MediaTypeInfo) GetIsVideoOk() (*bool, bool)`

GetIsVideoOk returns a tuple with the IsVideo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsVideo

`func (o *MediaTypeInfo) SetIsVideo(v bool)`

SetIsVideo sets IsVideo field to given value.

### HasIsVideo

`func (o *MediaTypeInfo) HasIsVideo() bool`

HasIsVideo returns a boolean if a field has been set.

### GetMultiPage

`func (o *MediaTypeInfo) GetMultiPage() bool`

GetMultiPage returns the MultiPage field if non-nil, zero value otherwise.

### GetMultiPageOk

`func (o *MediaTypeInfo) GetMultiPageOk() (*bool, bool)`

GetMultiPageOk returns a tuple with the MultiPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMultiPage

`func (o *MediaTypeInfo) SetMultiPage(v bool)`

SetMultiPage sets MultiPage field to given value.

### HasMultiPage

`func (o *MediaTypeInfo) HasMultiPage() bool`

HasMultiPage returns a boolean if a field has been set.

### GetRawThumbExifKey

`func (o *MediaTypeInfo) GetRawThumbExifKey() string`

GetRawThumbExifKey returns the RawThumbExifKey field if non-nil, zero value otherwise.

### GetRawThumbExifKeyOk

`func (o *MediaTypeInfo) GetRawThumbExifKeyOk() (*string, bool)`

GetRawThumbExifKeyOk returns a tuple with the RawThumbExifKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRawThumbExifKey

`func (o *MediaTypeInfo) SetRawThumbExifKey(v string)`

SetRawThumbExifKey sets RawThumbExifKey field to given value.

### HasRawThumbExifKey

`func (o *MediaTypeInfo) HasRawThumbExifKey() bool`

HasRawThumbExifKey returns a boolean if a field has been set.

### GetSupportsImgRecog

`func (o *MediaTypeInfo) GetSupportsImgRecog() bool`

GetSupportsImgRecog returns the SupportsImgRecog field if non-nil, zero value otherwise.

### GetSupportsImgRecogOk

`func (o *MediaTypeInfo) GetSupportsImgRecogOk() (*bool, bool)`

GetSupportsImgRecogOk returns a tuple with the SupportsImgRecog field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupportsImgRecog

`func (o *MediaTypeInfo) SetSupportsImgRecog(v bool)`

SetSupportsImgRecog sets SupportsImgRecog field to given value.

### HasSupportsImgRecog

`func (o *MediaTypeInfo) HasSupportsImgRecog() bool`

HasSupportsImgRecog returns a boolean if a field has been set.

### GetMime

`func (o *MediaTypeInfo) GetMime() string`

GetMime returns the Mime field if non-nil, zero value otherwise.

### GetMimeOk

`func (o *MediaTypeInfo) GetMimeOk() (*string, bool)`

GetMimeOk returns a tuple with the Mime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMime

`func (o *MediaTypeInfo) SetMime(v string)`

SetMime sets Mime field to given value.

### HasMime

`func (o *MediaTypeInfo) HasMime() bool`

HasMime returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


