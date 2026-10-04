package control

import (
	"errors"
	"os"
	"runtime"
	"syscall"
)

// Cache storage is optional. A permission or storage-capacity failure permits
// recomputation; invalid identities, unsafe paths and other failures do not.
func optionalCacheDirectory(path string, err error) (string, error) {
	if err == nil {
		return path, nil
	}
	// Windows reports ERROR_DISK_FULL or ERROR_HANDLE_DISK_FULL rather than
	// the portable ENOSPC value. Restrict these native codes to Windows.
	windowsFull := runtime.GOOS == "windows" && (errors.Is(err, syscall.Errno(112)) || errors.Is(err, syscall.Errno(39)))
	if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.ENOSPC) || windowsFull {
		return "", nil
	}
	return "", err
}
