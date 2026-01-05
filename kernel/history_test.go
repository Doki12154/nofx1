package kernel

import (
	"testing"
	"time"
)

// MockDecisionStore is a mock implementation for testing
type MockDecisionStore struct {
	records map[string][]*DecisionRecord
}

func NewMockDecisionStore() *MockDecisionStore {
	return &MockDecisionStore{
		records: make(map[string][]*DecisionRecord),
	}
}

func (m *MockDecisionStore) GetLatestRecordsBySymbol(traderID, symbol string, n int) ([]*DecisionRecord, error) {
	key := traderID + "_" + symbol
	records := m.records[key]

	// Return last n records
	if len(records) > n {
		return records[len(records)-n:], nil
	}
	return records, nil
}

func (m *MockDecisionStore) AddRecord(traderID, symbol, action, reasoning string) {
	key := traderID + "_" + symbol
	record := &DecisionRecord{
		Timestamp:   time.Now(),
		InputPrompt: "Test prompt",
		Decisions: []DecisionAction{
			{
				Symbol:    symbol,
				Action:    action,
				Reasoning: reasoning,
			},
		},
	}
	m.records[key] = append(m.records[key], record)
}

func TestDecisionStoreAdapter(t *testing.T) {
	_ = NewMockDecisionStore()
	adapter := &DecisionStoreAdapter{store: nil} // We'll use mock directly

	// Test that adapter implements the interface
	var _ DecisionStoreInterface = adapter
}

func TestHistoryRetrieval(t *testing.T) {
	mockStore := NewMockDecisionStore()

	// Add some test records
	mockStore.AddRecord("trader1", "BTCUSDT", "open_long", "Test reason 1")
	time.Sleep(1 * time.Millisecond)
	mockStore.AddRecord("trader1", "BTCUSDT", "close_long", "Test reason 2")
	time.Sleep(1 * time.Millisecond)
	mockStore.AddRecord("trader1", "ETHUSDT", "open_short", "Test reason 3")

	// Retrieve BTC records
	btcRecords, err := mockStore.GetLatestRecordsBySymbol("trader1", "BTCUSDT", 10)
	if err != nil {
		t.Fatalf("Failed to retrieve records: %v", err)
	}

	if len(btcRecords) != 2 {
		t.Errorf("Expected 2 BTC records, got %d", len(btcRecords))
	}

	// Verify the records are for the correct symbol
	for _, rec := range btcRecords {
		if len(rec.Decisions) == 0 {
			t.Error("Record has no decisions")
			continue
		}
		if rec.Decisions[0].Symbol != "BTCUSDT" {
			t.Errorf("Expected BTCUSDT, got %s", rec.Decisions[0].Symbol)
		}
	}

	// Retrieve ETH records
	ethRecords, err := mockStore.GetLatestRecordsBySymbol("trader1", "ETHUSDT", 10)
	if err != nil {
		t.Fatalf("Failed to retrieve ETH records: %v", err)
	}

	if len(ethRecords) != 1 {
		t.Errorf("Expected 1 ETH record, got %d", len(ethRecords))
	}
}

func TestHistoryLimit(t *testing.T) {
	mockStore := NewMockDecisionStore()

	// Add 15 records
	for i := 0; i < 15; i++ {
		mockStore.AddRecord("trader1", "BTCUSDT", "hold", "Reason "+string(rune(i)))
		time.Sleep(1 * time.Millisecond)
	}

	// Retrieve only 10
	records, err := mockStore.GetLatestRecordsBySymbol("trader1", "BTCUSDT", 10)
	if err != nil {
		t.Fatalf("Failed to retrieve records: %v", err)
	}

	if len(records) != 10 {
		t.Errorf("Expected 10 records (limit), got %d", len(records))
	}
}
