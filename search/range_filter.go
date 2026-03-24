package search

// RangeValue represents a range value with min and max.
type RangeValue struct {
	Min interface{}
	Max interface{}
}

// NewRangeFilter creates a new RangeFilter.
func NewRangeValue(min any, max any) *RangeValue {
	return &RangeValue{Min: min, Max: max}
}

// RangeFilter filters items by value range (min, max).
type RangeFilter struct {
	fieldName     string
	selfFiltering bool
	value         *RangeValue
}

// NewRangeFilter creates a new RangeFilter.
func NewRangeFilter(fieldName string, value *RangeValue) *RangeFilter {
	f := &RangeFilter{fieldName: fieldName, value: value}
	return f
}

// SetValue sets the filter range value.
func (f *RangeFilter) SetValue(value *RangeValue) {
	f.value = value
}

// GetValue returns the filter range value.
func (f *RangeFilter) GetValue() *RangeValue {
	return f.value
}

// GetFieldName returns the field name.
func (f *RangeFilter) GetFieldName() string {
	return f.fieldName
}

// SelfFiltering enables/disables self-filtering.
func (f *RangeFilter) SelfFiltering(enabled bool) *RangeFilter {
	f.selfFiltering = enabled
	return f
}

// HasSelfFiltering returns the self-filtering flag.
func (f *RangeFilter) HasSelfFiltering() bool {
	return f.selfFiltering
}

// GetMin returns the minimum value.
func (f *RangeFilter) GetMin() interface{} {
	return f.value.Min
}

// GetMax returns the maximum value.
func (f *RangeFilter) GetMax() interface{} {
	return f.value.Max
}

// FilterInput filters the faceted data by range.
func (f *RangeFilter) FilterInput(scanner ScannerInterface, inputIdKeys []int, excludeRecords map[int]struct{}) ([]int, error) {

	if f.value.Min == f.value.Max {
		return []int{}, nil
	}

	list, err := scanner.FindRangeIntersection(f.GetFieldName(), f.GetValue(), inputIdKeys, excludeRecords)
	return list, err
}
