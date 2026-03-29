package main

/*
 * Faceted search query server
 * HTTP server that processes faceted search queries
 */

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/k-samuel/faceted/search"
	"github.com/k-samuel/faceted/search/indexer"
)

// Product represents a product from database
/*
type Product struct {
	Id         int    `json:"id"`
	Brand      string `json:"brand"`
	Model      string `json:"model"`
	Color      string `json:"color"`
	Cam        string `json:"cam"`
	Diagonal   string `json:"diagonal"`
	Battery    string `json:"battery"`
	State      string `json:"state"`
	Price      int    `json:"price"`
	PriceRange int    `json:"price_range"`
	Ram        string `json:"ram"`
	Hd         string `json:"hd"`
}
*/

// Titles represents title mapping for display
var filterTitles = map[string]string{
	"brand":       "Brand",
	"price_range": "Price Range",
	"hd":          "Memory Storage, Gb",
	"state":       "Quality",
	"color":       "Color",
	"diagonal":    "Size",
	"battery":     "Battery",
	"cam":         "Cam resolution, MP",
	"ram":         "Memory RAM",
}

// List of field for numer sort
var numberFields = map[string]struct{}{
	"price_range": {},
	"price":       {},
	"hd":          {},
	"diagonal":    {},
	"battery":     {},
	"cam":         {},
	"ram":         {},
}

// file with db data
const dbPath = "./db/mobile-db-json.txt"
const pageLimit = 20
const defaultSort = "brand"

func main() {

	fmt.Println("k-samuel Faceted Search Server")
	fmt.Println("=============================")

	// Create Search Index using dependency settings
	db := search.NewContainer().NewDb()

	// Add RangeIndexer for price field (using price_range)
	priceIndexer, err := indexer.NewRangeIndexer(200)
	if err != nil {
		fmt.Printf("Error creating price indexer: %v\n", err)
		os.Exit(1)
	}

	storage := db.GetStorage()
	storage.AddIndexer("price_range", priceIndexer)

	// Load Data from file
	dbData := loadMobileDB(dbPath)

	// Add records to index
	for id, product := range dbData {
		// Add price_range field for range indexing
		if price, ok := product["price"].(int); ok {
			// create additional field for range indexing
			product["price_range"] = price
		}
		storage.AddRecord(id, product)
	}
	storage.Optimize()

	// ====== HTTP Server ======
	// static files
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// filters menu data
	http.HandleFunc("/filters", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")

		filters, querySort, err := extractQueryParams(r, defaultSort)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			fmt.Println("Invalid Request ", err.Error())
			return
		}

		// Find filters
		aggQuery := search.NewAggregationQuery().Filters(filters).CountItems(true).Sort(search.SortAsc, search.SortAsc)

		data, aerr := db.Aggregate(aggQuery)
		if aerr != nil {
			fmt.Println(aerr)
		}

		searchQuery := search.NewSearchQuery().Filters(filters)
		if querySort != nil {
			searchQuery.SortBy(querySort)
		}

		productData, qerr := db.Query(searchQuery)
		if qerr != nil {
			fmt.Println(qerr)
		}

		// Get first page of products
		resultItems := []map[string]interface{}{}

		if len(productData) > 0 {
			for _, id := range productData {
				if len(resultItems) == pageLimit {
					break
				}
				if p, ok := dbData[id]; ok {
					resultItems = append(resultItems, p)
				}
			}
		}

		dataCopy := make([]*search.AggregationResultField, 0, len(data))
		// Exclude fields that should not be used to filter data.
		for _, f := range data {
			// check if field exists in db
			if _, ok := filterTitles[f.Field]; ok {
				dataCopy = append(dataCopy, f)
			}
		}

		result := map[string]interface{}{
			"filters": map[string]interface{}{"data": dataCopy, "price_step": 200},
			"results": map[string]interface{}{"data": resultItems, "count": len(productData), "limit": pageLimit},
			"titles":  filterTitles,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// Start server
	fmt.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}

}

// loadMobileDB loads the mobile database from JSON Lines file
func loadMobileDB(filePath string) map[int]map[string]interface{} {
	db := make(map[int]map[string]interface{})

	// Load from JSON Lines file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening mobile DB: %v\n", err)
		return db
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var product map[string]interface{}
		if err := json.Unmarshal([]byte(line), &product); err != nil {
			fmt.Printf("Error reading DB: %v\n", err)
			continue
		}
		id := int(product["id"].(float64))
		delete(product, "id")

		db[id] = product
	}

	return db
}

func extractQueryParams(r *http.Request, defaultSortValue string) (filters []search.FilterInterface, sort *search.Sort, err error) {

	filters = make([]search.FilterInterface, 0)

	err = r.ParseForm()
	if err != nil {
		return nil, nil, err
	}

	s := r.FormValue("filters")
	if s != "" {
		var fMap map[string]map[string]map[string]bool

		err = json.Unmarshal([]byte(s), &fMap)

		if err != nil {
			return nil, nil, err
		}

		// parce include filters
		if v, ok := fMap["include"]; ok {
			for field, valMap := range v {

				// check if field exists in db
				if _, ok := filterTitles[field]; !ok {
					continue
				}

				if len(valMap) == 0 {
					continue
				}
				valList := make([]string, 0, 5)
				for valName, valFlag := range valMap {
					if valFlag {
						valList = append(valList, valName)
					}
				}
				if len(valList) > 0 {
					filters = append(filters, search.NewValueFilter(field, valList))
				}
			}
		}
		// parce exclude filters
		if v, ok := fMap["exclude"]; ok {
			for field, valMap := range v {

				// check if field exists in db
				if _, ok := filterTitles[field]; !ok {
					continue
				}

				if len(valMap) == 0 {
					continue
				}
				valList := make([]string, 0, 5)
				for valName, valFlag := range valMap {
					if valFlag {
						valList = append(valList, valName)
					}
				}
				if len(valList) > 0 {
					filters = append(filters, search.NewExcludeValueFilter(field, valList))
				}
			}
		}
	}

	pFromStr := r.FormValue("price_from")
	pToStr := r.FormValue("price_to")

	// Price range filter
	if pFromStr != "" || pToStr != "" {
		rangeValue := &search.RangeValue{}
		if pFromStr != "" {
			priceFrom, err := strconv.Atoi(pFromStr)
			if err == nil {
				rangeValue.Min = priceFrom
			}
		}

		if pToStr != "" {
			priceTo, err := strconv.Atoi(pToStr)
			if err == nil {
				rangeValue.Max = priceTo
			}
		}
		// Price range filter
		filters = append(filters, search.NewRangeFilter("price", rangeValue))
	}

	// Build sort config
	orderStr := r.FormValue("order")
	direction := search.SortAsc

	if orderStr == "" {
		orderStr = defaultSortValue
	}

	if orderStr != "" {
		directionStr := r.FormValue("dir")
		if directionStr == "desc" {
			direction = search.SortDesc
		}
	}

	var sortConfig *search.Sort
	if _, ok := numberFields[orderStr]; ok {
		sortConfig = search.NewSort(orderStr, direction, search.SortTypeNumbers)
	} else {
		sortConfig = search.NewSort(orderStr, direction, search.SortTypeStrings)
	}

	return filters, sortConfig, nil
}
