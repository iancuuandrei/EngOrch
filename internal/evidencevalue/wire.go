package evidencevalue

import (
	"errors"
	"strconv"
)

// Numeric preserves canonical-v1 compatibility without allowing JSON floats.
// These numbers are bounded advisory estimates, not authoritative quantities.
type Numeric string

func number(v float64) Numeric { return Numeric(strconv.FormatFloat(v, 'g', -1, 64)) }
func (n Numeric) value() (float64, error) {
	if len(n) < 1 || len(n) > 64 {
		return 0, errors.New("numeric estimate text outside bounds")
	}
	v, err := strconv.ParseFloat(string(n), 64)
	if err != nil || !valid(v) {
		return 0, errors.New("numeric estimate outside bounds")
	}
	return v, nil
}

// WireResource encodes a resource using canonical-compatible numeric strings.
type WireResource struct {
	Name            string  `json:"name"`
	Limit           Numeric `json:"limit"`
	Used            Numeric `json:"used"`
	Price           Numeric `json:"price"`
	GradientSquares Numeric `json:"gradient_squares"`
}

// WireOutcome encodes probabilities and utilities without JSON floats.
type WireOutcome struct {
	Probability Numeric   `json:"probability"`
	Utilities   []Numeric `json:"utilities"`
}

// WireAction preserves unavailable costs as explicit null values.
type WireAction struct {
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	Reliability Reliability         `json:"reliability"`
	Outcomes    []WireOutcome       `json:"outcomes"`
	Costs       map[string]*Numeric `json:"costs"`
}

// WireRequest is the canonical-compatible request representation.
type WireRequest struct {
	Version   int            `json:"version"`
	Binding   Binding        `json:"binding"`
	Resources []WireResource `json:"resources"`
	Actions   []WireAction   `json:"actions"`
}

// EncodeRequest converts estimates to wire values; it does not validate them.
func EncodeRequest(q Request) WireRequest {
	w := WireRequest{Version: q.Version, Binding: q.Binding, Resources: []WireResource{}, Actions: []WireAction{}}
	for _, r := range q.Resources {
		w.Resources = append(w.Resources, WireResource{r.Name, number(r.Limit), number(r.Used), number(r.Price), number(r.GradientSquares)})
	}
	for _, a := range q.Actions {
		v := WireAction{ID: a.ID, Kind: a.Kind, Reliability: a.Reliability, Outcomes: []WireOutcome{}, Costs: map[string]*Numeric{}}
		for _, o := range a.Outcomes {
			u := WireOutcome{Probability: number(o.Probability), Utilities: []Numeric{}}
			for _, n := range o.Utilities {
				u.Utilities = append(u.Utilities, number(n))
			}
			v.Outcomes = append(v.Outcomes, u)
		}
		for k, n := range a.Costs {
			if n == nil {
				v.Costs[k] = nil
			} else {
				x := number(*n)
				v.Costs[k] = &x
			}
		}
		w.Actions = append(w.Actions, v)
	}
	return w
}

// Decode bounds collections and numeric values before conversion. Evaluate
// performs the remaining model and resource validation.
func (w WireRequest) Decode() (Request, error) {
	q := Request{Version: w.Version, Binding: w.Binding, Resources: []Resource{}, Actions: []Action{}}
	if len(w.Resources) > 9 || len(w.Actions) > MaxActions {
		return q, errors.New("wire request outside bounds")
	}
	for _, r := range w.Resources {
		ns := []Numeric{r.Limit, r.Used, r.Price, r.GradientSquares}
		vs := make([]float64, 4)
		for i, n := range ns {
			v, err := n.value()
			if err != nil {
				return q, err
			}
			vs[i] = v
		}
		q.Resources = append(q.Resources, Resource{r.Name, vs[0], vs[1], vs[2], vs[3]})
	}
	for _, a := range w.Actions {
		v := Action{ID: a.ID, Kind: a.Kind, Reliability: a.Reliability, Outcomes: []Outcome{}, Costs: map[string]*float64{}}
		if len(a.Outcomes) > 16 || len(a.Costs) > 9 {
			return q, errors.New("wire action outside bounds")
		}
		for _, o := range a.Outcomes {
			p, err := o.Probability.value()
			if err != nil {
				return q, err
			}
			if len(o.Utilities) > 16 {
				return q, errors.New("wire utilities outside bounds")
			}
			out := Outcome{Probability: p, Utilities: []float64{}}
			for _, n := range o.Utilities {
				u, err := n.value()
				if err != nil {
					return q, err
				}
				out.Utilities = append(out.Utilities, u)
			}
			v.Outcomes = append(v.Outcomes, out)
		}
		for k, n := range a.Costs {
			if n == nil {
				v.Costs[k] = nil
			} else {
				x, err := n.value()
				if err != nil {
					return q, err
				}
				v.Costs[k] = &x
			}
		}
		q.Actions = append(q.Actions, v)
	}
	return q, nil
}

// WireEstimate encodes advisory results, including signed JEV values.
type WireEstimate struct {
	ID               string   `json:"id"`
	Kind             string   `json:"kind"`
	Status           string   `json:"status"`
	Reliability      Numeric  `json:"reliability"`
	InformationValue Numeric  `json:"information_value"`
	Cost             *Numeric `json:"normalized_shadow_cost"`
	JEV              *Numeric `json:"jev"`
	UnknownCosts     []string `json:"unknown_costs"`
}

// WireReport binds canonical-compatible estimates to their request and snapshot.
type WireReport struct {
	RequestHash string         `json:"request_hash,omitempty"`
	Version     int            `json:"version"`
	Binding     Binding        `json:"binding"`
	Provenance  string         `json:"provenance"`
	Authority   string         `json:"authority"`
	Selected    string         `json:"selected_action,omitempty"`
	StopReason  string         `json:"stop_reason"`
	Estimates   []WireEstimate `json:"estimates"`
}

// Wire converts a computed report without changing its advisory status.
func (r Report) Wire() WireReport {
	w := WireReport{Version: r.Version, Binding: r.Binding, Provenance: r.Provenance, Authority: r.Authority, Selected: r.Selected, StopReason: r.StopReason, Estimates: []WireEstimate{}}
	for _, e := range r.Estimates {
		v := WireEstimate{ID: e.ID, Kind: e.Kind, Status: e.Status, Reliability: number(e.Reliability), InformationValue: number(e.InformationValue), UnknownCosts: e.UnknownCosts}
		if e.Cost != nil {
			n := number(*e.Cost)
			v.Cost = &n
		}
		if e.JEV != nil {
			n := number(*e.JEV)
			v.JEV = &n
		}
		w.Estimates = append(w.Estimates, v)
	}
	return w
}
