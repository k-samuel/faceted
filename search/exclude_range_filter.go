package search

// ExcludeRangeFilter excludes items by value range.
type ExcludeRangeFilter struct {
	fieldName     string
	selfFiltering bool
	value         *RangeValue
}

// NewExcludeRangeFilter creates a new ExcludeRangeFilter.
func NewExcludeRangeFilter(fieldName string, value *RangeValue) *ExcludeRangeFilter {
	f := &ExcludeRangeFilter{fieldName: fieldName, value: value}

	return f
}

// GetValue returns the filter range value.
func (f *ExcludeRangeFilter) GetValue() *RangeValue {
	return f.value
}

// GetFieldName returns the field name.
func (f *ExcludeRangeFilter) GetFieldName() string {
	return f.fieldName
}

// SelfFiltering enables/disables self-filtering.
func (f *ExcludeRangeFilter) SelfFiltering(enabled bool) *ExcludeRangeFilter {
	f.selfFiltering = enabled
	return f
}

// HasSelfFiltering returns the self-filtering flag.
func (f *ExcludeRangeFilter) HasSelfFiltering() bool {
	return f.selfFiltering
}

// GetMin returns the minimum value.
func (f *ExcludeRangeFilter) GetMin() interface{} {
	return f.value.Min
}

// GetMax returns the maximum value.
func (f *ExcludeRangeFilter) GetMax() interface{} {
	return f.value.Max
}

// FilterInput filters the faceted data (same as RangeFilter for compatibility).
func (f *ExcludeRangeFilter) FilterInput(facetedData map[string][]int, inputIdKeys map[int]struct{}, excludeRecords map[int]struct{}) {
	// For exclude filters, FilterInput is typically not used directly
	// AddExcluded is used instead to populate excludeRecords
}

// AddExcluded adds records to the exclude list.
func (f *ExcludeRangeFilter) AddExcluded(scanner ScannerInterface, excludeRecords map[int]struct{}) error {

	if f.value.Min == f.value.Max {
		return nil
	}

	if excludeRecords == nil {
		excludeRecords = make(map[int]struct{})
	}

	err := scanner.FindInRange(f.GetFieldName(), f.GetValue(), excludeRecords)

	if err != nil {
		return err
	}
	return nil
}
