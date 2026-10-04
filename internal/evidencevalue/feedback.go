package evidencevalue

import (
	"errors"
	"sort"
	"strings"
)

// MaxObservations bounds one optional resource-feedback projection.
const MaxObservations = 128

// ResourceObservation contains costs verified by the caller for one immutable
// invocation. Null is unavailable; the pure evaluator cannot prove provenance.
type ResourceObservation struct {
	InvocationID string              `json:"invocation_id"`
	Costs        map[string]*Numeric `json:"costs"`
}

// FeedbackRequest retains caller-declared resource state, allocations and
// per-axis consumed invocation identities. It grants no budget authority.
type FeedbackRequest struct {
	Model       WireRequest         `json:"model"`
	Allocations map[string]Numeric  `json:"allocations"`
	Consumed    map[string][]string `json:"consumed"`
}

// FeedbackResult provides reusable advisory model state and exact new updates.
// Skipped axes have no new known observations, not measured zero consumption.
type FeedbackResult struct {
	Request FeedbackRequest     `json:"request"`
	Applied map[string][]string `json:"applied"`
	Skipped []string            `json:"skipped"`
}

// ApplyFeedback updates prices and advisory used quantities once per declared
// axis/invocation. Input observations remain in controller order. Unknown costs
// stay unconsumed; invalid cursors, numbers or excessive totals return an error.
// Caller data is copied and never mutated, including when validation fails.
func ApplyFeedback(request FeedbackRequest, observations []ResourceObservation) (FeedbackResult, error) {
	var result FeedbackResult
	model, err := request.Model.Decode()
	if err != nil {
		return result, err
	}
	if _, err := Evaluate(model); err != nil {
		return result, err
	}
	if len(observations) > MaxObservations || len(request.Allocations) != len(model.Resources) || len(request.Consumed) > len(model.Resources) {
		return result, errors.New("resource feedback outside bounds")
	}
	observedIDs := map[string]bool{}
	observedCosts := map[string]map[string]*Numeric{}
	for _, observation := range observations {
		if !feedbackID(observation.InvocationID) || observedIDs[observation.InvocationID] || len(observation.Costs) > 9 {
			return result, errors.New("invalid or duplicate resource observation")
		}
		observedIDs[observation.InvocationID] = true
		observedCosts[observation.InvocationID] = observation.Costs
		for axis, cost := range observation.Costs {
			if !resourceName(axis) {
				return result, errors.New("unsupported observed resource")
			}
			if cost != nil {
				if _, err := cost.value(); err != nil {
					return result, err
				}
			}
		}
	}
	result = FeedbackResult{Request: FeedbackRequest{Allocations: map[string]Numeric{}, Consumed: map[string][]string{}}, Applied: map[string][]string{}, Skipped: []string{}}
	axes := map[string]bool{}
	for index, resource := range model.Resources {
		axes[resource.Name] = true
		allocationText, ok := request.Allocations[resource.Name]
		if !ok {
			return FeedbackResult{}, errors.New("resource feedback allocation missing")
		}
		allocation, err := allocationText.value()
		if err != nil || allocation < 1 {
			return FeedbackResult{}, errors.New("resource feedback allocation invalid")
		}
		prior := request.Consumed[resource.Name]
		if len(prior) > MaxObservations {
			return FeedbackResult{}, errors.New("resource feedback cursor exceeds bound")
		}
		consumed := map[string]bool{}
		for _, id := range prior {
			if !observedIDs[id] || consumed[id] || observedCosts[id][resource.Name] == nil {
				return FeedbackResult{}, errors.New("resource feedback cursor is unknown or duplicated")
			}
			consumed[id] = true
		}
		applied := []string{}
		for _, observation := range observations {
			cost := observation.Costs[resource.Name]
			if consumed[observation.InvocationID] || cost == nil {
				continue
			}
			observed, err := cost.value()
			if err != nil {
				return FeedbackResult{}, err
			}
			resource, err = UpdatePrice(resource, observed, allocation)
			if err != nil {
				return FeedbackResult{}, err
			}
			resource.Used += observed
			if !valid(resource.Used) {
				return FeedbackResult{}, errors.New("observed resource total exceeds numerical bound")
			}
			consumed[observation.InvocationID] = true
			applied = append(applied, observation.InvocationID)
		}
		model.Resources[index] = resource
		result.Request.Allocations[resource.Name] = allocationText
		cursor := []string{}
		for id := range consumed {
			cursor = append(cursor, id)
		}
		sort.Strings(cursor)
		result.Request.Consumed[resource.Name] = cursor
		result.Applied[resource.Name] = applied
		if len(applied) == 0 {
			result.Skipped = append(result.Skipped, resource.Name)
		}
	}
	for axis := range request.Allocations {
		if !axes[axis] {
			return FeedbackResult{}, errors.New("resource feedback has undeclared allocation")
		}
	}
	for axis := range request.Consumed {
		if !axes[axis] {
			return FeedbackResult{}, errors.New("resource feedback has undeclared cursor")
		}
	}
	sort.Strings(result.Skipped)
	result.Request.Model = EncodeRequest(model)
	return result, nil
}

func feedbackID(id string) bool {
	if len(id) != 64 {
		return false
	}
	return strings.Trim(id, "0123456789abcdef") == ""
}
