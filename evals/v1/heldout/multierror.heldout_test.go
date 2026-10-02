package multierror

import (
	"errors"
	"testing"
)

func TestFabricV1Heldout(t *testing.T) {
	first, second := errors.New("first"), errors.New("second")
	e := Append(nil, first, second)
	snapshot := e.ErrorsSnapshot()
	if len(snapshot) != 2 || snapshot[0] != first || snapshot[1] != second {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	snapshot[0] = second
	snapshot = append(snapshot, errors.New("extra"))
	if len(e.Errors) != 2 || e.Errors[0] != first || e.Errors[1] != second {
		t.Fatalf("snapshot aliases source: %#v", e.Errors)
	}
	var nilError *Error
	if got := nilError.ErrorsSnapshot(); got != nil {
		t.Fatalf("nil receiver snapshot = %#v, want nil", got)
	}
}
