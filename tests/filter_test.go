package tests

import (
	"testing"

	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/value"
)

// TestValueFilterSetValue tests ValueFilter SetValue functionality.
func TestValueFilterSetValue(t *testing.T) {
	f := filter.NewValueFilter("color", "red", value.NewValueConverterDefault())

	values := f.GetValue()
	if len(values) != 1 || values[0] != "red" {
		t.Errorf("Expected single value 'red', got %v", values)
	}

	f.SetValue([]interface{}{"blue", "green"})
	values = f.GetValue()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
	if values[0] != "blue" || values[1] != "green" {
		t.Errorf("Expected ['blue', 'green'], got %v", values)
	}

	f.SetValue(42)
	values = f.GetValue()
	if len(values) != 1 || values[0] != "42" {
		t.Errorf("Expected single value 42, got %v", values)
	}

	f.SetValue([]int{1, 2, 3})
	values = f.GetValue()
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
}

// TestValueFilterSelfFiltering tests ValueFilter SelfFiltering functionality.
func TestValueFilterSelfFiltering(t *testing.T) {
	f := filter.NewValueFilter("color", "red", value.NewValueConverterDefault())

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}

	f.SelfFiltering(false)
	if f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be disabled")
	}
}

// TestExcludeValueFilterGetValue tests ExcludeValueFilter GetValue functionality.
func TestExcludeValueFilterGetValue(t *testing.T) {
	f := filter.NewExcludeValueFilter("color", []interface{}{"red", "blue"}, value.NewValueConverterDefault())

	values := f.GetValue()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
}

// TestExcludeValueFilterSelfFiltering tests ExcludeValueFilter self-filtering.
func TestExcludeValueFilterSelfFiltering(t *testing.T) {
	f := filter.NewExcludeValueFilter("color", []interface{}{"red"}, value.NewValueConverterDefault())

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering enabled")
	}

	f.SelfFiltering(false)
	if f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering disabled")
	}
}

// TestExcludeValueFilterGetValueWithDifferentTypes tests ExcludeValueFilter GetValue with different types.
func TestExcludeValueFilterGetValueWithDifferentTypes(t *testing.T) {
	f := filter.NewExcludeValueFilter("size", []interface{}{42, 44, 46}, value.NewValueConverterDefault())

	values := f.GetValue()
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
}

// TestExcludeValueFilterWithEmptyValues tests ExcludeValueFilter with empty values.
func TestExcludeValueFilterWithEmptyValues(t *testing.T) {
	f := filter.NewExcludeValueFilter("color", []interface{}{}, value.NewValueConverterDefault())

	values := f.GetValue()
	if len(values) != 0 {
		t.Errorf("Expected 0 values, got %d", len(values))
	}
}

// TestExcludeValueFilter tests ExcludeValueFilter functionality.
func TestExcludeValueFilter(t *testing.T) {
	f := filter.NewExcludeValueFilter("color", []interface{}{"red", "blue"}, value.NewValueConverterDefault())

	values := f.GetValue()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}
}

// TestRangeFilter tests RangeFilter functionality.
func TestRangeFilter(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	values := f.GetValue()
	if values.Min != 10 || values.Max != 100 {
		t.Errorf("Expected RangeValue{Min: 10, Max: 100}, got %v", values)
	}

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}
}

// TestRangeFilterGetValue tests RangeFilter GetValue functionality.
func TestRangeFilterGetValue(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	values := f.GetValue()
	if values.Min != 10 || values.Max != 100 {
		t.Errorf("Expected RangeValue{Min: 10, Max: 100}, got %v", values)
	}
}

// TestRangeFilterGetValueWithFloat tests RangeFilter GetValue with float values.
func TestRangeFilterGetValueWithFloat(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10.5, Max: 99.9})

	values := f.GetValue()
	if values.Min != 10.5 || values.Max != 99.9 {
		t.Errorf("Expected RangeValue{Min: 10.5, Max: 99.9}, got %v", values)
	}
}

// TestRangeFilterGetValueWithNegative tests RangeFilter GetValue with negative values.
func TestRangeFilterGetValueWithNegative(t *testing.T) {
	f := filter.NewRangeFilter("temperature", &filter.RangeValue{Min: -100, Max: -10})

	values := f.GetValue()
	if values.Min != -100 || values.Max != -10 {
		t.Errorf("Expected RangeValue{Min: -100, Max: -10}, got %v", values)
	}
}

// TestRangeFilterSetValue tests RangeFilter SetValue functionality.
func TestRangeFilterSetValue(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	f.SetValue(&filter.RangeValue{Min: 20, Max: 200})
	values := f.GetValue()
	if values.Min != 20 || values.Max != 200 {
		t.Errorf("Expected RangeValue{Min: 20, Max: 200}, got %v", values)
	}
}

// TestRangeFilterGetFieldName tests RangeFilter GetFieldName functionality.
func TestRangeFilterGetFieldName(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	if f.GetFieldName() != "price" {
		t.Errorf("Expected field name 'price', got %s", f.GetFieldName())
	}
}

// TestRangeFilterGetMin tests RangeFilter GetMin functionality.
func TestRangeFilterGetMin(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 25.5, Max: 100})

	if f.GetMin() != 25.5 {
		t.Errorf("Expected min 25.5, got %v", f.GetMin())
	}
}

// TestRangeFilterGetMax tests RangeFilter GetMax functionality.
func TestRangeFilterGetMax(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10, Max: 99.9})

	if f.GetMax() != 99.9 {
		t.Errorf("Expected max 99.9, got %v", f.GetMax())
	}
}

// TestRangeFilterWithZeroValue tests RangeFilter with zero values.
func TestRangeFilterWithZeroValue(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 0, Max: 0})

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}
}

// TestRangeFilterWithNegativeValues tests RangeFilter with negative values.
func TestRangeFilterWithNegativeValues(t *testing.T) {
	f := filter.NewRangeFilter("temperature", &filter.RangeValue{Min: -20, Max: -5})

	values := f.GetValue()
	if values.Min != -20 || values.Max != -5 {
		t.Errorf("Expected RangeValue{Min: -20, Max: -5}, got %v", values)
	}
}

// TestRangeFilterWithFloatValues tests RangeFilter with float values.
func TestRangeFilterWithFloatValues(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10.5, Max: 99.99})

	values := f.GetValue()
	if values.Min != 10.5 || values.Max != 99.99 {
		t.Errorf("Expected RangeValue{Min: 10.5, Max: 99.99}, got %v", values)
	}
}

// TestRangeFilterFilterInput tests RangeFilter FilterInput functionality.
func TestRangeFilterFilterInput(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	facetedData := map[string][]int{
		"5":   {1, 2},
		"15":  {3, 4},
		"50":  {5, 6, 7},
		"99":  {8, 9},
		"100": {10},
		"101": {11, 12},
	}

	inputRecords := map[int]bool{1: true, 3: true, 5: true, 8: true, 10: true, 11: true}
	excludeRecords := map[int]bool{5: true}

	f.FilterInput(facetedData, inputRecords, excludeRecords)

	// Expected: 3,4,8,9,10 (records in range [10,100]: 3,4,5,6,7,8,9,10)
	// After excluding 5: 3,4,6,7,8,9,10
	// After intersecting with inputRecords {1,3,5,8,10,11}: 3,8,10
	expectedCount := 3 // records 3, 8, 10
	if len(inputRecords) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(inputRecords))
	}

	if !inputRecords[3] {
		t.Errorf("Expected record 3 to be included")
	}
	if inputRecords[5] {
		t.Errorf("Expected record 5 to be excluded")
	}
	if !inputRecords[8] {
		t.Errorf("Expected record 8 to be included")
	}
	if !inputRecords[10] {
		t.Errorf("Expected record 10 to be included")
	}
}

// TestRangeFilterFilterInputWithNoMatches tests RangeFilter FilterInput with no matches.
func TestRangeFilterFilterInputWithNoMatches(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 200, Max: 300})

	facetedData := map[string][]int{
		"50":  {1, 2},
		"100": {3, 4},
	}

	inputRecords := map[int]bool{1: true, 2: true, 3: true}
	f.FilterInput(facetedData, inputRecords, map[int]bool{})

	if len(inputRecords) != 0 {
		t.Errorf("Expected 0 records after filtering, got %d", len(inputRecords))
	}
}

// TestRangeFilterFilterInputWithAllMatches tests RangeFilter FilterInput with all matches.
func TestRangeFilterFilterInputWithAllMatches(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 0, Max: 1000})

	facetedData := map[string][]int{
		"10": {1, 2},
		"50": {3, 4},
	}

	inputRecords := map[int]bool{1: true, 2: true, 3: true, 4: true}
	f.FilterInput(facetedData, inputRecords, map[int]bool{})

	if len(inputRecords) != 4 {
		t.Errorf("Expected 4 records, got %d", len(inputRecords))
	}
}

// TestRangeFilterFilterInputWithExcludes tests RangeFilter FilterInput with exclude records.
func TestRangeFilterFilterInputWithExcludes(t *testing.T) {
	f := filter.NewRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	facetedData := map[string][]int{
		"20": {1, 2, 3},
		"50": {4, 5},
	}

	inputRecords := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true}
	excludeRecords := map[int]bool{2: true, 5: true}

	f.FilterInput(facetedData, inputRecords, excludeRecords)

	if len(inputRecords) != 3 {
		t.Errorf("Expected 3 records, got %d", len(inputRecords))
	}
	if inputRecords[2] {
		t.Errorf("Expected record 2 to be excluded")
	}
	if inputRecords[5] {
		t.Errorf("Expected record 5 to be excluded")
	}
}

// TestExcludeRangeFilter tests ExcludeRangeFilter functionality.
func TestExcludeRangeFilter(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	values := f.GetValue()
	if values.Min != 10 || values.Max != 100 {
		t.Errorf("Expected RangeValue{Min: 10, Max: 100}, got %v", values)
	}

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}
}

// TestExcludeRangeFilterGetValue tests ExcludeRangeFilter GetValue functionality.
func TestExcludeRangeFilterGetValue(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 20, Max: 200})

	values := f.GetValue()
	if values.Min != 20 || values.Max != 200 {
		t.Errorf("Expected RangeValue{Min: 20, Max: 200}, got %v", values)
	}
}

// TestExcludeRangeFilterGetValueWithFloat tests ExcludeRangeFilter GetValue with float values.
func TestExcludeRangeFilterGetValueWithFloat(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 20.5, Max: 200.5})

	values := f.GetValue()
	if values.Min != 20.5 || values.Max != 200.5 {
		t.Errorf("Expected RangeValue{Min: 20.5, Max: 200.5}, got %v", values)
	}
}

// TestExcludeRangeFilterGetValueWithNegative tests ExcludeRangeFilter GetValue with negative values.
func TestExcludeRangeFilterGetValueWithNegative(t *testing.T) {
	f := filter.NewExcludeRangeFilter("temperature", &filter.RangeValue{Min: -200, Max: -20})

	values := f.GetValue()
	if values.Min != -200 || values.Max != -20 {
		t.Errorf("Expected RangeValue{Min: -200, Max: -20}, got %v", values)
	}
}

// TestExcludeRangeFilterSetValue tests ExcludeRangeFilter SetValue functionality.
func TestExcludeRangeFilterSetValue(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	f.SetValue(&filter.RangeValue{Min: 30, Max: 150})
	values := f.GetValue()
	if values.Min != 30 || values.Max != 150 {
		t.Errorf("Expected RangeValue{Min: 30, Max: 150}, got %v", values)
	}
}

// TestExcludeRangeFilterGetFieldName tests ExcludeRangeFilter GetFieldName functionality.
func TestExcludeRangeFilterGetFieldName(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	if f.GetFieldName() != "price" {
		t.Errorf("Expected field name 'price', got %s", f.GetFieldName())
	}
}

// TestExcludeRangeFilterGetMin tests ExcludeRangeFilter GetMin functionality.
func TestExcludeRangeFilterGetMin(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 25.5, Max: 100})

	if f.GetMin() != 25.5 {
		t.Errorf("Expected min 25.5, got %v", f.GetMin())
	}
}

// TestExcludeRangeFilterGetMax tests ExcludeRangeFilter GetMax functionality.
func TestExcludeRangeFilterGetMax(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 10, Max: 99.9})

	if f.GetMax() != 99.9 {
		t.Errorf("Expected max 99.9, got %v", f.GetMax())
	}
}

// TestExcludeRangeFilterFilterInput tests ExcludeRangeFilter FilterInput functionality.
func TestExcludeRangeFilterFilterInput(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	facetedData := map[string][]int{
		"5":   {1, 2},
		"15":  {3, 4},
		"50":  {5, 6, 7},
		"99":  {8, 9},
		"101": {11, 12},
	}

	inputRecords := map[int]bool{1: true, 3: true, 5: true, 8: true, 11: true}
	excludeRecords := map[int]bool{}

	f.FilterInput(facetedData, inputRecords, excludeRecords)

	// FilterInput for exclude filters doesn't modify inputRecords directly
	// The exclude records are added via AddExcluded
	if len(inputRecords) != 5 {
		t.Errorf("Expected 5 records (unchanged), got %d", len(inputRecords))
	}
}

// TestExcludeRangeFilterAddExcluded tests ExcludeRangeFilter AddExcluded functionality.
func TestExcludeRangeFilterAddExcluded(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 10, Max: 100})

	facetedData := map[string][]int{
		"5":   {1, 2},
		"15":  {3, 4},
		"50":  {5, 6, 7},
		"99":  {8, 9},
		"101": {10, 11},
	}

	excludeRecords := make(map[int]bool)

	f.AddExcluded(facetedData, &excludeRecords)

	// Records in range [10, 100] should be excluded: 3,4,5,6,7,8,9
	expectedExcluded := 7 // records 3,4,5,6,7,8,9
	if len(excludeRecords) != expectedExcluded {
		t.Errorf("Expected %d excluded records, got %d", expectedExcluded, len(excludeRecords))
	}

	if !excludeRecords[3] {
		t.Errorf("Expected record 3 to be excluded")
	}
	if !excludeRecords[9] {
		t.Errorf("Expected record 9 to be excluded")
	}
	// Records outside range should not be excluded
	if excludeRecords[1] {
		t.Errorf("Expected record 1 to NOT be excluded")
	}
	if excludeRecords[10] {
		t.Errorf("Expected record 10 to NOT be excluded")
	}
}

// TestExcludeRangeFilterAddExcludedWithZeroValue tests ExcludeRangeFilter AddExcluded with zero values.
func TestExcludeRangeFilterAddExcludedWithZeroValue(t *testing.T) {
	f := filter.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 0, Max: 0})

	excludeRecords := make(map[int]bool)
	f.AddExcluded(map[string][]int{}, &excludeRecords)

	// With zero values, AddExcluded should return without adding any records
	if len(excludeRecords) != 0 {
		t.Errorf("Expected 0 excluded records, got %d", len(excludeRecords))
	}
}
