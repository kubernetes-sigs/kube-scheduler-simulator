package v1alpha1

import (
	"testing"
)

func TestValidateOperations(t *testing.T) {
	cases := []struct {
		name    string
		op      ScenarioOperation
		wantErr bool
	}{
		{"one Op", ScenarioOperation{Create: &CreateOperation{}}, false},
		{"two Ops", ScenarioOperation{Create: &CreateOperation{}, Delete: &DeleteOperation{}}, true},
		{"zero Op", ScenarioOperation{}, true},
	}
	for _, tc := range cases {
		got := tc.op.ValidateCreate()
		if (got != nil) != tc.wantErr {
			t.Errorf("%s: unexpected error=%v", tc.name, got)
		}
	}
}
