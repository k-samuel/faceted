package filter

import (
	"github.com/k-samuel/faceted/pkg/value"
)

// ExcludeValueFilter excludes items by value.
type ExcludeValueFilter struct {
	fieldName     string
	selfFiltering bool
	values        []string
	converter     value.ValueConverter
}

// NewExcludeValueFilter creates a new ExcludeValueFilter.
func NewExcludeValueFilter(fieldName string, values interface{}, converter value.ValueConverter) *ExcludeValueFilter {
	f := &ExcludeValueFilter{fieldName: fieldName, converter: converter}
	f.SetValue(values)
	return f
}

// SetValue sets filter values.
func (f *ExcludeValueFilter) SetValue(value interface{}) {

	list, err := f.converter.ValueToStringSlice(value)
	if err == nil {
		f.values = list
	}
}

// GetValue returns filter values.
func (f *ExcludeValueFilter) GetValue() []string {
	return f.values
}

// GetFieldName returns the field name.
func (f *ExcludeValueFilter) GetFieldName() string {
	return f.fieldName
}

// SelfFiltering enables/disables self-filtering.
func (f *ExcludeValueFilter) SelfFiltering(enabled bool) *ExcludeValueFilter {
	f.selfFiltering = enabled
	return f
}

// HasSelfFiltering returns the self-filtering flag.
func (f *ExcludeValueFilter) HasSelfFiltering() bool {
	return f.selfFiltering
}

// FilterInput filters the faceted data (same as ValueFilter for compatibility).
func (f *ExcludeValueFilter) FilterInput(facetedData map[string][]int, inputIdKeys map[int]bool, excludeRecords map[int]bool) {
	// For exclude filters, FilterInput is typically not used directly
	// AddExcluded is used instead to populate excludeRecords
}

// AddExcluded adds records to the exclude list.
func (f *ExcludeValueFilter) AddExcluded(facetedData map[string][]int, excludeRecords *map[int]bool) {
	if *excludeRecords == nil {
		*excludeRecords = make(map[int]bool)
	}

	for _, item := range f.values {
		if records, ok := facetedData[item]; ok {
			if len(*excludeRecords) == 0 {
				// Fast fill for first value
				for _, recId := range records {
					(*excludeRecords)[recId] = true
				}
				continue
			}
			for _, recId := range records {
				(*excludeRecords)[recId] = true
			}
		}
	}
}
