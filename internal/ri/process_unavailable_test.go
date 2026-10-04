package ri

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestClientProcessLaunchFailureIsTypedOptionalCapability(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(bytes)
	client := Client{Executable: executable, ExecutableHash: hex.EncodeToString(digest[:])}

	_, err = client.Call(context.Background(), map[string]any{"operation": "test"})
	if !errors.Is(err, ErrProcessUnavailable) || !IsProcessUnavailableOnly(err) {
		t.Fatalf("invalid child-process response was not typed as optional unavailability: %T %v", err, err)
	}
	if err.Error() != "RI process unavailable" {
		t.Fatalf("process details leaked through the stable capability error: %q", err)
	}
}

func TestProcessUnavailableClassificationRequiresEveryErrorLeaf(t *testing.T) {
	optional := unavailableProcess()
	if !IsProcessUnavailableOnly(fmt.Errorf("wrapped: %w", optional)) {
		t.Fatal("wrapped process unavailability was not recognized")
	}
	if IsProcessUnavailableOnly(errors.Join(optional, errors.New("lease close failed"))) {
		t.Fatal("joined strict failure was classified as optional unavailability")
	}
	if IsProcessUnavailableOnly(errors.New("protocol mismatch")) {
		t.Fatal("untyped protocol failure was classified as optional unavailability")
	}
}
