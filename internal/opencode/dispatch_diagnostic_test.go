package opencode

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDispatchDiagnosticPreservesCauseWithoutProviderText(t *testing.T) {
	for _, cause := range []error{errors.New("TOKEN=private"), context.DeadlineExceeded, context.Canceled} {
		err := dispatchFailure("transcript_decode", cause)
		d := DispatchDiagnosticFromError(errors.Join(errors.New("outer private provider text"), err))
		if d == nil || d.Stage != "transcript_decode" || !errors.Is(err, cause) || strings.Contains(err.Error(), "private") {
			t.Fatal("diagnostic lost stage/cause or leaked text", err)
		}
	}
	if DispatchDiagnosticFromError(errors.New("message_post deadline_exceeded TOKEN=private")) != nil {
		t.Fatal("arbitrary error text became a trusted diagnostic")
	}
}

func TestSynchronousDispatchReportsDecodeStageWithoutResend(t *testing.T) {
	intent, brokerPath, _ := synchronousToolFixture(t)
	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"TOKEN":"private","malformed":true}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "fixture", "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "dispatch.jsonl")
	_, err = client.SubmitSynchronousToolTurn(ctx, path, brokerPath, intent)
	d := DispatchDiagnosticFromError(err)
	if d == nil || d.Stage != "response_decode" || posts.Load() != 1 || strings.Contains(err.Error(), "private") {
		t.Fatal("failed response was unexplained or leaked", err, posts.Load())
	}
	_, err = client.SubmitSynchronousToolTurn(ctx, path, brokerPath, intent)
	if err == nil || posts.Load() != 1 {
		t.Fatal("diagnostic authorized duplicate POST", err)
	}
}
