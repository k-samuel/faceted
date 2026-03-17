package index

import (
	"fmt"

	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/intersection"
	"github.com/k-samuel/faceted/pkg/query"
	"github.com/k-samuel/faceted/pkg/sort"
	"github.com/k-samuel/faceted/pkg/storage"
)

// Index implements IndexInterface for faceted search.
type Index struct {
	storage         storage.StorageInterface
	filterSort      *sort.Filters
	aggregationSort *sort.AggregationResults
	querySort       sort.QueryResultsInterface
	scanner         *storage.Scanner
	intersection    intersection.IntersectionInterface
	profiler        *Profile
}

// NewIndex creates a new Index.
func NewIndex(
	storage storage.StorageInterface,
	filterSort *sort.Filters,
	aggregationSort *sort.AggregationResults,
	querySort sort.QueryResultsInterface,
	scanner *storage.Scanner,
	intersection intersection.IntersectionInterface,
) *Index {
	return &Index{
		storage:         storage,
		filterSort:      filterSort,
		aggregationSort: aggregationSort,
		querySort:       querySort,
		scanner:         scanner,
		intersection:    intersection,
	}
}

// Query finds records using Query.
func (i *Index) Query(q *query.SearchQuery) []int {
	inputRecords := q.GetInRecords()
	filterList := q.GetFilters()

	filters := make([]filter.FilterInterface, 0, len(filterList))
	exceptFilters := make([]filter.ExcludeFilterInterface, 0)

	for _, item := range filterList {
		if ef, ok := item.(filter.ExcludeFilterInterface); ok {
			exceptFilters = append(exceptFilters, ef)
		} else {
			filters = append(filters, item)
		}
	}

	order := q.GetSort()

	// Convert input records to map
	var inputMap map[int]bool
	if len(inputRecords) > 0 {
		inputMap = i.mapInputArray(inputRecords)
	}

	// Optimize filter order - process filters with fewer matches first
	if len(inputMap) == 0 && len(filters) > 1 {
		filters = i.filterSort.ByCount(i.storage, filters)
	}

	// Build exclude map
	excludeMap := make(map[int]bool)
	if len(exceptFilters) > 0 {
		i.scanner.FindExcludeRecordsMap(i.storage, exceptFilters, &excludeMap)
	}

	// Find records
	mapResult := i.scanner.FindRecordsMap(i.storage, filters, inputMap, excludeMap)

	// Sort if needed
	if order != nil {
		result := i.querySort.Sort(i.storage, mapResult, order)
		return result
	}

	// Convert map to slice with pre-allocated capacity
	result := make([]int, 0, len(mapResult))
	for k := range mapResult {
		result = append(result, k)
	}
	return result
}

// Aggregate finds acceptable filter values.
func (i *Index) Aggregate(q *query.AggregationQuery) map[string]map[string]interface{} {
	input := q.GetInRecords()
	filterList := q.GetFilters()

	filters := make([]filter.FilterInterface, 0)
	exceptFilters := make([]filter.ExcludeFilterInterface, 0)

	for _, item := range filterList {
		if ef, ok := item.(filter.ExcludeFilterInterface); ok {
			exceptFilters = append(exceptFilters, ef)
		} else {
			filters = append(filters, item)
		}
	}

	countValues := q.GetCountItems()
	sortConfig := q.GetSort()

	// Build exclude map
	excludeMap := make(map[int]bool)
	if len(exceptFilters) > 0 {
		i.scanner.FindExcludeRecordsMap(i.storage, exceptFilters, &excludeMap)
	}

	// Return all values if no filters and no input
	if len(filters) == 0 && len(input) == 0 {
		var result map[string]map[string]interface{}
		if countValues {
			result = i.getValuesCount(excludeMap)
		} else {
			result = i.getValues(excludeMap)
		}

		if sortConfig != nil {
			i.aggregationSort.Sort(sortConfig, result)
		}
		return result
	}

	// Convert input records to map
	var inputMap map[int]bool
	if len(input) > 0 {
		inputMap = i.mapInputArray(input)
	}

	filteredRecords := make(map[int]bool)
	resultCache := make(map[string]map[int]bool)

	if len(filters) > 0 {
		// Optimize filter order
		if len(filters) > 1 {
			filters = i.filterSort.ByCount(i.storage, filters)
		}

		// Index filters by field and cache results
		for _, f := range filters {
			name := f.GetFieldName()
			resultCache[name] = i.scanner.FindRecordsMap(i.storage, []filter.FilterInterface{f}, inputMap, excludeMap)
		}

		// Merge results
		filteredRecords = i.mergeFilters(resultCache, "")
	} else if len(inputMap) > 0 {
		filteredRecords = i.scanner.FindRecordsMap(i.storage, []filter.FilterInterface{}, inputMap, excludeMap)
	}

	// Intersect index values and filtered records
	result := i.aggregationScan(
		resultCache,
		filteredRecords,
		countValues,
		inputMap,
		excludeMap,
		q.HasSelfFiltering(),
		filters,
	)

	if sortConfig != nil {
		i.aggregationSort.Sort(sortConfig, result)
	}
	return result
}

// aggregationScan performs the aggregation scan.
// Optimized version with minimal allocations.
func (i *Index) aggregationScan(
	resultCache map[string]map[int]bool,
	filteredRecords map[int]bool,
	countRecords bool,
	input map[int]bool,
	exclude map[int]bool,
	selfFiltering bool,
	filters []filter.FilterInterface,
) map[string]map[string]interface{} {

	result := make(map[string]map[string]interface{})
	cacheCount := len(resultCache)

	// Index filters by field name
	indexedFilters := make(map[string]filter.FilterInterface)
	for _, f := range filters {
		indexedFilters[f.GetFieldName()] = f
	}

	// Scan storage
	for kv := range i.storage.Scan() {
		filterName := kv.Key
		filterValues := kv.Value

		// Check if self-filtering is needed
		needSelfFiltering := selfFiltering
		if f, ok := indexedFilters[filterName]; ok && f.HasSelfFiltering() {
			needSelfFiltering = true
		}

		var recordIds map[int]bool
		if _, ok := resultCache[filterName]; ok {
			// Use cached result
			if cacheCount > 1 {
				if needSelfFiltering {
					// For self-filtering, include all filters (no skip)
					recordIds = i.mergeFilters(resultCache, "")
				} else {
					// For non-self-filtering, exclude current field's cache
					recordIds = i.mergeFilters(resultCache, filterName)
				}
			} else {
				// Single filter - no need to merge
				if needSelfFiltering {
					recordIds = i.scanner.FindRecordsMap(i.storage, filters, input, exclude)
				} else {
					recordIds = i.scanner.FindRecordsMap(i.storage, []filter.FilterInterface{}, input, exclude)
				}
			}
		} else {
			recordIds = filteredRecords
		}

		// Process filter values with optimized intersection
		fieldResult := make(map[string]interface{}, len(filterValues))

		if countRecords {
			for filterValue, data := range filterValues {
				intersectCount := i.intersection.GetIntersectMapCount(data, recordIds)
				if intersectCount > 0 {
					fieldResult[fmt.Sprintf("%v", filterValue)] = intersectCount
				}
			}
		} else {
			for filterValue, data := range filterValues {
				if i.intersection.HasIntersectIntMap(data, recordIds) {
					fieldResult[fmt.Sprintf("%v", filterValue)] = true
				}
			}
		}

		if len(fieldResult) > 0 {
			result[filterName] = fieldResult
		}
	}
	return result
}

// getValues returns all values from index.
func (i *Index) getValues(excludeMap map[int]bool) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	if len(excludeMap) == 0 {
		for kv := range i.storage.Scan() {
			filterName := kv.Key
			filterValues := kv.Value
			result[filterName] = make(map[string]interface{})
			for key := range filterValues {
				result[filterName][key] = true
			}
		}
	} else {
		for kv := range i.storage.Scan() {
			filterName := kv.Key
			filterValues := kv.Value
			result[filterName] = make(map[string]interface{})
			for key, list := range filterValues {
				for _, value := range list {
					if !excludeMap[value] {
						result[filterName][fmt.Sprintf("%v", key)] = true
						break
					}
				}
			}
		}
	}
	return result
}

// getValuesCount returns all values with their counts.
func (i *Index) getValuesCount(excludeMap map[int]bool) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	if len(excludeMap) == 0 {
		for kv := range i.storage.Scan() {
			filterName := kv.Key
			filterValues := kv.Value
			result[filterName] = make(map[string]interface{})
			for key, list := range filterValues {
				result[filterName][fmt.Sprintf("%v", key)] = len(list)
			}
		}
	} else {
		for kv := range i.storage.Scan() {
			filterName := kv.Key
			filterValues := kv.Value
			result[filterName] = make(map[string]interface{})
			for key, list := range filterValues {
				count := 0
				for _, value := range list {
					if !excludeMap[value] {
						count++
					}
				}
				result[filterName][fmt.Sprintf("%v", key)] = count
			}
		}
	}
	return result
}

// mergeFilters merges filter results using optimized intersection.
// Iterates over the smallest map for better performance.
func (i *Index) mergeFilters(maps map[string]map[int]bool, skipKey string) map[int]bool {
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
		return make(map[int]bool)
	}

	// Start with the smallest map
	result := make(map[int]bool, smallestSize)
	for k, v := range maps[smallestKey] {
		result[k] = v
	}

	// Intersect with other maps
	for key, mapData := range maps {
		if key == smallestKey || (skipKey != "" && key == skipKey) {
			continue
		}

		for k := range result {
			if !mapData[k] {
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

// mapInputArray converts input array to map.
func (i *Index) mapInputArray(inputRecords []int) map[int]bool {
	input := make(map[int]bool)
	for _, v := range inputRecords {
		input[v] = true
	}
	return input
}

// GetCount returns count of unique records.
func (i *Index) GetCount() int {
	return len(i.scanner.GetAllRecordIdMap(i.storage))
}

// SetProfiler sets time profiler.
func (i *Index) SetProfiler(profiler *Profile) {
	i.profiler = profiler
}

// GetStorage returns index storage.
func (i *Index) GetStorage() storage.StorageInterface {
	return i.storage
}

// GetScanner returns index scanner.
func (i *Index) GetScanner() *storage.Scanner {
	return i.scanner
}

// SetData loads saved data.
func (i *Index) SetData(data map[string]map[string][]int) {
	i.storage.SetData(data)
}

// Export exports facet index data.
func (i *Index) Export() map[string]map[string][]int {
	return i.storage.Export()
}

// Optimize optimizes index structure.
func (i *Index) Optimize() {
	i.storage.Optimize()
}
