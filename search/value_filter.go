package search

// ValueFilter filters items by value (OR condition for multiple values).
type ValueFilter struct {
	fieldName     string
	selfFiltering bool
	values        interface{}
}

// NewValueFilter creates a new ValueFilter.
func NewValueFilter(fieldName string, values interface{}) *ValueFilter {
	f := &ValueFilter{fieldName: fieldName}
	f.values = values
	return f
}

// GetValue returns filter values.
func (f *ValueFilter) GetValue() interface{} {
	return f.values
}

// GetValue set filter values
func (f *ValueFilter) SetValue(value interface{}) {
	f.values = value
}

// GetFieldName returns the field name.
func (f *ValueFilter) GetFieldName() string {
	return f.fieldName
}

// SelfFiltering enables/disables self-filtering.
func (f *ValueFilter) SelfFiltering(enabled bool) *ValueFilter {
	f.selfFiltering = enabled
	return f
}

// HasSelfFiltering returns the self-filtering flag.
func (f *ValueFilter) HasSelfFiltering() bool {
	return f.selfFiltering
}

// FilterInput filters the faceted data using mark-and-sweep optimization.
// Uses flag value 2 to mark matching entries instead of allocating new arrays.
// This matches the PHP implementation for consistent behavior.
func (f *ValueFilter) FilterInput(scanner ScannerInterface, limitRecords []int, excludeRecords map[int]struct{}) ([]int, error) {

	res, err := scanner.IntersectFilterValues(f.GetFieldName(), f.GetValue(), limitRecords, excludeRecords)
	return res, err
}
