package main

import (
	"encoding/json"
	"fmt"

	"github.com/k-samuel/faceted/search"
)

func main() {
	fmt.Println("Faceted Search Library - Go Port")
	fmt.Println("=================================")

	// Create index using Factory
	provider := search.NewContainer()
	db := provider.NewDb()

	storage := db.GetStorage()

	// Sample data - similar to PHP example
	data := []map[string]interface{}{
		{"id": 7, "color": "black", "price": 100, "sale": true, "size": 36},
		{"id": 9, "color": "green", "price": 100, "sale": true, "size": 40},
		{"id": 10, "color": "black", "price": 150, "sale": false, "size": 38},
		{"id": 11, "color": "blue", "price": 200, "sale": true, "size": 42},
		{"id": 12, "color": "green", "price": 120, "sale": false, "size": 36},
	}

	// Add records to index
	for _, item := range data {
		recordId := int(item["id"].(int))
		delete(item, "id")
		storage.AddRecord(recordId, item)
	}

	// Optimize index
	storage.Optimize()

	// Export index data
	indexData := storage.Export()
	jsonData, _ := json.MarshalIndent(indexData, "", "  ")
	fmt.Printf("\nIndex Data (first 500 chars):\n")
	if len(jsonData) > 500 {
		fmt.Printf("%s...\n", string(jsonData[:500]))
	} else {
		fmt.Printf("%s\n", string(jsonData))
	}

	// Example 1: Search with filters
	fmt.Println("\n=== Example 1: Search with Filters ===")
	filters := []search.FilterInterface{
		search.NewValueFilter("color", []interface{}{"black", "green"}),
		search.NewRangeFilter("size", search.NewRangeValue(36, 40)),
	}

	searchQuery := search.NewSearchQuery().Filters(filters)
	records, _ := db.Query(searchQuery)
	fmt.Printf("Found records: %v\n", records)

	// Example 2: Aggregation without count
	fmt.Println("\n=== Example 2: Aggregation (without count) ===")
	aggQuery := search.NewAggregationQuery().Filters(filters)
	aggData, _ := db.Aggregate(aggQuery)
	fmt.Printf("Available filters: %v\n", aggData)

	// Example 3: Aggregation with count and sorting
	fmt.Println("\n=== Example 3: Aggregation (with count & sort) ===")
	aggQuery2 := search.NewAggregationQuery().Filters(filters).CountItems(true).Sort(search.SortAsc, search.SortAsc)
	aggData2, _ := db.Aggregate(aggQuery2)
	fmt.Printf("Available filters with count: %v\n", aggData2)

	// Example 4: Aggregation without filters (all values)
	fmt.Println("\n=== Example 4: Aggregation (all values) ===")
	aggQuery3 := search.NewAggregationQuery().CountItems(true)
	aggData3, _ := db.Aggregate(aggQuery3)
	fmt.Printf("All available values: %v\n", aggData3)

	// Example 5: Search with sorting
	fmt.Println("\n=== Example 5: Search with Sorting ===")
	searchQuery4 := search.NewSearchQuery().Filters(filters).Sort("price", search.SortDesc, search.SortTypeNumbers)
	sortedRecords, _ := db.Query(searchQuery4)
	fmt.Printf("Sorted records (by price DESC): %v\n", sortedRecords)

	// Example 7: Exclude filter
	fmt.Println("\n=== Example 7: Exclude Filter ===")
	filtersWithExclude := []search.FilterInterface{
		search.NewValueFilter("sale", []interface{}{1}),
		search.NewExcludeValueFilter("color", []interface{}{"blue"}),
	}
	searchQuery6 := search.NewSearchQuery().Filters(filtersWithExclude)
	recordsWithExclude, _ := db.Query(searchQuery6)
	fmt.Printf("Records (sale=1, color!=blue): %v\n", recordsWithExclude)

	// Example 8: ValueIntersectionFilter (AND condition)
	fmt.Println("\n=== Example 8: Value Intersection Filter (AND) ===")
	// For records with multiple values per field

	search2 := provider.NewDb()

	// Add record with multiple purposes
	search2.GetStorage().AddRecord(1, map[string]interface{}{
		"brand":   "Pony",
		"purpose": []interface{}{"hunting", "fishing", "sports"},
	})
	search2.GetStorage().AddRecord(2, map[string]interface{}{
		"brand":   "Nike",
		"purpose": []interface{}{"hunting", "sports"},
	})
	search2.GetStorage().AddRecord(3, map[string]interface{}{
		"brand":   "Adidas",
		"purpose": []interface{}{"fishing", "sports"},
	})

	// Find items that have BOTH hunting AND fishing
	intersectionFilter := search.NewValueIntersectionFilter("purpose", []interface{}{"hunting", "fishing"})
	searchQuery7 := search.NewSearchQuery().Filter(intersectionFilter)
	intersectionRecords, _ := search2.Query(searchQuery7)
	fmt.Printf("Records with hunting AND fishing: %v\n", intersectionRecords)

	// Example 9: Get record count
	fmt.Println("\n=== Example 9: Record Count ===")
	fmt.Printf("Total records in index: %d\n", search2.GetStorage().GetCount())

	fmt.Println("\n=================================")
	fmt.Println("Examples completed successfully!")
}
