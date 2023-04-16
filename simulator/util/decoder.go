package util

import (
	"reflect"
	"strings"
)

// PrivateFieldsDecoder is a function that provides access to the values of private fields
// within a specified struct. The stru parameter should be a pointer to a struct containing
// the private field(s). The privateFieldName parameter should be the name of the private
// field you wish to access. If you need to access a subfield, specify the field names
// separated by dots (e.g. "privateField2.privateField").
//
// The function returns a reflect.Value object containing the value of the specified private field.
// See TestPrivateFieldsDecoder in decoder_test.go to know how to access your wanted field via reflect.Value.
func PrivateFieldsDecoder(stru interface{}, privateFieldName string) reflect.Value {
	value := reflect.ValueOf(stru).Elem()
	fieldNames := strings.Split(privateFieldName, ".")

	for _, fieldName := range fieldNames {
		value = value.FieldByName(fieldName)
	}

	return value
}
