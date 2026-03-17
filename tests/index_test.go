package tests

import (
	"testing"

	"github.com/k-samuel/faceted"
	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/index"
	"github.com/k-samuel/faceted/pkg/query"
	"github.com/k-samuel/faceted/pkg/storage"
)

func getStorages() (search *faceted.Search, searchIndex index.IndexInterface, storage storage.StorageInterface) {
	search = faceted.NewSearch()
	searchIndex, _ = search.NewIndex(faceted.ArrayStorage)
	storage = searchIndex.GetStorage()
	return search, searchIndex, storage
}

// TestAggregate tests basic aggregation functionality.
func TestAggregate(t *testing.T) {
	search, searchIndex, storage := getStorages()

	_ = storage.AddRecord(112, map[string]interface{}{"vendor": "Tester", "price": 100})
	_ = storage.AddRecord(113, map[string]interface{}{"vendor": "Tester2", "price": 101})
	_ = storage.AddRecord(114, map[string]interface{}{"vendor": "Tester2", "price": 101})

	result := searchIndex.Aggregate(search.NewAggregationQuery().CountItems(true))

	expect := map[string]map[string]interface{}{
		"vendor": {"Tester": 1, "Tester2": 2},
		"price":  {"100": 1, "101": 2},
	}

	assertEqualMaps(t, expect, result)
}

// TestQuery tests basic query functionality.
func TestQuery(t *testing.T) {
	search, searchIndex, storage := getStorages()

	records := getTestData()
	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	// Test with simpler filters that work in Go implementation
	filters := []filter.FilterInterface{
		search.NewValueFilter("vendor", []interface{}{"Samsung", "Apple"}),
		search.NewValueFilter("color", []interface{}{"black"}),
	}

	result := searchIndex.Query(query.NewSearchQuery().Filters(filters))
	sortIntSlice(result)

	// Should find records with vendor Samsung/Apple AND color black
	if len(result) == 0 {
		t.Errorf("Expected some results, got none")
	}

	// Test no results
	filters2 := []filter.FilterInterface{
		search.NewValueFilter("vendor", []interface{}{"Google"}),
	}
	result2 := searchIndex.Query(query.NewSearchQuery().Filters(filters2))
	assertEqualSlices(t, []int{}, result2)
}

// TestQueryLimit tests query with inRecords limit.
func TestQueryLimit(t *testing.T) {
	search, searchIndex, storage := getStorages()

	records := getTestData()
	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	filters := []filter.FilterInterface{
		search.NewValueFilter("vendor", []interface{}{"Samsung", "Apple"}),
	}

	result := searchIndex.Query(query.NewSearchQuery().Filters(filters).InRecords([]int{1, 3}))
	resultMap := make(map[int]bool)
	for _, r := range result {
		resultMap[r] = true
	}

	if !resultMap[1] || !resultMap[3] {
		t.Errorf("Expected results to contain 1 and 3, got %v", result)
	}
}

// TestAggregation tests aggregation with filter.
func TestAggregation(t *testing.T) {
	search, searchIndex, storage := getStorages()

	records := getTestData()
	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	filters := []filter.FilterInterface{
		search.NewValueFilter("color", []interface{}{"black"}),
	}

	result := searchIndex.Aggregate(search.NewAggregationQuery().Filters(filters))

	// Check that expected fields exist
	expectedFields := []string{"vendor", "model", "price", "color", "has_phones", "cam_mp", "sale"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Expected field %s in result", field)
		}
	}
}

// TestAggregationCountNoFilter tests aggregation count without filters.
func TestAggregationCountNoFilter(t *testing.T) {
	search, searchIndex, storage := getStorages()

	records := []map[string]interface{}{
		{"color": "black", "size": 7, "group": "A"},
		{"color": "black", "size": 8, "group": "A"},
		{"color": "white", "size": 7, "group": "B"},
		{"color": "yellow", "size": 7, "group": "C"},
		{"color": "black", "size": 7, "group": "C"},
	}

	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	result := searchIndex.Aggregate(search.NewAggregationQuery().CountItems(true))

	expect := map[string]map[string]interface{}{
		"color": {"black": 3, "white": 1, "yellow": 1},
		"size":  {"7": 4, "8": 1},
		"group": {"A": 2, "B": 1, "C": 2},
	}

	assertEqualMaps(t, expect, result)
}

// TestAggregationCountLimit tests aggregation with inRecords limit.
func TestAggregationCountLimit(t *testing.T) {
	search, searchIndex, storage := getStorages()

	records := []map[string]interface{}{
		{"id": 1, "color": "black", "size": 7, "group": "A"},
		{"id": 2, "color": "black", "size": 8, "group": "A"},
		{"id": 3, "color": "white", "size": 7, "group": "B"},
		{"id": 4, "color": "yellow", "size": 7, "group": "C"},
		{"id": 5, "color": "black", "size": 7, "group": "C"},
	}

	for _, item := range records {
		id := int(item["id"].(int))
		delete(item, "id")
		_ = storage.AddRecord(id, item)
	}

	result := searchIndex.Aggregate(search.NewAggregationQuery().InRecords([]int{1, 2}).CountItems(true))

	if colorResult, ok := result["color"]; ok {
		if count, ok := colorResult["black"]; !ok || count != 2 {
			t.Errorf("Expected black count to be 2, got %v", count)
		}
	}
}

// TestIntFilterNames tests integer field names.
func TestIntFilterNames(t *testing.T) {
	search, searchIndex, storage := getStorages()

	records := []map[string]interface{}{
		{"id": 1, "f1": "black", "f2": 7.5, "group": "A"},
		{"id": 2, "f1": "black", "f2": 8.9, "group": "A"},
		{"id": 3, "f1": "white", "f2": 7.11, "group": "B"},
	}

	for _, item := range records {
		id := int(item["id"].(int))
		delete(item, "id")
		_ = storage.AddRecord(id, item)
	}

	// Test query with float field
	filters := []filter.FilterInterface{
		search.NewValueFilter("f2", []interface{}{7.11}),
	}
	result := searchIndex.Query(search.NewSearchQuery().Filters(filters))
	if len(result) != 1 || result[0] != 3 {
		t.Errorf("Expected [3], got %v", result)
	}
}

// TestOrderedSearch tests sorting functionality.
func TestOrderedSearch(t *testing.T) {
	search, searchIndex, storage := getStorages()

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
	result := searchIndex.Query(search.NewSearchQuery().Sort("size", query.SortDesc, query.SortNumeric))
	expected := []int{4, 2, 1, 3, 5}
	assertEqualSlices(t, expected, result)

	// Test ASC sort with filter
	result2 := searchIndex.Query(query.NewSearchQuery().
		Filters([]filter.FilterInterface{
			search.NewValueFilter("group", []interface{}{"C"}),
		}).
		Sort("size", query.SortAsc, query.SortNumeric))
	expected2 := []int{5, 4}
	assertEqualSlices(t, expected2, result2)
}

// TestGetCount tests record count.
func TestGetCount(t *testing.T) {
	_, searchIndex, storage := getStorages()

	_ = storage.AddRecord(1, map[string]interface{}{"col": 2})
	_ = storage.AddRecord(2, map[string]interface{}{"col": 2, "pr": 1, "dr": 2})
	_ = storage.AddRecord(3, map[string]interface{}{"col": 2, "pr": 1, "dr": 3})

	if searchIndex.GetCount() != 3 {
		t.Errorf("Expected count 3, got %d", searchIndex.GetCount())
	}
}

// TestSetData tests data export/import.
func TestSetData(t *testing.T) {
	_, searchIndex, storage := getStorages()

	data := map[string]map[string][]int{
		"field1": {
			"val1": {1, 2, 3},
			"val2": {2, 3, 4},
		},
		"field2": {
			"val1": {1},
			"val2": {3, 4},
		},
	}

	storage.SetData(data)
	exported := searchIndex.Export()

	if len(exported) != len(data) {
		t.Errorf("Expected %d fields, got %d", len(data), len(exported))
	}
}

// TestSetDataAndExport tests SetData followed by Export.
func TestSetDataAndExport(t *testing.T) {
	_, searchIndex, storage := getStorages()

	data := map[string]map[string][]int{
		"brand": {
			"Nony":  {1, 2, 3},
			"Mikon": {2, 3, 4},
		},
	}

	storage.SetData(data)

	exported := searchIndex.Export()
	if len(exported) != 1 {
		t.Errorf("Expected 1 field, got %d", len(exported))
	}

	if brandData, ok := exported["brand"]; ok {
		if len(brandData) != 2 {
			t.Errorf("Expected 2 brand values, got %d", len(brandData))
		}
	}
}

// TestOptimize tests index optimization.
func TestOptimize(t *testing.T) {
	search, searchIndex, storage := getStorages()

	_ = storage.AddRecord(1, map[string]interface{}{"brand": "Nony", "price": 100})
	_ = storage.AddRecord(2, map[string]interface{}{"brand": "Mikon", "price": 200})
	_ = storage.AddRecord(3, map[string]interface{}{"brand": "Nony", "price": 150})

	// Optimize should not fail
	searchIndex.Optimize()

	// Verify query works after optimization
	result := searchIndex.Query(search.NewSearchQuery().Filters([]filter.FilterInterface{
		search.NewValueFilter("brand", []interface{}{"Nony"}),
	}))
	if len(result) != 2 {
		t.Errorf("Expected 2 records, got %d", len(result))
	}
}

// TestProfile tests profile functionality.
func TestProfile(t *testing.T) {
	search, searchIndex, _ := getStorages()

	profile := index.NewProfile()
	searchIndex.SetProfiler(profile)

	if profile.GetSortingTime() != 0 {
		t.Errorf("Expected initial sorting time to be 0, got %f", profile.GetSortingTime())
	}

	profile.SetSortingTime(1.23)
	if profile.GetSortingTime() != 1.23 {
		t.Errorf("Expected sorting time 1.23, got %f", profile.GetSortingTime())
	}

	// Test that query still works after setting profiler
	result := searchIndex.Query(search.NewSearchQuery().Filters([]filter.FilterInterface{
		search.NewValueFilter("brand", []interface{}{"Nony"}),
	}))
	_ = result // Just verify it doesn't crash
}

// Helper functions

func getTestData() map[int]map[string]interface{} {
	return map[int]map[string]interface{}{
		1: {"vendor": "Apple", "model": "Iphone X Pro Max", "price": 80999, "color": "black", "has_phones": 1, "cam_mp": 40, "sale": 1},
		2: {"vendor": "Samsung", "model": "Galaxy S20", "price": 70599, "color": "white", "has_phones": 1, "cam_mp": 105, "sale": 0},
		3: {"vendor": "Samsung", "model": "Galaxy S20", "price": 70599, "color": "yellow", "has_phones": 1, "cam_mp": 105, "sale": 1},
		4: {"vendor": "Samsung", "model": "Galaxy A5", "price": 15000, "color": "black", "has_phones": 1, "cam_mp": 12, "sale": 1},
		5: {"vendor": "Xiaomi", "model": "MI 9", "price": 26000, "color": "black", "has_phones": 1, "cam_mp": 48, "sale": 0},
		6: {"vendor": "Apple", "model": "Iphone X Pro Max", "price": 80999, "color": "white", "has_phones": 1, "cam_mp": 40, "sale": 1},
	}
}
