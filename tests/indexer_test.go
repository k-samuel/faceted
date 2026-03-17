package tests

import (
	"testing"

	"github.com/k-samuel/faceted"
	"github.com/k-samuel/faceted/pkg/indexer"
)

// TestRangeIndexerAddRecord tests RangeIndexer AddRecord functionality.
func TestRangeIndexerAddRecord(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

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

	actual := storage.GetData()
	assertEqualFacetDataStringSlice(t, expected, actual)
}

// TestRangeListIndexerAddRecord tests RangeListIndexer AddRecord functionality.
func TestRangeListIndexerAddRecord(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	indexer, _ := search.NewRangeListIndexer([]int{100, 200, 150, 500})
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

	actual := storage.GetData()
	assertEqualFacetDataStringSlice(t, expected, actual)
}
