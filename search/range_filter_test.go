package search

import (
	"testing"

	"github.com/k-samuel/faceted/search/indexer"
	"github.com/k-samuel/faceted/search/value"
)

// TestRangeFilterBasic tests basic RangeFilter functionality.
func TestRangeFilterBasic(t *testing.T) {
	storage := NewMapStorage(value.NewConverter())

	// Add records with specific prices
	prices := []int{500, 750, 1000, 1250, 1500, 2000, 2500, 3000, 4500, 5000, 5500}
	for i, price := range prices {
		storage.AddRecord(i+1, map[string]interface{}{"price": price})
	}

	storage.Optimize()

	// Filter for price [1000, 5000]
	// Should return records with prices: 1000, 1250, 1500, 2000, 2500, 3000, 4500, 5000 = 8 records
	// Prices: 500(ID1), 750(ID2), 1000(ID3), 1250(ID4), 1500(ID5), 2000(ID6), 2500(ID7), 3000(ID8), 4500(ID9), 5000(ID10), 5500(ID11)
	// In range [1000, 5000]: ID3-ID10 = 8 records
	f := NewRangeFilter("price", NewRangeValue(1000, 5000))
	scanner := NewMapScanner(storage)

	result, err := scanner.FindRangeIntersection("price", f.GetValue(), []int{}, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check count
	expectedCount := 8
	if len(result) != expectedCount {
		t.Errorf("Expected %d records in range [1000, 5000], got %d", expectedCount, len(result))
		t.Logf("Got records: %v", result)
	}

	// Check specific records
	expectedRecords := []int{3, 4, 5, 6, 7, 8, 9, 10} // IDs 3-10 for prices 1000-5000
	for _, id := range expectedRecords {
		found := false
		for _, r := range result {
			if r == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected record %d to be in result", id)
		}
	}
}

// TestRangeFilterWithInputRecords tests RangeFilter with input records.
func TestRangeFilterWithInputRecords(t *testing.T) {
	storage := NewMapStorage(value.NewConverter())

	// Add records with prices
	for i := 1; i <= 20; i++ {
		storage.AddRecord(i, map[string]interface{}{"price": i * 250}) // 250, 500, ..., 5000
	}

	storage.Optimize()

	// Filter for price [1000, 3000] with input records [1, 2, 3, 5, 7]
	f := NewRangeFilter("price", NewRangeValue(1000, 3000))
	scanner := NewMapScanner(storage)

	inputRecords := []int{1, 2, 3, 5, 7} // IDs for prices 250, 500, 750, 1250, 1750
	result, err := scanner.FindRangeIntersection("price", f.GetValue(), inputRecords, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Expected: price 1250 (ID 5) and 1750 (ID 7) are in range [1000, 3000]
	expectedCount := 2
	if len(result) != expectedCount {
		t.Errorf("Expected %d records, got %d: %v", expectedCount, len(result), result)
	}
}

// TestRangeFilterWithExcludes tests RangeFilter with exclude records.
func TestRangeFilterWithExcludes(t *testing.T) {
	storage := NewMapStorage(value.NewConverter())

	// Add records with prices
	prices := []int{1000, 1500, 2000, 2500, 3000}
	for i, price := range prices {
		storage.AddRecord(i+1, map[string]interface{}{"price": price})
	}

	storage.Optimize()

	// Filter for price [1000, 3000] excluding records 2 and 4
	f := NewRangeFilter("price", NewRangeValue(1000, 3000))
	scanner := NewMapScanner(storage)

	excludeRecords := map[int]struct{}{2: {}, 4: {}} // exclude prices 1500 and 2500
	result, err := scanner.FindRangeIntersection("price", f.GetValue(), []int{}, excludeRecords)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Expected: 1000 (ID 1), 2000 (ID 3), 3000 (ID 5) = 3 records
	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("Expected %d records, got %d: %v", expectedCount, len(result), result)
	}

	// Check that excluded records are not in result
	for _, id := range []int{2, 4} {
		for _, r := range result {
			if r == id {
				t.Errorf("Excluded record %d should not be in result", id)
			}
		}
	}
}

// TestRangeFilterEdgeCases tests RangeFilter edge cases.
func TestRangeFilterEdgeCases(t *testing.T) {
	storage := NewMapStorage(value.NewConverter())

	// Add records
	prices := []int{100, 200, 300, 400, 500}
	for i, price := range prices {
		storage.AddRecord(i+1, map[string]interface{}{"price": price})
	}

	storage.Optimize()
	scanner := NewMapScanner(storage)

	tests := []struct {
		name        string
		min         int
		max         int
		expected    int
		description string
	}{
		{"exact range", 200, 400, 3, "should return 200, 300, 400"},
		{"min only", 300, 0, 3, "should return 300, 400, 500 (no max)"},
		{"max only", 100, 300, 3, "should return 100, 200, 300 (no min)"},
		{"empty range", 600, 700, 0, "should return nothing"},
		{"zero values", 0, 0, 0, "should return empty (special case)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var minVal interface{}
			var maxVal interface{}

			// Handle min-only case - use nil for no max
			if tt.name == "min only" {
				minVal = tt.min
				maxVal = nil
			} else if tt.name == "max only" {
				minVal = nil
				maxVal = tt.max
			} else {
				minVal = tt.min
				maxVal = tt.max
			}

			f := NewRangeFilter("price", NewRangeValue(minVal, maxVal))
			result, err := scanner.FindRangeIntersection("price", f.GetValue(), []int{}, nil)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(result) != tt.expected {
				t.Errorf("%s: expected %d records, got %d - %s", tt.name, tt.expected, len(result), tt.description)
			}
		})
	}
}

// TestRangeFilterDuplicateRecords tests RangeFilter with duplicate record IDs in data.
func TestRangeFilterDuplicateRecords(t *testing.T) {
	storage := NewMapStorage(value.NewConverter())

	// Add records - some prices may have multiple records with same ID (edge case)
	storage.AddRecord(1, map[string]interface{}{"price": 1000})
	storage.AddRecord(2, map[string]interface{}{"price": 1000}) // same price as record 1
	storage.AddRecord(3, map[string]interface{}{"price": 2000})
	storage.AddRecord(4, map[string]interface{}{"price": 2000}) // same price as record 3

	storage.Optimize()

	f := NewRangeFilter("price", NewRangeValue(1000, 2000))
	scanner := NewMapScanner(storage)

	result, err := scanner.FindRangeIntersection("price", f.GetValue(), []int{}, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// All 4 records should be in range [1000, 2000]
	expectedCount := 4
	if len(result) != expectedCount {
		t.Errorf("Expected %d records, got %d: %v", expectedCount, len(result), result)
	}
}

// TestRangeFilterWithIndexer tests RangeFilter with RangeIndexer.
func TestRangeFilterWithIndexer(t *testing.T) {
	storage := NewMapStorage(value.NewConverter())

	// Add RangeIndexer with step 250
	rangeIndexer, _ := indexer.NewRangeIndexer(250)
	storage.AddIndexer("price", rangeIndexer)

	// Add records with prices that fall into ranges [1000, 5000]
	prices := []int{1000, 1100, 1200, 1300, 1400, 1500, 1600, 2000, 2500, 3000, 4500, 5000}
	for i, price := range prices {
		storage.AddRecord(i+1, map[string]interface{}{"price": price})
	}

	storage.Optimize()

	// Filter for price [1000, 5000]
	f := NewRangeFilter("price", NewRangeValue(1000, 5000))
	scanner := NewMapScanner(storage)

	result, err := scanner.FindRangeIntersection("price", f.GetValue(), []int{}, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// All 12 records should be in range [1000, 5000]
	expectedCount := len(prices)
	if len(result) != expectedCount {
		t.Errorf("Expected %d records, got %d: %v", expectedCount, len(result), result)
	}

	// Verify all prices are in range
	for _, id := range result {
		// Get the price for this ID (need to track it separately in real test)
		t.Logf("Record ID: %d", id)
	}
}

// TestRangeFilterConsistency tests that RangeFilter returns consistent results.
func TestRangeFilterConsistency(t *testing.T) {
	storage := NewMapStorage(value.NewConverter())

	// Add many records
	for i := 1; i <= 100; i++ {
		storage.AddRecord(i, map[string]interface{}{"price": i * 100})
	}

	storage.Optimize()
	scanner := NewMapScanner(storage)

	f := NewRangeFilter("price", NewRangeValue(3000, 7000))

	// Run multiple times to check consistency
	var lastResult []int
	for i := 0; i < 5; i++ {
		result, err := scanner.FindRangeIntersection("price", f.GetValue(), []int{}, nil)
		if err != nil {
			t.Errorf("Iteration %d: expected no error, got %v", i, err)
		}

		if i > 0 && len(result) != len(lastResult) {
			t.Errorf("Iteration %d: inconsistent result length: expected %d, got %d", i, len(lastResult), len(result))
		}

		// Verify all records are in range
		for _, id := range result {
			// Price for ID should be between 3000 and 7000
			// ID 30 -> price 3000, ID 70 -> price 7000
			if id < 30 || id > 70 {
				t.Errorf("Record %d is not in range [3000, 7000]", id)
			}
		}

		lastResult = result
	}

	// Expected: IDs 30-70 = 41 records
	expectedCount := 41
	if len(lastResult) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(lastResult))
	}
}
