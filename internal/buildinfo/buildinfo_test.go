package buildinfo

import "testing"

func TestCurrentDefaultsAreUsable(t *testing.T) {
	info, err := Current()
	if err != nil {
		t.Fatal(err)
	}
	if info != (Info{Version: "devel", Commit: "unknown", Date: "unknown"}) {
		t.Fatalf("Current() = %#v", info)
	}
}

func TestCurrentRejectsControlCharacters(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })
	Version = "v1\nforged"
	if _, err := Current(); err == nil {
		t.Fatal("Current accepted a newline-bearing linker value")
	}
}
