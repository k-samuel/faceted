package search

import (
	"testing"

	"github.com/k-samuel/faceted/search/indexer"
)

// sortStringSlice sorts a slice of strings in-place.
func sortStringSlice(a []string) {
	for i := 0; i < len(a)-1; i++ {
		for j := i + 1; j < len(a); j++ {
			if a[i] > a[j] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}

// equalStringSlices checks if two string slices have the same elements.
func equalStringSlices(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i := range expected {
		if expected[i] != actual[i] {
			return false
		}
	}
	return true
}

// TestAggregationSort tests AggregationSort functionality.
func TestAggregationSort(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	records := []map[string]interface{}{
		{"size": 7, "color": "yellow", "group": "C"},
		{"color": "black", "size": 7, "group": "C"},
		{"color": "black", "size": 7, "group": "A"},
		{"color": "black", "size": 8, "group": "A"},
		{"color": "white", "size": 7, "group": "B"},
	}

	for id, item := range records {
		id := int(id) + 1
		delete(item, "id")
		_ = storage.AddRecord(id, item)
	}

	result, _ := db.Aggregate(NewAggregationQuery().CountItems(true).Sort(SortAsc, SortAsc))

	expectedKeys := []string{"color", "group", "size"}
	actualKeys := make([]string, 0, len(result))
	for _, field := range result {
		actualKeys = append(actualKeys, field.Field)
	}
	sortStringSlice(actualKeys)

	if !equalStringSlices(expectedKeys, actualKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, actualKeys)
	}

	// Check DESC sort
	resultDesc, _ := db.Aggregate(NewAggregationQuery().CountItems(true).Sort(SortDesc, SortDesc))

	actualKeysDesc := make([]string, 0, len(resultDesc))
	for _, field := range resultDesc {
		actualKeysDesc = append(actualKeysDesc, field.Field)
	}
	sortStringSlice(actualKeysDesc)

	if !equalStringSlices(expectedKeys, actualKeysDesc) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, actualKeysDesc)
	}
}

// TestOrder tests Order functionality.
func TestOrder(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	records := []map[string]interface{}{
		{"id": 1, "color": "black", "size": 7.5, "group": "A"},
		{"id": 2, "color": "black", "size": 8.9, "group": "A"},
		{"id": 3, "color": "white", "size": 7.11, "group": "B"},
		{"id": 4, "color": "white", "size": 9, "group": "C"},
		{"id": 5, "color": "white", "size": 3, "group": "C"},
	}

	for _, item := range records {
		id := int(item["id"].(int))
		delete(item, "id")
		_ = storage.AddRecord(id, item)
	}

	// Test DESC sort
	resultDesc, _ := db.Query(NewSearchQuery().Sort("size", SortDesc, SortTypeNumbers))
	expectedDesc := []int{4, 2, 1, 3, 5}
	assertEqualSlices(t, expectedDesc, resultDesc)

	// Test ASC sort with filter
	resultAsc, _ := db.Query(NewSearchQuery().
		Filters([]FilterInterface{
			NewValueFilter("group", []interface{}{"C"}),
		}).
		Sort("size", SortAsc, SortTypeNumbers))
	expectedAsc := []int{5, 4}
	assertEqualSlices(t, expectedAsc, resultAsc)
}

// TestSortRangeIndexer tests sorting with RangeIndexer.
func TestSortRangeIndexer(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	rangeIndexer, _ := indexer.NewRangeIndexer(100)
	storage.AddIndexer("price", rangeIndexer)

	_ = storage.AddRecord(1, map[string]interface{}{"price": 50})
	_ = storage.AddRecord(2, map[string]interface{}{"price": 107})
	_ = storage.AddRecord(3, map[string]interface{}{"price": 103})
	_ = storage.AddRecord(4, map[string]interface{}{"price": 112})
	_ = storage.AddRecord(5, map[string]interface{}{"price": 210})

	storage.Optimize()

	filters := []FilterInterface{
		NewValueFilter("price", "100"),
	}

	// Test ASC sort
	resultAsc, _ := db.Query(NewSearchQuery().Filters(filters).Sort("price", SortAsc, SortTypeNumbers))
	expectedAsc := []int{3, 2, 4}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc, _ := db.Query(NewSearchQuery().
		Filters(filters).
		Sort("price", SortDesc, SortTypeNumbers))
	expectedDesc := []int{4, 2, 3}
	assertEqualSlices(t, expectedDesc, resultDesc)
}

// TestSortRangeListIndexer tests sorting with RangeListIndexer.
func TestSortRangeListIndexer(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	rangeIndexer, _ := indexer.NewRangeListIndexer([]int{0, 100, 200})
	storage.AddIndexer("price", rangeIndexer)

	_ = storage.AddRecord(1, map[string]interface{}{"price": 50})
	_ = storage.AddRecord(2, map[string]interface{}{"price": 107})
	_ = storage.AddRecord(3, map[string]interface{}{"price": 103})
	_ = storage.AddRecord(4, map[string]interface{}{"price": 112})
	_ = storage.AddRecord(5, map[string]interface{}{"price": 210})

	storage.Optimize()

	filters := []FilterInterface{
		NewValueFilter("price", "100"),
	}

	// Test ASC sort
	resultAsc, _ := db.Query(NewSearchQuery().
		Filters(filters).
		Sort("price", SortAsc, SortTypeNumbers))
	expectedAsc := []int{3, 2, 4}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc, _ := db.Query(NewSearchQuery().
		Filters(filters).
		Sort("price", SortDesc, SortTypeNumbers))
	expectedDesc := []int{4, 2, 3}
	assertEqualSlices(t, expectedDesc, resultDesc)
}

// TestSortRange tests basic sorting with numeric values.
func TestSortRange(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	records := []map[string]interface{}{
		{"id": 1, "size": 7.5, "color": "black"},
		{"id": 2, "size": 8.9, "color": "black"},
		{"id": 3, "size": 7.11, "color": "white"},
	}

	for _, item := range records {
		id := int(item["id"].(int))
		delete(item, "id")
		_ = storage.AddRecord(id, item)
	}

	// Test ASC sort
	resultAsc, _ := db.Query(NewSearchQuery().Sort("size", SortAsc, SortTypeNumbers))
	expectedAsc := []int{3, 1, 2}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc, _ := db.Query(NewSearchQuery().Sort("size", SortDesc, SortTypeNumbers))
	expectedDesc := []int{2, 1, 3}
	assertEqualSlices(t, expectedDesc, resultDesc)
}

// TestSortRangeList tests sorting with range values.
func TestSortRangeList(t *testing.T) {
	provider := NewContainer()
	db := provider.NewDb()
	storage := db.GetStorage()

	rangeIndexer, _ := indexer.NewRangeListIndexer([]int{0, 10, 20})
	storage.AddIndexer("size", rangeIndexer)

	_ = storage.AddRecord(1, map[string]interface{}{"size": 7.5, "color": "black"})
	_ = storage.AddRecord(2, map[string]interface{}{"size": 8.9, "color": "black"})
	_ = storage.AddRecord(3, map[string]interface{}{"size": 7.11, "color": "white"})

	storage.Optimize()

	filters := []FilterInterface{
		NewValueFilter("color", []interface{}{"black"}),
	}

	// Test ASC sort
	resultAsc, _ := db.Query(NewSearchQuery().
		Filters(filters).
		Sort("size", SortAsc, SortTypeNumbers))
	expectedAsc := []int{1, 2}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc, _ := db.Query(NewSearchQuery().
		Filters(filters).
		Sort("size", SortDesc, SortTypeNumbers))
	expectedDesc := []int{2, 1}
	assertEqualSlices(t, expectedDesc, resultDesc)
}
