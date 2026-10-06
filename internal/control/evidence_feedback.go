package control

import (
	"errors"
	"strconv"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/codexusage"
	"harness.local/engorch/internal/evidencevalue"
)

// EvidenceFeedbackReport binds advisory price updates to receipt-verified usage.
// No run budgets, effects, policies, acceptance or journals are modified.
type EvidenceFeedbackReport struct {
	Provenance             string                              `json:"provenance"`
	UsageHash              string                              `json:"usage_hash"`
	Scope                  string                              `json:"scope"`
	Feedback               evidencevalue.FeedbackResult        `json:"feedback"`
	Observations           []evidencevalue.ResourceObservation `json:"observations"`
	UnavailableInvocations []string                            `json:"unavailable_invocations"`
}

func evidenceTokenObservations(usage RunUsage) ([]evidencevalue.ResourceObservation, []string, error) {
	observations := []evidencevalue.ResourceObservation{}
	unavailable := []string{}
	for _, entry := range usage.Invocations {
		observation := evidencevalue.ResourceObservation{InvocationID: entry.InvocationID, Costs: map[string]*evidencevalue.Numeric{}}
		observations = append(observations, observation)
		if !entry.JournalPresent || !entry.ReceiptMatched || entry.Usage == nil || !entry.Usage.Completed || entry.Usage.PendingCalls != 0 {
			unavailable = append(unavailable, entry.InvocationID)
			continue
		}
		accounting := entry.Usage.ProviderUsage.Accounting
		if accounting == nil || accounting.Source != codexusage.SourceCodexAppServer || accounting.Coverage != codexusage.CoverageObserved || len(accounting.Anomalies) != 0 {
			unavailable = append(unavailable, entry.InvocationID)
			continue
		}
		tokens := accounting.Delta
		if tokens.InputTokens < 0 || tokens.CachedInputTokens < 0 || tokens.CachedInputTokens > tokens.InputTokens || tokens.OutputTokens < 0 || tokens.ReasoningOutputTokens < 0 || tokens.ReasoningOutputTokens > tokens.OutputTokens {
			return nil, nil, errors.New("observed token resource subsets invalid")
		}
		for name, count := range map[string]int64{
			"uncached_input_tokens":   tokens.InputTokens - tokens.CachedInputTokens,
			"cached_input_tokens":     tokens.CachedInputTokens,
			"ordinary_output_tokens":  tokens.OutputTokens - tokens.ReasoningOutputTokens,
			"reasoning_output_tokens": tokens.ReasoningOutputTokens,
		} {
			value := evidencevalue.Numeric(strconv.FormatInt(count, 10))
			observation.Costs[name] = &value
		}
	}
	return observations, unavailable, nil
}

// RepriceEvidenceResources derives advisory token feedback from exact runtime
// receipts. The supplied model must bind the same inspected controller prefix.
// Missing typed accounting never becomes zero; non-token axes remain unknown.
func RepriceEvidenceResources(path string, request evidencevalue.FeedbackRequest) (EvidenceFeedbackReport, error) {
	var report EvidenceFeedbackReport
	raw, err := canonical.Bytes(request)
	if err != nil || len(raw) > evidencevalue.MaxBytes {
		return report, errors.New("resource feedback request exceeds encoding bounds")
	}
	s, head, err := InspectWithHead(path)
	if err != nil {
		return report, err
	}
	binding, err := EvidenceBinding(s, head)
	if err != nil {
		return report, err
	}
	if request.Model.Binding != binding {
		return report, errors.New("resource feedback binding is stale or substituted")
	}
	usage, err := MeasureRunUsage(path)
	if err != nil {
		return report, err
	}
	if usage.RunID != s.RunID || usage.ControllerHead != head {
		return report, errors.New("controller changed during resource feedback inspection")
	}
	report.Observations, report.UnavailableInvocations, err = evidenceTokenObservations(usage)
	if err != nil {
		return report, err
	}
	report.Feedback, err = evidencevalue.ApplyFeedback(request, report.Observations)
	if err != nil {
		return EvidenceFeedbackReport{}, err
	}
	report.UsageHash, err = canonical.Hash("harness.evidence-resource-usage.v1", usage)
	report.Scope = "receipt_matched_completed_codex_typed_tokens"
	report.Provenance = "observed_token_feedback_with_caller_supplied_advisory_baseline_and_allocations"
	return report, err
}
