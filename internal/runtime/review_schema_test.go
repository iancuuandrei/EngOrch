package runtime

import (
	"encoding/json"
	"regexp"
	"testing"
)

func TestReviewSchemaFindingPath(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(ReviewOutputSchema(), &schema); err != nil {
		t.Fatal(err)
	}
	p := schema["properties"].(map[string]any)["findings"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)["path"].(map[string]any)["pattern"].(string)
	re, err := regexp.Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"", "bytes.go", "internal/parse_test.go", ".gitignore", "path with spaces/code.go"} {
		if !re.MatchString(path) {
			t.Fatal("valid finding path rejected", path)
		}
	}
	for _, path := range []string{"go build -v ./...", "../file.go", "./file.go", "/absolute", `C:\file.go`, "file. ", "dir//file.go"} {
		if re.MatchString(path) {
			t.Fatal("ambiguous finding path admitted", path)
		}
	}
}
