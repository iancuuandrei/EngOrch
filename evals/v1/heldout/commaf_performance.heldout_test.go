package humanize

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

var fabricV1CommafSink string

func referenceFabricCommaf(v float64) string {
	if math.IsNaN(v) {
		return "NaN"
	}
	if math.IsInf(v, 1) {
		return "+Inf"
	}
	if math.IsInf(v, -1) {
		return "-Inf"
	}
	negative := math.Signbit(v)
	if negative {
		v = math.Abs(v)
	}
	parts := strings.SplitN(strconv.FormatFloat(v, 'f', -1, 64), ".", 2)
	digits := parts[0]
	var b strings.Builder
	if negative {
		b.WriteByte('-')
	}
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(digits[i])
	}
	if len(parts) == 2 {
		b.WriteByte('.')
		b.WriteString(parts[1])
	}
	return b.String()
}

func checkFabricV1HeldoutCommafOutputEquivalence(t *testing.T) {
	values := []float64{
		0, 10.11, 1000, 1234567890.83584, -123456789.875,
		math.MaxFloat64, math.SmallestNonzeroFloat64,
		math.Copysign(0, -1), math.Inf(1), math.Inf(-1), math.NaN(),
	}
	r := rand.New(rand.NewSource(20261003))
	for i := 0; i < 1000; i++ {
		values = append(values, math.Float64frombits(r.Uint64()))
	}
	for i, v := range values {
		got, want := Commaf(v), referenceFabricCommaf(v)
		if got != want {
			t.Errorf("case %d (%g): got %q, want %q", i, v, got, want)
		}
	}
}

func checkFabricV1HeldoutCommafAllocationBudget(t *testing.T) {
	inputs := []struct {
		name  string
		value float64
	}{
		{"ordinary", 1234567890.83584},
		{"fraction", 1000.125},
		{"negative", -123456789.875},
		{"largest-finite", math.MaxFloat64},
		{"smallest-subnormal", math.SmallestNonzeroFloat64},
		{"negative-zero", math.Copysign(0, -1)},
	}
	for _, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			got := testing.AllocsPerRun(1000, func() { fabricV1CommafSink = Commaf(input.value) })
			t.Logf("allocs/op=%.0f", got)
			if got > 2 {
				t.Fatalf("Commaf allocs/op = %.0f, want <= 2", got)
			}
		})
	}
}

func TestFabricV1Heldout(t *testing.T) {
	t.Run("CommafOutputEquivalence", checkFabricV1HeldoutCommafOutputEquivalence)
	t.Run("CommafAllocationBudget", checkFabricV1HeldoutCommafAllocationBudget)
}
