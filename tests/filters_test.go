package tests

import (
	"testing"

	"github.com/k-samuel/faceted"
	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/index"
	"github.com/k-samuel/faceted/pkg/indexer"
	"github.com/k-samuel/faceted/pkg/query"
	"github.com/k-samuel/faceted/pkg/value"
)

// TestValueIntersectionFilter tests ValueIntersectionFilter functionality.
func TestValueIntersectionFilter(t *testing.T) {
	getIndex := func() (search *faceted.Search, searchIndex index.IndexInterface) {
		search = faceted.NewSearch()
		searchIndex, _ = search.NewIndex(faceted.ArrayStorage)
		storage := searchIndex.GetStorage()

		data := []map[string]interface{}{
			{"id": 1, "brand": "Nony", "first_usage": []interface{}{"weddings", "wildlife"}, "second_usage": []interface{}{"wildlife", "portraits"}},
			{"id": 2, "brand": "Mikon", "first_usage": []interface{}{"weddings", "streetphoto"}, "second_usage": []interface{}{"wildlife", "streetphoto"}},
			{"id": 3, "brand": "Common", "first_usage": []interface{}{"streetphoto", "portraits"}, "second_usage": []interface{}{"streetphoto", "portraits"}},
			{"id": 4, "brand": "Digma", "first_usage": []interface{}{"streetphoto", "portraits", "weddings"}, "second_usage": []interface{}{"streetphoto", "portraits"}},
			{"id": 5, "brand": "Digma", "first_usage": []interface{}{"streetphoto"}, "second_usage": []interface{}{"portraits"}},
			{"id": 6, "brand": "Mikon", "first_usage": []interface{}{"weddings", "wildlife"}, "second_usage": []interface{}{"wildlife", "portraits"}},
		}

		for _, item := range data {
			id := int(item["id"].(int))
			delete(item, "id")
			_ = storage.AddRecord(id, item)
		}
		storage.Optimize()
		return search, searchIndex
	}

	// Test Query
	t.Run("Query", func(t *testing.T) {
		search, searchIndex := getIndex()

		// Query 1
		filters1 := []filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Nony", "Digma", "Mikon", "Common"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "weddings"}),
		}
		result1 := searchIndex.Query(query.NewSearchQuery().Filters(filters1))
		sortIntSlice(result1)
		assertEqualSlices(t, []int{2, 4}, result1)

		// Query 2
		filters2 := []filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Mikon", "Digma"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "weddings"}),
			search.NewValueIntersectionFilter("second_usage", []interface{}{"streetphoto", "portraits"}),
		}
		result2 := searchIndex.Query(query.NewSearchQuery().Filters(filters2))
		assertEqualSlices(t, []int{4}, result2)

		// Query 3
		filters3 := []filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Nony", "Digma", "Mikon", "Common"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "weddings"}),
			search.NewExcludeValueFilter("brand", []interface{}{"Digma"}),
		}
		result3 := searchIndex.Query(query.NewSearchQuery().Filters(filters3))
		assertEqualSlices(t, []int{2}, result3)

		// Query 4 with inRecords
		filters4 := []filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Nony", "Digma", "Mikon", "Common"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "weddings"}),
		}
		result4 := searchIndex.Query(query.NewSearchQuery().Filters(filters4).InRecords([]int{1, 3, 4}))
		assertEqualSlices(t, []int{4}, result4)
	})

	// Test Aggregate
	t.Run("Aggregate", func(t *testing.T) {
		search, searchIndex := getIndex()

		query1 := search.NewAggregationQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Mikon", "Digma"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "weddings"}),
			search.NewValueIntersectionFilter("second_usage", []interface{}{"streetphoto", "portraits"}),
		}).CountItems(true).Sort(query.SortAsc, query.SortRegular)

		result1 := searchIndex.Aggregate(query1)

		// Check brand field
		if brandResult, ok := result1["brand"]; ok {
			if count, ok := brandResult["Digma"]; !ok || count != 1 {
				t.Errorf("Expected Digma count to be 1, got %v", count)
			}
		} else {
			t.Error("Expected brand field in result")
		}

		// Check first_usage field
		if firstUsageResult, ok := result1["first_usage"]; ok {
			expectedFirstUsage := map[string]int{"streetphoto": 1, "weddings": 1, "portraits": 1}
			for k, v := range expectedFirstUsage {
				if count, ok := firstUsageResult[k]; !ok || count != v {
					t.Errorf("Expected first_usage[%s] to be %d, got %v", k, v, count)
				}
			}
		} else {
			t.Error("Expected first_usage field in result")
		}

		// Check second_usage field
		if secondUsageResult, ok := result1["second_usage"]; ok {
			expectedSecondUsage := map[string]int{"streetphoto": 2, "wildlife": 1, "portraits": 1}
			for k, v := range expectedSecondUsage {
				if count, ok := secondUsageResult[k]; !ok || count != v {
					t.Errorf("Expected second_usage[%s] to be %d, got %v", k, v, count)
				}
			}
		} else {
			t.Error("Expected second_usage field in result")
		}

		// Test Aggregate with selfFiltering
		query2 := search.NewAggregationQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Mikon", "Digma"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "weddings"}),
			search.NewValueIntersectionFilter("second_usage", []interface{}{"streetphoto", "portraits"}),
		}).CountItems(true).Sort(query.SortAsc, query.SortRegular).SelfFiltering(true)

		result2 := searchIndex.Aggregate(query2)

		// Check brand field
		if brandResult, ok := result2["brand"]; ok {
			if count, ok := brandResult["Digma"]; !ok || count != 1 {
				t.Errorf("Expected Digma count to be 1 in self-filtering, got %v", count)
			}
		} else {
			t.Error("Expected brand field in result2")
		}

		// Check first_usage field
		if firstUsageResult, ok := result2["first_usage"]; ok {
			expectedFirstUsage := map[string]int{"streetphoto": 1, "weddings": 1, "portraits": 1}
			for k, v := range expectedFirstUsage {
				if count, ok := firstUsageResult[k]; !ok || count != v {
					t.Errorf("Expected first_usage[%s] to be %d in self-filtering, got %v", k, v, count)
				}
			}
		} else {
			t.Error("Expected first_usage field in result2")
		}

		// Check second_usage field (should exclude wildlife due to self-filtering)
		if secondUsageResult, ok := result2["second_usage"]; ok {
			expectedSecondUsage := map[string]int{"streetphoto": 1, "portraits": 1}
			for k, v := range expectedSecondUsage {
				if count, ok := secondUsageResult[k]; !ok || count != v {
					t.Errorf("Expected second_usage[%s] to be %d in self-filtering, got %v", k, v, count)
				}
			}
			// Wildlife should not be present due to self-filtering
			if _, ok := secondUsageResult["wildlife"]; ok {
				t.Error("Expected wildlife to be excluded in self-filtering")
			}
		} else {
			t.Error("Expected second_usage field in result2")
		}

		// Test Aggregate with ValueIntersectionFilter only (no brand filter)
		query3 := search.NewAggregationQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Mikon", "Digma"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"wildlife", "weddings", "portraits"}),
		}).CountItems(true).Sort(query.SortAsc, query.SortRegular)

		result3 := searchIndex.Aggregate(query3)

		if firstUsageResult, ok := result3["first_usage"]; ok {
			expectedFirstUsage := map[string]int{"portraits": 1, "streetphoto": 3, "weddings": 3, "wildlife": 1}
			for k, v := range expectedFirstUsage {
				if count, ok := firstUsageResult[k]; !ok || count != v {
					t.Errorf("Expected first_usage[%s] to be %d in query3, got %v", k, v, count)
				}
			}
		} else {
			t.Error("Expected first_usage field in result3")
		}

		// Test Aggregate with selfFiltering only on ValueIntersectionFilter
		query4 := search.NewAggregationQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Mikon", "Digma"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"wildlife", "weddings", "portraits"}).SelfFiltering(true),
		}).CountItems(true).Sort(query.SortAsc, query.SortRegular)

		result4 := searchIndex.Aggregate(query4)

		if len(result4) > 0 {
			t.Error("Expected empty result")
		}

		// Test SelfFiltering with ValueIntersectionFilter only
		query5 := search.NewAggregationQuery().Filters([]filter.FilterInterface{
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "portraits", "weddings"}),
		}).CountItems(true).Sort(query.SortAsc, query.SortRegular).SelfFiltering(true)

		result5 := searchIndex.Aggregate(query5)

		expected5 := map[string]map[string]int{
			"brand":        {"Digma": 1},
			"first_usage":  {"portraits": 1, "streetphoto": 1, "weddings": 1},
			"second_usage": {"streetphoto": 1, "portraits": 1},
		}

		for field, expectedValues := range expected5 {
			if fieldResult, ok := result5[field]; ok {
				for k, v := range expectedValues {
					if count, ok := fieldResult[k]; !ok || count != v {
						t.Errorf("Expected %s[%s] to be %d in query5, got %v", field, k, v, count)
					}
				}
			} else {
				t.Errorf("Expected %s field in result5", field)
			}
		}

		// Test SelfFiltering with ValueIntersectionFilter on filter level
		query6 := search.NewAggregationQuery().Filters([]filter.FilterInterface{
			search.NewValueIntersectionFilter("first_usage", []interface{}{"streetphoto", "portraits", "weddings"}).SelfFiltering(true),
		}).CountItems(true).Sort(query.SortAsc, query.SortRegular)

		result6 := searchIndex.Aggregate(query6)

		for field, expectedValues := range expected5 {
			if fieldResult, ok := result6[field]; ok {
				for k, v := range expectedValues {
					if count, ok := fieldResult[k]; !ok || count != v {
						t.Errorf("Expected %s[%s] to be %d in query6, got %v", field, k, v, count)
					}
				}
			} else {
				t.Errorf("Expected %s field in result6", field)
			}
		}
	})
}

// TestValueIntersectionFilterGetValue tests ValueIntersectionFilter GetValue functionality.
func TestValueIntersectionFilterGetValue(t *testing.T) {
	f := filter.NewValueIntersectionFilter("tag", []interface{}{"a", "b", "c"}, value.NewValueConverterDefault())

	values := f.GetValue()
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
	if values[0] != "a" || values[1] != "b" || values[2] != "c" {
		t.Errorf("Expected ['a', 'b', 'c'], got %v", values)
	}
}

// TestValueIntersectionFilterFilterInput tests ValueIntersectionFilter FilterInput functionality.
func TestValueIntersectionFilterFilterInput(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"tag": []interface{}{"a", "b", "c"}})
	_ = storage.AddRecord(2, map[string]interface{}{"tag": []interface{}{"a", "b"}})
	_ = storage.AddRecord(3, map[string]interface{}{"tag": []interface{}{"a", "c"}})
	_ = storage.AddRecord(4, map[string]interface{}{"tag": []interface{}{"b", "c"}})
	_ = storage.AddRecord(5, map[string]interface{}{"tag": []interface{}{"a"}})
	storage.Optimize()

	f := search.NewValueIntersectionFilter("tag", []interface{}{"a", "b"})
	result := searchIndex.Query(query.NewSearchQuery().Filters([]filter.FilterInterface{f}))

	// Records 1 and 2 have both "a" and "b"
	sortIntSlice(result)
	assertEqualSlices(t, []int{1, 2}, result)
}

// TestValueIntersectionFilterFilterInputWithExclude tests ValueIntersectionFilter FilterInput with exclude records.
func TestValueIntersectionFilterFilterInputWithExclude(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"tag": []interface{}{"a", "b", "c"}})
	_ = storage.AddRecord(2, map[string]interface{}{"tag": []interface{}{"a", "b"}})
	_ = storage.AddRecord(3, map[string]interface{}{"tag": []interface{}{"a", "c"}})
	_ = storage.AddRecord(4, map[string]interface{}{"tag": []interface{}{"b", "c"}})
	storage.Optimize()

	f := search.NewValueIntersectionFilter("tag", []interface{}{"a", "b"})
	excludeFilter := search.NewExcludeValueFilter("tag", []interface{}{"c"})

	result := searchIndex.Query(query.NewSearchQuery().Filters([]filter.FilterInterface{f, excludeFilter}))

	// Record 1 has "c" so should be excluded, record 2 should remain
	assertEqualSlices(t, []int{2}, result)
}

// TestValueIntersectionFilterFilterInputNoMatches tests ValueIntersectionFilter FilterInput with no matches.
func TestValueIntersectionFilterFilterInputNoMatches(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"tag": []interface{}{"a", "b"}})
	_ = storage.AddRecord(2, map[string]interface{}{"tag": []interface{}{"a", "c"}})
	storage.Optimize()

	f := search.NewValueIntersectionFilter("tag", []interface{}{"a", "d"})
	result := searchIndex.Query(query.NewSearchQuery().Filters([]filter.FilterInterface{f}))

	// No records have both "a" and "d"
	assertEqualSlices(t, []int{}, result)
}

// TestValueIntersectionFilterFilterInputEmptyInput tests ValueIntersectionFilter FilterInput with empty input.
func TestValueIntersectionFilterFilterInputEmptyInput(t *testing.T) {
	f := filter.NewValueIntersectionFilter("tag", []interface{}{"a", "b"}, value.NewValueConverterDefault())

	facetedData := map[string][]int{
		"a": {1, 2, 3},
		"b": {2, 3, 4},
	}

	inputRecords := map[int]bool{}
	excludeRecords := map[int]bool{}

	f.FilterInput(facetedData, inputRecords, excludeRecords)

	// With empty input, should find intersection: records 2 and 3
	if len(inputRecords) != 2 {
		t.Errorf("Expected 2 records, got %d", len(inputRecords))
	}
}

// TestValueIntersectionFilterFilterInputWithInput tests ValueIntersectionFilter FilterInput with existing input.
func TestValueIntersectionFilterFilterInputWithInput(t *testing.T) {
	f := filter.NewValueIntersectionFilter("tag", []interface{}{"a", "b"}, value.NewValueConverterDefault())

	facetedData := map[string][]int{
		"a": {1, 2, 3, 4},
		"b": {2, 3, 5},
	}

	inputRecords := map[int]bool{1: true, 2: true, 3: true, 6: true}
	excludeRecords := map[int]bool{}

	f.FilterInput(facetedData, inputRecords, excludeRecords)

	// Input has 1,2,3,6. Intersection of "a" and "b" is 2,3. So result should be 2,3.
	if len(inputRecords) != 2 {
		t.Errorf("Expected 2 records, got %d", len(inputRecords))
	}
	if !inputRecords[2] || !inputRecords[3] {
		t.Errorf("Expected records 2 and 3")
	}
}

// TestSelfFilter tests self-filtering functionality.
func TestSelfFilter(t *testing.T) {
	getIndex := func() (search *faceted.Search, searchIndex index.IndexInterface) {
		search = faceted.NewSearch()
		searchIndex, _ = search.NewIndex(faceted.ArrayStorage)
		storage := searchIndex.GetStorage()

		data := []map[string]interface{}{
			{"id": 1, "brand": "Nony", "first_usage": []interface{}{"weddings", "wildlife"}, "second_usage": []interface{}{"wildlife", "portraits"}},
			{"id": 2, "brand": "Mikon", "first_usage": []interface{}{"weddings", "streetphoto"}, "second_usage": []interface{}{"wildlife", "streetphoto"}},
			{"id": 3, "brand": "Common", "first_usage": []interface{}{"streetphoto", "portraits"}, "second_usage": []interface{}{"streetphoto", "portraits"}},
			{"id": 4, "brand": "Digma", "first_usage": []interface{}{"streetphoto", "portraits", "weddings"}, "second_usage": []interface{}{"streetphoto", "portraits"}},
			{"id": 5, "brand": "Digma", "first_usage": []interface{}{"streetphoto"}, "second_usage": []interface{}{"portraits"}},
			{"id": 6, "brand": "Mikon", "first_usage": []interface{}{"weddings", "wildlife"}, "second_usage": []interface{}{"wildlife", "portraits"}},
		}

		for _, item := range data {
			id := int(item["id"].(int))
			delete(item, "id")
			_ = storage.AddRecord(id, item)
		}
		storage.Optimize()
		return search, searchIndex
	}

	t.Run("MixedFiltering", func(t *testing.T) {
		search, searchIndex := getIndex()

		query1 := search.NewAggregationQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("brand", []interface{}{"Nony", "Digma", "Mikon"}),
			search.NewValueIntersectionFilter("first_usage", []interface{}{"weddings"}).SelfFiltering(true),
		}).CountItems(true).Sort(query.SortAsc, query.SortRegular)

		result := searchIndex.Aggregate(query1)

		t.Logf("Result: %+v", result)

		// Check that result has expected fields
		if _, ok := result["brand"]; !ok {
			t.Error("Expected brand field in result")
		}
		if _, ok := result["first_usage"]; !ok {
			t.Error("Expected first_usage field in result")
		}
	})
}

// TestRangeIndexer tests RangeIndexer functionality.
func TestRangeIndexer(t *testing.T) {
	t.Run("AddRecord", func(t *testing.T) {

		search := faceted.NewSearch()
		searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
		storage := searchIndex.GetStorage()

		rangeIndexer, _ := search.NewRangeIndexer(100)
		storage.AddIndexer("price", rangeIndexer)

		_ = storage.AddRecord(2, map[string]interface{}{"price": 90})
		_ = storage.AddRecord(3, map[string]interface{}{"price": 100})
		_ = storage.AddRecord(4, map[string]interface{}{"price": 110})
		_ = storage.AddRecord(5, map[string]interface{}{"price": 1000})

		data := storage.GetData()
		if priceData, ok := data["price"]; ok {
			if _, ok := priceData["0"]; !ok {
				t.Error("Expected price range 0")
			}
			if _, ok := priceData["100"]; !ok {
				t.Error("Expected price range 100")
			}
			if _, ok := priceData["1000"]; !ok {
				t.Error("Expected price range 1000")
			}
		}
	})

	t.Run("CombinationTest", func(t *testing.T) {
		search := faceted.NewSearch()
		searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
		storage := searchIndex.GetStorage()

		rangeIndexer, _ := indexer.NewRangeIndexer(50)
		storage.AddIndexer("price", rangeIndexer)

		_ = storage.AddRecord(1, map[string]interface{}{"price": 90})
		_ = storage.AddRecord(2, map[string]interface{}{"price": 110})
		_ = storage.AddRecord(3, map[string]interface{}{"price": 140})
		_ = storage.AddRecord(4, map[string]interface{}{"price": 200})

		storage.Optimize()

		filters := []filter.FilterInterface{
			search.NewExcludeRangeFilter("price", &filter.RangeValue{Min: 100, Max: 150}),
		}

		result := searchIndex.Query(query.NewSearchQuery().Filters(filters))
		sortIntSlice(result)
		assertEqualSlices(t, []int{1, 4}, result)
	})
}

// TestStorage tests storage operations.
func TestStorage(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	t.Run("AddRecord", func(t *testing.T) {
		records := map[int]map[string]interface{}{
			112: {"vendor": "Tester", "price": 100},
			113: {"vendor": "Tester2", "price": 101},
			114: {"vendor": "Tester2", "price": 101},
		}

		for id, val := range records {
			err := storage.AddRecord(id, val)
			if err != nil {
				t.Errorf("Failed to add record %d", id)
			}
			if id == 112 {
				storage.Optimize()
			}
		}

		data := storage.GetData()
		if vendorData, ok := data["vendor"]; ok {
			if testerRecords, ok := vendorData["Tester"]; ok {
				if len(testerRecords) != 1 || testerRecords[0] != 112 {
					t.Errorf("Expected Tester to have record [112], got %v", testerRecords)
				}
			}
		}
	})

	t.Run("HasField", func(t *testing.T) {
		if !storage.HasField("vendor") {
			t.Error("Expected vendor field to exist")
		}
		if storage.HasField("undefined_field") {
			t.Error("Expected undefined_field to not exist")
		}
	})

	t.Run("RecordsCount", func(t *testing.T) {
		if count := storage.GetRecordsCount("price", "101"); count != 2 {
			t.Errorf("Expected price '101' count to be 2, got %d", count)
		}
		if count := storage.GetRecordsCount("price", "500"); count != 0 {
			t.Errorf("Expected price '500' count to be 0, got %d", count)
		}
	})

	t.Run("DeleteRecord", func(t *testing.T) {
		search := faceted.NewSearch()
		searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
		storage := searchIndex.GetStorage()

		_ = storage.AddRecord(10, map[string]interface{}{"color": "red"})
		_ = storage.AddRecord(11, map[string]interface{}{"color": "blue"})

		result := searchIndex.Query(query.NewSearchQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("color", []interface{}{"red"}),
		}))
		assertEqualSlices(t, []int{10}, result)

		storage.DeleteRecord(10)

		result2 := searchIndex.Query(query.NewSearchQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("color", []interface{}{"red"}),
		}))
		assertEqualSlices(t, []int{}, result2)
	})

	t.Run("ReplaceRecord", func(t *testing.T) {
		search := faceted.NewSearch()
		searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
		storage := searchIndex.GetStorage()

		_ = storage.AddRecord(10, map[string]interface{}{"color": "red", "size": 100})
		_ = storage.AddRecord(11, map[string]interface{}{"color": "blue"})

		// Replace record 10
		storage.ReplaceRecord(10, map[string]interface{}{"color": "blue", "size": 150})

		// Red should be gone
		result := searchIndex.Query(query.NewSearchQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("color", []interface{}{"red"}),
		}))
		assertEqualSlices(t, []int{}, result)

		// Size 150 should exist
		result2 := searchIndex.Query(query.NewSearchQuery().Filters([]filter.FilterInterface{
			search.NewValueFilter("size", []interface{}{150}),
		}))
		if len(result2) != 1 {
			t.Errorf("Expected 1 result for size 150, got %d", len(result2))
		}
	})
}

// TestScanner tests scanner functionality.
func TestScanner(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(112, map[string]interface{}{"vendor": "Tester", "price": 100})
	_ = storage.AddRecord(113, map[string]interface{}{"vendor": "Tester2", "price": 101})
	_ = storage.AddRecord(114, map[string]interface{}{"vendor": "Tester2", "price": 101})

	scanner := searchIndex.GetScanner()
	allRecords := scanner.GetAllRecordIdMap(storage)

	if len(allRecords) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allRecords))
	}

	if !allRecords[112] || !allRecords[113] || !allRecords[114] {
		t.Errorf("Expected all records to be present")
	}
}

// TestAggregateWithNoFilters tests aggregation with no filters.
func TestAggregateWithNoFilters(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"brand": "Nony", "price": 100})
	_ = storage.AddRecord(2, map[string]interface{}{"brand": "Mikon", "price": 200})
	_ = storage.AddRecord(3, map[string]interface{}{"brand": "Digma", "price": 150})
	storage.Optimize()

	// Aggregation with no filters should return all values
	result := searchIndex.Aggregate(search.NewAggregationQuery())

	if brandResult, ok := result["brand"]; ok {
		if _, hasNony := brandResult["Nony"]; !hasNony {
			t.Errorf("Expected brand Nony in result")
		}
		if _, hasMikon := brandResult["Mikon"]; !hasMikon {
			t.Errorf("Expected brand Mikon in result")
		}
		if _, hasDigma := brandResult["Digma"]; !hasDigma {
			t.Errorf("Expected brand Digma in result")
		}
	} else {
		t.Error("Expected brand field in result")
	}

	if priceResult, ok := result["price"]; ok {
		if _, has100 := priceResult["100"]; !has100 {
			t.Errorf("Expected price 100 in result")
		}
		if _, has200 := priceResult["200"]; !has200 {
			t.Errorf("Expected price 200 in result")
		}
		if _, has150 := priceResult["150"]; !has150 {
			t.Errorf("Expected price 150 in result")
		}
	} else {
		t.Error("Expected price field in result")
	}
}

// TestAggregateWithNoFiltersAndCount tests aggregation with no filters and count.
func TestAggregateWithNoFiltersAndCount(t *testing.T) {
	search := faceted.NewSearch()
	searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
	storage := searchIndex.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"brand": "Nony"})
	_ = storage.AddRecord(2, map[string]interface{}{"brand": "Mikon"})
	_ = storage.AddRecord(3, map[string]interface{}{"brand": "Digma"})
	storage.Optimize()

	// Aggregation with count
	result := searchIndex.Aggregate(search.NewAggregationQuery().CountItems(true))

	if brandResult, ok := result["brand"]; ok {
		if count, ok := brandResult["Nony"]; ok {
			if count != 1 {
				t.Errorf("Expected count 1 for Nony, got %v", count)
			}
		} else {
			t.Errorf("Expected Nony in brand result")
		}
	}
}

// TestAggregationQueryFilter tests AggregationQuery Filter method.
func TestAggregationQueryFilter(t *testing.T) {
	query := faceted.NewSearch().NewAggregationQuery()
	query.Filter(faceted.NewSearch().NewValueFilter("brand", []interface{}{"Nony"}))

	filters := query.GetFilters()
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}
}

// TestSearchQueryFilter tests SearchQuery Filter method.
func TestSearchQueryFilter(t *testing.T) {
	query := faceted.NewSearch().NewSearchQuery()
	query.Filter(faceted.NewSearch().NewValueFilter("brand", []interface{}{"Nony"}))

	filters := query.GetFilters()
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}
}

// TestSearchQuerySortBy tests SearchQuery SortBy method.
func TestSearchQuerySortBy(t *testing.T) {
	search := faceted.NewSearch()
	query := search.NewSearchQuery()
	sort := search.NewQuerySort("price", 1, 1) // SortDesc=1, SortNumeric=1
	query.SortBy(sort)

	resultSort := query.GetSort()
	if resultSort == nil {
		t.Errorf("Expected sort to be set")
	}
	if resultSort.GetField() != "price" {
		t.Errorf("Expected field 'price', got %s", resultSort.GetField())
	}
}
