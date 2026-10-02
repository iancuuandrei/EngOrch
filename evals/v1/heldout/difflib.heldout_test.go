package difflib

import (
	"reflect"
	"testing"
)

func TestFabricV1Heldout(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", []string{}},
		{"alpha", []string{"alpha\n"}},
		{"alpha\n", []string{"alpha\n"}},
		{"alpha\nbeta\n", []string{"alpha\n", "beta\n"}},
	}
	for _, tc := range cases {
		if got := SplitLines(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("SplitLines(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}
