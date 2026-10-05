package worktree

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDestinationCountsWindowsUTF16WorkingDirectoryBound(t *testing.T) {
	for _, tc := range []struct {
		path     string
		rejected bool
	}{
		{strings.Repeat("a", 259), false},
		{strings.Repeat("a", 260), true},
		{strings.Repeat("a", 258) + "😀", true},
		{strings.Repeat("a", 257) + "😀", false},
	} {
		err := checkDestinationPath(tc.path, "windows")
		if errors.Is(err, ErrDestinationUnsupported) != tc.rejected {
			t.Fatal("wrong Windows path boundary", len(tc.path), err)
		}
		if err := checkDestinationPath(tc.path, "linux"); err != nil {
			t.Fatal("Windows bound applied to another platform", err)
		}
	}
}

func TestDestinationPreflightPreservesHistoricalRequestValidation(t *testing.T) {
	_, r := fixture(t)
	r.Source.Root = filepath.Join(r.Source.Root, strings.Repeat("a", 180))
	var err error
	r, err = Prepare(r.RunID, r.Source)
	if err != nil {
		t.Fatal("historical long destination stopped validating", err)
	}
	if runtime.GOOS == "windows" {
		if _, err := Create(context.Background(), r); !errors.Is(err, ErrDestinationUnsupported) {
			t.Fatal("unsupported destination reached Git effects", err)
		}
	}
}
