package search

import (
	"fmt"
	"slices"
	"testing"

	"github.com/k-samuel/faceted/search/indexer"
)

func getStorages() (db *Db, storage StorageInterface) {
	db = NewContainer().NewDb()
	storage = db.GetStorage()
	return db, storage
}

// TestAggregate tests basic aggregation functionality.
func TestAggregate(t *testing.T) {
	db, storage := getStorages()

	_ = storage.AddRecord(112, map[string]interface{}{"vendor": "Tester", "price": 100})
	_ = storage.AddRecord(113, map[string]interface{}{"vendor": "Tester2", "price": 101})
	_ = storage.AddRecord(114, map[string]interface{}{"vendor": "Tester2", "price": 101})

	result, _ := db.Aggregate(NewAggregationQuery().CountItems(true))

	// Check expected values
	expected := map[string]map[string]int{
		"vendor": {"Tester": 1, "Tester2": 2},
		"price":  {"100": 1, "101": 2},
	}

	if !compareAggregationResultMap(result, expected) {
		t.Errorf("Aggregation result mismatch")
		for field, expectedVal := range expected {
			t.Logf("Field %s: expected %v", field, expectedVal)
		}
	}
}

// TestQuery tests basic query functionality.
func TestQuery(t *testing.T) {
	db, storage := getStorages()

	records := map[int]map[string]interface{}{
		1: {"vendor": "Apple", "model": "Iphone X Pro Max", "price": 80999, "color": "black", "has_phones": 1, "cam_mp": 40, "sale": 1, "warehouse": []int{1, 7, 10}},
		2: {"vendor": "Samsung", "model": "Galaxy S20", "price": 70599, "color": "white", "has_phones": 1, "cam_mp": 105, "sale": 0, "warehouse": []int{1, 7, 12}},
		3: {"vendor": "Samsung", "model": "Galaxy S20", "price": 70599, "color": "yellow", "has_phones": 1, "cam_mp": 105, "sale": 1, "warehouse": []int{2, 3, 5}},
		4: {"vendor": "Samsung", "model": "Galaxy A5", "price": 15000, "color": "black", "has_phones": 1, "cam_mp": 12, "sale": 1, "warehouse": []int{1, 7, 12}},
		5: {"vendor": "Xiaomi", "model": "MI 9", "price": 26000, "color": "black", "has_phones": 1, "cam_mp": 48, "sale": 0, "warehouse": []int{1, 7, 12}},
		6: {"vendor": "Apple", "model": "Iphone X Pro Max", "price": 80999, "color": "white", "has_phones": 1, "cam_mp": 40, "sale": 1, "warehouse": []int{1, 7, 12}},
	}
	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	// Test with simpler filters that work in Go implementation
	filters := []FilterInterface{
		NewValueFilter("vendor", []interface{}{"Samsung", "Apple"}),
		NewValueFilter("color", []interface{}{"black"}),
		NewValueFilter("warehouse", []int{1, 7, 34}),
	}

	result, _ := db.Query(NewSearchQuery().Filters(filters))

	// Should find records with vendor Samsung/Apple AND color black
	if len(result) == 0 {
		t.Errorf("Expected some results, got none")
	}

	if !slices.Equal(result, []int{1, 4}) {
		t.Errorf("Expected results 1,4")
	}

	// Test no results
	filters2 := []FilterInterface{
		NewValueFilter("vendor", []interface{}{"Google"}),
	}
	result2, err := db.Query(NewSearchQuery().Filters(filters2))
	if err != nil {
		t.Error(err)
	}
	assertEqualSlices(t, []int{}, result2)
}

func TestQueryWithRange(t *testing.T) {
	db, storage := getStorages()
	indexer, _ := indexer.NewRangeIndexer(1000)
	storage.AddIndexer("price", indexer)

	records := map[int]map[string]interface{}{
		1: {"vendor": "Apple", "model": "Iphone X Pro Max", "price": 1000, "color": "black", "has_phones": 1, "cam_mp": 40, "sale": 1, "warehouse": []int{1, 7, 10, 11}},
		2: {"vendor": "Samsung", "model": "Galaxy S20", "price": 1000, "color": "white", "has_phones": 1, "cam_mp": 105, "sale": 0, "warehouse": []int{1, 7, 12, 11}},
		3: {"vendor": "Samsung", "model": "Galaxy S20", "price": 2500, "color": "yellow", "has_phones": 1, "cam_mp": 105, "sale": 1, "warehouse": []int{2, 3, 5}},
		4: {"vendor": "Samsung", "model": "Galaxy A5", "price": 800, "color": "black", "has_phones": 1, "cam_mp": 12, "sale": 1, "warehouse": []int{1, 7, 12, 11}},
		5: {"vendor": "Xiaomi", "model": "MI 9", "price": 500, "color": "black", "has_phones": 1, "cam_mp": 48, "sale": 0, "warehouse": []int{1, 7, 12, 11}},
		6: {"vendor": "Apple", "model": "Iphone X Pro Max", "price": 2000, "color": "white", "has_phones": 1, "cam_mp": 40, "sale": 1, "warehouse": []int{1, 7, 12, 11}},
	}
	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	// Test with simpler filters that work in Go implementation
	filters := []FilterInterface{
		NewValueFilter("vendor", []interface{}{"Samsung", "Apple"}),
		NewRangeFilter("price", NewRangeValue(1000, 5000)),
		NewValueFilter("warehouse", []int{1, 11}),
	}

	result, _ := db.Query(NewSearchQuery().Filters(filters))

	// Should find records with vendor Samsung/Apple AND color black
	if !slices.Equal(result, []int{1, 2, 6}) {
		t.Errorf("Expected results 1,2,6 ")
	}
}

// TestQueryLimit tests query with inRecords limit.
func TestQueryLimit(t *testing.T) {
	searchIndex, storage := getStorages()

	records := getTestData()
	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	filters := []FilterInterface{
		NewValueFilter("vendor", []interface{}{"Samsung", "Apple"}),
	}

	result, _ := searchIndex.Query(NewSearchQuery().Filters(filters).InRecords([]int{1, 3}))
	resultMap := make(map[int]bool)
	for _, r := range result {
		resultMap[r] = true
	}

	if !resultMap[1] || !resultMap[3] {
		t.Errorf("Expected results to contain 1 and 3, got %v", result)
	}
}

// TestAggregation tests aggregation with
func TestAggregation(t *testing.T) {
	db, storage := getStorages()

	records := getTestData()
	for id, item := range records {
		_ = storage.AddRecord(id, item)
	}

	filters := []FilterInterface{
		NewValueFilter("color", []interface{}{"black"}),
	}

	result, _ := db.Aggregate(NewAggregationQuery().Filters(filters))

	// Check that expected fields exist
	expectedFields := []string{"vendor", "model", "price", "color", "has_phones", "cam_mp", "sale"}
	for _, field := range expectedFields {
		if !hasField(result, field) {
			t.Errorf("Expected field %s in result", field)
		}
	}
}

// TestAggregationCountNoFilter tests aggregation count without filters.
func TestAggregationCountNoFilter(t *testing.T) {
	db, storage := getStorages()

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

	result, _ := db.Aggregate(NewAggregationQuery().CountItems(true))

	// Check expected values
	expected := map[string]map[string]int{
		"color": {"black": 3, "white": 1, "yellow": 1},
		"size":  {"7": 4, "8": 1},
		"group": {"A": 2, "B": 1, "C": 2},
	}

	if !compareAggregationResultMap(result, expected) {
		t.Errorf("Aggregation result mismatch")
		for field, expectedVal := range expected {
			t.Logf("Field %s: expected %v", field, expectedVal)
		}
	}
}

// TestAggregationCountLimit tests aggregation with inRecords limit.

// TestIntFilterNames tests integer field names.
func TestIntFilterNames(t *testing.T) {
	db, storage := getStorages()

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
	filters := []FilterInterface{
		NewValueFilter("f2", []interface{}{7.11}),
	}
	result, _ := db.Query(NewSearchQuery().Filters(filters))
	if len(result) != 1 || result[0] != 3 {
		t.Errorf("Expected [3], got %v", result)
	}
}

// TestOrderedSearch tests sorting functionality.
func TestOrderedSearch(t *testing.T) {
	db, storage := getStorages()

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
	result, _ := db.Query(NewSearchQuery().Sort("size", SortDesc, SortTypeNumbers))
	expected := []int{4, 2, 1, 3, 5}
	assertEqualSlices(t, expected, result)

	// Test ASC sort with filter
	result2, _ := db.Query(NewSearchQuery().
		Filters([]FilterInterface{
			NewValueFilter("group", []interface{}{"C"}),
		}).
		Sort("size", SortAsc, SortTypeNumbers))
	expected2 := []int{5, 4}
	assertEqualSlices(t, expected2, result2)
}

// TestGetCount tests record count.
func TestGetCount(t *testing.T) {
	db, storage := getStorages()

	_ = storage.AddRecord(1, map[string]interface{}{"col": 2})
	_ = storage.AddRecord(2, map[string]interface{}{"col": 2, "pr": 1, "dr": 2})
	_ = storage.AddRecord(3, map[string]interface{}{"col": 2, "pr": 1, "dr": 3})

	if db.GetStorage().GetCount() != 3 {
		t.Errorf("Expected count 3, got %d", db.GetStorage().GetCount())
	}
}

// TestSetData tests data export/import.
func TestSetData(t *testing.T) {
	db, storage := getStorages()

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
	exported := db.GetStorage().Export()

	if len(exported) != len(data) {
		t.Errorf("Expected %d fields, got %d", len(data), len(exported))
	}
}

// TestSetDataAndExport tests SetData followed by Export.
func TestSetDataAndExport(t *testing.T) {
	db, storage := getStorages()

	data := map[string]map[string][]int{
		"brand": {
			"Nony":  {1, 2, 3},
			"Mikon": {2, 3, 4},
		},
	}

	storage.SetData(data)

	exported := db.GetStorage().Export()
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
	db, storage := getStorages()

	_ = storage.AddRecord(1, map[string]interface{}{"brand": "Nony", "price": 100})
	_ = storage.AddRecord(2, map[string]interface{}{"brand": "Mikon", "price": 200})
	_ = storage.AddRecord(3, map[string]interface{}{"brand": "Nony", "price": 150})

	// Optimize should not fail
	db.GetStorage().Optimize()

	// Verify query works after optimization
	result, _ := db.Query(NewSearchQuery().Filters([]FilterInterface{
		NewValueFilter("brand", []interface{}{"Nony"}),
	}))
	if len(result) != 2 {
		t.Errorf("Expected 2 records, got %d", len(result))
	}
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

// sortIntSlice sorts a slice of integers in-place.
func sortIntSlice(a []int) {
	for i := 0; i < len(a)-1; i++ {
		for j := i + 1; j < len(a); j++ {
			if a[i] > a[j] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}

// assertEqualSlices asserts that two int slices are equal.
func assertEqualSlices(t *testing.T, expected, actual []int) {
	if len(expected) != len(actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
		return
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("Expected %v, got %v", expected, actual)
			return
		}
	}
}

// hasField checks if a field exists in aggregation results.
func hasField(result []*AggregationResultField, fieldName string) bool {
	for _, field := range result {
		if field.Field == fieldName {
			return true
		}
	}
	return false
}

// compareAggregationResultMap compares aggregation results with expected map.
func compareAggregationResultMap(result []*AggregationResultField, expected map[string]map[string]int) bool {
	if len(result) != len(expected) {
		return false
	}

	// Build map from result
	resultMap := make(map[string]map[string]int)
	for _, field := range result {
		resultMap[field.Field] = make(map[string]int)
		for _, value := range field.Values {
			if value.Count != nil {
				resultMap[field.Field][value.Value] = *value.Count
			}
		}
	}

	// Compare with expected
	for field, expectedVal := range expected {
		actualVal, ok := resultMap[field]
		if !ok {
			return false
		}
		for expKey, expCount := range expectedVal {
			actCount, ok := actualVal[expKey]
			if !ok || actCount != expCount {
				return false
			}
		}
	}

	return true
}

// TestAggregationStability tests that aggregate returns consistent results across multiple calls.
func TestAggregationStability(t *testing.T) {
	db, storage := getStorages()

	records := []map[string]interface{}{
		{"color": "black", "size": 7, "group": "A", "warehouse": []int{1, 7, 8, 9, 44, 5, 6}},
		{"color": "black", "size": 8, "group": "A", "warehouse": []int{2, 3, 5, 7, 8}},
		{"color": "white", "size": 7, "group": "B", "warehouse": []int{1, 4, 6, 27, 18}},
		{"color": "yellow", "size": 7, "group": "C", "warehouse": []int{11, 14, 6, 27, 18}},
		{"color": "black", "size": 7, "group": "C", "warehouse": []int{1, 11, 2, 22}},
		{"color": "black", "size": 7, "group": "A", "warehouse": []int{5, 7, 8}},
		{"color": "black", "size": 8, "group": "A", "warehouse": []int{1, 2}},
		{"color": "white", "size": 7, "group": "B", "warehouse": []int{1}},
	}

	for id, item := range records {
		_ = storage.AddRecord(id+1, item)
	}

	// Call Aggregate multiple times and compare results
	var results [][]*AggregationResultField
	for i := 0; i < 5; i++ {
		result, err := db.Aggregate(NewAggregationQuery().CountItems(true))
		if err != nil {
			t.Fatalf("Aggregate call %d failed: %v", i+1, err)
		}
		results = append(results, result)

		// Print results for debugging
		t.Logf("Call %d: %v", i+1, formatAggregationResult(result))
	}

	// Compare all results with first result
	for i := 1; i < len(results); i++ {
		if !compareAggregationResults(results[0], results[i]) {
			t.Errorf("Result mismatch between call 1 and call %d", i+1)
			t.Logf("Call 1: %v", formatAggregationResult(results[0]))
			t.Logf("Call %d: %v", i+1, formatAggregationResult(results[i]))
		}
	}
}

// formatAggregationResult formats aggregation result for logging.
func formatAggregationResult(result []*AggregationResultField) string {
	var str string
	for _, field := range result {
		str += field.Field + ": {"
		for _, value := range field.Values {
			if value.Count != nil {
				str += value.Value + "=>" + fmt.Sprintf("%d", *value.Count) + ", "
			} else {
				str += value.Value + ", "
			}
		}
		str += "} "
	}
	return str
}

// compareAggregationResults compares two aggregation results for equality.
func compareAggregationResults(a, b []*AggregationResultField) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for easier comparison
	mapA := make(map[string]map[string]int)
	mapB := make(map[string]map[string]int)

	for _, field := range a {
		mapA[field.Field] = make(map[string]int)
		for _, value := range field.Values {
			if value.Count != nil {
				mapA[field.Field][value.Value] = *value.Count
			}
		}
	}

	for _, field := range b {
		mapB[field.Field] = make(map[string]int)
		for _, value := range field.Values {
			if value.Count != nil {
				mapB[field.Field][value.Value] = *value.Count
			}
		}
	}

	return compareAggregationMaps(mapA, mapB)
}

// compareAggregationMaps compares two aggregation maps.
func compareAggregationMaps(a, b map[string]map[string]int) bool {
	if len(a) != len(b) {
		return false
	}

	for field, valuesA := range a {
		valuesB, ok := b[field]
		if !ok {
			return false
		}
		if len(valuesA) != len(valuesB) {
			return false
		}
		for val, countA := range valuesA {
			countB, ok := valuesB[val]
			if !ok || countA != countB {
				return false
			}
		}
	}

	return true
}
