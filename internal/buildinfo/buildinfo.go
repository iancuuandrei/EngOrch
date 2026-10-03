// Package buildinfo exposes the immutable identity stamped into a Fabric binary.
package buildinfo

import (
	"fmt"
	"strings"
)

// These values are deliberately variables so a release build can set them with
// -ldflags. Development builds retain honest, non-release defaults.
var (
	Version = "devel"
	Commit  = "unknown"
	Date    = "unknown"
)

// Info is the complete identity printed by a Fabric executable.
type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// Current returns the build identity after checking that linker values did not
// introduce control characters into the CLI or machine-readable output.
func Current() (Info, error) {
	info := Info{Version: Version, Commit: Commit, Date: Date}
	for name, value := range map[string]string{
		"version": info.Version,
		"commit":  info.Commit,
		"date":    info.Date,
	} {
		if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\r\n\x00") {
			return Info{}, fmt.Errorf("invalid build %s", name)
		}
	}
	return info, nil
}
