package worktree

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type errCountingCancelContext struct {
	context.Context
	calls    int
	cancelAt int
	cancel   context.CancelFunc
}

func (c *errCountingCancelContext) Err() error {
	c.calls++
	if c.cancelAt > 0 && c.calls >= c.cancelAt {
		c.cancel()
	}
	return c.Context.Err()
}

func readFixture(t testing.TB) (Request, Binding, *Lease) {
	t.Helper()
	_, request := fixture(t)
	lease, err := Acquire(request)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := Create(context.Background(), request)
	if err != nil {
		_ = lease.Close()
		t.Fatal(err)
	}
	return request, binding, lease
}

func readCandidate(t testing.TB, binding Binding) Candidate {
	t.Helper()
	candidate, err := Fingerprint(context.Background(), binding)
	if err != nil {
		t.Fatal(err)
	}
	return candidate
}

func TestCandidateReadTracksModifiedBytesAndDrift(t *testing.T) {
	request, binding, lease := readFixture(t)
	defer lease.Close()
	content := []byte{'x', 0xff, 'y', '\n'}
	if err := os.WriteFile(filepath.Join(request.Path, "source.txt"), content, 0600); err != nil {
		t.Fatal(err)
	}
	candidate := readCandidate(t, binding)
	first, err := ReadSource(context.Background(), binding, candidate, "source.txt", 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if first.ContentBase64 != base64.StdEncoding.EncodeToString(content[:2]) || first.ContentUTF8 != nil || first.NextOffset == nil || *first.NextOffset != 2 {
		t.Fatal("binary page mismatch", first)
	}
	last, err := ReadSource(context.Background(), binding, candidate, "source.txt", 2, 2)
	if err != nil || last.ContentUTF8 == nil || *last.ContentUTF8 != "y\n" || last.NextOffset != nil || last.SHA256 != first.SHA256 {
		t.Fatal("final page mismatch", err)
	}
	for _, path := range []string{"../source.txt", ".git", "absent"} {
		if _, err := ReadSource(context.Background(), binding, candidate, path, 0, 2); err == nil {
			t.Fatal("invalid candidate path admitted", path)
		}
	}
	if _, err := ReadSource(context.Background(), binding, candidate, "source.txt", 5, 2); err == nil {
		t.Fatal("past-EOF offset admitted")
	}
	if err := os.WriteFile(filepath.Join(request.Path, "other.txt"), []byte("drift"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSource(context.Background(), binding, candidate, "source.txt", 0, 2); err == nil {
		t.Fatal("unrelated candidate drift admitted")
	}
}

func TestReadSourcesBoundsAndSelectedFileMetadata(t *testing.T) {
	request, binding, lease := readFixture(t)
	defer lease.Close()
	content := []byte("first-file-content")
	second := []byte("second")
	if err := os.WriteFile(filepath.Join(request.Path, "source.txt"), content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(request.Path, "other.txt"), second, 0600); err != nil {
		t.Fatal(err)
	}
	candidate := readCandidate(t, binding)
	got, err := ReadSources(context.Background(), binding, candidate, []string{"source.txt", "other.txt"}, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Path != "source.txt" || got[0].CandidateID == "" || string(got[0].Content) != string(content[:4]) || got[0].Size != int64(len(content)) || got[0].NextOffset == nil || *got[0].NextOffset != 4 {
		t.Fatalf("first file mismatch: %#v", got)
	}
	if got[1].Path != "other.txt" || string(got[1].Content) != string(second[:4]) || got[1].NextOffset == nil || *got[1].NextOffset != 4 || got[1].SHA256 == "" {
		t.Fatalf("second file mismatch: %#v", got[1])
	}
	for _, tc := range []struct {
		name  string
		paths []string
		limit int
	}{
		{name: "empty", paths: nil, limit: 4},
		{name: "too many", paths: make([]string, 25), limit: 4},
		{name: "invalid limit", paths: []string{"source.txt"}, limit: 32769},
		{name: "duplicate", paths: []string{"source.txt", "source.txt"}, limit: 4},
		{name: "protected", paths: []string{".git/config"}, limit: 4},
		{name: "absent", paths: []string{"absent.txt"}, limit: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ReadSources(context.Background(), binding, candidate, tc.paths, tc.limit); err == nil {
				t.Fatal("invalid batch admitted")
			}
		})
	}
}

func TestReadSourcesRejectsCandidateDriftAndSymlink(t *testing.T) {
	t.Run("unrelated drift", func(t *testing.T) {
		request, binding, lease := readFixture(t)
		defer lease.Close()
		if err := os.WriteFile(filepath.Join(request.Path, "other.txt"), []byte("before"), 0600); err != nil {
			t.Fatal(err)
		}
		candidate := readCandidate(t, binding)
		if err := os.WriteFile(filepath.Join(request.Path, "other.txt"), []byte("after"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadSources(context.Background(), binding, candidate, []string{"source.txt"}, 8); err == nil {
			t.Fatal("unrelated candidate drift admitted")
		}
	})
	t.Run("selected symlink", func(t *testing.T) {
		request, binding, lease := readFixture(t)
		defer lease.Close()
		if err := os.WriteFile(filepath.Join(request.Path, "source.txt"), []byte("regular"), 0600); err != nil {
			t.Fatal(err)
		}
		candidate := readCandidate(t, binding)
		outside := filepath.Join(t.TempDir(), "outside.txt")
		if err := os.WriteFile(outside, []byte("must not be read"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(request.Path, "source.txt")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(request.Path, "source.txt")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		if _, err := ReadSources(context.Background(), binding, candidate, []string{"source.txt"}, 8); err == nil {
			t.Fatal("symlinked source admitted")
		}
	})
}

func TestReadSourcesCancellationReturnsNoPartialBatch(t *testing.T) {
	request, binding, lease := readFixture(t)
	defer lease.Close()
	if err := os.WriteFile(filepath.Join(request.Path, "other.txt"), []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	candidate := readCandidate(t, binding)
	probeBase, probeCancel := context.WithCancel(context.Background())
	probe := &errCountingCancelContext{Context: probeBase}
	if _, _, err := Capture(probe, binding); err != nil {
		probeCancel()
		t.Fatal(err)
	}
	probeCancel()
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &errCountingCancelContext{Context: base, cancelAt: probe.calls + 1, cancel: cancel}
	files, err := ReadSources(ctx, binding, candidate, []string{"source.txt", "other.txt"}, 16)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation after candidate capture, got files=%d err=%v", len(files), err)
	}
	if files != nil {
		t.Fatalf("canceled batch returned partial results: %d", len(files))
	}
}

func BenchmarkReadSourcesAgainstIndividualReads(b *testing.B) {
	const candidateFiles = 256
	const candidateFileBytes = 4096
	request, binding, lease := readFixture(b)
	defer lease.Close()
	content := make([]byte, candidateFileBytes)
	for i := range content {
		content[i] = 'x'
	}
	paths := make([]string, candidateFiles)
	for i := range paths {
		paths[i] = fmt.Sprintf("bench-%04d.go", i)
		if err := os.WriteFile(filepath.Join(request.Path, paths[i]), content, 0600); err != nil {
			b.Fatal(err)
		}
	}
	candidate := readCandidate(b, binding)
	ctx := context.Background()
	for _, count := range []int{1, 8, 24} {
		selected := paths[:count]
		b.Run(fmt.Sprintf("selected_%02d/individual", count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, path := range selected {
					chunk, err := ReadSource(ctx, binding, candidate, path, 0, 32768)
					if err != nil {
						b.Fatal(err)
					}
					if _, err := base64.StdEncoding.DecodeString(chunk.ContentBase64); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
		b.Run(fmt.Sprintf("selected_%02d/batch", count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := ReadSources(ctx, binding, candidate, selected, 32768); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
