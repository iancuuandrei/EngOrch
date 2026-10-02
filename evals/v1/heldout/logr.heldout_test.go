package logr

import "testing"

type fabricV1Sink struct{ withValues int }

func (*fabricV1Sink) Init(RuntimeInfo)                 {}
func (*fabricV1Sink) Enabled(int) bool                 { return true }
func (*fabricV1Sink) Info(int, string, ...any)         {}
func (*fabricV1Sink) Error(error, string, ...any)      {}
func (s *fabricV1Sink) WithValues(...any) LogSink      { s.withValues++; return s }
func (s *fabricV1Sink) WithName(string) LogSink        { return s }

func TestFabricV1Heldout(t *testing.T) {
	sink := &fabricV1Sink{}
	logger := New(sink)
	got := logger.WithValues()
	if sink.withValues != 0 {
		t.Fatalf("empty WithValues delegated %d times", sink.withValues)
	}
	if got.GetV() != logger.GetV() {
		t.Fatalf("empty WithValues changed verbosity")
	}
	_ = logger.WithValues("k", "v")
	if sink.withValues != 1 {
		t.Fatalf("non-empty WithValues delegation count = %d, want 1", sink.withValues)
	}
}
