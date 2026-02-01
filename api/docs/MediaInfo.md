# MediaInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContentID** | Pointer to **string** | Hash of the file content, to ensure that the same files don&#39;t get duplicated | [optional] 
**CreateDate** | Pointer to **int32** |  | [optional] 
**Duration** | Pointer to **int32** | Total time, in milliseconds, of a video | [optional] 
**Enabled** | Pointer to **bool** | If the media disabled. This can happen when the backing file(s) are deleted, but the media stays behind because it can be re-used if needed. | [optional] 
**FileIDs** | Pointer to **[]string** | Slices of files whos content hash to the contentId | [optional] 
**HdirScore** | Pointer to **float32** | Similarity score from HDIR search | [optional] 
**Height** | Pointer to **int32** |  | [optional] 
**Hidden** | Pointer to **bool** | If the media is hidden from the timeline TODO - make this per user | [optional] 
**Imported** | Pointer to **bool** |  | [optional] 
**LikedBy** | Pointer to **[]string** |  | [optional] 
**Location** | Pointer to **[]float32** |  | [optional] 
**MimeType** | Pointer to **string** | Mime-type key of the media | [optional] 
**Owner** | Pointer to **string** | User who owns the file that resulted in this media being created | [optional] 
**PageCount** | Pointer to **int32** | Number of pages (typically 1, 0 in not a valid page count) | [optional] 
**RecognitionTags** | Pointer to **[]string** | Tags from the ML image scan so searching for particular objects in the images can be done | [optional] 
**Width** | Pointer to **int32** | Full-res image dimensions | [optional] 

## Methods

### NewMediaInfo

`func NewMediaInfo() *MediaInfo`

NewMediaInfo instantiates a new MediaInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMediaInfoWithDefaults

`func NewMediaInfoWithDefaults() *MediaInfo`

NewMediaInfoWithDefaults instantiates a new MediaInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContentID

`func (o *MediaInfo) GetContentID() string`

GetContentID returns the ContentID field if non-nil, zero value otherwise.

### GetContentIDOk

`func (o *MediaInfo) GetContentIDOk() (*string, bool)`

GetContentIDOk returns a tuple with the ContentID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentID

`func (o *MediaInfo) SetContentID(v string)`

SetContentID sets ContentID field to given value.

### HasContentID

`func (o *MediaInfo) HasContentID() bool`

HasContentID returns a boolean if a field has been set.

### GetCreateDate

`func (o *MediaInfo) GetCreateDate() int32`

GetCreateDate returns the CreateDate field if non-nil, zero value otherwise.

### GetCreateDateOk

`func (o *MediaInfo) GetCreateDateOk() (*int32, bool)`

GetCreateDateOk returns a tuple with the CreateDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateDate

`func (o *MediaInfo) SetCreateDate(v int32)`

SetCreateDate sets CreateDate field to given value.

### HasCreateDate

`func (o *MediaInfo) HasCreateDate() bool`

HasCreateDate returns a boolean if a field has been set.

### GetDuration

`func (o *MediaInfo) GetDuration() int32`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *MediaInfo) GetDurationOk() (*int32, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *MediaInfo) SetDuration(v int32)`

SetDuration sets Duration field to given value.

### HasDuration

`func (o *MediaInfo) HasDuration() bool`

HasDuration returns a boolean if a field has been set.

### GetEnabled

`func (o *MediaInfo) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *MediaInfo) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *MediaInfo) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *MediaInfo) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetFileIDs

`func (o *MediaInfo) GetFileIDs() []string`

GetFileIDs returns the FileIDs field if non-nil, zero value otherwise.

### GetFileIDsOk

`func (o *MediaInfo) GetFileIDsOk() (*[]string, bool)`

GetFileIDsOk returns a tuple with the FileIDs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileIDs

`func (o *MediaInfo) SetFileIDs(v []string)`

SetFileIDs sets FileIDs field to given value.

### HasFileIDs

`func (o *MediaInfo) HasFileIDs() bool`

HasFileIDs returns a boolean if a field has been set.

### GetHdirScore

`func (o *MediaInfo) GetHdirScore() float32`

GetHdirScore returns the HdirScore field if non-nil, zero value otherwise.

### GetHdirScoreOk

`func (o *MediaInfo) GetHdirScoreOk() (*float32, bool)`

GetHdirScoreOk returns a tuple with the HdirScore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHdirScore

`func (o *MediaInfo) SetHdirScore(v float32)`

SetHdirScore sets HdirScore field to given value.

### HasHdirScore

`func (o *MediaInfo) HasHdirScore() bool`

HasHdirScore returns a boolean if a field has been set.

### GetHeight

`func (o *MediaInfo) GetHeight() int32`

GetHeight returns the Height field if non-nil, zero value otherwise.

### GetHeightOk

`func (o *MediaInfo) GetHeightOk() (*int32, bool)`

GetHeightOk returns a tuple with the Height field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeight

`func (o *MediaInfo) SetHeight(v int32)`

SetHeight sets Height field to given value.

### HasHeight

`func (o *MediaInfo) HasHeight() bool`

HasHeight returns a boolean if a field has been set.

### GetHidden

`func (o *MediaInfo) GetHidden() bool`

GetHidden returns the Hidden field if non-nil, zero value otherwise.

### GetHiddenOk

`func (o *MediaInfo) GetHiddenOk() (*bool, bool)`

GetHiddenOk returns a tuple with the Hidden field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHidden

`func (o *MediaInfo) SetHidden(v bool)`

SetHidden sets Hidden field to given value.

### HasHidden

`func (o *MediaInfo) HasHidden() bool`

HasHidden returns a boolean if a field has been set.

### GetImported

`func (o *MediaInfo) GetImported() bool`

GetImported returns the Imported field if non-nil, zero value otherwise.

### GetImportedOk

`func (o *MediaInfo) GetImportedOk() (*bool, bool)`

GetImportedOk returns a tuple with the Imported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImported

`func (o *MediaInfo) SetImported(v bool)`

SetImported sets Imported field to given value.

### HasImported

`func (o *MediaInfo) HasImported() bool`

HasImported returns a boolean if a field has been set.

### GetLikedBy

`func (o *MediaInfo) GetLikedBy() []string`

GetLikedBy returns the LikedBy field if non-nil, zero value otherwise.

### GetLikedByOk

`func (o *MediaInfo) GetLikedByOk() (*[]string, bool)`

GetLikedByOk returns a tuple with the LikedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLikedBy

`func (o *MediaInfo) SetLikedBy(v []string)`

SetLikedBy sets LikedBy field to given value.

### HasLikedBy

`func (o *MediaInfo) HasLikedBy() bool`

HasLikedBy returns a boolean if a field has been set.

### GetLocation

`func (o *MediaInfo) GetLocation() []float32`

GetLocation returns the Location field if non-nil, zero value otherwise.

### GetLocationOk

`func (o *MediaInfo) GetLocationOk() (*[]float32, bool)`

GetLocationOk returns a tuple with the Location field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocation

`func (o *MediaInfo) SetLocation(v []float32)`

SetLocation sets Location field to given value.

### HasLocation

`func (o *MediaInfo) HasLocation() bool`

HasLocation returns a boolean if a field has been set.

### GetMimeType

`func (o *MediaInfo) GetMimeType() string`

GetMimeType returns the MimeType field if non-nil, zero value otherwise.

### GetMimeTypeOk

`func (o *MediaInfo) GetMimeTypeOk() (*string, bool)`

GetMimeTypeOk returns a tuple with the MimeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMimeType

`func (o *MediaInfo) SetMimeType(v string)`

SetMimeType sets MimeType field to given value.

### HasMimeType

`func (o *MediaInfo) HasMimeType() bool`

HasMimeType returns a boolean if a field has been set.

### GetOwner

`func (o *MediaInfo) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *MediaInfo) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *MediaInfo) SetOwner(v string)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *MediaInfo) HasOwner() bool`

HasOwner returns a boolean if a field has been set.

### GetPageCount

`func (o *MediaInfo) GetPageCount() int32`

GetPageCount returns the PageCount field if non-nil, zero value otherwise.

### GetPageCountOk

`func (o *MediaInfo) GetPageCountOk() (*int32, bool)`

GetPageCountOk returns a tuple with the PageCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPageCount

`func (o *MediaInfo) SetPageCount(v int32)`

SetPageCount sets PageCount field to given value.

### HasPageCount

`func (o *MediaInfo) HasPageCount() bool`

HasPageCount returns a boolean if a field has been set.

### GetRecognitionTags

`func (o *MediaInfo) GetRecognitionTags() []string`

GetRecognitionTags returns the RecognitionTags field if non-nil, zero value otherwise.

### GetRecognitionTagsOk

`func (o *MediaInfo) GetRecognitionTagsOk() (*[]string, bool)`

GetRecognitionTagsOk returns a tuple with the RecognitionTags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecognitionTags

`func (o *MediaInfo) SetRecognitionTags(v []string)`

SetRecognitionTags sets RecognitionTags field to given value.

### HasRecognitionTags

`func (o *MediaInfo) HasRecognitionTags() bool`

HasRecognitionTags returns a boolean if a field has been set.

### GetWidth

`func (o *MediaInfo) GetWidth() int32`

GetWidth returns the Width field if non-nil, zero value otherwise.

### GetWidthOk

`func (o *MediaInfo) GetWidthOk() (*int32, bool)`

GetWidthOk returns a tuple with the Width field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWidth

`func (o *MediaInfo) SetWidth(v int32)`

SetWidth sets Width field to given value.

### HasWidth

`func (o *MediaInfo) HasWidth() bool`

HasWidth returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


