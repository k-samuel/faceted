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
func (f *ValueFilter) FilterInput(scanner ScannerInterface, inputIdKeys map[int]struct{}, excludeRecords map[int]struct{}) error {

	if len(inputIdKeys) == 0 {
		err := scanner.FindUniqueRecords(f.GetFieldName(), f.GetValue(), inputIdKeys, excludeRecords)
		if err != nil {
			return err
		}
		return nil
	}

	flagMap, err := scanner.FindIntersection(f.GetFieldName(), f.GetValue(), inputIdKeys)

	if err != nil {
		return err
	}

	// Remove non-matching records (sweep phase)
	for recId := range inputIdKeys {
		if _, ok := flagMap[recId]; !ok {
			delete(inputIdKeys, recId)
		}
	}
	return nil
}
