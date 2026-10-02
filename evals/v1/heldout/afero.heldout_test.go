package afero_test

import (
	"errors"
	"syscall"
	"testing"

	"github.com/spf13/afero"
)

// External test package on purpose: the fixture lives inside the afero
// repository, so `package afero` plus an import of
// "github.com/spf13/afero" is a self-import (import cycle). The external
// test package exercises the same behavior without the cycle.
func TestFabricV1Heldout(t *testing.T) {
	base, layer := afero.NewMemMapFs(), afero.NewMemMapFs()
	if err := afero.WriteFile(base, "dir/base.txt", []byte("base"), 0600); err != nil {
		t.Fatal(err)
	}
	fs := afero.NewCopyOnWriteFs(base, layer)
	err := fs.RemoveAll("dir")
	if !errors.Is(err, syscall.EPERM) {
		t.Fatalf("base-only RemoveAll error = %v, want EPERM", err)
	}
	if got, e := afero.ReadFile(base, "dir/base.txt"); e != nil || string(got) != "base" {
		t.Fatalf("base changed: %q, %v", got, e)
	}
}
