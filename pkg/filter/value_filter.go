package filter

import "github.com/k-samuel/faceted/pkg/value"

// ValueFilter filters items by value (OR condition for multiple values).
type ValueFilter struct {
	fieldName     string
	selfFiltering bool
	values        []string
	converter     value.ValueConverter
}

// NewValueFilter creates a new ValueFilter.
func NewValueFilter(fieldName string, values interface{}, converter value.ValueConverter) *ValueFilter {
	f := &ValueFilter{fieldName: fieldName, converter: converter}
	f.SetValue(values)
	return f
}

// SetValue sets filter values.
func (f *ValueFilter) SetValue(value interface{}) {
	list, err := f.converter.ValueToStringSlice(value)
	if err == nil {
		f.values = list
	}
}

// GetValue returns filter values.
func (f *ValueFilter) GetValue() []string {
	return f.values
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
func (f *ValueFilter) FilterInput(facetedData map[string][]int, inputIdKeys map[int]bool, excludeRecords map[int]bool) {
	emptyExclude := len(excludeRecords) == 0

	if len(inputIdKeys) == 0 {
		// Create new result map
		result := make(map[int]bool)
		for _, item := range f.values {
			records, ok := facetedData[item]
			if !ok {
				continue
			}

			// Fast fill for first value with no excludes
			if len(result) == 0 && emptyExclude {
				for _, recId := range records {
					result[recId] = true
				}
				continue
			}

			for _, recId := range records {
				if emptyExclude || !excludeRecords[recId] {
					result[recId] = true
				}
			}
		}
		// Copy result to inputIdKeys
		for k := range inputIdKeys {
			delete(inputIdKeys, k)
		}
		for k, v := range result {
			inputIdKeys[k] = v
		}
		return
	}

	// Use a separate flag map to mark matching entries (same as PHP's flag=2 optimization)
	// This avoids type conversion overhead
	flagMap := make(map[int]bool)
	for _, item := range f.values {
		records, ok := facetedData[item]
		if !ok {
			continue
		}

		for _, recId := range records {
			if inputIdKeys[recId] && (emptyExclude || !excludeRecords[recId]) {
				flagMap[recId] = true // Mark matching entries
			}
		}
	}

	// Remove non-matching records (sweep phase)
	for recId := range inputIdKeys {
		if !flagMap[recId] {
			delete(inputIdKeys, recId)
		}
	}
}
