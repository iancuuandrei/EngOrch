package godotenv

import "testing"

func TestFabricV1Heldout(t *testing.T) {
	t.Run("quoted multiline values preserve quote-specific expansion and line breaks", func(t *testing.T) {
		input := "NAME=world\n" +
			"INLINE_SINGLE='${NAME}'\n" +
			"INLINE_DOUBLE=\"${NAME}\"\n" +
			"SINGLE='first line\n\nsecond ${NAME}' # single comment\n" +
			"DOUBLE=\"first line\nsecond ${NAME}\" # double comment\n" +
			"ESCAPED=\"first line\nquote: \\\"inside\\\"\nlast line\" # close comment\n" +
			"UNICODE='π first line\n\nsecond 🧭' # unicode comment\n" +
			"NEXT=value # assignment comment\n"
		got, err := Unmarshal(input)
		if err != nil {
			t.Fatalf("Unmarshal multiline input: %v", err)
		}
		want := map[string]string{
			"NAME":    "world",
			"INLINE_SINGLE": "${NAME}",
			"INLINE_DOUBLE": "world",
			"SINGLE":  "first line\n\nsecond ${NAME}",
			"DOUBLE":  "first line\nsecond world",
			"ESCAPED": "first line\nquote: \"inside\"\nlast line",
			"UNICODE": "π first line\n\nsecond 🧭",
			"NEXT":    "value",
		}
		if len(got) != len(want) {
			t.Fatalf("parsed %d assignments, want %d", len(got), len(want))
		}
		for key, expected := range want {
			if actual, ok := got[key]; !ok || actual != expected {
				t.Errorf("%s = %q (present=%t), want %q", key, actual, ok, expected)
			}
		}
	})

	t.Run("CRLF input normalizes multiline values and resumes assignments", func(t *testing.T) {
		got, err := Unmarshal("CRLF=\"first line\r\nsecond line\" # comment\r\nAFTER=ok\r\n")
		if err != nil {
			t.Fatalf("Unmarshal CRLF input: %v", err)
		}
		if got["CRLF"] != "first line\nsecond line" {
			t.Errorf("CRLF = %q, want LF-normalized multiline value", got["CRLF"])
		}
		if got["AFTER"] != "ok" {
			t.Errorf("AFTER = %q, want %q", got["AFTER"], "ok")
		}
		for key, value := range got {
			for _, r := range value {
				if r == '\r' {
					t.Errorf("%s retained a carriage return", key)
					break
				}
			}
		}
	})
}
