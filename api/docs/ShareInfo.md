# ShareInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Accessors** | Pointer to [**[]UserInfo**](UserInfo.md) |  | [optional] 
**Enabled** | Pointer to **bool** |  | [optional] 
**Expires** | Pointer to **int32** |  | [optional] 
**FileId** | Pointer to **string** |  | [optional] 
**Owner** | Pointer to **string** |  | [optional] 
**Permissions** | Pointer to [**map[string]PermissionsInfo**](PermissionsInfo.md) |  | [optional] 
**Public** | Pointer to **bool** |  | [optional] 
**ShareId** | Pointer to **string** |  | [optional] 
**ShareName** | Pointer to **string** |  | [optional] 
**ShareType** | Pointer to **string** |  | [optional] 
**TimelineOnly** | Pointer to **bool** |  | [optional] 
**Updated** | Pointer to **int32** |  | [optional] 
**Wormhole** | Pointer to **bool** |  | [optional] 

## Methods

### NewShareInfo

`func NewShareInfo() *ShareInfo`

NewShareInfo instantiates a new ShareInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewShareInfoWithDefaults

`func NewShareInfoWithDefaults() *ShareInfo`

NewShareInfoWithDefaults instantiates a new ShareInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessors

`func (o *ShareInfo) GetAccessors() []UserInfo`

GetAccessors returns the Accessors field if non-nil, zero value otherwise.

### GetAccessorsOk

`func (o *ShareInfo) GetAccessorsOk() (*[]UserInfo, bool)`

GetAccessorsOk returns a tuple with the Accessors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessors

`func (o *ShareInfo) SetAccessors(v []UserInfo)`

SetAccessors sets Accessors field to given value.

### HasAccessors

`func (o *ShareInfo) HasAccessors() bool`

HasAccessors returns a boolean if a field has been set.

### GetEnabled

`func (o *ShareInfo) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *ShareInfo) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *ShareInfo) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *ShareInfo) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetExpires

`func (o *ShareInfo) GetExpires() int32`

GetExpires returns the Expires field if non-nil, zero value otherwise.

### GetExpiresOk

`func (o *ShareInfo) GetExpiresOk() (*int32, bool)`

GetExpiresOk returns a tuple with the Expires field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpires

`func (o *ShareInfo) SetExpires(v int32)`

SetExpires sets Expires field to given value.

### HasExpires

`func (o *ShareInfo) HasExpires() bool`

HasExpires returns a boolean if a field has been set.

### GetFileId

`func (o *ShareInfo) GetFileId() string`

GetFileId returns the FileId field if non-nil, zero value otherwise.

### GetFileIdOk

`func (o *ShareInfo) GetFileIdOk() (*string, bool)`

GetFileIdOk returns a tuple with the FileId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileId

`func (o *ShareInfo) SetFileId(v string)`

SetFileId sets FileId field to given value.

### HasFileId

`func (o *ShareInfo) HasFileId() bool`

HasFileId returns a boolean if a field has been set.

### GetOwner

`func (o *ShareInfo) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *ShareInfo) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *ShareInfo) SetOwner(v string)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *ShareInfo) HasOwner() bool`

HasOwner returns a boolean if a field has been set.

### GetPermissions

`func (o *ShareInfo) GetPermissions() map[string]PermissionsInfo`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *ShareInfo) GetPermissionsOk() (*map[string]PermissionsInfo, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *ShareInfo) SetPermissions(v map[string]PermissionsInfo)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *ShareInfo) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetPublic

`func (o *ShareInfo) GetPublic() bool`

GetPublic returns the Public field if non-nil, zero value otherwise.

### GetPublicOk

`func (o *ShareInfo) GetPublicOk() (*bool, bool)`

GetPublicOk returns a tuple with the Public field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublic

`func (o *ShareInfo) SetPublic(v bool)`

SetPublic sets Public field to given value.

### HasPublic

`func (o *ShareInfo) HasPublic() bool`

HasPublic returns a boolean if a field has been set.

### GetShareId

`func (o *ShareInfo) GetShareId() string`

GetShareId returns the ShareId field if non-nil, zero value otherwise.

### GetShareIdOk

`func (o *ShareInfo) GetShareIdOk() (*string, bool)`

GetShareIdOk returns a tuple with the ShareId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShareId

`func (o *ShareInfo) SetShareId(v string)`

SetShareId sets ShareId field to given value.

### HasShareId

`func (o *ShareInfo) HasShareId() bool`

HasShareId returns a boolean if a field has been set.

### GetShareName

`func (o *ShareInfo) GetShareName() string`

GetShareName returns the ShareName field if non-nil, zero value otherwise.

### GetShareNameOk

`func (o *ShareInfo) GetShareNameOk() (*string, bool)`

GetShareNameOk returns a tuple with the ShareName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShareName

`func (o *ShareInfo) SetShareName(v string)`

SetShareName sets ShareName field to given value.

### HasShareName

`func (o *ShareInfo) HasShareName() bool`

HasShareName returns a boolean if a field has been set.

### GetShareType

`func (o *ShareInfo) GetShareType() string`

GetShareType returns the ShareType field if non-nil, zero value otherwise.

### GetShareTypeOk

`func (o *ShareInfo) GetShareTypeOk() (*string, bool)`

GetShareTypeOk returns a tuple with the ShareType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShareType

`func (o *ShareInfo) SetShareType(v string)`

SetShareType sets ShareType field to given value.

### HasShareType

`func (o *ShareInfo) HasShareType() bool`

HasShareType returns a boolean if a field has been set.

### GetTimelineOnly

`func (o *ShareInfo) GetTimelineOnly() bool`

GetTimelineOnly returns the TimelineOnly field if non-nil, zero value otherwise.

### GetTimelineOnlyOk

`func (o *ShareInfo) GetTimelineOnlyOk() (*bool, bool)`

GetTimelineOnlyOk returns a tuple with the TimelineOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimelineOnly

`func (o *ShareInfo) SetTimelineOnly(v bool)`

SetTimelineOnly sets TimelineOnly field to given value.

### HasTimelineOnly

`func (o *ShareInfo) HasTimelineOnly() bool`

HasTimelineOnly returns a boolean if a field has been set.

### GetUpdated

`func (o *ShareInfo) GetUpdated() int32`

GetUpdated returns the Updated field if non-nil, zero value otherwise.

### GetUpdatedOk

`func (o *ShareInfo) GetUpdatedOk() (*int32, bool)`

GetUpdatedOk returns a tuple with the Updated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdated

`func (o *ShareInfo) SetUpdated(v int32)`

SetUpdated sets Updated field to given value.

### HasUpdated

`func (o *ShareInfo) HasUpdated() bool`

HasUpdated returns a boolean if a field has been set.

### GetWormhole

`func (o *ShareInfo) GetWormhole() bool`

GetWormhole returns the Wormhole field if non-nil, zero value otherwise.

### GetWormholeOk

`func (o *ShareInfo) GetWormholeOk() (*bool, bool)`

GetWormholeOk returns a tuple with the Wormhole field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWormhole

`func (o *ShareInfo) SetWormhole(v bool)`

SetWormhole sets Wormhole field to given value.

### HasWormhole

`func (o *ShareInfo) HasWormhole() bool`

HasWormhole returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


