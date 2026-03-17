package filter

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

// SetValue sets the filter range value.
func (f *ExcludeRangeFilter) SetValue(value *RangeValue) {
	f.value = value
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
func (f *ExcludeRangeFilter) FilterInput(facetedData map[string][]int, inputIdKeys map[int]bool, excludeRecords map[int]bool) {
	// For exclude filters, FilterInput is typically not used directly
	// AddExcluded is used instead to populate excludeRecords
}

// AddExcluded adds records to the exclude list.
func (f *ExcludeRangeFilter) AddExcluded(facetedData map[string][]int, excludeRecords *map[int]bool) {
	if f.value.Min == 0 && f.value.Max == 0 {
		return
	}

	if *excludeRecords == nil {
		*excludeRecords = make(map[int]bool)
	}

	for value, records := range facetedData {
		valueFloat := toFloat64(value)

		if f.value.Min != 0 && valueFloat < toFloat64(f.value.Min) {
			continue
		}
		if f.value.Max != nil && valueFloat > toFloat64(f.value.Max) {
			continue
		}

		for _, recId := range records {
			(*excludeRecords)[recId] = true
		}
	}
}
