/*
Weblens API

Programmatic access to the Weblens server

API version: 1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// checks if the BackupInfo type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &BackupInfo{}

// BackupInfo struct for BackupInfo
type BackupInfo struct {
	FileHistory []FileActionInfo `json:"fileHistory,omitempty"`
	Instances []TowerInfo `json:"instances,omitempty"`
	LifetimesCount *int32 `json:"lifetimesCount,omitempty"`
	Tokens []TokenInfo `json:"tokens,omitempty"`
	Users []UserInfoArchive `json:"users,omitempty"`
}

// NewBackupInfo instantiates a new BackupInfo object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewBackupInfo() *BackupInfo {
	this := BackupInfo{}
	return &this
}

// NewBackupInfoWithDefaults instantiates a new BackupInfo object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewBackupInfoWithDefaults() *BackupInfo {
	this := BackupInfo{}
	return &this
}

// GetFileHistory returns the FileHistory field value if set, zero value otherwise.
func (o *BackupInfo) GetFileHistory() []FileActionInfo {
	if o == nil || IsNil(o.FileHistory) {
		var ret []FileActionInfo
		return ret
	}
	return o.FileHistory
}

// GetFileHistoryOk returns a tuple with the FileHistory field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BackupInfo) GetFileHistoryOk() ([]FileActionInfo, bool) {
	if o == nil || IsNil(o.FileHistory) {
		return nil, false
	}
	return o.FileHistory, true
}

// HasFileHistory returns a boolean if a field has been set.
func (o *BackupInfo) HasFileHistory() bool {
	if o != nil && !IsNil(o.FileHistory) {
		return true
	}

	return false
}

// SetFileHistory gets a reference to the given []FileActionInfo and assigns it to the FileHistory field.
func (o *BackupInfo) SetFileHistory(v []FileActionInfo) {
	o.FileHistory = v
}

// GetInstances returns the Instances field value if set, zero value otherwise.
func (o *BackupInfo) GetInstances() []TowerInfo {
	if o == nil || IsNil(o.Instances) {
		var ret []TowerInfo
		return ret
	}
	return o.Instances
}

// GetInstancesOk returns a tuple with the Instances field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BackupInfo) GetInstancesOk() ([]TowerInfo, bool) {
	if o == nil || IsNil(o.Instances) {
		return nil, false
	}
	return o.Instances, true
}

// HasInstances returns a boolean if a field has been set.
func (o *BackupInfo) HasInstances() bool {
	if o != nil && !IsNil(o.Instances) {
		return true
	}

	return false
}

// SetInstances gets a reference to the given []TowerInfo and assigns it to the Instances field.
func (o *BackupInfo) SetInstances(v []TowerInfo) {
	o.Instances = v
}

// GetLifetimesCount returns the LifetimesCount field value if set, zero value otherwise.
func (o *BackupInfo) GetLifetimesCount() int32 {
	if o == nil || IsNil(o.LifetimesCount) {
		var ret int32
		return ret
	}
	return *o.LifetimesCount
}

// GetLifetimesCountOk returns a tuple with the LifetimesCount field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BackupInfo) GetLifetimesCountOk() (*int32, bool) {
	if o == nil || IsNil(o.LifetimesCount) {
		return nil, false
	}
	return o.LifetimesCount, true
}

// HasLifetimesCount returns a boolean if a field has been set.
func (o *BackupInfo) HasLifetimesCount() bool {
	if o != nil && !IsNil(o.LifetimesCount) {
		return true
	}

	return false
}

// SetLifetimesCount gets a reference to the given int32 and assigns it to the LifetimesCount field.
func (o *BackupInfo) SetLifetimesCount(v int32) {
	o.LifetimesCount = &v
}

// GetTokens returns the Tokens field value if set, zero value otherwise.
func (o *BackupInfo) GetTokens() []TokenInfo {
	if o == nil || IsNil(o.Tokens) {
		var ret []TokenInfo
		return ret
	}
	return o.Tokens
}

// GetTokensOk returns a tuple with the Tokens field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BackupInfo) GetTokensOk() ([]TokenInfo, bool) {
	if o == nil || IsNil(o.Tokens) {
		return nil, false
	}
	return o.Tokens, true
}

// HasTokens returns a boolean if a field has been set.
func (o *BackupInfo) HasTokens() bool {
	if o != nil && !IsNil(o.Tokens) {
		return true
	}

	return false
}

// SetTokens gets a reference to the given []TokenInfo and assigns it to the Tokens field.
func (o *BackupInfo) SetTokens(v []TokenInfo) {
	o.Tokens = v
}

// GetUsers returns the Users field value if set, zero value otherwise.
func (o *BackupInfo) GetUsers() []UserInfoArchive {
	if o == nil || IsNil(o.Users) {
		var ret []UserInfoArchive
		return ret
	}
	return o.Users
}

// GetUsersOk returns a tuple with the Users field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BackupInfo) GetUsersOk() ([]UserInfoArchive, bool) {
	if o == nil || IsNil(o.Users) {
		return nil, false
	}
	return o.Users, true
}

// HasUsers returns a boolean if a field has been set.
func (o *BackupInfo) HasUsers() bool {
	if o != nil && !IsNil(o.Users) {
		return true
	}

	return false
}

// SetUsers gets a reference to the given []UserInfoArchive and assigns it to the Users field.
func (o *BackupInfo) SetUsers(v []UserInfoArchive) {
	o.Users = v
}

func (o BackupInfo) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o BackupInfo) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.FileHistory) {
		toSerialize["fileHistory"] = o.FileHistory
	}
	if !IsNil(o.Instances) {
		toSerialize["instances"] = o.Instances
	}
	if !IsNil(o.LifetimesCount) {
		toSerialize["lifetimesCount"] = o.LifetimesCount
	}
	if !IsNil(o.Tokens) {
		toSerialize["tokens"] = o.Tokens
	}
	if !IsNil(o.Users) {
		toSerialize["users"] = o.Users
	}
	return toSerialize, nil
}

type NullableBackupInfo struct {
	value *BackupInfo
	isSet bool
}

func (v NullableBackupInfo) Get() *BackupInfo {
	return v.value
}

func (v *NullableBackupInfo) Set(val *BackupInfo) {
	v.value = val
	v.isSet = true
}

func (v NullableBackupInfo) IsSet() bool {
	return v.isSet
}

func (v *NullableBackupInfo) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableBackupInfo(val *BackupInfo) *NullableBackupInfo {
	return &NullableBackupInfo{value: val, isSet: true}
}

func (v NullableBackupInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableBackupInfo) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


