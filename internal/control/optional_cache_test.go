package control

import (
	"errors"
	"os"
	"runtime"
	"syscall"
	"testing"
)

func TestOptionalCacheDirectoryRecomputesOnlyStorageFailures(t *testing.T) {
	unsafe := errors.New("cache path escapes owned directory")
	for _, test := range []struct {
		err      error
		wantErr  bool
		wantPath string
	}{
		{nil, false, "cache"}, {os.ErrPermission, false, ""}, {syscall.ENOSPC, false, ""}, {unsafe, true, ""},
		{&os.PathError{Op: "mkdir", Path: "cache", Err: os.ErrPermission}, false, ""},
	} {
		p, e := optionalCacheDirectory("cache", test.err)
		if (e != nil) != test.wantErr || p != test.wantPath {
			t.Fatalf("unexpected cache disposition: path=%q err=%v", p, e)
		}
	}
}

func TestOptionalCacheDirectoryNativeWindowsDiskFull(t *testing.T) {
	for _, code := range []syscall.Errno{112, 39} {
		path, err := optionalCacheDirectory("cache", &os.PathError{Op: "mkdir", Path: "cache", Err: code})
		if path != "" || (err == nil) != (runtime.GOOS == "windows") {
			t.Fatalf("native disk full classification: code=%d path=%q err=%v", code, path, err)
		}
	}
}
