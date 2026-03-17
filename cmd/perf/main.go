package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/k-samuel/faceted"
	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/index"
	"github.com/k-samuel/faceted/pkg/query"
)

const (
	outputDir = "./tests/data"
	dataFile  = "data.json"
)

// WarehouseData represents warehouse data that can be either array or object.
type WarehouseData map[string]int

// Record represents a product record from the dataset.
type Record struct {
	ID        int             `json:"id"`
	Color     string          `json:"color"`
	BackColor string          `json:"back_color"`
	Size      int             `json:"size"`
	Brand     string          `json:"brand"`
	Price     int             `json:"price"`
	Discount  int             `json:"discount"`
	Combined  int             `json:"combined"`
	Quantity  int             `json:"quantity"`
	Warehouse json.RawMessage `json:"warehouse"`
	Type      string          `json:"type"`
}

// TestResult represents a single test result.
type TestResult struct {
	Method  string
	Time    float64
	Records int
	Extra   string
}

// IndexStat represents index statistics.
type IndexStat struct {
	Method string
	Value  string
	Extra  string
}

func main() {

	sz := flag.Int("size", 100000, "dataset size")
	flag.Parse()
	resultSize := *sz

	fmt.Println("Faceted Search - Performance Test (Go)")
	fmt.Println("======================================")
	fmt.Println()

	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	var loadTime float64
	var optTime float64

	filePath := fmt.Sprintf("%s/%d/%s", outputDir, resultSize, dataFile)

	// Always create index from raw data
	fmt.Printf("Creating index from %s...\n", filePath)
	t := time.Now()

	records, err := loadData(filePath)
	if err != nil {
		fmt.Printf("Error loading data: %v\n", err)
		return
	}

	loadTime = time.Since(t).Seconds()
	fmt.Printf("Loaded %d records in %.6f s\n", len(records), loadTime)

	// Create index with FastStorage for better performance
	search := faceted.NewSearch()
	searchIndex, errs := search.NewIndex(faceted.ArrayStorage)
	if errs != nil {
		panic(errs)
	}

	storage := searchIndex.GetStorage()

	// Add RangeIndexer for price field (same as PHP version with step 250)
	rangeIndexer, _ := search.NewRangeIndexer(250)
	storage.AddIndexer("price", rangeIndexer)

	// Add records to index
	fmt.Println("Indexing records...")
	t = time.Now()
	for _, rec := range records {
		// Parse warehouse - can be array or object
		var warehouse []int
		if len(rec.Warehouse) > 0 {
			// Try to parse as array first
			var arr []int
			if err := json.Unmarshal(rec.Warehouse, &arr); err == nil {
				warehouse = arr
			} else {
				// Parse as object and convert to array
				var obj map[string]int
				if err := json.Unmarshal(rec.Warehouse, &obj); err == nil {
					for _, v := range obj {
						warehouse = append(warehouse, v)
					}
				}
			}
		}

		recordValues := map[string]interface{}{
			"color":      rec.Color,
			"back_color": rec.BackColor,
			"size":       rec.Size,
			"brand":      rec.Brand,
			"price":      rec.Price,
			"discount":   rec.Discount,
			"combined":   rec.Combined,
			"quantity":   rec.Quantity,
			"warehouse":  interfaceSlice(warehouse),
			"type":       rec.Type,
		}
		storage.AddRecord(rec.ID, recordValues)
	}
	indexTime := time.Since(t).Seconds()
	fmt.Printf("Indexing completed in %.6f s\n", indexTime)

	// Optimize
	fmt.Println("Optimizing index...")
	t = time.Now()
	storage.Optimize()
	optTime = time.Since(t).Seconds()
	fmt.Printf("Optimization completed in %.6f s\n", optTime)

	runtime.GC()
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	fmt.Printf("Total memory used: %d Mb\n", (memEnd.Alloc-memStart.Alloc)/1024/1024)

	// Index stats
	resultData := []IndexStat{
		{"Records", fmt.Sprintf("%d", searchIndex.GetCount()), ""},
		{"Index memory usage", fmt.Sprintf("%d Mb", memEnd.Alloc/1024/1024), ""},
		{"Loading time", fmt.Sprintf("%.6f s", loadTime), ""},
		{"Indexing time", fmt.Sprintf("%.6f s", indexTime), ""},
		{"Optimize time", fmt.Sprintf("%.6f s", optTime), ""},
	}

	// Define filters (same as PHP find.php)
	filters := []filter.FilterInterface{
		search.NewValueFilter("color", []interface{}{"black"}),
		search.NewValueFilter("warehouse", []interface{}{789, 45, 65, 1, 10}),
		search.NewValueFilter("type", []interface{}{"normal", "middle"}),
	}

	filters2 := []filter.FilterInterface{
		search.NewValueFilter("color", []interface{}{"black"}),
		search.NewValueFilter("warehouse", []interface{}{789, 45, 65, 1, 10}),
		search.NewRangeFilter("price", search.NewRangeValue(1000, 5000)),
	}

	filters3 := []filter.FilterInterface{
		search.NewValueFilter("color", []interface{}{"black"}),
		search.NewValueFilter("warehouse", []interface{}{789, 45, 65, 1, 10}),
		search.NewExcludeValueFilter("type", []interface{}{"good"}),
	}

	// Test functions
	find := func(s index.IndexInterface, f []filter.FilterInterface) TestResult {
		t := time.Now()
		results := s.Query(search.NewSearchQuery().Filters(f))
		return TestResult{"Find", time.Since(t).Seconds(), len(results), ""}
	}

	findAndSort := func(s index.IndexInterface, f []filter.FilterInterface) TestResult {
		t := time.Now()

		// Create query
		queryObj := search.NewSearchQuery().Filters(f).Sort("quantity", query.SortDesc, query.SortNumeric)

		// Query with timing
		t1 := time.Now()
		results := s.Query(queryObj)
		queryTime := time.Since(t1)

		totalTime := time.Since(t)
		return TestResult{"Find & Sort", totalTime.Seconds(), len(results), fmt.Sprintf("query=%.3f", queryTime.Seconds())}
	}

	aggregate := func(s index.IndexInterface, f []filter.FilterInterface) TestResult {
		t := time.Now()
		_ = s.Aggregate(query.NewAggregationQuery().Filters(f))
		return TestResult{"Filters", time.Since(t).Seconds(), len(f), ""}
	}

	aggregateAndCount := func(s index.IndexInterface, f []filter.FilterInterface) TestResult {
		t := time.Now()
		_ = s.Aggregate(query.NewAggregationQuery().Filters(f).CountItems(true))
		return TestResult{"Filters & count", time.Since(t).Seconds(), len(f), ""}
	}

	aggregateAndCountWithExclude := func(s index.IndexInterface, f []filter.FilterInterface) TestResult {
		t := time.Now()
		_ = s.Aggregate(query.NewAggregationQuery().Filters(f).CountItems(true))
		return TestResult{"Filters & count & exc", time.Since(t).Seconds(), len(f), ""}
	}

	findWithRange := func(s index.IndexInterface, f []filter.FilterInterface) TestResult {
		t := time.Now()
		results := s.Query(query.NewSearchQuery().Filters(f))
		return TestResult{"Find (ranges)", time.Since(t).Seconds(), len(results), ""}
	}

	findWithExclude := func(s index.IndexInterface, f []filter.FilterInterface) TestResult {
		t := time.Now()
		results := s.Query(query.NewSearchQuery().Filters(f))
		return TestResult{"Find (unsets)", time.Since(t).Seconds(), len(results), ""}
	}

	// Run tests

	tests := []struct {
		name string
		fn   func(index.IndexInterface, []filter.FilterInterface) TestResult
		f    []filter.FilterInterface
	}{
		{"find", find, filters},
		{"findAndSort", findAndSort, filters},
		{"findWithExclude", findWithExclude, filters3},
		{"findWithRange", findWithRange, filters2},
		{"aggregate", aggregate, filters},
		{"aggregateAndCount", aggregateAndCount, filters},
		{"aggregateAndCountWithExclude", aggregateAndCountWithExclude, filters3},
	}

	testResultData := []TestResult{}
	runtime.GC()

	for _, test := range tests {
		var memBefore runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memBefore)
		memBeforeMb := int64(memBefore.Alloc) / 1024 / 1024

		result := test.fn(searchIndex, test.f)

		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		memAfterMb := int64(memAfter.Alloc) / 1024 / 1024

		memDiff := memAfterMb - memBeforeMb
		if memDiff < 0 {
			memDiff = 0
		}
		result.Extra = fmt.Sprintf("%d / %d", memDiff, memAfterMb)
		testResultData = append(testResultData, result)
	}

	// Print results
	colLen := [4]int{25, 12, 10, 20}
	totalLen := colLen[0] + colLen[1] + colLen[2] + colLen[3] + 8

	fmt.Println()
	fmt.Println("Index Info")
	fmt.Println(prepareIndexStat(colLen[:], totalLen, resultData))
	fmt.Println()

	fmt.Println("Perf Results")
	fmt.Println(prepareResultsHeader(colLen[:], totalLen, []string{"Method", "Time, s.", "Records", "Extra / Total Mb"}))
	fmt.Println(prepareResults(colLen[:], totalLen, testResultData))
	fmt.Println()
}

// loadData reads JSON lines from file and returns records.
func loadData(filename string) ([]Record, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []Record
	scanner := bufio.NewScanner(file)
	lineNum := 0
	skipCount := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var rec Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			skipCount++
			continue
		}
		records = append(records, rec)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if skipCount > 0 {
		fmt.Printf("Skipped %d invalid lines\n", skipCount)
	}

	return records, nil
}

// interfaceSlice converts []int to []interface{}.
func interfaceSlice(ints []int) []interface{} {
	result := make([]interface{}, len(ints))
	for i, v := range ints {
		result[i] = v
	}
	return result
}

func prepareIndexStat(colLen []int, totalLen int, data []IndexStat) string {
	result := repeatString("-", totalLen) + "\n"
	for _, cols := range data {
		for i := 0; i < 2; i++ {
			if i == 0 {
				result += "| " + padRight(cols.Method, colLen[i])
			} else {
				result += "| " + padRight(cols.Value, colLen[i])
			}
		}
		result += "|\n"
	}
	result += repeatString("-", totalLen) + "\n"
	return result
}

func prepareResultsHeader(colLen []int, totalLen int, data []string) string {
	result := repeatString("-", totalLen) + "\n"

	for i, title := range data {
		result += "| " + padRight(title, colLen[i])
	}
	result += "|\n"
	result += repeatString("-", totalLen)
	return result
}

func prepareResults(colLen []int, totalLen int, data []TestResult) string {
	result := ""
	for _, cols := range data {
		result += "| " + padRight(cols.Method, colLen[0])
		result += "| " + padRight(fmt.Sprintf("%.6f", cols.Time), colLen[1])
		result += "| " + padRight(fmt.Sprintf("%d", cols.Records), colLen[2])
		result += "| " + padRight(cols.Extra, colLen[3])
		result += "|\n"
	}
	result += repeatString("-", totalLen) + "\n"
	return result
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + repeatString(" ", length-len(s))
}
