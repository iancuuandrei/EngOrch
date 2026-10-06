// Package evidencevalue evaluates finite advisory information purchases.
// Estimates and recommendations never grant execution or acceptance authority.
package evidencevalue

import (
	"errors"
	"math"
	"sort"
)

// MaxActions bounds the finite acquisition frontier.
const MaxActions = 16

// MaxBytes bounds an encoded advisory request before decoding.
const MaxBytes = 64 << 10

// Binding identifies the controller snapshot; the caller validates provenance.
type Binding struct {
	RunID       string `json:"run_id"`
	SourceID    string `json:"source_id"`
	CandidateID string `json:"candidate_id"`
	JournalHead string `json:"journal_head"`
}

// Resource is one declared constraint. Price is derived dual state, not a
// measured monetary price. Used may exceed Limit; this does not reopen it.
type Resource struct {
	Name            string  `json:"name"`
	Limit           float64 `json:"limit"`
	Used            float64 `json:"used"`
	Price           float64 `json:"price"`
	GradientSquares float64 `json:"gradient_squares"`
}

// Outcome describes conditional utilities for the shared finite decision model.
type Outcome struct {
	Probability float64   `json:"probability"`
	Utilities   []float64 `json:"utilities"`
}

// Reliability supplies an ordinal prior and attributed binary observation counts.
type Reliability struct {
	Ordinal   int `json:"ordinal"`
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

// Action models one possible acquisition; nil costs remain unavailable.
type Action struct {
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	Reliability Reliability         `json:"reliability"`
	Outcomes    []Outcome           `json:"outcomes"`
	Costs       map[string]*float64 `json:"costs"`
}

// Request supplies bounded advisory estimates, never execution permissions.
type Request struct {
	Version   int        `json:"version"`
	Binding   Binding    `json:"binding"`
	Resources []Resource `json:"resources"`
	Actions   []Action   `json:"actions"`
}

// Estimate reports eligibility and value; unavailable costs leave JEV absent.
type Estimate struct {
	ID               string   `json:"id"`
	Kind             string   `json:"kind"`
	Status           string   `json:"status"`
	Reliability      float64  `json:"reliability"`
	InformationValue float64  `json:"information_value"`
	Cost             *float64 `json:"normalized_shadow_cost"`
	JEV              *float64 `json:"jev"`
	UnknownCosts     []string `json:"unknown_costs"`
}

// Report recommends an acquisition without granting controller admission.
type Report struct {
	Version    int        `json:"version"`
	Binding    Binding    `json:"binding"`
	Provenance string     `json:"provenance"`
	Authority  string     `json:"authority"`
	Selected   string     `json:"selected_action,omitempty"`
	StopReason string     `json:"stop_reason"`
	Estimates  []Estimate `json:"estimates"`
}

func valid(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) && x >= 0 && x <= 1e12 }
func resourceName(s string) bool {
	switch s {
	case "uncached_input_tokens", "cached_input_tokens", "ordinary_output_tokens", "reasoning_output_tokens", "wall_ms", "money_minor_units", "local_compute_ms", "human_interventions", "runtime_slots":
		return true
	}
	return false
}
func kind(s string) bool {
	switch s {
	case "source_read", "graph_query", "targeted_test", "coverage", "spawn_explorer", "stronger_model", "patch_candidate":
		return true
	}
	return false
}
func validID(s string) bool {
	if len(s) < 1 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// Posterior uses a fixed symmetric logistic ordinal prior with two equivalent
// observations. Counts must represent attributed binary outcomes, not prose.
// These supplied counts are estimates; the caller still owns provenance checks.
func Posterior(r Reliability) (float64, error) {
	if r.Ordinal < 0 || r.Ordinal > 4 || r.Successes < 0 || r.Failures < 0 || r.Successes > 1000000 || r.Failures > 1000000 || r.Successes+r.Failures > 1000000 {
		return 0, errors.New("reliability outside bounds")
	}
	p := 1 / (1 + math.Exp(-float64(r.Ordinal-2)))
	return (2*p + float64(r.Successes)) / (2 + float64(r.Successes) + float64(r.Failures)), nil
}

// InformationValue calculates finite EVSI from supplied conditional utilities:
// E[max decision utility after outcome] - max E[utility before outcome].
// It is a model estimate, not an empirical quality measurement.
func InformationValue(outcomes []Outcome) (float64, error) {
	if len(outcomes) < 1 || len(outcomes) > 16 || len(outcomes[0].Utilities) < 1 || len(outcomes[0].Utilities) > 16 {
		return 0, errors.New("decision model outside bounds")
	}
	baseline := make([]float64, len(outcomes[0].Utilities))
	total, after := 0.0, 0.0
	for _, o := range outcomes {
		if !valid(o.Probability) || o.Probability > 1 || len(o.Utilities) != len(baseline) {
			return 0, errors.New("invalid outcome model")
		}
		best := 0.0
		for i, u := range o.Utilities {
			if !valid(u) || u > 1 {
				return 0, errors.New("utility outside unit bounds")
			}
			baseline[i] += o.Probability * u
			best = math.Max(best, u)
		}
		total += o.Probability
		after += o.Probability * best
	}
	if math.Abs(total-1) > 1e-12 {
		return 0, errors.New("outcome probabilities must sum to one")
	}
	before := 0.0
	for _, u := range baseline {
		before = math.Max(before, u)
	}
	return math.Max(0, after-before), nil
}

// UpdatePrice performs a projected scalar AdaGrad dual update, normalized by
// the explicitly supplied per-observation allocation. It never changes budget.
func UpdatePrice(r Resource, observed, allocation float64) (Resource, error) {
	if !resourceName(r.Name) || !valid(r.Limit) || r.Limit < 1 || !valid(r.Used) || !valid(r.Price) || !valid(r.GradientSquares) || !valid(observed) || !valid(allocation) || allocation < 1 {
		return r, errors.New("invalid resource observation")
	}
	g := observed/allocation - 1
	squares := r.GradientSquares + g*g
	if !valid(squares) {
		return r, errors.New("resource gradient outside bounds")
	}
	if squares > 0 {
		r.Price = math.Max(0, r.Price+g/math.Sqrt(squares))
	}
	if !valid(r.Price) {
		return r, errors.New("resource price outside bounds")
	}
	r.GradientSquares = squares
	return r, nil
}

// Evaluate validates the finite model and selects the highest positive known
// JEV, with lexical tie ordering. Invalid or undeclared costs return an error.
// It does not execute actions or mutate controller state.
func Evaluate(q Request) (Report, error) {
	r := Report{Version: 1, Binding: q.Binding, Provenance: "caller_supplied_unvalidated_decision_model", Authority: "advisory_only_requires_controller_admission", StopReason: "no_positive_known_jev", Estimates: []Estimate{}}
	if q.Version != 1 || len(q.Actions) > MaxActions || len(q.Resources) > 9 || len(q.Actions) > 0 && len(q.Resources) < 1 {
		return r, errors.New("evidence request outside bounds")
	}
	resources := map[string]Resource{}
	names := []string{}
	for _, v := range q.Resources {
		if !resourceName(v.Name) || !valid(v.Limit) || v.Limit < 1 || !valid(v.Used) || !valid(v.Price) || !valid(v.GradientSquares) {
			return r, errors.New("invalid resource constraint")
		}
		if _, ok := resources[v.Name]; ok {
			return r, errors.New("duplicate resource constraint")
		}
		resources[v.Name] = v
		names = append(names, v.Name)
	}
	sort.Strings(names)
	ids := map[string]bool{}
	best := 0.0
	var currentUtilities []float64
	actions := append([]Action{}, q.Actions...)
	sort.Slice(actions, func(i, j int) bool { return actions[i].ID < actions[j].ID })
	for _, a := range actions {
		if !validID(a.ID) || !kind(a.Kind) || ids[a.ID] || len(a.Costs) > 9 {
			return r, errors.New("invalid or duplicate evidence action")
		}
		ids[a.ID] = true
		reliability, err := Posterior(a.Reliability)
		if err != nil {
			return r, err
		}
		information, err := InformationValue(a.Outcomes)
		if err != nil {
			return r, err
		}
		marginal := make([]float64, len(a.Outcomes[0].Utilities))
		for _, o := range a.Outcomes {
			for i, u := range o.Utilities {
				marginal[i] += o.Probability * u
			}
		}
		if currentUtilities == nil {
			currentUtilities = marginal
		} else {
			if len(currentUtilities) != len(marginal) {
				return r, errors.New("actions must share the current decision model")
			}
			for i, u := range marginal {
				if math.Abs(u-currentUtilities[i]) > 1e-12 {
					return r, errors.New("actions must share current expected utilities")
				}
			}
		}
		e := Estimate{ID: a.ID, Kind: a.Kind, Status: "resource_eligible", Reliability: reliability, InformationValue: information, UnknownCosts: []string{}}
		for name, value := range a.Costs {
			if !resourceName(name) || value != nil && !valid(*value) {
				return r, errors.New("invalid evidence cost")
			}
			if _, declared := resources[name]; !declared {
				return r, errors.New("evidence cost lacks a declared resource constraint")
			}
			if value == nil {
				e.UnknownCosts = append(e.UnknownCosts, name)
			}
		}
		cost := 0.0
		unknown := map[string]bool{}
		for _, name := range e.UnknownCosts {
			unknown[name] = true
		}
		for _, name := range names {
			v := resources[name]
			value, ok := a.Costs[name]
			if !ok || value == nil {
				e.Status = "active_cost_unavailable"
				unknown[name] = true
				continue
			}
			if *value > v.Limit-v.Used && e.Status != "active_cost_unavailable" {
				e.Status = "resource_budget_exceeded"
			}
			cost += v.Price * (*value / v.Limit)
		}
		e.UnknownCosts = []string{}
		for name := range unknown {
			e.UnknownCosts = append(e.UnknownCosts, name)
		}
		sort.Strings(e.UnknownCosts)
		if e.Status == "resource_eligible" {
			score := reliability*information - cost
			e.Cost = &cost
			e.JEV = &score
			if score > best {
				best = score
				r.Selected = a.ID
				r.StopReason = "positive_estimated_jev_requires_admission"
			}
		}
		r.Estimates = append(r.Estimates, e)
	}
	return r, nil
}
