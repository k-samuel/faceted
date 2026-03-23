package search

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
func (sc *MapScanner) FindRecordsMap(filters []FilterInterface, inputRecords map[int]struct{}, excludeRecords map[int]struct{}) (map[int]struct{}, error) {
	// Initialize inputRecords if nil
	if inputRecords == nil {
		inputRecords = make(map[int]struct{})
	}

	// If no filters, find by input records only
	if len(filters) == 0 {
		return sc.FindInput(inputRecords, excludeRecords), nil
	}

	// Process each filter
	for _, f := range filters {
		if !sc.storage.HasField(f.GetFieldName()) {
			return make(map[int]struct{}), nil
		}

		// Apply filter
		err := f.FilterInput(sc, inputRecords, excludeRecords)
		if err != nil {
			return nil, err
		}

		// Early exit if no matches
		if len(inputRecords) == 0 {
			return inputRecords, nil
		}
	}

	return inputRecords, nil
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

// FindInput finds records without filters.
func (sc *MapScanner) FindInput(inputRecords map[int]struct{}, excludeRecords map[int]struct{}) map[int]struct{} {

	total := sc.GetAllRecordIdMap()

	if len(inputRecords) == 0 && len(excludeRecords) == 0 {
		return total
	}

	// Intersect with input records
	if len(inputRecords) > 0 {
		intersected := make(map[int]struct{}, minInt(len(total), len(inputRecords)))
		if len(total) > len(inputRecords) {
			for recId := range inputRecords {
				if _, ok := total[recId]; ok {
					intersected[recId] = struct{}{}
				}
			}
		} else {
			for recId := range total {
				if _, ok := inputRecords[recId]; ok {
					intersected[recId] = struct{}{}
				}
			}
		}
		total = intersected
	}

	// Remove excluded records
	if len(excludeRecords) > 0 {
		if len(total) > len(excludeRecords) {
			for recId := range excludeRecords {
				delete(total, recId)
			}
		} else {
			for recId := range total {
				if _, ok := excludeRecords[recId]; ok {
					delete(total, recId)
				}
			}
		}
	}

	return total
}

// GetAllRecordIdMap returns all record IDs from the index.
func (sc *MapScanner) GetAllRecordIdMap() map[int]struct{} {
	data := sc.storage.GetData()

	// Estimate capacity
	totalRecords := 0
	for _, values := range data {
		for _, list := range values {
			totalRecords += len(list)
			break
		}
		break
	}

	// Count unique fields for better estimation
	fieldCount := len(data)
	if fieldCount > 0 {
		totalRecords /= fieldCount
	}

	result := make(map[int]struct{}, totalRecords)

	for _, v := range sc.storage.GetData() {
		for _, list := range v {
			for _, recId := range list {
				result[recId] = struct{}{}
			}
		}
	}

	return result
}

// aggregationScan performs the aggregation scan.
func (sc *MapScanner) AggregationScan(
	resultCache *ResultCache,
	filteredRecords map[int]struct{},
	countRecords bool,
	input map[int]struct{},
	exclude map[int]struct{},
	selfFiltering bool,
	filters []FilterInterface,
) (map[string]map[string]interface{}, error) {

	result := make(map[string]map[string]interface{})
	cacheCount := len(resultCache.Filters)

	// Index filters by field name
	indexedFilters := make(map[string]FilterInterface)
	for _, f := range filters {
		indexedFilters[f.GetFieldName()] = f
	}

	// Scan storage
	for filterName, filterValues := range sc.storage.GetData() {

		// Check if self-filtering is needed
		needSelfFiltering := selfFiltering
		if f, ok := indexedFilters[filterName]; ok && f.HasSelfFiltering() {
			needSelfFiltering = true
		}

		var recordIds map[int]struct{}
		var err error
		if _, ok := resultCache.Filters[filterName]; ok {
			// Use cached result
			if cacheCount > 1 {
				if needSelfFiltering {
					// For self-filtering, include all filters (no skip)
					recordIds = mergeFilters(resultCache, "")
				} else {
					// For non-self-filtering, exclude current field's cache
					recordIds = mergeFilters(resultCache, filterName)
				}
			} else {
				// Single filter - no need to merge
				if needSelfFiltering {
					recordIds, err = sc.FindRecordsMap(filters, input, exclude)
				} else {
					recordIds, err = sc.FindRecordsMap([]FilterInterface{}, input, exclude)
				}
				if err != nil {
					return nil, err
				}
			}
		} else {
			recordIds = filteredRecords
		}

		// Process filter values with optimized intersection
		fieldResult := make(map[string]interface{}, len(filterValues))

		if countRecords {
			for filterValue, data := range filterValues {
				intersectCount := getIntersectMapCount(data, recordIds)
				if intersectCount > 0 {
					fieldResult[filterValue] = intersectCount
				}
			}
		} else {
			for filterValue, data := range filterValues {
				if hasIntersectIntMap(data, recordIds) {
					fieldResult[filterValue] = true
				}
			}
		}

		if len(fieldResult) > 0 {
			result[filterName] = fieldResult
		}
	}
	return result, nil
}

// getValues returns all values from index.
func (sc *MapScanner) GetAllValues(excludeMap map[int]struct{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	if len(excludeMap) == 0 {
		for filterName, filterValues := range sc.storage.GetData() {
			result[filterName] = make(map[string]interface{})
			for key := range filterValues {
				result[filterName][key] = true
			}
		}
	} else {
		for filterName, filterValues := range sc.storage.GetData() {
			result[filterName] = make(map[string]interface{})
			for key, list := range filterValues {
				for _, value := range list {
					if _, ok := excludeMap[value]; !ok {
						result[filterName][key] = struct{}{}
						break
					}
				}
			}
		}
	}
	return result
}

// getValuesCount returns all values with their counts.
func (sc *MapScanner) GetAllValuesCount(excludeMap map[int]struct{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	if len(excludeMap) == 0 {
		for filterName, filterValues := range sc.storage.GetData() {

			result[filterName] = make(map[string]interface{})
			for key, list := range filterValues {
				result[filterName][key] = len(list)
			}
		}
	} else {
		for filterName, filterValues := range sc.storage.GetData() {
			result[filterName] = make(map[string]interface{})
			for key, list := range filterValues {
				count := 0
				for _, value := range list {
					if _, ok := excludeMap[value]; !ok {
						count++
					}
				}
				result[filterName][key] = count
			}
		}
	}
	return result
}
func (sc *MapScanner) GetFieldValueRecords(field string) map[string][]int {
	return sc.storage.GetFieldData(field)
}

func (sc *MapScanner) FindIntersection(field string, values interface{}, inputRecords map[int]struct{}) (result map[int]struct{}, err error) {
	result = make(map[int]struct{})

	if !sc.storage.HasField(field) {
		return
	}

	val, err := sc.storage.converter.ValueToStringSlice(values)
	if err != nil {
		return nil, err
	}

	data := sc.storage.GetFieldData(field)

	for _, item := range val {
		records, ok := data[item]
		if !ok {
			continue
		}

		for _, recId := range records {

			_, inInput := inputRecords[recId]
			if inInput {
				result[recId] = struct{}{} // Mark matching entries
			}
		}
	}
	return result, nil
}

func (sc *MapScanner) FindRangeIntersection(field string, value *RangeValue, inputRecords map[int]struct{}, excludeRecords map[int]struct{}) (err error) {

	if !sc.storage.HasField(field) {
		return
	}

	emptyExclude := len(excludeRecords) == 0

	var limit []int

	var minValue string
	var maxValue string

	hasMin := false
	hasMax := false

	if value.Min != nil {
		minValue, err = sc.storage.converter.GetValueString(value.Min)
		if err != nil {
			return err
		}
		hasMin = true
	}

	if value.Max != nil {
		maxValue, err = sc.storage.converter.GetValueString(value.Max)
		if err != nil {
			return err
		}
		hasMax = true
	}

	data := sc.storage.GetFieldData(field)

	for value, records := range data {

		if hasMin && sc.storage.converter.CompareNumStrings(value, minValue) == -1 {
			continue
		}

		if hasMax && sc.storage.converter.CompareNumStrings(value, maxValue) == 1 {
			continue
		}

		if emptyExclude {
			limit = append(limit, records...)
		} else {
			for _, recId := range records {
				if _, ok := excludeRecords[recId]; !ok {
					limit = append(limit, recId)
				}
			}
		}
	}

	if len(limit) == 0 {
		for recId := range inputRecords {
			delete(inputRecords, recId)
		}
		return nil
	}

	emptyInput := len(inputRecords) == 0

	if emptyInput {
		// Fill inputIdKeys with limit records
		for _, recId := range limit {
			inputRecords[recId] = struct{}{}
		}
		return
	}

	// Use flag map for optimization (matching PHP implementation)
	// First pass: collect matching records into a map
	matchingMap := make(map[int]struct{})
	for _, recId := range limit {
		matchingMap[recId] = struct{}{}
	}

	// Second pass: mark matching records (O(n) instead of O(n*m))
	for recId := range inputRecords {
		if _, ok := matchingMap[recId]; !ok {
			delete(inputRecords, recId)
		}
	}

	return nil
}

func (sc *MapScanner) FindValueIntersection(field string, values interface{}, inputRecords map[int]struct{}) (result map[int]struct{}, err error) {

	if !sc.storage.HasField(field) {
		return
	}

	val, err := sc.storage.converter.ValueToStringSlice(values)
	if err != nil {
		return nil, err
	}

	data := sc.storage.GetFieldData(field)
	result = make(map[int]struct{})
	isFirst := true

	for _, item := range val {
		records, ok := data[item]
		if !ok {
			// No records for this value - intersection is empty
			for k := range inputRecords {
				delete(inputRecords, k)
			}
			return
		}

		if isFirst {
			for _, recId := range records {
				result[recId] = struct{}{}
			}
			isFirst = false
		}

		tmp := make(map[int]struct{})
		for _, recId := range records {

			_, inResult := result[recId]
			_, inInput := inputRecords[recId]
			if inResult && inInput {
				tmp[recId] = struct{}{} // Mark matching entries
			}
		}
		result = tmp
	}
	return result, nil
}

func (sc *MapScanner) FindUniqueRecords(field string, values interface{}, resultRecords map[int]struct{}, excludeRecords map[int]struct{}) (err error) {

	if !sc.storage.HasField(field) {
		return
	}

	val, err := sc.storage.converter.ValueToStringSlice(values)
	if err != nil {
		return err
	}

	data := sc.storage.GetFieldData(field)
	emptyExclude := len(excludeRecords) == 0

	for _, item := range val {

		records, ok := data[item]
		if !ok {
			continue
		}

		// Fast fill for first value with no excludes
		if len(resultRecords) == 0 && emptyExclude {
			for _, recId := range records {
				resultRecords[recId] = struct{}{}
			}
			continue
		}

		for _, recId := range records {
			_, inExluded := excludeRecords[recId]
			if emptyExclude || !inExluded {
				resultRecords[recId] = struct{}{}
			}
		}
	}
	return nil
}

func (sc *MapScanner) FindRecordsIntersection(field string, values interface{}, excludeRecords map[int]struct{}) (result map[int]struct{}, err error) {

	result = make(map[int]struct{})

	if !sc.storage.HasField(field) {
		return
	}

	val, err := sc.storage.converter.ValueToStringSlice(values)
	if err != nil {
		return nil, err
	}

	data := sc.storage.GetFieldData(field)
	emptyExclude := len(excludeRecords) == 0
	isFirst := true
	for _, item := range val {

		records, ok := data[item]
		if !ok {
			continue
		}

		if isFirst {
			if emptyExclude {
				for _, recId := range records {
					result[recId] = struct{}{}
				}
			} else {
				for _, recId := range records {
					if _, ok := excludeRecords[recId]; !ok {
						result[recId] = struct{}{}
					}
				}
			}
			isFirst = false
			continue
		}

		// Intersect with current value's records
		tmp := make(map[int]struct{})
		for _, recId := range records {
			_, inResult := result[recId]
			inExlude := false
			if len(excludeRecords) > 0 {
				_, inExlude = excludeRecords[recId]
			}
			if inResult && !inExlude {
				tmp[recId] = struct{}{}
			}
		}
		result = tmp
	}
	return result, nil
}
func (sc *MapScanner) FindInRange(field string, value *RangeValue, result map[int]struct{}) (err error) {

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
			return err
		}
		hasMin = true
	}

	if value.Max != nil {
		maxValue, err = sc.storage.converter.GetValueString(value.Max)
		if err != nil {
			return err
		}
		hasMax = true
	}

	data := sc.storage.GetFieldData(field)

	for value, records := range data {
		if hasMin && sc.storage.converter.CompareNumStrings(value, minValue) == -1 {
			continue
		}

		if hasMax && sc.storage.converter.CompareNumStrings(value, maxValue) == 1 {
			continue
		}

		for _, recId := range records {
			result[recId] = struct{}{}
		}
	}
	return nil
}

func (sc *MapScanner) FindInValues(field string, value interface{}, result map[int]struct{}) (err error) {
	if !sc.storage.HasField(field) {
		return
	}

	val, err := sc.storage.converter.ValueToStringSlice(value)
	if err != nil {
		return err
	}

	data := sc.storage.GetFieldData(field)

	for _, v := range val {
		records, ok := data[v]
		if !ok {
			continue
		}
		for _, recId := range records {
			result[recId] = struct{}{}
		}
	}
	return
}

// getIntersectMapCount returns the count of intersecting elements.
func getIntersectMapCount(a []int, b map[int]struct{}) int {
	intersectLen := 0
	for _, key := range a {
		if _, ok := b[key]; ok {
			intersectLen++
		}
	}
	return intersectLen
}

// hasIntersectIntMap checks if two collections have any intersection.
func hasIntersectIntMap(a []int, b map[int]struct{}) bool {
	for _, key := range a {
		if _, ok := b[key]; ok {
			return true
		}
	}
	return false
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
