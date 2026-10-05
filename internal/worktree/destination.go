package worktree

import (
	"errors"
	"runtime"
	"unicode/utf16"
)

// ErrDestinationUnsupported means the installed Git working-directory route
// cannot observe the proposed workspace. It never settles an existing effect.
var ErrDestinationUnsupported = errors.New("Git workspace path exceeds the Windows working-directory limit; use a shorter repository checkout path")

// CheckDestination is a pre-effect platform check, deliberately separate from
// Request.Validate so historical pending requests remain replayable/observable.
func CheckDestination(r Request) error {
	if err := r.Validate(); err != nil {
		return err
	}
	return checkDestinationPath(r.Path, runtime.GOOS)
}

func checkDestinationPath(path, goos string) error {
	// Git for Windows -C still rejects a working directory at MAX_PATH even
	// with core.longpaths=true. Count UTF-16 units, including the terminating NUL.
	if goos == "windows" && len(utf16.Encode([]rune(path))) >= 260 {
		return ErrDestinationUnsupported
	}
	return nil
}
