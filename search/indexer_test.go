package search

import (
	"testing"

	"github.com/k-samuel/faceted/search/indexer"
)

// TestRangeIndexerAddRecord tests RangeIndexer AddRecord functionality.
func TestRangeIndexerAddRecord(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	indexer, _ := indexer.NewRangeIndexer(100)
	storage.AddIndexer("price", indexer)

	if err := storage.AddRecord(2, map[string]interface{}{"price": 90}); err != nil {
		t.Errorf("Failed to add record 2")
	}
	if err := storage.AddRecord(3, map[string]interface{}{"price": 100}); err != nil {
		t.Errorf("Failed to add record 3")
	}
	if err := storage.AddRecord(4, map[string]interface{}{"price": 110}); err != nil {
		t.Errorf("Failed to add record 4")
	}
	if err := storage.AddRecord(5, map[string]interface{}{"price": 1000}); err != nil {
		t.Errorf("Failed to add record 5")
	}

	expected := map[string]map[string][]int{
		"price": {
			"0":    []int{2},
			"100":  []int{3, 4},
			"1000": []int{5},
		},
	}

	actual := storage.Export()
	assertEqualFacetDataStringSlice(t, expected, actual)
}

// TestRangeListIndexerAddRecord tests RangeListIndexer AddRecord functionality.
func TestRangeListIndexerAddRecord(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	indexer, _ := indexer.NewRangeListIndexer([]int{100, 200, 150, 500})
	storage.AddIndexer("price", indexer)

	if err := storage.AddRecord(2, map[string]interface{}{"price": 90}); err != nil {
		t.Errorf("Failed to add record 2")
	}
	if err := storage.AddRecord(3, map[string]interface{}{"price": 100}); err != nil {
		t.Errorf("Failed to add record 3")
	}
	if err := storage.AddRecord(4, map[string]interface{}{"price": 110}); err != nil {
		t.Errorf("Failed to add record 4")
	}
	if err := storage.AddRecord(5, map[string]interface{}{"price": 1000}); err != nil {
		t.Errorf("Failed to add record 5")
	}

	expected := map[string]map[string][]int{
		"price": {
			"0":   []int{2},
			"100": []int{3, 4},
			"500": []int{5},
		},
	}

	actual := storage.Export()
	assertEqualFacetDataStringSlice(t, expected, actual)
}

func assertEqualFacetDataStringSlice(t *testing.T, expected, actual map[string]map[string][]int) {
	for field, expectedVal := range expected {
		actualVal, ok := actual[field]
		if !ok {
			t.Errorf("Missing field %s in result", field)
			continue
		}
		for expKey, expRecords := range expectedVal {
			actRecords, ok := actualVal[expKey]
			if !ok {
				t.Errorf("Field %s: missing key %v", field, expKey)
				continue
			}
			if len(expRecords) != len(actRecords) {
				t.Errorf("Field %s[%v]: expected %d records, got %d", field, expKey, len(expRecords), len(actRecords))
				continue
			}
			recordMap := make(map[int]bool)
			for _, r := range actRecords {
				recordMap[r] = true
			}
			for _, r := range expRecords {
				if !recordMap[r] {
					t.Errorf("Field %s[%v]: missing record %d", field, expKey, r)
				}
			}
		}
	}
}
