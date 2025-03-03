package recorder

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestHoge(t *testing.T) {
	bytes := `[{"time":"2025-03-03T16:28:00.957176231+09:00","event":"Add","resource":{"Object":{"apiVersion":"v1","kind":"Pod","metadata":{"name":"pod-1","namespace":"default"}}}}]`

	record := []Record{}
	if err := json.Unmarshal([]byte(bytes), &record); err != nil {
		t.Fatalf("failed to unmarshal record: %v", err)
	}

	fmt.Println(record)
}
