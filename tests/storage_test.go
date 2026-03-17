package tests

import (
	"testing"

	"github.com/k-samuel/faceted"
)

// TestStorageAddRecord tests Storage AddRecord functionality.
func TestStorageAddRecord(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	if err := storage.AddRecord(112, map[string]interface{}{"vendor": "Tester", "price": 100}); err != nil {
		t.Errorf("Failed to add record 112: %v", err)
	}
	if err := storage.AddRecord(113, map[string]interface{}{"vendor": "Tester2", "price": 101}); err != nil {
		t.Errorf("Failed to add record 113: %v", err)
	}
	if err := storage.AddRecord(114, map[string]interface{}{"vendor": "Tester2", "price": 101}); err != nil {
		t.Errorf("Failed to add record 114: %v", err)
	}

	expected := map[string]map[string][]int{
		"vendor": {
			"Tester":  []int{112},
			"Tester2": []int{113, 114},
		},
		"price": {
			"100": []int{112},
			"101": []int{113, 114},
		},
	}

	actualData := storage.GetData()
	assertEqualFacetDataStringSlice(t, expected, actualData)
}

// TestStorageHasField tests Storage HasField functionality.
func TestStorageHasField(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"vendor": "Tester", "price": 100})

	if !storage.HasField("vendor") {
		t.Errorf("Expected field 'vendor' to exist")
	}
	if !storage.HasField("price") {
		t.Errorf("Expected field 'price' to exist")
	}
	if storage.HasField("nonexistent") {
		t.Errorf("Expected field 'nonexistent' not to exist")
	}
}

// TestStorageExport tests Storage Export functionality.
func TestStorageExport(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"vendor": "Tester", "price": 100})
	_ = storage.AddRecord(2, map[string]interface{}{"vendor": "Tester2", "price": 101})

	exported := storage.Export()
	if len(exported) != 2 {
		t.Errorf("Expected 2 fields in export, got %d", len(exported))
	}
}

// TestStorageRecordsCount tests Storage GetRecordsCount functionality.
func TestStorageRecordsCount(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"vendor": "Tester", "price": 100})
	_ = storage.AddRecord(2, map[string]interface{}{"vendor": "Tester", "price": 100})
	_ = storage.AddRecord(3, map[string]interface{}{"vendor": "Tester2", "price": 100})

	if count := storage.GetRecordsCount("vendor", "Tester"); count != 2 {
		t.Errorf("Expected vendor 'Tester' count to be 2, got %d", count)
	}
	if count := storage.GetRecordsCount("price", "100"); count != 3 {
		t.Errorf("Expected price '100' count to be 3, got %d", count)
	}
}

// TestStorageSetData tests Storage SetData functionality.
func TestStorageSetData(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	data := map[string]map[string][]int{
		"vendor": {
			"Tester":  []int{1},
			"Tester2": []int{2},
		},
	}

	storage.SetData(data)

	result := storage.GetData()
	if len(result) != 1 {
		t.Errorf("Expected 1 field after SetData, got %d", len(result))
	}
}

// TestStorageOptimize tests Storage Optimize functionality.
func TestStorageOptimize(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"vendor": "Tester"})

	// Optimize should not panic
	storage.Optimize()
}

// TestStorageDeleteRecord tests Storage DeleteRecord functionality.
func TestStorageDeleteRecord(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"vendor": "Tester"})
	_ = storage.AddRecord(2, map[string]interface{}{"vendor": "Tester"})

	storage.DeleteRecord(1)

	result := storage.GetData()
	if len(result["vendor"]["Tester"]) != 1 {
		t.Errorf("Expected vendor 'Tester' count to be 1 after deletion, got %d",
			len(result["vendor"]["Tester"]))
	}
}

// TestStorageReplaceRecord tests Storage ReplaceRecord functionality.
func TestStorageReplaceRecord(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"vendor": "Tester", "price": 100})

	if err := storage.ReplaceRecord(1, map[string]interface{}{"vendor": "Tester2", "price": 200}); err != nil {
		t.Errorf("Failed to replace record 1: %v", err)
	}

	result := storage.GetData()
	if len(result["vendor"]["Tester"]) != 0 {
		t.Errorf("Expected vendor 'Tester' count to be 0 after replacement, got %d",
			len(result["vendor"]["Tester"]))
	}
	if len(result["vendor"]["Tester2"]) != 1 {
		t.Errorf("Expected vendor 'Tester2' count to be 1 after replacement, got %d",
			len(result["vendor"]["Tester2"]))
	}
}
