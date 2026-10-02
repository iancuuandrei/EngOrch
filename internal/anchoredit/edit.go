// Package anchoredit applies bounded, exact UTF-8 replacements to original
// candidate bytes while preserving every byte outside the selected anchors.
package anchoredit

import (
	"errors"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	// MaxSourceBytes bounds existing and composed source files.
	MaxSourceBytes = 256 << 10
	// MaxEditBytes bounds the combined before and after text in one change set.
	MaxEditBytes = 64 << 10
	// MaxEdits bounds the number of replacements in one file.
	MaxEdits = 64
)

// Edit replaces the exact Before anchor with After. Before is located exactly
// once in the original source; edits never search output produced by another.
type Edit struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type replacement struct {
	start int
	end   int
	after string
}

// Apply returns source with the unique, non-overlapping anchors replaced.
func Apply(source []byte, edits []Edit) ([]byte, error) {
	if len(source) == 0 || len(source) > MaxSourceBytes || !utf8.Valid(source) {
		return nil, errors.New("anchored edit source must be bounded nonempty UTF-8")
	}
	if len(edits) == 0 || len(edits) > MaxEdits {
		return nil, errors.New("anchored edit count outside bounds")
	}
	total := 0
	replacements := make([]replacement, 0, len(edits))
	text := string(source)
	for _, edit := range edits {
		if edit.Before == "" || !utf8.ValidString(edit.Before) || !utf8.ValidString(edit.After) {
			return nil, errors.New("anchored edit requires a nonempty UTF-8 before anchor")
		}
		total += len(edit.Before) + len(edit.After)
		if total > MaxEditBytes {
			return nil, errors.New("anchored edit payload exceeds bound")
		}
		start := strings.Index(text, edit.Before)
		if start < 0 || strings.Index(text[start+1:], edit.Before) >= 0 {
			return nil, errors.New("before anchor must occur exactly once in original source")
		}
		replacements = append(replacements, replacement{start: start, end: start + len(edit.Before), after: edit.After})
	}
	ordered := append([]replacement(nil), replacements...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].start < ordered[j].start })
	for i := 1; i < len(ordered); i++ {
		if ordered[i].start < ordered[i-1].end {
			return nil, errors.New("anchored edits overlap in original source")
		}
	}
	sort.Slice(replacements, func(i, j int) bool { return replacements[i].start > replacements[j].start })
	result := append([]byte(nil), source...)
	for _, item := range replacements {
		result = append(result[:item.start], append([]byte(item.after), result[item.end:]...)...)
		if len(result) > MaxSourceBytes {
			return nil, errors.New("anchored edit result exceeds source bound")
		}
	}
	if !utf8.Valid(result) {
		return nil, errors.New("anchored edit result is not valid UTF-8")
	}
	return result, nil
}

// ValidateNewFile bounds explicit content for an absent path. Empty text is a
// valid new empty file; null content is never a deletion operation.
func ValidateNewFile(content string) error {
	if len(content) > MaxSourceBytes || !utf8.ValidString(content) {
		return errors.New("new file content must be bounded UTF-8")
	}
	return nil
}
