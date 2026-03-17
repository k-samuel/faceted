package indexer

import (
	"testing"
)

// TestNewRangeIndexer tests creating a new RangeIndexer.
func TestNewRangeIndexer(t *testing.T) {
	// Valid step
	ri, err := NewRangeIndexer(100)
	if err != nil {
		t.Errorf("Expected no error for valid step, got %v", err)
	}
	if ri.step != 100 {
		t.Errorf("Expected step 100, got %d", ri.step)
	}

	// Invalid step (zero)
	_, err = NewRangeIndexer(0)
	if err == nil {
		t.Errorf("Expected error for zero step")
	}

	// Invalid step (negative)
	_, err = NewRangeIndexer(-10)
	if err == nil {
		t.Errorf("Expected error for negative step")
	}
}

// TestRangeIndexerAdd tests adding records to RangeIndexer.
func TestRangeIndexerAdd(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	indexContainer := make(map[string][]int)

	// Test adding records with various values
	err := ri.Add(&indexContainer, 1, []string{"50"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = ri.Add(&indexContainer, 2, []string{"150"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = ri.Add(&indexContainer, 3, []string{"250"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check ranges: 0-99, 100-199, 200-299
	if len(indexContainer) != 3 {
		t.Errorf("Expected 3 ranges, got %d", len(indexContainer))
	}

	// Check range 0-99
	if _, ok := indexContainer["0"]; !ok {
		t.Errorf("Expected range 0-99 to exist")
	}

	// Check range 100-199
	if _, ok := indexContainer["100"]; !ok {
		t.Errorf("Expected range 100-199 to exist")
	}
}

// TestRangeIndexerAddMultipleValues tests adding records with multiple values.
func TestRangeIndexerAddMultipleValues(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	indexContainer := make(map[string][]int)

	err := ri.Add(&indexContainer, 1, []string{"50", "150", "250"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should have 3 ranges
	if len(indexContainer) != 3 {
		t.Errorf("Expected 3 ranges, got %d", len(indexContainer))
	}
}

// TestRangeIndexerAddInvalidValue tests adding invalid values.
func TestRangeIndexerAddInvalidValue(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	indexContainer := make(map[string][]int)

	err := ri.Add(&indexContainer, 1, []string{"invalid"})
	if err == nil {
		t.Errorf("Expected error for invalid value")
	}
}

// TestRangeIndexerOptimize tests optimizing the range index.
func TestRangeIndexerOptimize(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	indexContainer := make(map[string][]int)

	// Add records without optimization first
	ri.hasUnsorted = true
	ri.unsortedBuf["0"] = make(map[string][]int)
	ri.unsortedBuf["0"]["50"] = []int{1, 2}
	ri.unsortedBuf["0"]["30"] = []int{3}
	ri.unsortedBuf["0"]["70"] = []int{4, 5}

	ri.Optimize(&indexContainer)

	// After optimization, unsortedBuf should be cleared
	if len(ri.unsortedBuf) != 0 {
		t.Errorf("Expected unsortedBuf to be cleared after optimization")
	}

	// Check that hasUnsorted is reset
	if ri.hasUnsorted {
		t.Errorf("Expected hasUnsorted to be false after optimization")
	}
}

// TestRangeIndexerOptimizeNoUnsorted tests optimizing when no unsorted data exists.
func TestRangeIndexerOptimizeNoUnsorted(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	ri.hasUnsorted = false
	indexContainer := make(map[string][]int)

	// Should not panic
	ri.Optimize(&indexContainer)
}

// TestDetectRangeKey tests detecting range keys.
func TestRangeIndexerDetectRangeKey(t *testing.T) {
	ri, _ := NewRangeIndexer(100)

	tests := []struct {
		value    float64
		expected int
	}{
		{50, 0},    // 0-99
		{99, 0},    // 0-99
		{100, 100}, // 100-199
		{150, 100}, // 100-199
		{200, 200}, // 200-299
		{299, 200}, // 200-299
		{1000, 1000},
	}

	for _, tt := range tests {
		result := ri.detectRangeKey(tt.value)
		if result != tt.expected {
			t.Errorf("detectRangeKey(%v): expected %d, got %d", tt.value, tt.expected, result)
		}
	}
}

// TestRangeIndexerStepSizes tests different step sizes.
func TestRangeIndexerStepSizes(t *testing.T) {
	steps := []int{50, 100, 200, 500}

	for _, step := range steps {
		ri, err := NewRangeIndexer(step)
		if err != nil {
			t.Errorf("Failed to create RangeIndexer with step %d: %v", step, err)
			continue
		}

		if ri.step != step {
			t.Errorf("Expected step %d, got %d", step, ri.step)
		}
	}
}

// TestRangeIndexerEdgeCases tests edge cases.
func TestRangeIndexerEdgeCases(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	indexContainer := make(map[string][]int)

	// Test with zero value
	err := ri.Add(&indexContainer, 1, []string{"0"})
	if err != nil {
		t.Errorf("Expected no error for zero value, got %v", err)
	}

	// Test with large value
	err = ri.Add(&indexContainer, 2, []string{"999999"})
	if err != nil {
		t.Errorf("Expected no error for large value, got %v", err)
	}

	// Check that ranges were created
	if len(indexContainer) != 2 {
		t.Errorf("Expected 2 ranges, got %d", len(indexContainer))
	}
}

// TestRangeIndexerDuplicateRecords tests handling of duplicate record IDs.
func TestRangeIndexerDuplicateRecords(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	indexContainer := make(map[string][]int)

	// Add same record multiple times
	ri.Add(&indexContainer, 1, []string{"50"})
	ri.Add(&indexContainer, 1, []string{"150"})

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

// TestRangeIndexerUnsortedFlag tests the hasUnsorted flag.
func TestRangeIndexerUnsortedFlag(t *testing.T) {
	ri, _ := NewRangeIndexer(100)
	indexContainer := make(map[string][]int)

	// Initially hasUnsorted should be false
	if ri.hasUnsorted {
		t.Errorf("Expected hasUnsorted to be false initially")
	}

	// After adding, it should be true
	ri.Add(&indexContainer, 1, []string{"50"})
	if !ri.hasUnsorted {
		t.Errorf("Expected hasUnsorted to be true after adding")
	}
}
