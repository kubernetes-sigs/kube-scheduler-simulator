package reset

import "testing"

func TestService_Reset(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{
			name: "a",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
