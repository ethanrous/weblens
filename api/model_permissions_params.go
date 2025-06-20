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

// checks if the PermissionsParams type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PermissionsParams{}

// PermissionsParams struct for PermissionsParams
type PermissionsParams struct {
	CanDelete *bool `json:"canDelete,omitempty"`
	CanDownload *bool `json:"canDownload,omitempty"`
	CanEdit *bool `json:"canEdit,omitempty"`
}

// NewPermissionsParams instantiates a new PermissionsParams object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPermissionsParams() *PermissionsParams {
	this := PermissionsParams{}
	return &this
}

// NewPermissionsParamsWithDefaults instantiates a new PermissionsParams object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPermissionsParamsWithDefaults() *PermissionsParams {
	this := PermissionsParams{}
	return &this
}

// GetCanDelete returns the CanDelete field value if set, zero value otherwise.
func (o *PermissionsParams) GetCanDelete() bool {
	if o == nil || IsNil(o.CanDelete) {
		var ret bool
		return ret
	}
	return *o.CanDelete
}

// GetCanDeleteOk returns a tuple with the CanDelete field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PermissionsParams) GetCanDeleteOk() (*bool, bool) {
	if o == nil || IsNil(o.CanDelete) {
		return nil, false
	}
	return o.CanDelete, true
}

// HasCanDelete returns a boolean if a field has been set.
func (o *PermissionsParams) HasCanDelete() bool {
	if o != nil && !IsNil(o.CanDelete) {
		return true
	}

	return false
}

// SetCanDelete gets a reference to the given bool and assigns it to the CanDelete field.
func (o *PermissionsParams) SetCanDelete(v bool) {
	o.CanDelete = &v
}

// GetCanDownload returns the CanDownload field value if set, zero value otherwise.
func (o *PermissionsParams) GetCanDownload() bool {
	if o == nil || IsNil(o.CanDownload) {
		var ret bool
		return ret
	}
	return *o.CanDownload
}

// GetCanDownloadOk returns a tuple with the CanDownload field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PermissionsParams) GetCanDownloadOk() (*bool, bool) {
	if o == nil || IsNil(o.CanDownload) {
		return nil, false
	}
	return o.CanDownload, true
}

// HasCanDownload returns a boolean if a field has been set.
func (o *PermissionsParams) HasCanDownload() bool {
	if o != nil && !IsNil(o.CanDownload) {
		return true
	}

	return false
}

// SetCanDownload gets a reference to the given bool and assigns it to the CanDownload field.
func (o *PermissionsParams) SetCanDownload(v bool) {
	o.CanDownload = &v
}

// GetCanEdit returns the CanEdit field value if set, zero value otherwise.
func (o *PermissionsParams) GetCanEdit() bool {
	if o == nil || IsNil(o.CanEdit) {
		var ret bool
		return ret
	}
	return *o.CanEdit
}

// GetCanEditOk returns a tuple with the CanEdit field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PermissionsParams) GetCanEditOk() (*bool, bool) {
	if o == nil || IsNil(o.CanEdit) {
		return nil, false
	}
	return o.CanEdit, true
}

// HasCanEdit returns a boolean if a field has been set.
func (o *PermissionsParams) HasCanEdit() bool {
	if o != nil && !IsNil(o.CanEdit) {
		return true
	}

	return false
}

// SetCanEdit gets a reference to the given bool and assigns it to the CanEdit field.
func (o *PermissionsParams) SetCanEdit(v bool) {
	o.CanEdit = &v
}

func (o PermissionsParams) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PermissionsParams) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.CanDelete) {
		toSerialize["canDelete"] = o.CanDelete
	}
	if !IsNil(o.CanDownload) {
		toSerialize["canDownload"] = o.CanDownload
	}
	if !IsNil(o.CanEdit) {
		toSerialize["canEdit"] = o.CanEdit
	}
	return toSerialize, nil
}

type NullablePermissionsParams struct {
	value *PermissionsParams
	isSet bool
}

func (v NullablePermissionsParams) Get() *PermissionsParams {
	return v.value
}

func (v *NullablePermissionsParams) Set(val *PermissionsParams) {
	v.value = val
	v.isSet = true
}

func (v NullablePermissionsParams) IsSet() bool {
	return v.isSet
}

func (v *NullablePermissionsParams) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePermissionsParams(val *PermissionsParams) *NullablePermissionsParams {
	return &NullablePermissionsParams{value: val, isSet: true}
}

func (v NullablePermissionsParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePermissionsParams) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


