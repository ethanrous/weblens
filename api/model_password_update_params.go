/*
Weblens API

Programmatic access to the Weblens server

API version: 1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the PasswordUpdateParams type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PasswordUpdateParams{}

// PasswordUpdateParams struct for PasswordUpdateParams
type PasswordUpdateParams struct {
	NewPassword string `json:"newPassword"`
	OldPassword *string `json:"oldPassword,omitempty"`
}

type _PasswordUpdateParams PasswordUpdateParams

// NewPasswordUpdateParams instantiates a new PasswordUpdateParams object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPasswordUpdateParams(newPassword string) *PasswordUpdateParams {
	this := PasswordUpdateParams{}
	this.NewPassword = newPassword
	return &this
}

// NewPasswordUpdateParamsWithDefaults instantiates a new PasswordUpdateParams object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPasswordUpdateParamsWithDefaults() *PasswordUpdateParams {
	this := PasswordUpdateParams{}
	return &this
}

// GetNewPassword returns the NewPassword field value
func (o *PasswordUpdateParams) GetNewPassword() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.NewPassword
}

// GetNewPasswordOk returns a tuple with the NewPassword field value
// and a boolean to check if the value has been set.
func (o *PasswordUpdateParams) GetNewPasswordOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NewPassword, true
}

// SetNewPassword sets field value
func (o *PasswordUpdateParams) SetNewPassword(v string) {
	o.NewPassword = v
}

// GetOldPassword returns the OldPassword field value if set, zero value otherwise.
func (o *PasswordUpdateParams) GetOldPassword() string {
	if o == nil || IsNil(o.OldPassword) {
		var ret string
		return ret
	}
	return *o.OldPassword
}

// GetOldPasswordOk returns a tuple with the OldPassword field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PasswordUpdateParams) GetOldPasswordOk() (*string, bool) {
	if o == nil || IsNil(o.OldPassword) {
		return nil, false
	}
	return o.OldPassword, true
}

// HasOldPassword returns a boolean if a field has been set.
func (o *PasswordUpdateParams) HasOldPassword() bool {
	if o != nil && !IsNil(o.OldPassword) {
		return true
	}

	return false
}

// SetOldPassword gets a reference to the given string and assigns it to the OldPassword field.
func (o *PasswordUpdateParams) SetOldPassword(v string) {
	o.OldPassword = &v
}

func (o PasswordUpdateParams) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PasswordUpdateParams) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["newPassword"] = o.NewPassword
	if !IsNil(o.OldPassword) {
		toSerialize["oldPassword"] = o.OldPassword
	}
	return toSerialize, nil
}

func (o *PasswordUpdateParams) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"newPassword",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varPasswordUpdateParams := _PasswordUpdateParams{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varPasswordUpdateParams)

	if err != nil {
		return err
	}

	*o = PasswordUpdateParams(varPasswordUpdateParams)

	return err
}

type NullablePasswordUpdateParams struct {
	value *PasswordUpdateParams
	isSet bool
}

func (v NullablePasswordUpdateParams) Get() *PasswordUpdateParams {
	return v.value
}

func (v *NullablePasswordUpdateParams) Set(val *PasswordUpdateParams) {
	v.value = val
	v.isSet = true
}

func (v NullablePasswordUpdateParams) IsSet() bool {
	return v.isSet
}

func (v *NullablePasswordUpdateParams) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePasswordUpdateParams(val *PasswordUpdateParams) *NullablePasswordUpdateParams {
	return &NullablePasswordUpdateParams{value: val, isSet: true}
}

func (v NullablePasswordUpdateParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePasswordUpdateParams) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


