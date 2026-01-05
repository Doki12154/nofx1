package kernel

import (
	"nofx/store"
)

// DecisionStoreAdapter adapts store.DecisionStore to DecisionStoreInterface
type DecisionStoreAdapter struct {
	store *store.DecisionStore
}

// NewDecisionStoreAdapter creates a new adapter
func NewDecisionStoreAdapter(s *store.DecisionStore) *DecisionStoreAdapter {
	return &DecisionStoreAdapter{store: s}
}

// GetLatestRecordsBySymbol implements DecisionStoreInterface
func (a *DecisionStoreAdapter) GetLatestRecordsBySymbol(traderID, symbol string, n int) ([]*DecisionRecord, error) {
	storeRecords, err := a.store.GetLatestRecordsBySymbol(traderID, symbol, n)
	if err != nil {
		return nil, err
	}

	// Convert store.DecisionRecord to kernel.DecisionRecord
	records := make([]*DecisionRecord, 0, len(storeRecords))
	for _, sr := range storeRecords {
		kr := &DecisionRecord{
			Timestamp:   sr.Timestamp,
			InputPrompt: sr.InputPrompt,
			Decisions:   make([]DecisionAction, 0, len(sr.Decisions)),
		}

		// Convert decisions
		for _, sd := range sr.Decisions {
			kr.Decisions = append(kr.Decisions, DecisionAction{
				Symbol:    sd.Symbol,
				Action:    sd.Action,
				Reasoning: sd.Reasoning,
			})
		}

		records = append(records, kr)
	}

	return records, nil
}

// Ensure adapter implements the interface
var _ DecisionStoreInterface = (*DecisionStoreAdapter)(nil)
