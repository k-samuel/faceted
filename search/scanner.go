package search

type ScannerInterface interface {
	// FindRecordsMap finds records by filters using optimized batch processing.
	// All filters are processed in a single pass through the data.
	FindRecordsMap(filters []FilterInterface, inputRecords map[int]struct{}, excludeRecords map[int]struct{}) (map[int]struct{}, error)

	// FindExcludeRecordsMap finds records by exclude filters.
	FindExcludeRecordsMap(filters []ExcludeFilterInterface, excludeRecords map[int]struct{})

	// FindInput finds records without filters.
	FindInput(inputRecords map[int]struct{}, excludeRecords map[int]struct{}) map[int]struct{}

	// GetAllRecordIdMap returns all record IDs from the index.
	GetAllRecordIdMap() map[int]struct{}

	// Aggregate results
	AggregationScan(
		resultCache *ResultCache,
		filteredRecords map[int]struct{},
		countRecords bool,
		input map[int]struct{},
		exclude map[int]struct{},
		selfFiltering bool,
		filters []FilterInterface,
	) (map[string]map[string]interface{}, error)

	// GetAllValuesCount
	GetAllValuesCount(excludeMap map[int]struct{}) map[string]map[string]interface{}
	// GetAllValues
	GetAllValues(excludeMap map[int]struct{}) map[string]map[string]interface{}
	// GetFieldValueRecords get all field values with recordId in it
	GetFieldValueRecords(field string) map[string][]int

	// Filters Actions
	// FindUniqueRecords and place them into resultMap
	FindUniqueRecords(field string, values interface{}, resultMap map[int]struct{}, excludeRecords map[int]struct{}) (err error)
	FindRecordsIntersection(field string, values interface{}, excludeRecords map[int]struct{}) (result map[int]struct{}, err error)
	FindIntersection(field string, values interface{}, inputRecords map[int]struct{}) (result map[int]struct{}, err error)
	FindRangeIntersection(field string, value *RangeValue, inputRecords map[int]struct{}, excludeRecords map[int]struct{}) (err error)
	FindValueIntersection(field string, values interface{}, inputRecords map[int]struct{}) (result map[int]struct{}, err error)
	//FindInRange find records with Range, insert into result map
	FindInRange(field string, value *RangeValue, result map[int]struct{}) (err error)
	//FindInValues find records with value, insert into result map
	FindInValues(field string, value interface{}, result map[int]struct{}) (err error)
}
