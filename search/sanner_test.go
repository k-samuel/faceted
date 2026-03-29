package search

import (
	"slices"
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
	result := scanner.GetAllRecordId(nil, nil)

	expected := []int{1, 2, 3}
	if len(result) != len(expected) {
		t.Errorf("Expected %d records, got %d", len(expected), len(result))
	}

	if !slices.Equal(expected, result) {
		t.Errorf("Expected slices to be equal")
	}

}

func TestDeduplicate(t *testing.T) {
	list := []int{1, 3, 4, 4, 4}
	list = Deduplicate(list)
	if !slices.Equal(list, []int{1, 3, 4}) {
		t.Errorf("Expected slices to be equal")
	}
}
