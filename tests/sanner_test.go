package tests

import (
	"testing"

	"github.com/k-samuel/faceted"
)

// TestScannerGetAllRecordIdMap tests Scanner GetAllRecordIdMap functionality.
func TestScannerGetAllRecordIdMap(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"col": 2})
	_ = storage.AddRecord(2, map[string]interface{}{"col": 2, "pr": 1})
	_ = storage.AddRecord(3, map[string]interface{}{"col": 2, "pr": 1, "dr": 2})

	scanner := searchIndex.GetScanner()
	result := scanner.GetAllRecordIdMap(storage)

	expected := map[int]bool{1: true, 2: true, 3: true}
	if len(result) != len(expected) {
		t.Errorf("Expected %d records, got %d", len(expected), len(result))
	}
	for id := range expected {
		if !result[id] {
			t.Errorf("Missing record %d", id)
		}
	}
}
