package util

import (
	"reflect"
	"testing"
)

func TestPrivateFieldsDecoder(t *testing.T) {
	t.Parallel()
	type args struct {
		stru             interface{}
		privateFieldName string
	}
	type NestedEmbedded struct {
		nestedHiddenField int
	}
	type Embedded struct {
		hiddenField string
		nested      NestedEmbedded
	}
	type MyStruct struct {
		privateInt    int
		privateString string
		privateSlice  []int
		privateMap    map[string]int
		embedded      Embedded
	}
	tests := []struct {
		name     string
		args     args
		assertFn func(a reflect.Value) bool
	}{
		{
			name: "Accessing private int field",
			args: args{
				stru:             &MyStruct{privateInt: 42},
				privateFieldName: "privateInt",
			},
			assertFn: func(a reflect.Value) bool {
				return a.Int() == 42
			},
		},
		{
			name: "Accessing private string field",
			args: args{
				stru:             &MyStruct{privateString: "Hello, world!"},
				privateFieldName: "privateString",
			},
			assertFn: func(a reflect.Value) bool {
				return a.String() == "Hello, world!"
			},
		},
		{
			name: "Accessing private slice field",
			args: args{
				stru:             &MyStruct{privateSlice: []int{1, 2, 3}},
				privateFieldName: "privateSlice",
			},
			assertFn: func(a reflect.Value) bool {
				return a.Len() == 3 && a.Index(0).Int() == 1 && a.Index(1).Int() == 2 && a.Index(2).Int() == 3
			},
		},
		{
			name: "Accessing private map field",
			args: args{
				stru:             &MyStruct{privateMap: map[string]int{"one": 1, "two": 2}},
				privateFieldName: "privateMap",
			},
			assertFn: func(a reflect.Value) bool {
				expectedMap := map[string]int{"one": 1, "two": 2}
				mapKeys := a.MapKeys()

				if len(mapKeys) != len(expectedMap) {
					return false
				}

				for _, key := range mapKeys {
					expectedValue, ok := expectedMap[key.String()]
					if !ok || a.MapIndex(key).Int() != int64(expectedValue) {
						return false
					}
				}

				return true
			},
		},
		{
			name: "Accessing embedded private field",
			args: args{
				stru:             &MyStruct{embedded: Embedded{hiddenField: "Hidden message"}},
				privateFieldName: "embedded.hiddenField",
			},
			assertFn: func(a reflect.Value) bool {
				return a.String() == "Hidden message"
			},
		},
		{
			name: "Accessing nested private struct field",
			args: args{
				stru: &MyStruct{
					embedded: Embedded{
						nested: NestedEmbedded{
							nestedHiddenField: 99,
						},
					},
				},
				privateFieldName: "embedded.nested.nestedHiddenField",
			},
			assertFn: func(a reflect.Value) bool {
				return a.Int() == 99
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := PrivateFieldsDecoder(tt.args.stru, tt.args.privateFieldName)
			if !tt.assertFn(got) {
				t.Errorf("PrivateFieldsDecoder() = %v", got)
			}
		})
	}
}
