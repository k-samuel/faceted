package filter

import (
	"strconv"
)

// RangeFilter filters items by value range (min, max).
type RangeFilter struct {
	fieldName     string
	selfFiltering bool
	value         *RangeValue
}

// RangeValue represents a range value with min and max.
type RangeValue struct {
	Min interface{}
	Max interface{}
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
func (f *RangeFilter) FilterInput(facetedData map[string][]int, inputIdKeys map[int]bool, excludeRecords map[int]bool) {
	if f.value.Min == 0 && f.value.Max == 0 {
		// Invalid range - clear all
		for k := range inputIdKeys {
			delete(inputIdKeys, k)
		}
		return
	}

	emptyExclude := len(excludeRecords) == 0

	// Collect all record IDs within range
	var limit []int
	for value, records := range facetedData {
		valueFloat := toFloat64(value)

		if f.value.Min != nil && valueFloat < toFloat64(f.value.Min) {
			continue
		}
		if f.value.Min != nil && valueFloat > toFloat64(f.value.Max) {
			continue
		}

		if emptyExclude {
			limit = append(limit, records...)
		} else {
			for _, recId := range records {
				if !excludeRecords[recId] {
					limit = append(limit, recId)
				}
			}
		}
	}

	if len(limit) == 0 {
		// No records in range - clear all
		for k := range inputIdKeys {
			delete(inputIdKeys, k)
		}
		return
	}

	emptyInput := len(inputIdKeys) == 0

	if emptyInput {
		// Fill inputIdKeys with limit records
		for _, recId := range limit {
			inputIdKeys[recId] = true
		}
		return
	}

	// Use flag map for optimization (matching PHP implementation)
	// First pass: collect matching records into a map
	matchingMap := make(map[int]bool)
	for _, recId := range limit {
		matchingMap[recId] = true
	}

	// Second pass: mark matching records (O(n) instead of O(n*m))
	flagMap := make(map[int]bool)
	for recId := range inputIdKeys {
		if matchingMap[recId] {
			flagMap[recId] = true
		}
	}

	// Third pass: sweep and keep only matching records
	for recId := range inputIdKeys {
		if !flagMap[recId] {
			delete(inputIdKeys, recId)
		}
	}
}

// toFloat64 converts interface{} to float64 for comparison.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		// Try to parse string as float
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			f = 0
		}
		return f
	default:
		return 0
	}
}
