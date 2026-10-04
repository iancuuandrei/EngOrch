package decoder

import (
	"strings"
	"testing"
)

func TestExistingSmallRecords(t *testing.T) {
	records, err := Decode(strings.NewReader("{\"n\":1}\n\n[2,3]\n"))
	if err != nil || len(records) != 2 || string(records[0]) != "{\"n\":1}" || string(records[1]) != "[2,3]" {
		t.Fatalf("unexpected records: %v, %v", records, err)
	}
}
