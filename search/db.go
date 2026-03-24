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

	if len(inputRecords) > 0 {
		sort.Slice(inputRecords, func(i, j int) bool { return inputRecords[i] < inputRecords[j] })
	}

	var err error
	// Optimize filter order - process filters with fewer matches first
	if len(inputRecords) == 0 && len(filters) > 1 {
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
	res, err := i.scanner.FindRecords(filters, inputRecords, excludeMap)
	if err != nil {
		return nil, err
	}

	// Sort if needed

	if order != nil && len(res) > 1 {
		velueMap := i.scanner.GetFieldValueRecords(order.GetField())
		result := sortQuery(velueMap, res, order)
		return result, nil
	}

	return res, nil
}

// Aggregate finds acceptable filter values.
func (i *Db) Aggregate(q *AggregationQuery) ([]*AggregationResultField, error) {
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

	var result []*AggregationResultField

	// Return all values if no filters and no input
	if len(filters) == 0 && len(input) == 0 {

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

	if len(input) > 0 {
		sort.Slice(input, func(i, j int) bool { return input[i] < input[j] })
	}

	filteredRecords := make([]int, 0, 0)

	var err error

	if len(filters) > 0 {
		// Optimize filter order
		if len(filters) > 1 {
			filters, err = sortFilters(i.storage, filters)
			if err != nil {
				return nil, err
			}
		}

		// Merge results
		filteredRecords, err = i.scanner.FindRecords(filters, input, excludeMap)
		if err != nil {
			return nil, err
		}
	} else if len(input) > 0 {
		res, err := i.scanner.FindRecords([]FilterInterface{}, input, excludeMap)
		if err != nil {
			return nil, err
		}
		filteredRecords = res
	}

	// Intersect index values and filtered records
	result, err = i.scanner.AggregationScan(
		filteredRecords,
		countValues,
		input,
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

func sortFilters(storage StorageInterface, filters []FilterInterface) ([]FilterInterface, error) {

	type filterWithIndex struct {
		index int
		count int
	}

	counts := make([]filterWithIndex, len(filters))

	for i, flt := range filters {
		// Non-ValueFilter types get lowest priority
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
func sortAggregarion(sortConfig *AggregationSort, result []*AggregationResultField) {

	for _, v := range result {
		if sortConfig.ValueDirection == SortAsc {
			slices.SortFunc(v.Values, func(a, b *AggregationResultValue) int {
				return strings.Compare(a.Value, b.Value)
			})
		} else {
			slices.SortFunc(v.Values, func(a, b *AggregationResultValue) int {
				return strings.Compare(b.Value, a.Value)
			})
		}
	}

	if sortConfig.FieldDirection == SortAsc {
		slices.SortFunc(result, func(a, b *AggregationResultField) int {
			return strings.Compare(a.Field, b.Field)
		})
	} else {
		slices.SortFunc(result, func(a, b *AggregationResultField) int {
			return strings.Compare(b.Field, a.Field)
		})
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

func sortQuery(values map[string][]int, results []int, order *Sort) []int {

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
	processed := make(map[int]struct{}, len(results))
	for _, k := range results {
		processed[k] = struct{}{}
	}

	for _, value := range sortedValues {
		records := values[value]
		if order.GetDirection() == SortAsc {
			for _, recId := range records {

				if _, ok := processed[recId]; ok {
					sorted = append(sorted, recId)
					delete(processed, recId)
				}
			}
		} else {
			// Reverse order for descending
			for i := len(records) - 1; i >= 0; i-- {
				recId := records[i]
				if _, ok := processed[recId]; ok {
					sorted = append(sorted, recId)
					delete(processed, recId)
				}
			}
		}
	}

	return sorted
}
