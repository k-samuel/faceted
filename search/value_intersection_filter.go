package search

// ValueIntersectionFilter filters items by multiple values (AND condition).
// All specified values must be present in the record.
type ValueIntersectionFilter struct {
	fieldName     string
	selfFiltering bool
	values        interface{}
}

// NewValueIntersectionFilter creates a new ValueIntersectionFilter.
func NewValueIntersectionFilter(fieldName string, values interface{}) *ValueIntersectionFilter {
	f := &ValueIntersectionFilter{fieldName: fieldName}
	f.values = values
	return f
}

// GetValue returns filter values.
func (f *ValueIntersectionFilter) GetValue() interface{} {
	return f.values
}

// GetFieldName returns the field name.
func (f *ValueIntersectionFilter) GetFieldName() string {
	return f.fieldName
}

// SelfFiltering enables/disables self-filtering.
func (f *ValueIntersectionFilter) SelfFiltering(enabled bool) *ValueIntersectionFilter {
	f.selfFiltering = enabled
	return f
}

// HasSelfFiltering returns the self-filtering flag.
func (f *ValueIntersectionFilter) HasSelfFiltering() bool {
	return f.selfFiltering
}

// FilterInput filters the faceted data with AND condition.
func (f *ValueIntersectionFilter) FilterInput(scanner ScannerInterface, inputIdKeys map[int]struct{}, excludeRecords map[int]struct{}) error {
	emptyInput := len(inputIdKeys) == 0

	if emptyInput {
		result, err := scanner.FindRecordsIntersection(f.GetFieldName(), f.GetValue(), excludeRecords)
		if err != nil {
			return err
		}

		// Copy result to inputIdKeys
		for k := range inputIdKeys {
			delete(inputIdKeys, k)
		}
		for k, v := range result {
			inputIdKeys[k] = v
		}
		return nil
	}

	flagMap, err := scanner.FindValueIntersection(f.GetFieldName(), f.GetValue(), inputIdKeys)

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
