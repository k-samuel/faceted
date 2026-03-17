package search

import (
	"testing"
)

// TestScannerGetAllRecordIdMap tests Scanner GetAllRecordIdMap functionality.
func TestScannerGetAllRecordIdMap(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"col": 2})
	_ = storage.AddRecord(2, map[string]interface{}{"col": 2, "pr": 1})
	_ = storage.AddRecord(3, map[string]interface{}{"col": 2, "pr": 1, "dr": 2})

	scanner := db.GetScanner()
	result := scanner.GetAllRecordIdMap()

	expected := map[int]bool{1: true, 2: true, 3: true}
	if len(result) != len(expected) {
		t.Errorf("Expected %d records, got %d", len(expected), len(result))
	}
	for id := range expected {
		if _, ok := result[id]; !ok {
			t.Errorf("Missing record %d", id)
		}
	}
}
