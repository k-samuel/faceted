package filter

import (
	"github.com/k-samuel/faceted/pkg/value"
)

// ValueIntersectionFilter filters items by multiple values (AND condition).
// All specified values must be present in the record.
type ValueIntersectionFilter struct {
	fieldName     string
	selfFiltering bool
	values        []string
	converter     value.ValueConverter
}

// NewValueIntersectionFilter creates a new ValueIntersectionFilter.
func NewValueIntersectionFilter(fieldName string, values interface{}, converter value.ValueConverter) *ValueIntersectionFilter {
	f := &ValueIntersectionFilter{fieldName: fieldName, converter: converter}
	f.SetValue(values)
	return f
}

// SetValue sets filter values.
func (f *ValueIntersectionFilter) SetValue(value interface{}) {
	list, err := f.converter.ValueToStringSlice(value)
	if err == nil {
		f.values = list
	}
}

// GetValue returns filter values.
func (f *ValueIntersectionFilter) GetValue() []string {
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
func (f *ValueIntersectionFilter) FilterInput(facetedData map[string][]int, inputIdKeys map[int]bool, excludeRecords map[int]bool) {
	emptyInput := len(inputIdKeys) == 0
	emptyExclude := len(excludeRecords) == 0

	if emptyInput {
		// Create new result map with intersection of all values
		var result map[int]bool
		isFirst := true

		for _, item := range f.values {
			records, ok := facetedData[item]
			if !ok {
				// No records for this value - intersection is empty
				return
			}

			if isFirst {
				result = make(map[int]bool)
				for _, recId := range records {
					if emptyExclude || !excludeRecords[recId] {
						result[recId] = true
					}
				}
				isFirst = false
				continue
			}

			// Intersect with current value's records
			tmp := make(map[int]bool)
			for _, recId := range records {
				if result[recId] && (emptyExclude || !excludeRecords[recId]) {
					tmp[recId] = true
				}
			}
			result = tmp

			if len(result) == 0 {
				return
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

	// Intersection with existing inputIdKeys
	result := make(map[int]bool)
	isFirst := true

	for _, item := range f.values {
		records, ok := facetedData[item]
		if !ok {
			// No records for this value - intersection is empty
			for k := range inputIdKeys {
				delete(inputIdKeys, k)
			}
			return
		}

		if isFirst {
			for _, recId := range records {
				if inputIdKeys[recId] && (emptyExclude || !excludeRecords[recId]) {
					result[recId] = true
				}
			}
			isFirst = false
			continue
		}

		// Intersect with current value's records
		tmp := make(map[int]bool)
		for _, recId := range records {
			if result[recId] && inputIdKeys[recId] && (emptyExclude || !excludeRecords[recId]) {
				tmp[recId] = true
			}
		}
		result = tmp

		if len(result) == 0 {
			for k := range inputIdKeys {
				delete(inputIdKeys, k)
			}
			return
		}
	}

	// Copy result to inputIdKeys
	for k := range inputIdKeys {
		delete(inputIdKeys, k)
	}
	for k, v := range result {
		inputIdKeys[k] = v
	}
}
