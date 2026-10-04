// Package toolcontent contains presentation rules for model-visible tool
// results. It does not change the underlying source or candidate observations.
package toolcontent

import (
	"bytes"
	"encoding/base64"
	"errors"
	"unicode/utf8"
)

// UTF8First omits redundant Base64 only when the text is valid UTF-8 and is
// byte-for-byte equal to a canonical Base64 decoding. Binary content keeps its
// Base64 representation unchanged.
func UTF8First(encoded string, text *string) (string, *string, error) {
	decoded, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(decoded) != encoded {
		return "", nil, errors.New("tool base64 representation is not canonical")
	}
	if text == nil {
		return encoded, nil, nil
	}
	if !utf8.ValidString(*text) || !bytes.Equal(decoded, []byte(*text)) {
		return "", nil, errors.New("tool text and byte representations differ")
	}
	return "", text, nil
}
