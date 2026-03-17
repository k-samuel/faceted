package tests

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/k-samuel/faceted"
	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/index"
	"github.com/k-samuel/faceted/pkg/query"
	//	"github.com/k-samuel/faceted/pkg/query"
)

// -----
// go test ./tests/perf -bench . -benchmem
// go test ./tests/perf -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1 performance_test.go
// go tool pprof -callgrind -output callgrind.c.out cpu.out
// go tool pprof -callgrind -output callgrind.m.out mem.out

var testIndex index.IndexInterface
var search *faceted.Search
var datasetFilePrefix = ".test.dataset."
var results = 100000
var datasetFile string

func init() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	datasetFile = datasetFilePrefix + strconv.Itoa(results)
	if _, err := os.Stat(datasetFile); errors.Is(err, os.ErrNotExist) {
		CreateDataset()
	}
	search = faceted.NewSearch()
	testIndex = CreateIndex()
}

func CreateDataset() {
	start := time.Now()
	colors := []string{"red", "green", "blue", "yellow", "black", "white"}
	brands := []string{
		"Nike",
		"H&M",
		"Zara",
		"Adidas",
		"Louis Vuitton",
		"Cartier",
		"Hermes",
		"Gucci",
		"Uniqlo",
		"Rolex",
		"Coach",
		"Victoria\"s Secret",
		"Chow Tai Fook",
		"Tiffany & Co.",
		"Burberry",
		"Christian Dior",
		"Polo Ralph Lauren",
		"Prada",
		"Under Armour",
		"Armani",
		"Puma",
		"Ray-Ban"}

	warehouses := []int{1, 10, 23, 345, 43, 5476, 34, 675, 34, 24, 789, 45, 65, 34, 54, 511, 512, 520}
	itemType := []string{"normal", "middle", "good"}

	f, err := os.Create(datasetFile)
	check(err)
	defer f.Close()

	for i := 1; i < results+1; i++ {

		countWh := rand.Intn(len(warehouses))
		wh := make([]int, 0)
		for j := 0; j < int(countWh); j++ {
			wh = append(wh, rand.Intn(len(warehouses))-1)
		}

		randType := rand.Intn(int(len(itemType) - 1))

		record := map[string]interface{}{
			"id":         i,
			"color":      colors[rand.Intn(5)],
			"back_color": colors[rand.Intn(5)],
			"size":       randNum(34, 50),
			"brand":      brands[rand.Intn(len(brands))],
			"price":      randNum(10000, 100000),
			"discount":   rand.Intn(10),
			"combined":   rand.Intn(2),
			"quantity":   rand.Intn(100),
			"warehouse":  unique(wh),
			"type":       itemType[randType],
		}

		s, e := json.Marshal(record)
		check(e)
		f.Write(s)
		f.WriteString("\n")
	}
	fmt.Println("Dataset: ", time.Since(start))
}

func createFilters() []filter.FilterInterface {
	return []filter.FilterInterface{
		search.NewValueFilter("color", []string{"black"}),
		search.NewValueFilter("warehouse", []string{"789", "45", "65", "1", "10"}),
		search.NewValueFilter("type", []string{"normal", "middle"}),
	}
}

func CreateIndex() index.IndexInterface {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startM := m.Alloc
	start := time.Now()
	var result map[string]interface{}

	var localIndex index.IndexInterface
	localIndex, _ = search.NewIndex(faceted.ArrayStorage)
	storage := localIndex.GetStorage()

	file, err := os.Open(datasetFile)
	check(err)
	defer file.Close()
	counter := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		json.Unmarshal([]byte(scanner.Text()), &result)
		id := int(result["id"].(float64))
		delete(result, "id")
		storage.AddRecord(id, result)
		counter++
	}

	runtime.GC()
	runtime.ReadMemStats(&m)

	fmt.Printf("Alloc: %v MiB for %v items ", bToMb(m.Alloc-startM), counter)
	fmt.Println("Load: ", time.Since(start))

	return localIndex
}

func randNum(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func BenchmarkFind(b *testing.B) {
	filters := createFilters()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testIndex.Query(search.NewSearchQuery().Filters(filters))
	}
}

func BenchmarkAggregateFilters(b *testing.B) {

	filters := createFilters()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testIndex.Aggregate(search.NewAggregationQuery().Filters(filters))
	}
}

func BenchmarkFindAndSort(b *testing.B) {
	filters := createFilters()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testIndex.Query(search.NewSearchQuery().Filters(filters).Sort("quantity", query.SortDesc, query.SortRegular))
	}
}

func BenchmarkSearch(b *testing.B) {

	filters := createFilters()

	start := time.Now()
	res := testIndex.Query(search.NewSearchQuery().Filters(filters))
	duration := time.Since(start)
	b.Log(" Find: ", duration, " Results: ", len(res))

	runtime.GC()

	start = time.Now()
	filterRes := testIndex.Aggregate(search.NewAggregationQuery().Filters(filters))
	duration = time.Since(start)
	b.Log(" Aggregate filters: ", duration, " filters: ", len(filterRes))

	runtime.GC()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
