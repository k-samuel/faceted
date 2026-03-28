package search

type ScannerInterface interface {
	// FindRecordsMap finds records by filters using optimized batch processing.
	// All filters are processed in a single pass through the data.
	FindRecords(filters []FilterInterface, limitRecords []int, excludeRecords map[int]struct{}) ([]int, error)

	// FindExcludeRecordsMap finds records by exclude filters.
	FindExcludeRecordsMap(filters []ExcludeFilterInterface, excludeRecords map[int]struct{})

	// GetAllRecordIdMap returns all record IDs from the index.
	GetAllRecordId(inputRecords []int, excludeRecords map[int]struct{}) []int

	// Aggregate results
	AggregationScan(
		countRecords bool,
		input []int,
		exclude map[int]struct{},
		selfFiltering bool,
		filters []FilterInterface,
	) ([]*AggregationResultField, error)

	// GetAllValuesCount
	GetAllValuesCount(excludeMap map[int]struct{}) []*AggregationResultField
	// GetAllValues
	GetAllValues(excludeMap map[int]struct{}) []*AggregationResultField
	// GetFieldValueRecords get all field values with recordId in it
	GetFieldValueRecords(field string) map[string][]int

	// Filters Actions
	IntersectFilterValues(field string, values interface{}, inputRecords []int, excludeRecords map[int]struct{}) ([]int, error)
	// AddExcludedValues Find records by filters and add them into excludeRecords Map
	AddExcludedValues(field string, values interface{}, excludeRecords map[int]struct{}) (err error)
	FindRangeIntersection(field string, value *RangeValue, inputRecords []int, excludeRecords map[int]struct{}) ([]int, error)
	FindValueIntersection(field string, values interface{}, inputRecords []int, excludeRecords map[int]struct{}) ([]int, error)
	//FindInRange find records with Range
	FindInRange(field string, value *RangeValue) ([]int, error)
}
