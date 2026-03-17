package search

import (
	"math"
	"slices"
	"sort"
	"strings"
)

// Index implements IndexInterface for faceted search.
type Db struct {
	storage StorageInterface
	scanner ScannerInterface
}

// NewIndex creates a new Index.
func NewDb(
	storage StorageInterface,
	scanner ScannerInterface,
) *Db {
	return &Db{
		storage: storage,
		scanner: scanner,
	}
}

// Query finds records using Query.
func (i *Db) Query(q *SearchQuery) ([]int, error) {
	inputRecords := q.GetInRecords()
	filterList := q.GetFilters()

	filters := make([]FilterInterface, 0, len(filterList))
	exceptFilters := make([]ExcludeFilterInterface, 0)

	for _, item := range filterList {
		if ef, ok := item.(ExcludeFilterInterface); ok {
			exceptFilters = append(exceptFilters, ef)
		} else {
			filters = append(filters, item)
		}
	}

	order := q.GetSort()

	// Convert input records to map
	var inputMap map[int]struct{}
	if len(inputRecords) > 0 {
		inputMap = mapInputArray(inputRecords)
	}

	var err error
	// Optimize filter order - process filters with fewer matches first
	if len(inputMap) == 0 && len(filters) > 1 {
		filters, err = sortFilters(i.storage, filters)
		if err != nil {
			return nil, err
		}
	}

	// Build exclude map
	excludeMap := make(map[int]struct{})
	if len(exceptFilters) > 0 {
		i.scanner.FindExcludeRecordsMap(exceptFilters, excludeMap)
	}

	// Find records
	mapResult, err := i.scanner.FindRecordsMap(filters, inputMap, excludeMap)
	if err != nil {
		return nil, err
	}

	// Sort if needed
	if order != nil {
		velueMap := i.scanner.GetFieldValueRecords(order.GetField())
		result := sortQuery(velueMap, mapResult, order)
		return result, nil
	}

	// Convert map to slice with pre-allocated capacity
	result := make([]int, 0, len(mapResult))
	for k := range mapResult {
		result = append(result, k)
	}
	return result, nil
}

// Aggregate finds acceptable filter values.
func (i *Db) Aggregate(q *AggregationQuery) (map[string]map[string]interface{}, error) {
	input := q.GetInRecords()
	filterList := q.GetFilters()

	filters := make([]FilterInterface, 0)
	exceptFilters := make([]ExcludeFilterInterface, 0)

	for _, item := range filterList {
		if ef, ok := item.(ExcludeFilterInterface); ok {
			exceptFilters = append(exceptFilters, ef)
		} else {
			filters = append(filters, item)
		}
	}

	countValues := q.GetCountItems()
	sortConfig := q.GetSort()

	// Build exclude map
	excludeMap := make(map[int]struct{})
	if len(exceptFilters) > 0 {
		i.scanner.FindExcludeRecordsMap(exceptFilters, excludeMap)
	}

	// Return all values if no filters and no input
	if len(filters) == 0 && len(input) == 0 {
		var result map[string]map[string]interface{}
		if countValues {
			result = i.scanner.GetAllValuesCount(excludeMap)
		} else {
			result = i.scanner.GetAllValues(excludeMap)
		}

		if sortConfig != nil {
			sortAggregarion(sortConfig, result)
		}
		return result, nil
	}

	// Convert input records to map
	var inputMap map[int]struct{}
	if len(input) > 0 {
		inputMap = mapInputArray(input)
	}

	filteredRecords := make(map[int]struct{})
	resultCache := make(map[string]map[int]struct{})

	var err error

	if len(filters) > 0 {
		// Optimize filter order
		if len(filters) > 1 {
			filters, err = sortFilters(i.storage, filters)
			if err != nil {
				return nil, err
			}
		}

		// Index filters by field and cache results
		for _, f := range filters {
			name := f.GetFieldName()
			res, err := i.scanner.FindRecordsMap([]FilterInterface{f}, inputMap, excludeMap)
			if err != nil {
				return nil, err
			}
			resultCache[name] = res
		}

		// Merge results
		filteredRecords = mergeFilters(resultCache, "")
	} else if len(inputMap) > 0 {
		res, err := i.scanner.FindRecordsMap([]FilterInterface{}, inputMap, excludeMap)
		if err != nil {
			return nil, err
		}
		filteredRecords = res
	}

	// Intersect index values and filtered records
	result, err := i.scanner.AggregationScan(
		resultCache,
		filteredRecords,
		countValues,
		inputMap,
		excludeMap,
		q.HasSelfFiltering(),
		filters,
	)
	if err != nil {
		return nil, err
	}

	if sortConfig != nil {
		sortAggregarion(sortConfig, result)
	}
	return result, nil
}

// GetStorage returns index storage.
func (i *Db) GetStorage() StorageInterface {
	return i.storage
}

// GetStorage returns index storage.
func (i *Db) GetScanner() ScannerInterface {
	return i.scanner
}

// mapInputArray converts input array to map.
func mapInputArray(inputRecords []int) map[int]struct{} {
	input := make(map[int]struct{})
	for _, v := range inputRecords {
		input[v] = struct{}{}
	}
	return input
}

// mergeFilters merges filter results using optimized intersection.
// Iterates over the smallest map for better performance.
func mergeFilters(maps map[string]map[int]struct{}, skipKey string) map[int]struct{} {
	// Find the smallest map to start with
	var smallestKey string
	smallestSize := int(^uint(0) >> 1)

	for key, mapData := range maps {
		if skipKey != "" && key == skipKey {
			continue
		}
		if len(mapData) < smallestSize {
			smallestSize = len(mapData)
			smallestKey = key
		}
	}

	if smallestKey == "" {
		return make(map[int]struct{})
	}

	// Start with the smallest map
	result := make(map[int]struct{}, smallestSize)
	for k, v := range maps[smallestKey] {
		result[k] = v
	}

	// Intersect with other maps
	for key, mapData := range maps {
		if key == smallestKey || (skipKey != "" && key == skipKey) {
			continue
		}

		for k := range result {
			if _, ok := mapData[k]; !ok {
				delete(result, k)
			}
		}

		// Early exit if result is empty
		if len(result) == 0 {
			return result
		}
	}

	return result
}

func sortFilters(storage StorageInterface, filters []FilterInterface) ([]FilterInterface, error) {

	type filterWithIndex struct {
		index int
		count int
	}

	counts := make([]filterWithIndex, len(filters))

	for i, flt := range filters {
		// Non-ValueFilter types get max priority
		vf, ok := flt.(*ValueFilter)
		if !ok {
			counts[i] = filterWithIndex{index: i, count: math.MaxInt}
			continue
		}

		fieldName := vf.GetFieldName()

		if !storage.HasField(fieldName) {
			counts[i] = filterWithIndex{index: i, count: 0}
			continue
		}

		filterValues, err := storage.GetValueConverter().ValueToStringSlice(vf.GetValue())
		if err != nil {
			return filters, err
		}
		filterValuesCount := make(map[interface{}]int)
		valuesInFilter := len(filterValues)
		count := 0

		for _, value := range filterValues {
			cnt := storage.GetRecordsCount(fieldName, value)
			if valuesInFilter > 1 {
				filterValuesCount[value] = cnt
			}

			if cnt > count {
				count = cnt
			}
		}

		counts[i] = filterWithIndex{index: i, count: count}

		// Sort filter values by records count
		if valuesInFilter > 1 {
			sortedValues := make([]interface{}, 0, len(filterValuesCount))
			for value := range filterValuesCount {
				sortedValues = append(sortedValues, value)
			}
			sort.Slice(sortedValues, func(i, j int) bool {
				return filterValuesCount[sortedValues[i]] < filterValuesCount[sortedValues[j]]
			})
			vf.SetValue(sortedValues)
		}
	}

	// Sort filters by count
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count < counts[j].count
	})

	// Build result
	result := make([]FilterInterface, len(counts))
	for i, fi := range counts {
		result[i] = filters[fi.index]
	}

	return result, nil
}

// Sort sorts aggregation result fields and values.
// result: map[fieldName]map[fieldValue]count|true
func sortAggregarion(sortConfig *AggregationSort, result map[string]map[string]interface{}) {

	// Sort outer map keys (field names)
	fieldNames := make([]string, 0, len(result))
	for fieldName := range result {
		fieldNames = append(fieldNames, fieldName)
	}

	if sortConfig.FieldDirection == SortAsc {
		slices.SortStableFunc(fieldNames, func(i, j string) int {
			return strings.Compare(i, j)
		})
	} else {
		slices.SortStableFunc(fieldNames, func(i, j string) int {
			return strings.Compare(j, i)
		})

	}

	// Rebuild result in sorted order
	sortedResult := make(map[string]map[string]interface{})
	for _, fieldName := range fieldNames {
		values := result[fieldName]
		sortedValues := sortValues(values, false)
		sortedResult[fieldName] = sortedValues
	}

	// Copy back to result
	for k := range result {
		delete(result, k)
	}
	for k, v := range sortedResult {
		result[k] = v
	}
}

// sortValuesAscending sorts values in ascending order.
func sortValues(values map[string]interface{}, reverse bool) map[string]interface{} {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	if reverse {
		slices.SortStableFunc(keys, func(i, j string) int {
			return strings.Compare(j, i)
		})
	} else {
		slices.SortStableFunc(keys, func(i, j string) int {
			return strings.Compare(i, j)
		})
	}

	sortedValues := make(map[string]interface{})
	for _, k := range keys {
		sortedValues[k] = values[k]
	}
	return sortedValues
}

// Sort sorts results by field value.
func sortQuery(values map[string][]int, resultsMap map[int]struct{}, order *Sort) []int {

	// Determine sort type and create typed slice for efficient sorting

	sortedValues := make([]string, len(values))
	for k := range values {
		sortedValues = append(sortedValues, k)
	}

	if order.GetDirection() == SortAsc {
		slices.SortStableFunc(sortedValues, func(i, j string) int {
			return strings.Compare(i, j)
		})

	} else {
		// Descending order
		slices.SortStableFunc(sortedValues, func(i, j string) int {
			return strings.Compare(j, i)
		})
	}

	// Build sorted result
	sorted := make([]int, 0)

	for _, value := range sortedValues {
		records := values[value]
		if order.GetDirection() == SortAsc {
			for _, recId := range records {
				if _, ok := resultsMap[recId]; ok {
					sorted = append(sorted, recId)
					delete(resultsMap, recId)
				}
			}
		} else {
			// Reverse order for descending
			for i := len(records) - 1; i >= 0; i-- {
				recId := records[i]
				if _, ok := resultsMap[recId]; ok {
					sorted = append(sorted, recId)
					delete(resultsMap, recId)
				}
			}
		}
	}

	return sorted
}
