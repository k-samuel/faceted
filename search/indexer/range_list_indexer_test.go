package indexer

import (
	"testing"
)

// TestNewRangeListIndexer tests creating a new RangeListIndexer.
func TestNewRangeListIndexer(t *testing.T) {
	// Valid ranges
	ri, err := NewRangeListIndexer([]int{100, 500, 1000})
	if err != nil {
		t.Errorf("Expected no error for valid ranges, got %v", err)
	}
	if len(ri.ranges) != 3 {
		t.Errorf("Expected 3 ranges, got %d", len(ri.ranges))
	}

	// Invalid ranges (too few)
	_, err = NewRangeListIndexer([]int{100})
	if err == nil {
		t.Errorf("Expected error for too few ranges")
	}

	// Test with unsorted input (should be sorted internally)
	ri2, err := NewRangeListIndexer([]int{500, 100, 1000})
	if err != nil {
		t.Errorf("Expected no error for unsorted ranges, got %v", err)
	}
	// Check that ranges are sorted
	if ri2.ranges[0] != 100 || ri2.ranges[1] != 500 || ri2.ranges[2] != 1000 {
		t.Errorf("Expected sorted ranges [100, 500, 1000], got %v", ri2.ranges)
	}
}

// TestRangeListIndexerAdd tests adding records to RangeListIndexer.
func TestRangeListIndexerAdd(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	// Test adding records with various values
	err := ri.Add(indexContainer, 1, []string{"50"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = ri.Add(indexContainer, 2, []string{"150"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = ri.Add(indexContainer, 3, []string{"550"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = ri.Add(indexContainer, 4, []string{"1500"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check ranges: 0-99, 100-499, 500-999, 1000+
	if len(indexContainer) != 4 {
		t.Errorf("Expected 4 ranges, got %d", len(indexContainer))
	}

	// Check range 0-99
	if _, ok := indexContainer["0"]; !ok {
		t.Errorf("Expected range 0-99 to exist")
	}

	// Check range 100-499
	if _, ok := indexContainer["100"]; !ok {
		t.Errorf("Expected range 100-499 to exist")
	}

	// Check range 500-999
	if _, ok := indexContainer["500"]; !ok {
		t.Errorf("Expected range 500-999 to exist")
	}

	// Check range 1000+
	if _, ok := indexContainer["1000"]; !ok {
		t.Errorf("Expected range 1000+ to exist")
	}
}

// TestRangeListIndexerAddMultipleValues tests adding records with multiple values.
func TestRangeListIndexerAddMultipleValues(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	err := ri.Add(indexContainer, 1, []string{"50", "150", "550"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should have 3 ranges
	if len(indexContainer) != 3 {
		t.Errorf("Expected 3 ranges, got %d", len(indexContainer))
	}
}

// TestRangeListIndexerAddInvalidValue tests adding invalid values.
func TestRangeListIndexerAddInvalidValue(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	err := ri.Add(indexContainer, 1, []string{"invalid"})
	if err == nil {
		t.Errorf("Expected error for invalid value")
	}
}

// TestRangeListIndexerOptimize tests optimizing the range index.
func TestRangeListIndexerOptimize(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	// Add records without optimization first
	ri.hasUnsorted = true
	ri.unsortedBuf["0"] = make(map[string][]int)
	ri.unsortedBuf["0"]["80"] = []int{1, 2}
	ri.unsortedBuf["0"]["20"] = []int{3}
	ri.unsortedBuf["0"]["60"] = []int{4, 5}

	ri.Optimize(indexContainer)

	// After optimization, unsortedBuf should be cleared
	if len(ri.unsortedBuf) != 0 {
		t.Errorf("Expected unsortedBuf to be cleared after optimization")
	}

	// Check that hasUnsorted is reset
	if ri.hasUnsorted {
		t.Errorf("Expected hasUnsorted to be false after optimization")
	}
}

// TestRangeListIndexerOptimizeNoUnsorted tests optimizing when no unsorted data exists.
func TestRangeListIndexerOptimizeNoUnsorted(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	ri.hasUnsorted = false
	indexContainer := make(map[string][]int)

	// Should not panic
	ri.Optimize(indexContainer)
}

// TestRangeListIndexerDetectRangeKey tests detecting range keys.
func TestRangeListIndexerDetectRangeKey(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})

	tests := []struct {
		value    float64
		expected int
	}{
		{50, 0},      // 0-99
		{99, 0},      // 0-99
		{100, 100},   // 100-499
		{250, 100},   // 100-499
		{499, 100},   // 100-499
		{500, 500},   // 500-999
		{750, 500},   // 500-999
		{999, 500},   // 500-999
		{1000, 1000}, // 1000+
		{2000, 1000}, // 1000+
	}

	for _, tt := range tests {
		result := ri.detectRangeKey(tt.value)
		if result != tt.expected {
			t.Errorf("detectRangeKey(%v): expected %d, got %d", tt.value, tt.expected, result)
		}
	}
}

// TestRangeListIndexerEdgeCases tests edge cases.
func TestRangeListIndexerEdgeCases(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	// Test with zero value
	err := ri.Add(indexContainer, 1, []string{"0"})
	if err != nil {
		t.Errorf("Expected no error for zero value, got %v", err)
	}

	// Test with large value
	err = ri.Add(indexContainer, 2, []string{"999999"})
	if err != nil {
		t.Errorf("Expected no error for large value, got %v", err)
	}

	// Check that ranges were created
	if len(indexContainer) != 2 {
		t.Errorf("Expected 2 ranges, got %d", len(indexContainer))
	}
}

// TestRangeListIndexerDuplicateRecords tests handling of duplicate record IDs.
func TestRangeListIndexerDuplicateRecords(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	// Add same record multiple times
	ri.Add(indexContainer, 1, []string{"50"})
	ri.Add(indexContainer, 1, []string{"150"})

	// Should have 2 ranges
	if len(indexContainer) != 2 {
		t.Errorf("Expected 2 ranges, got %d", len(indexContainer))
	}

	// Record 1 should appear in both ranges
	if len(indexContainer["0"]) != 1 || indexContainer["0"][0] != 1 {
		t.Errorf("Expected record 1 in range 0")
	}

	if len(indexContainer["100"]) != 1 || indexContainer["100"][0] != 1 {
		t.Errorf("Expected record 1 in range 100")
	}
}

// TestRangeListIndexerCustomRanges tests custom range configurations.
func TestRangeListIndexerCustomRanges(t *testing.T) {
	ranges := []int{10, 25, 50, 100}
	ri, _ := NewRangeListIndexer(ranges)

	tests := []struct {
		value    float64
		expected int
	}{
		{5, 0},     // 0-9
		{10, 10},   // 10-24
		{20, 10},   // 10-24
		{25, 25},   // 25-49
		{40, 25},   // 25-49
		{50, 50},   // 50-99
		{99, 50},   // 50-99
		{100, 100}, // 100+
	}

	for _, tt := range tests {
		result := ri.detectRangeKey(tt.value)
		if result != tt.expected {
			t.Errorf("detectRangeKey(%v) with ranges %v: expected %d, got %d", tt.value, ranges, tt.expected, result)
		}
	}
}

// TestRangeListIndexerUnsortedFlag tests the hasUnsorted flag.
func TestRangeListIndexerUnsortedFlag(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	// Initially hasUnsorted should be false
	if ri.hasUnsorted {
		t.Errorf("Expected hasUnsorted to be false initially")
	}

	// After adding, it should be true
	ri.Add(indexContainer, 1, []string{"50"})
	if !ri.hasUnsorted {
		t.Errorf("Expected hasUnsorted to be true after adding")
	}
}

// TestRangeListIndexerNegativeValues tests handling of negative values.
func TestRangeListIndexerNegativeValues(t *testing.T) {
	ri, _ := NewRangeListIndexer([]int{100, 500, 1000})
	indexContainer := make(map[string][]int)

	err := ri.Add(indexContainer, 1, []string{"-50"})
	if err != nil {
		t.Errorf("Expected no error for negative value, got %v", err)
	}

	err = ri.Add(indexContainer, 2, []string{"-10"})
	if err != nil {
		t.Errorf("Expected no error for negative value, got %v", err)
	}

	// Negative values fall into the first range (0) since detectRangeKey
	// returns lastKey when value < first boundary
	// The ranges are: 0-99, 100-499, 500-999, 1000+
	// So -50 and -10 would fall into range 0 (0-99)
	if len(indexContainer) != 1 {
		t.Errorf("Expected 1 range for negative values, got %d", len(indexContainer))
	}
}
