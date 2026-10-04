package ri

import (
	"context"
	"errors"
	"strings"
)

const (
	semanticSnapshotReferences      = "references"
	semanticSnapshotImplementations = "implementations"
)

// SemanticSnapshotQuery selects one bounded, producer-scoped semantic view of
// an immutable RI snapshot. References are direct SCIP occurrences;
// implementations are only explicitly declared IMPLEMENTS graph edges.
type SemanticSnapshotQuery struct {
	Vocabulary string `json:"vocabulary"`
	Symbol     string `json:"symbol,omitempty"`
	Node       string `json:"node,omitempty"`
	Producer   string `json:"producer"`
	Direction  string `json:"direction,omitempty"`
	Limit      int    `json:"limit"`
}

// SemanticSnapshotResult preserves one delegated semantic page with its exact
// snapshot and producer binding. Coverage is PARTIAL for direct occurrences;
// implementation coverage is the producer's exact directional declaration.
// Empty results never imply semantic absence unless AbsenceProven is true.
type SemanticSnapshotResult struct {
	Version       int          `json:"version"`
	Vocabulary    string       `json:"vocabulary"`
	SnapshotID    string       `json:"snapshot_id"`
	Source        Source       `json:"source"`
	Producer      string       `json:"producer"`
	QueryID       string       `json:"query_id"`
	Coverage      string       `json:"coverage"`
	AbsenceProven bool         `json:"absence_proven"`
	Occurrences   []Occurrence `json:"occurrences"`
	Edges         []Edge       `json:"edges"`
	NextAfter     *string      `json:"next_after"`
}

// QuerySemanticSnapshot reads either direct semantic references or explicit
// implementation edges from one pinned snapshot. It does not resolve calls,
// infer implementations, merge producers, or execute an indexer.
func (c Client) QuerySemanticSnapshot(ctx context.Context, ref SnapshotRef, query SemanticSnapshotQuery, after *string) (SemanticSnapshotResult, error) {
	if err := ValidateSemanticSnapshotQuery(query); err != nil {
		return SemanticSnapshotResult{}, err
	}
	var result SemanticSnapshotResult
	switch query.Vocabulary {
	case semanticSnapshotReferences:
		page, err := c.Occurrences(ctx, ref, OccurrenceQuery{Symbol: query.Symbol, Producer: query.Producer, Definitions: false, Limit: query.Limit}, after)
		if err != nil {
			return SemanticSnapshotResult{}, err
		}
		result = SemanticSnapshotResult{Version: 1, Vocabulary: query.Vocabulary, SnapshotID: ref.ID, Source: ref.Source, Producer: query.Producer, QueryID: page.QueryID, Coverage: "PARTIAL", AbsenceProven: page.AbsenceProven, Occurrences: append([]Occurrence{}, page.Occurrences...), Edges: []Edge{}, NextAfter: copySemanticCursor(page.NextAfter)}
	case semanticSnapshotImplementations:
		page, err := c.Neighbors(ctx, ref, NeighborQuery{Node: query.Node, Relation: "IMPLEMENTS", Direction: query.Direction, Producer: query.Producer, Limit: query.Limit}, after)
		if err != nil {
			return SemanticSnapshotResult{}, err
		}
		result = SemanticSnapshotResult{Version: 1, Vocabulary: query.Vocabulary, SnapshotID: ref.ID, Source: ref.Source, Producer: query.Producer, QueryID: page.QueryID, Coverage: page.Evidence.Completeness, AbsenceProven: page.Evidence.AbsenceProven, Occurrences: []Occurrence{}, Edges: append([]Edge{}, page.Evidence.Edges...), NextAfter: copySemanticCursor(page.NextAfter)}
	}
	if err := validateSemanticSnapshotResult(result, ref, query); err != nil {
		return SemanticSnapshotResult{}, err
	}
	return result, nil
}

// ValidateSemanticSnapshotQuery rejects unsupported semantic vocabulary and
// selector combinations before a pinned RI process is invoked.
func ValidateSemanticSnapshotQuery(query SemanticSnapshotQuery) error {
	if query.Limit < 1 || query.Limit > 128 || query.Producer == "" || len(query.Producer) > 256 || strings.TrimSpace(query.Producer) != query.Producer {
		return errors.New("invalid semantic snapshot producer or limit")
	}
	switch query.Vocabulary {
	case semanticSnapshotReferences:
		if query.Symbol == "" || len(query.Symbol) > 256 || query.Node != "" || query.Direction != "" {
			return errors.New("references query requires one exact symbol")
		}
	case semanticSnapshotImplementations:
		if query.Node == "" || len(query.Node) > 256 || query.Symbol != "" || (query.Direction != "OUTGOING" && query.Direction != "INCOMING") {
			return errors.New("implementations query requires one exact node and direction")
		}
	default:
		return errors.New("unsupported semantic snapshot vocabulary")
	}
	return nil
}

func validateSemanticSnapshotResult(result SemanticSnapshotResult, ref SnapshotRef, query SemanticSnapshotQuery) error {
	if result.Version != 1 || result.Vocabulary != query.Vocabulary || result.SnapshotID != ref.ID || result.Source != ref.Source || result.Producer != query.Producer || result.QueryID == "" || result.Occurrences == nil || result.Edges == nil {
		return errors.New("semantic snapshot result identity or shape mismatch")
	}
	switch query.Vocabulary {
	case semanticSnapshotReferences:
		if result.Coverage != "PARTIAL" || result.AbsenceProven || len(result.Edges) != 0 {
			return errors.New("semantic reference coverage or edge shape mismatch")
		}
	case semanticSnapshotImplementations:
		if len(result.Occurrences) != 0 {
			return errors.New("semantic implementation occurrence shape mismatch")
		}
		switch result.Coverage {
		case "UNKNOWN", "PARTIAL", "COMPLETE":
		default:
			return errors.New("semantic implementation coverage is invalid")
		}
		if result.AbsenceProven && (len(result.Edges) != 0 || result.Coverage != "COMPLETE") {
			return errors.New("semantic implementation absence claim is invalid")
		}
		for _, edge := range result.Edges {
			if edge.Relation != "IMPLEMENTS" || edge.Producer != query.Producer || (query.Direction == "OUTGOING" && edge.From != query.Node) || (query.Direction == "INCOMING" && edge.To != query.Node) {
				return errors.New("semantic implementation edge scope mismatch")
			}
		}
	}
	return nil
}

func copySemanticCursor(cursor *string) *string {
	if cursor == nil {
		return nil
	}
	copy := *cursor
	return &copy
}
