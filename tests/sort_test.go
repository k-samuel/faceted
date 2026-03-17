package tests

import (
	"testing"

	"github.com/k-samuel/faceted"
	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/indexer"
	"github.com/k-samuel/faceted/pkg/query"
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
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

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

	result := searchIndex.Aggregate(search.NewAggregationQuery().CountItems(true).Sort(query.SortAsc, query.SortRegular))

	expectedKeys := []string{"color", "group", "size"}
	actualKeys := make([]string, 0, len(result))
	for k := range result {
		actualKeys = append(actualKeys, k)
	}
	sortStringSlice(actualKeys)

	if !equalStringSlices(expectedKeys, actualKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, actualKeys)
	}

	// Check DESC sort
	resultDesc := searchIndex.Aggregate(search.NewAggregationQuery().CountItems(true).Sort(query.SortDesc, query.SortRegular))

	actualKeysDesc := make([]string, 0, len(resultDesc))
	for k := range resultDesc {
		actualKeysDesc = append(actualKeysDesc, k)
	}
	sortStringSlice(actualKeysDesc)

	if !equalStringSlices(expectedKeys, actualKeysDesc) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, actualKeysDesc)
	}
}

// TestOrder tests Order functionality.
func TestOrder(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

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
	resultDesc := searchIndex.Query(search.NewSearchQuery().Sort("size", query.SortDesc, query.SortNumeric))
	expectedDesc := []int{4, 2, 1, 3, 5}
	assertEqualSlices(t, expectedDesc, resultDesc)

	// Test ASC sort with filter
	resultAsc := searchIndex.Query(query.NewSearchQuery().
		Filters([]filter.FilterInterface{
			search.NewValueFilter("group", []interface{}{"C"}),
		}).
		Sort("size", query.SortAsc, query.SortNumeric))
	expectedAsc := []int{5, 4}
	assertEqualSlices(t, expectedAsc, resultAsc)
}

// TestSortRangeIndexer tests sorting with RangeIndexer.
func TestSortRangeIndexer(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	rangeIndexer, _ := indexer.NewRangeIndexer(100)
	storage.AddIndexer("price", rangeIndexer)

	_ = storage.AddRecord(1, map[string]interface{}{"price": 50})
	_ = storage.AddRecord(2, map[string]interface{}{"price": 107})
	_ = storage.AddRecord(3, map[string]interface{}{"price": 103})
	_ = storage.AddRecord(4, map[string]interface{}{"price": 112})
	_ = storage.AddRecord(5, map[string]interface{}{"price": 210})

	storage.Optimize()

	filters := []filter.FilterInterface{
		search.NewValueFilter("price", "100"),
	}

	// Test ASC sort
	resultAsc := searchIndex.Query(search.NewSearchQuery().
		Filters(filters).
		Sort("price", query.SortAsc, query.SortNumeric))
	expectedAsc := []int{3, 2, 4}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc := searchIndex.Query(search.NewSearchQuery().
		Filters(filters).
		Sort("price", query.SortDesc, query.SortNumeric))
	expectedDesc := []int{4, 2, 3}
	assertEqualSlices(t, expectedDesc, resultDesc)
}

// TestSortRangeListIndexer tests sorting with RangeListIndexer.
func TestSortRangeListIndexer(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	rangeIndexer, _ := search.NewRangeListIndexer([]int{0, 100, 200})
	storage.AddIndexer("price", rangeIndexer)

	_ = storage.AddRecord(1, map[string]interface{}{"price": 50})
	_ = storage.AddRecord(2, map[string]interface{}{"price": 107})
	_ = storage.AddRecord(3, map[string]interface{}{"price": 103})
	_ = storage.AddRecord(4, map[string]interface{}{"price": 112})
	_ = storage.AddRecord(5, map[string]interface{}{"price": 210})

	storage.Optimize()

	filters := []filter.FilterInterface{
		search.NewValueFilter("price", "100"),
	}

	// Test ASC sort
	resultAsc := searchIndex.Query(query.NewSearchQuery().
		Filters(filters).
		Sort("price", query.SortAsc, query.SortNumeric))
	expectedAsc := []int{3, 2, 4}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc := searchIndex.Query(query.NewSearchQuery().
		Filters(filters).
		Sort("price", query.SortDesc, query.SortNumeric))
	expectedDesc := []int{4, 2, 3}
	assertEqualSlices(t, expectedDesc, resultDesc)
}

// TestSortRange tests basic sorting with numeric values.
func TestSortRange(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

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
	resultAsc := searchIndex.Query(search.NewSearchQuery().Sort("size", query.SortAsc, query.SortNumeric))
	expectedAsc := []int{3, 1, 2}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc := searchIndex.Query(query.NewSearchQuery().Sort("size", query.SortDesc, query.SortNumeric))
	expectedDesc := []int{2, 1, 3}
	assertEqualSlices(t, expectedDesc, resultDesc)
}

// TestSortRangeList tests sorting with range values.
func TestSortRangeList(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	rangeIndexer, _ := search.NewRangeListIndexer([]int{0, 10, 20})
	storage.AddIndexer("size", rangeIndexer)

	_ = storage.AddRecord(1, map[string]interface{}{"size": 7.5, "color": "black"})
	_ = storage.AddRecord(2, map[string]interface{}{"size": 8.9, "color": "black"})
	_ = storage.AddRecord(3, map[string]interface{}{"size": 7.11, "color": "white"})

	storage.Optimize()

	filters := []filter.FilterInterface{
		search.NewValueFilter("color", []interface{}{"black"}),
	}

	// Test ASC sort
	resultAsc := searchIndex.Query(query.NewSearchQuery().
		Filters(filters).
		Sort("size", query.SortAsc, query.SortNumeric))
	expectedAsc := []int{1, 2}
	assertEqualSlices(t, expectedAsc, resultAsc)

	// Test DESC sort
	resultDesc := searchIndex.Query(query.NewSearchQuery().
		Filters(filters).
		Sort("size", query.SortDesc, query.SortNumeric))
	expectedDesc := []int{2, 1}
	assertEqualSlices(t, expectedDesc, resultDesc)
}
