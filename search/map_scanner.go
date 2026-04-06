package search

import (
	"slices"
	"sort"
)

// Scanner provides methods for scanning and filtering index data.
type MapScanner struct {
	storage *MapStorage
}

// NewScanner creates a new Scanner.
func NewMapScanner(storage *MapStorage) *MapScanner {
	return &MapScanner{
		storage: storage,
	}
}

// FindRecordsMap finds records by filters using optimized batch processing.
// All filters are processed in a single pass through the data.
func (sc *MapScanner) FindRecords(filters []FilterInterface, inputRecords []int, excludeRecords map[int]struct{}) ([]int, error) {
	var err error
	// Initialize inputRecords if nil
	if inputRecords == nil {
		inputRecords = make([]int, 0, 0)
	}

	// If no filters, find by input records only
	if len(filters) == 0 {
		total := sc.GetAllRecordId(inputRecords, excludeRecords)
		// Intersect with input records
		if len(excludeRecords) == 0 {
			if len(inputRecords) > 0 {
				data := IntersectSortedInt(total, inputRecords)
				data = Deduplicate(data)
				return data, nil
			} else {
				return total, nil
			}

		} else {
			list := make([]int, 0, len(total))
			for _, v := range total {
				if _, ok := excludeRecords[v]; !ok {
					list = append(list, v)
				}
			}
			return list, nil
		}
	}

	result := make([]int, 0, len(inputRecords))
	if len(inputRecords) > 0 {
		result = append(result, inputRecords...)
	}
	// Process each filter
	for _, f := range filters {

		// Break if storage has no filtered field
		if !sc.storage.HasField(f.GetFieldName()) {
			return make([]int, 0, 0), nil
		}

		// Apply filter recursively
		result, err = f.FilterInput(sc, result, excludeRecords)
		if err != nil {
			return nil, err
		}

		// Early exit if no matches
		if len(result) == 0 {
			return result, nil
		}
	}

	return result, nil
}

// FindExcludeRecordsMap finds records by exclude filters.
func (sc *MapScanner) FindExcludeRecordsMap(filters []ExcludeFilterInterface, excludeRecords map[int]struct{}) {

	if len(filters) == 0 {
		return
	}

	for _, f := range filters {
		f.AddExcluded(sc, excludeRecords)
	}
}

// GetAllRecordIdMap returns all record IDs from the index.
func (sc *MapScanner) GetAllRecordId(inputRecords []int, excludeRecords map[int]struct{}) []int {

	data := make(map[int]struct{}, sc.storage.GetTotalCount())
	result := make([]int, 0, sc.storage.GetTotalCount())

	hasInput := len(inputRecords) > 0
	hasExcluded := len(excludeRecords) > 0

	var inputMap map[int]struct{}
	if hasInput {
		inputMap = make(map[int]struct{}, len(inputRecords))
		for _, v := range inputRecords {
			inputMap[v] = struct{}{}
		}
	}

	for _, v := range sc.storage.GetData() {
		for _, list := range v {
			for _, recId := range list {
				if _, ok := data[recId]; !ok {
					if hasExcluded {
						if _, ok := excludeRecords[recId]; ok {
							continue
						}
					}
					if hasInput {
						if _, ok := inputMap[recId]; !ok {
							continue
						}
					}
					data[recId] = struct{}{}
					result = append(result, recId)
				}
			}
		}
	}
	slices.Sort(result)
	return result
}

// aggregationScan performs the aggregation scan.
func (sc *MapScanner) AggregationScan(
	countRecords bool,
	input []int,
	exclude map[int]struct{},
	selfFiltering bool,
	filters []FilterInterface,
) ([]*AggregationResultField, error) {

	result := make([]*AggregationResultField, 0, 0)

	// filter results cache
	filtersCache := make(map[string][]int, len(filters))
	filtersOrder := make([]string, 0, len(filters))

	// Index filters by field name
	indexedFilters := make(map[string]FilterInterface)
	for _, f := range filters {
		indexedFilters[f.GetFieldName()] = f
		recordIds, err := sc.FindRecords([]FilterInterface{f}, input, exclude)
		if err != nil {
			return result, err
		}

		filtersCache[f.GetFieldName()] = recordIds
		filtersOrder = append(filtersOrder, f.GetFieldName())
	}

	data := sc.storage.GetData()
	needSelfFiltering := selfFiltering
	if !needSelfFiltering {
		for filterName := range data {
			if f, ok := indexedFilters[filterName]; ok && f.HasSelfFiltering() {
				needSelfFiltering = true
				break
			}
		}
	}

	var filteredData []int
	var err error

	// Optimization for fields without filters
	filteredData, err = sc.FindRecords(filters, input, exclude)
	if err != nil {
		return nil, err
	}

	// Scan storage
	for filterName, filterValues := range sc.storage.GetData() {

		// Process filter values with optimized intersection
		fieldResult := make([]*AggregationResultValue, 0, len(filterValues))

		// fast path
		if len(filters) == 0 && len(input) == 0 {
			for val, list := range filterValues {
				fRes := &AggregationResultValue{Value: val}
				if countRecords {
					cnt := len(list)
					fRes.Count = &cnt
				}
				fieldResult = append(fieldResult, fRes)
			}
			continue
		}

		// Check if self-filtering is needed
		needSelfFiltering = selfFiltering
		if f, ok := indexedFilters[filterName]; ok && f.HasSelfFiltering() {
			needSelfFiltering = true
		}

		var recordIds []int
		_, hasFilter := indexedFilters[filterName]
		// Single filter - no need to merge
		if needSelfFiltering || !hasFilter {
			recordIds = filteredData
		} else {
			if len(filtersCache) > 1 {
				recordIds = mergeFiltersCache(filtersCache, filterName, filtersOrder)
			} else {
				recordIds, err = sc.FindRecords([]FilterInterface{}, input, exclude)
				if err != nil {
					return nil, err
				}
			}
		}

		if countRecords {
			for filterValue, data := range filterValues {
				intersectCount := countIntersectionSortedInt(data, recordIds)
				if intersectCount > 0 {
					fieldResult = append(fieldResult, &AggregationResultValue{Value: filterValue, Count: &intersectCount})
				}
			}
		} else {
			for filterValue, data := range filterValues {
				if hasIntersectionSortedInt(data, recordIds) {
					fieldResult = append(fieldResult, &AggregationResultValue{Value: filterValue})
				}
			}
		}

		if len(fieldResult) > 0 {
			result = append(result, &AggregationResultField{Field: filterName, Values: fieldResult})
		}
	}
	return result, nil
}

func mergeFiltersCache(cache map[string][]int, skipKey string, filtersOrder []string) []int {
	result := make([]int, 0, 100)
	start := true
	for _, name := range filtersOrder {
		if name == skipKey {
			continue
		}

		if start {
			result = append(result, cache[name]...)
			start = false
			continue
		}
		result = IntersectSortedInt(result, cache[name])
	}
	return result
}

// getValues returns all values from index.
func (sc *MapScanner) GetAllValues(excludeMap map[int]struct{}) []*AggregationResultField {

	data := sc.storage.GetData()
	result := make([]*AggregationResultField, 0, len(data))

	if len(excludeMap) == 0 {
		for filterName, filterValues := range sc.storage.GetData() {
			resField := &AggregationResultField{
				Field:  filterName,
				Values: make([]*AggregationResultValue, 0, len(filterValues)),
			}
			for key := range filterValues {
				resField.Values = append(resField.Values, &AggregationResultValue{Value: key})
			}
			result = append(result, resField)
		}
	} else {
		for filterName, filterValues := range sc.storage.GetData() {
			resField := &AggregationResultField{
				Field:  filterName,
				Values: make([]*AggregationResultValue, 0, len(filterValues)),
			}
			for key, list := range filterValues {
				for _, value := range list {
					if _, ok := excludeMap[value]; !ok {
						resField.Values = append(resField.Values, &AggregationResultValue{Value: key})
						break
					}
				}
			}
			result = append(result, resField)
		}
	}
	return result
}

// getValuesCount returns all values with their counts.
func (sc *MapScanner) GetAllValuesCount(excludeMap map[int]struct{}) []*AggregationResultField {

	data := sc.storage.GetData()
	result := make([]*AggregationResultField, 0, len(data))

	if len(excludeMap) == 0 {
		for filterName, filterValues := range sc.storage.GetData() {
			resField := &AggregationResultField{
				Field:  filterName,
				Values: make([]*AggregationResultValue, 0, len(filterValues)),
			}
			for key, list := range filterValues {
				cLen := len(list)
				resField.Values = append(resField.Values, &AggregationResultValue{Value: key, Count: &cLen})
			}
			result = append(result, resField)
		}
	} else {
		for filterName, filterValues := range sc.storage.GetData() {
			resField := &AggregationResultField{
				Field:  filterName,
				Values: make([]*AggregationResultValue, 0, len(filterValues)),
			}
			for key, list := range filterValues {
				count := 0
				for _, value := range list {
					if _, ok := excludeMap[value]; !ok {
						count++
					}
				}
				resField.Values = append(resField.Values, &AggregationResultValue{Value: key, Count: &count})
			}
			result = append(result, resField)
		}
	}
	return result
}
func (sc *MapScanner) GetFieldValueRecords(field string) map[string][]int {
	return sc.storage.GetFieldData(field)
}

func (sc *MapScanner) GetSortedFieldValues(field string) []string {

	data := sc.storage.GetFieldData(field)
	list := make([]string, 0, len(data))
	for k := range data {
		list = append(list, k)
	}

	// sort values
	slices.SortFunc(list, func(a, b string) int {
		return sc.storage.converter.CompareNumStrings(a, b)
	})
	return list
}

func (sc *MapScanner) IntersectFilterValues(field string, values interface{}, input []int, excludeRecords map[int]struct{}) (result []int, err error) {

	result = make([]int, 0, len(input))

	if !sc.storage.HasField(field) {
		return
	}

	val, err := sc.storage.converter.ValueToStringSlice(values)
	if err != nil {
		return
	}

	data := sc.storage.GetFieldData(field)

	var list []int

	for _, item := range val {
		records, ok := data[item]
		if !ok {
			continue
		}

		if len(excludeRecords) > 0 {
			list = make([]int, 0, len(records))
			for _, v := range records {
				if _, ok := excludeRecords[v]; !ok {
					list = append(list, v)
				}
			}
		} else {
			list = records
		}

		if len(input) > 0 {
			list = IntersectSortedInt(list, input)
			if len(list) > 1 {
				list = Deduplicate(list)
			}
		}

		if len(result) == 0 {
			result = append(result, list...)
		} else {
			result = MergeSortedSlices(result, list)
			if len(result) > 1 {
				result = Deduplicate(result)
			}
		}
	}
	return result, nil
}

func (sc *MapScanner) AddExcludedValues(field string, values interface{}, excludeRecords map[int]struct{}) (err error) {

	if !sc.storage.HasField(field) {
		return err
	}

	val, err := sc.storage.converter.ValueToStringSlice(values)
	if err != nil {
		return err
	}

	data := sc.storage.GetFieldData(field)

	for _, item := range val {
		records, ok := data[item]
		if !ok {
			continue
		}

		for _, recId := range records {
			excludeRecords[recId] = struct{}{}
		}
	}
	return nil
}

func (sc *MapScanner) FindRangeIntersection(field string, value *RangeValue, limitRecords []int, excludeRecords map[int]struct{}) (result []int, err error) {

	if !sc.storage.HasField(field) {
		return
	}

	var minValue string
	var maxValue string

	hasMin := false
	hasMax := false

	if value.Min != nil {
		minValue, err = sc.storage.converter.GetValueString(value.Min)
		if err != nil {
			return
		}
		hasMin = true
	}

	if value.Max != nil {
		maxValue, err = sc.storage.converter.GetValueString(value.Max)
		if err != nil {
			return
		}
		hasMax = true
	}

	data := sc.GetFieldValueRecords(field)
	sortedValues := sc.GetSortedFieldValues(field)

	var limitMap map[int]struct{}
	var list map[int]struct{}

	hasExclude := len(excludeRecords) > 0
	hasLimit := len(limitRecords) > 0

	if hasLimit {
		limitMap = make(map[int]struct{}, len(limitRecords))
		for _, v := range limitRecords {
			limitMap[v] = struct{}{}
		}
	}

	for _, value := range sortedValues {

		if hasMin && sc.storage.converter.CompareNumStrings(value, minValue) == -1 {
			continue
		}

		if hasMax && sc.storage.converter.CompareNumStrings(value, maxValue) == 1 {
			break
		}

		records := data[value]
		if list == nil {
			list = make(map[int]struct{}, len(records))
		}

		for _, recId := range records {

			if _, ok := list[recId]; ok {
				continue
			}

			if hasExclude {
				if _, ok := excludeRecords[recId]; ok {
					continue
				}
			}
			if hasLimit {
				if _, ok := limitMap[recId]; !ok {
					continue
				}
			}
			list[recId] = struct{}{}
		}
	}

	res := make([]int, 0, len(list))
	for k := range list {
		res = append(res, k)
	}
	slices.Sort(res)

	return res, nil
}

func (sc *MapScanner) FindValueIntersection(field string, values interface{}, inputRecords []int, excludeRecords map[int]struct{}) (result []int, err error) {

	result = make([]int, 0, len(inputRecords))

	if !sc.storage.HasField(field) {
		return
	}

	val, err := sc.storage.converter.ValueToStringSlice(values)
	if err != nil {
		return
	}
	var list []int
	data := sc.storage.GetFieldData(field)

	for _, item := range val {
		records, ok := data[item]
		if !ok {
			result = make([]int, 0, 0)
			return result, nil
		}

		if len(excludeRecords) > 0 {
			for _, v := range records {
				if _, ok := excludeRecords[v]; !ok {
					list = append(list, v)
				}
			}
		} else {
			list = records
		}

		if len(inputRecords) > 0 {
			list = IntersectSortedInt(list, inputRecords)
			if len(list) > 1 {
				list = Deduplicate(list)
			}
		}

		if len(list) == 0 {
			return []int{}, nil
		}

		if len(result) > 0 {
			result = IntersectSortedInt(result, list)
		} else {
			result = append(result, list...)
		}
	}
	return result, nil
}

func (sc *MapScanner) FindInRange(field string, value *RangeValue) (result []int, err error) {

	result = make([]int, 0, 0)

	if !sc.storage.HasField(field) {
		return
	}

	if value.Min == value.Max {
		return
	}

	var minValue string
	var maxValue string

	hasMin := false
	hasMax := false

	if value.Min != nil {
		minValue, err = sc.storage.converter.GetValueString(value.Min)
		if err != nil {
			return
		}
		hasMin = true
	}

	if value.Max != nil {
		maxValue, err = sc.storage.converter.GetValueString(value.Max)
		if err != nil {
			return
		}
		hasMax = true
	}

	data := sc.GetFieldValueRecords(field)
	sortedValues := sc.GetSortedFieldValues(field)

	for _, value := range sortedValues {
		records := data[value]
		if hasMin && sc.storage.converter.CompareNumStrings(value, minValue) == -1 {
			continue
		}

		if hasMax && sc.storage.converter.CompareNumStrings(value, maxValue) == 1 {
			break
		}

		if len(result) == 0 {
			result = append(result, records...)
		} else {
			result = MergeSortedSlices(result, records)
			result = Deduplicate(result)
		}
	}
	return result, nil
}

// IntersectSortedInt intersect sorted int slices
func IntersectSortedInt(a, b []int) []int {
	if len(a) == 0 || len(b) == 0 {
		return []int{}
	}

	compareCount := len(b)
	comparePointer := 0

	result := make([]int, 0, 100)

	for _, value := range a {

		if comparePointer >= compareCount {
			break
		}

		if value < b[comparePointer] {
			continue
		}
		for ; comparePointer < compareCount; comparePointer++ {
			if b[comparePointer] < value {
				continue
			}

			if b[comparePointer] == value {
				result = append(result, value)
				break
			}

			if b[comparePointer] > value {
				break
			}
		}
	}
	return result
}

func IntersectMapInt(a map[int]struct{}, b []int) []int {
	res := make([]int, 0, len(b))
	for _, v := range b {
		if _, ok := a[v]; ok {
			res = append(res, v)
		}
	}
	return res
}

func MapAddValues(data map[int]struct{}, values []int) {
	for _, v := range values {
		data[v] = struct{}{}
	}
}

// Deduplicate - remove duplicates from int slice
func Deduplicate(in []int) []int {
	sort.Slice(in, func(i, j int) bool { return in[i] < in[j] })
	// In-place deduplicate https://github.com/golang/go/wiki/SliceTricks
	j := 0
	for i := 1; i < len(in); i++ {
		if in[j] == in[i] {
			continue
		}
		j++
		// preserve the original data
		// in[i], in[j] = in[j], in[i]
		// only set what is required
		in[j] = in[i]
	}
	result := in[:j+1]
	return result
}

func MergeSortedSlices(a, b []int) []int {
	res := make([]int, 0, len(a)+len(b)) // Предварительное выделение памяти критично для скорости
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			res = append(res, a[i])
			i++
		} else {
			res = append(res, b[j])
			j++
		}
	}
	// Добавляем остатки (один из слайсов гарантированно пуст)
	res = append(res, a[i:]...)
	return append(res, b[j:]...)
}

// IntersectCountSortedInt get intersect count for sorted int slices
func countIntersectionSortedInt(a, b []int) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	result := 0

	compareCount := len(b)
	comparePointer := 0

	for _, value := range a {
		if comparePointer >= compareCount {
			break
		}
		if value < b[comparePointer] {
			continue
		}
		for ; comparePointer < compareCount; comparePointer++ {
			if b[comparePointer] < value {
				continue
			}

			if b[comparePointer] == value {
				result++
				break
			}

			if b[comparePointer] > value {
				break
			}
		}
	}
	return result
}

// IntersectCountSortedInt get intersect count for sorted int slices
func hasIntersectionSortedInt(a, b []int) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	compareCount := len(b)
	comparePointer := 0

	for _, value := range a {
		if comparePointer >= compareCount {
			break
		}
		if value < b[comparePointer] {
			continue
		}
		for ; comparePointer < compareCount; comparePointer++ {
			if b[comparePointer] < value {
				continue
			}

			if b[comparePointer] == value {
				return true
			}

			if b[comparePointer] > value {
				break
			}
		}
	}
	return false
}
