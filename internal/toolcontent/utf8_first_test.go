package toolcontent

import (
	"encoding/base64"
	"testing"
)

func TestUTF8FirstProjectsExactTextAndPreservesBinary(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		text *string
	}{
		{name: "valid text", data: []byte("hello 🦊"), text: stringPtr("hello 🦊")},
		{name: "binary", data: []byte{0, 0xff, 'x'}},
		{name: "empty eof", data: []byte{}, text: stringPtr("")},
		{name: "rune split page prefix", data: []byte{0xf0}},
		{name: "rune split page suffix", data: []byte{0x9f, 0xa6, 0x8a}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded := base64.StdEncoding.EncodeToString(tc.data)
			gotBase64, gotText, err := UTF8First(encoded, tc.text)
			if err != nil {
				t.Fatal(err)
			}
			if tc.text == nil {
				if gotBase64 != encoded || gotText != nil {
					t.Fatal("binary or incomplete UTF-8 page changed", gotBase64, gotText)
				}
			} else if gotBase64 != "" || gotText == nil || *gotText != *tc.text {
				t.Fatal("valid UTF-8 page was not projected exactly", gotBase64, gotText)
			}
		})
	}
}

func TestUTF8FirstRejectsInvalidOrMismatchedRepresentations(t *testing.T) {
	for _, tc := range []struct {
		name    string
		encoded string
		text    *string
	}{
		{name: "malformed binary base64", encoded: "%%%"},
		{name: "noncanonical base64", encoded: "YR=="},
		{name: "text byte mismatch", encoded: base64.StdEncoding.EncodeToString([]byte("actual")), text: stringPtr("other")},
		{name: "invalid utf8 text", encoded: base64.StdEncoding.EncodeToString([]byte{0xff}), text: stringPtr("\xff")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := UTF8First(tc.encoded, tc.text); err == nil {
				t.Fatal("invalid text/base64 representation accepted")
			}
		})
	}
}

func stringPtr(value string) *string { return &value }
