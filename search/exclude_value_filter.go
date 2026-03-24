package search

// ExcludeValueFilter excludes items by value.
type ExcludeValueFilter struct {
	fieldName     string
	selfFiltering bool
	values        interface{}
}

// NewExcludeValueFilter creates a new ExcludeValueFilter.
func NewExcludeValueFilter(fieldName string, values interface{}) *ExcludeValueFilter {
	f := &ExcludeValueFilter{fieldName: fieldName}
	f.values = values
	return f
}

// GetValue returns filter values.
func (f *ExcludeValueFilter) GetValue() interface{} {
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
func (f *ExcludeValueFilter) FilterInput(scanner ScannerInterface, inputIdKeys []int, excludeRecords map[int]struct{}) ([]int, error) {
	// For exclude filters, FilterInput is typically not used directly
	// AddExcluded is used instead to populate excludeRecords
	panic("Method ExcludeValueFilter::FilterInput should not be called")
}

// AddExcluded adds records to the exclude list.
func (f *ExcludeValueFilter) AddExcluded(scanner ScannerInterface, excludeRecords map[int]struct{}) error {

	var input []int

	if excludeRecords == nil {
		excludeRecords = make(map[int]struct{})
		input = make([]int, 0, 0)
	} else {
		input = make([]int, 0, len(excludeRecords))
		for k := range excludeRecords {
			input = append(input, k)
		}
	}

	list, err := scanner.MergeExcludedValues(f.GetFieldName(), f.GetValue())
	if err != nil {
		return err
	}
	for _, k := range list {
		excludeRecords[k] = struct{}{}
	}
	return nil
}
