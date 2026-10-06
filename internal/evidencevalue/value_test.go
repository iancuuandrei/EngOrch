package evidencevalue

import (
	"math"
	"reflect"
	"testing"
)

func fixture() Request {
	cost := 10.0
	return Request{Version: 1, Resources: []Resource{{Name: "wall_ms", Limit: 100, Price: 1}}, Actions: []Action{{ID: "read", Kind: "source_read", Reliability: Reliability{Ordinal: 2}, Costs: map[string]*float64{"wall_ms": &cost}, Outcomes: []Outcome{{Probability: .5, Utilities: []float64{1, 0}}, {Probability: .5, Utilities: []float64{0, 1}}}}}}
}

func TestFiniteInformationValueAndStopping(t *testing.T) {
	empty, err := Evaluate(Request{Version: 1})
	if err != nil || empty.Selected != "" {
		t.Fatal("empty frontier did not stop safely", empty, err)
	}
	q := fixture()
	r, err := Evaluate(q)
	if err != nil || r.Selected != "read" || math.Abs(*r.Estimates[0].JEV-.15) > 1e-12 {
		t.Fatal(r, err)
	}
	q.Actions[0].Outcomes = []Outcome{{Probability: 1, Utilities: []float64{1, 0}}}
	r, err = Evaluate(q)
	if err != nil || r.Selected != "" || r.Estimates[0].InformationValue != 0 {
		t.Fatal("purchased evidence that cannot change decision", r, err)
	}
	q = fixture()
	q.Resources[0].Price = 3
	r, err = Evaluate(q)
	if err != nil || r.Selected != "" {
		t.Fatal("purchased negative JEV", r, err)
	}
}

func TestFiniteUnknownCostsAndBudgetPrecedeValue(t *testing.T) {
	q := fixture()
	q.Actions[0].Costs["wall_ms"] = nil
	r, err := Evaluate(q)
	if err != nil || r.Selected != "" || r.Estimates[0].Cost != nil || r.Estimates[0].JEV != nil || r.Estimates[0].Status != "active_cost_unavailable" {
		t.Fatal(r, err)
	}
	q = fixture()
	q.Resources = append(q.Resources, Resource{Name: "money_minor_units", Limit: 1})
	r, err = Evaluate(q)
	if err != nil || r.Selected != "" || r.Estimates[0].Cost != nil {
		t.Fatal("unknown monetary cost became zero", r, err)
	}
	q = fixture()
	q.Resources[0].Used = 100
	r, err = Evaluate(q)
	if err != nil || r.Selected != "" || r.Estimates[0].Status != "resource_budget_exceeded" {
		t.Fatal("reopened exhausted budget", r, err)
	}
	q = fixture()
	q.Actions[0].Costs["money_minor_units"] = nil
	r, err = Evaluate(q)
	if err == nil {
		t.Fatal("accepted undeclared unknown cost", r)
	}
	money := 100.0
	q.Actions[0].Costs["money_minor_units"] = &money
	if _, err = Evaluate(q); err == nil {
		t.Fatal("silently omitted undeclared known cost")
	}
}

func TestFiniteStableOrderingAndPurity(t *testing.T) {
	q := fixture()
	other := q.Actions[0]
	other.ID = "aaa"
	q.Actions = append(q.Actions, other)
	before := fixture()
	r, err := Evaluate(q)
	if err != nil || r.Selected != "aaa" {
		t.Fatal(r, err)
	}
	q.Actions[0], q.Actions[1] = q.Actions[1], q.Actions[0]
	next, err := Evaluate(q)
	if err != nil || !reflect.DeepEqual(r, next) {
		t.Fatal("order changed result", r, next, err)
	}
	if !reflect.DeepEqual(before.Resources, q.Resources) {
		t.Fatal("mutated resource state")
	}
}

func TestFiniteRejectsMalformedModels(t *testing.T) {
	mutations := []func(*Request){
		func(q *Request) { q.Version = 2 }, func(q *Request) { q.Actions[0].Kind = "approve" },
		func(q *Request) { q.Actions[0].Outcomes[0].Probability = .6 },
		func(q *Request) { q.Actions[0].Outcomes[1].Utilities = []float64{1} },
		func(q *Request) { q.Actions[0].Outcomes[0].Utilities[0] = math.NaN() },
		func(q *Request) { q.Resources[0].Price = math.Inf(1) },
		func(q *Request) { q.Resources = append(q.Resources, q.Resources[0]) },
		func(q *Request) { q.Actions = append(q.Actions, q.Actions[0]) },
		func(q *Request) { q.Actions[0].Reliability.Successes = 1000001 },
		func(q *Request) { x := -1.0; q.Actions[0].Costs["wall_ms"] = &x },
		func(q *Request) { q.Resources[0].Limit = 1e-300 },
	}
	for i, mutate := range mutations {
		q := fixture()
		mutate(&q)
		if _, err := Evaluate(q); err == nil {
			t.Fatal("accepted invalid model", i)
		}
	}
	q := fixture()
	other := q.Actions[0]
	other.ID = "different"
	other.Outcomes = []Outcome{{Probability: 1, Utilities: []float64{1, 0}}}
	q.Actions = append(q.Actions, other)
	if _, err := Evaluate(q); err == nil {
		t.Fatal("compared different prior decision states")
	}
}

func TestAdaptivePricesAndReliability(t *testing.T) {
	r := Resource{Name: "wall_ms", Limit: 100}
	next, err := UpdatePrice(r, 200, 100)
	if err != nil || next.Price != 1 || next.GradientSquares != 1 || next.Limit != r.Limit || next.Used != r.Used {
		t.Fatal(next, err)
	}
	down, err := UpdatePrice(next, 0, 100)
	if err != nil || math.Abs(down.Price-(1-1/math.Sqrt(2))) > 1e-12 {
		t.Fatal(down, err)
	}
	zero, err := UpdatePrice(r, 100, 100)
	if err != nil || zero != r {
		t.Fatal("zero gradient changed state", zero, err)
	}
	if _, err := UpdatePrice(r, 1, 0); err == nil {
		t.Fatal("accepted zero allocation")
	}
	p, _ := Posterior(Reliability{Ordinal: 2, Successes: 1})
	if math.Abs(p-2.0/3) > 1e-12 {
		t.Fatal(p)
	}
	low, _ := Posterior(Reliability{Ordinal: 1})
	high, _ := Posterior(Reliability{Ordinal: 3})
	if math.Abs(low+high-1) > 1e-12 {
		t.Fatal("prior is not symmetric", low, high)
	}
}
