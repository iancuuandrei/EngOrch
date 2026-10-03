package atomic

import (
	"encoding"
	"math"
	"strconv"
	"testing"
)

func requireTextContracts(t *testing.T, value any) (encoding.TextMarshaler, encoding.TextUnmarshaler) {
	t.Helper()
	marshaler, ok := value.(encoding.TextMarshaler)
	if !ok {
		t.Fatalf("%T must implement encoding.TextMarshaler", value)
	}
	unmarshaler, ok := value.(encoding.TextUnmarshaler)
	if !ok {
		t.Fatalf("%T must implement encoding.TextUnmarshaler", value)
	}
	return marshaler, unmarshaler
}

func exerciseNumericTextContract[T comparable](
	t *testing.T,
	values []T,
	format func(T) string,
	parse func(string) (T, error),
	newAtom func(T) (any, func() T),
	inputs []string,
	sentinel T,
) {
	t.Helper()

	for _, value := range values {
		atom, load := newAtom(value)
		marshaler, unmarshaler := requireTextContracts(t, atom)

		got, err := marshaler.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText(%v): %v", value, err)
		}
		want := format(value)
		if string(got) != want {
			t.Fatalf("MarshalText(%v) = %q, want canonical %q", value, got, want)
		}

		if err := unmarshaler.UnmarshalText(got); err != nil {
			t.Fatalf("UnmarshalText(%q): %v", got, err)
		}
		if gotValue := load(); gotValue != value {
			t.Fatalf("round trip = %v, want %v", gotValue, value)
		}
	}

	for _, input := range inputs {
		want, wantErr := parse(input)
		atom, load := newAtom(sentinel)
		_, unmarshaler := requireTextContracts(t, atom)
		err := unmarshaler.UnmarshalText([]byte(input))
		if (err == nil) != (wantErr == nil) {
			t.Fatalf("UnmarshalText(%q) error = %v, strconv parse error = %v", input, err, wantErr)
		}
		if wantErr == nil {
			if got := load(); got != want {
				t.Fatalf("UnmarshalText(%q) = %v, want %v", input, got, want)
			}
		} else if got := load(); got != sentinel {
			t.Fatalf("UnmarshalText(%q) changed value to %v after error %v", input, got, err)
		}
	}
}

func TestFabricV1Heldout(t *testing.T) {
	t.Run("Int64", func(t *testing.T) {
		exerciseNumericTextContract(
			t,
			[]int64{math.MinInt64, -1, 0, 1, math.MaxInt64},
			func(value int64) string { return strconv.FormatInt(value, 10) },
			func(input string) (int64, error) { return strconv.ParseInt(input, 10, 64) },
			func(value int64) (any, func() int64) {
				atom := NewInt64(value)
				return atom, atom.Load
			},
			[]string{
				"0", "+0", "-0", "0007", "+7", "-7",
				"", "word", "9223372036854775808", "-9223372036854775809", " 7", "7 ",
			},
			int64(91),
		)
	})

	t.Run("Uint64", func(t *testing.T) {
		exerciseNumericTextContract(
			t,
			[]uint64{0, 1, math.MaxUint64},
			func(value uint64) string { return strconv.FormatUint(value, 10) },
			func(input string) (uint64, error) { return strconv.ParseUint(input, 10, 64) },
			func(value uint64) (any, func() uint64) {
				atom := NewUint64(value)
				return atom, atom.Load
			},
			[]string{
				"0", "+0", "0007", "+7", "-0", "-7",
				"", "word", "18446744073709551616", " 7", "7 ",
			},
			uint64(91),
		)
	})
}
