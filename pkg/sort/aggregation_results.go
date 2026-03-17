package sort

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/k-samuel/faceted/pkg/query"
)

// AggregationResults provides sorting for aggregation results.
type AggregationResults struct{}

// NewAggregationResults creates a new AggregationResults.
func NewAggregationResults() *AggregationResults {
	return &AggregationResults{}
}

// Sort sorts aggregation result fields and values.
// result: map[fieldName]map[fieldValue]count|true
func (a *AggregationResults) Sort(sortConfig *query.AggregationSort, result map[string]map[string]interface{}) {
	sortFlags := sortConfig.GetSortFlags()

	if sortConfig.GetDirection() == query.SortAsc {
		// Sort keys ascending
		a.sortKeysAscending(result, sortFlags)
	} else {
		// Sort keys descending
		a.sortKeysDescending(result, sortFlags)
	}
}

// sortKeysAscending sorts result keys in ascending order.
func (a *AggregationResults) sortKeysAscending(result map[string]map[string]interface{}, sortFlags int) {
	// Sort outer map keys (field names)
	fieldNames := make([]string, 0, len(result))
	for fieldName := range result {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)

	// Rebuild result in sorted order
	sortedResult := make(map[string]map[string]interface{})
	for _, fieldName := range fieldNames {
		values := result[fieldName]
		sortedValues := a.sortValuesAscending(values, sortFlags)
		sortedResult[fieldName] = sortedValues
	}

	// Copy back to result
	for k := range result {
		delete(result, k)
	}
	for k, v := range sortedResult {
		result[k] = v
	}
}

// sortKeysDescending sorts result keys in descending order.
func (a *AggregationResults) sortKeysDescending(result map[string]map[string]interface{}, sortFlags int) {
	// Sort outer map keys (field names)
	fieldNames := make([]string, 0, len(result))
	for fieldName := range result {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)

	// Reverse order
	for i, j := 0, len(fieldNames)-1; i < j; i, j = i+1, j-1 {
		fieldNames[i], fieldNames[j] = fieldNames[j], fieldNames[i]
	}

	// Rebuild result in sorted order
	sortedResult := make(map[string]map[string]interface{})
	for _, fieldName := range fieldNames {
		values := result[fieldName]
		sortedValues := a.sortValuesDescending(values, sortFlags)
		sortedResult[fieldName] = sortedValues
	}

	// Copy back to result
	for k := range result {
		delete(result, k)
	}
	for k, v := range sortedResult {
		result[k] = v
	}
}

// sortValuesAscending sorts values in ascending order.
func (a *AggregationResults) sortValuesAscending(values map[string]interface{}, sortFlags int) map[string]interface{} {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	a.sortKeys(keys, sortFlags, false)

	sortedValues := make(map[string]interface{})
	for _, k := range keys {
		sortedValues[fmt.Sprintf("%v", k)] = values[k]
	}
	return sortedValues
}

// sortValuesDescending sorts values in descending order.
func (a *AggregationResults) sortValuesDescending(values map[string]interface{}, sortFlags int) map[string]interface{} {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	a.sortKeys(keys, sortFlags, true)

	sortedValues := make(map[string]interface{})
	for _, k := range keys {
		sortedValues[fmt.Sprintf("%v", k)] = values[k]
	}
	return sortedValues
}

// sortKeys sorts keys based on sort flags.
func (a *AggregationResults) sortKeys(keys []string, sortFlags int, descending bool) {
	sort.Slice(keys, func(i, j int) bool {
		less := a.compare(keys[i], keys[j], sortFlags)
		if descending {
			return !less
		}
		return less
	})
}

// compare compares two values based on sort flags.
func (a *AggregationResults) compare(i, j interface{}, sortFlags int) bool {
	switch sortFlags {
	case query.SortNumeric:
		return toFloat64Value(i) < toFloat64Value(j)
	case query.SortString:
		return toStringValue(i) < toStringValue(j)
	default:
		// SortRegular - use default comparison
		return compareDefault(i, j)
	}
}

// toFloat64Value converts interface{} to float64.
func toFloat64Value(v interface{}) float64 {
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
		// Parse as float
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// toStringValue converts interface{} to string.
func toStringValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return ""
	}
}

// compareDefault provides default comparison.
func compareDefault(i, j interface{}) bool {
	switch vi := i.(type) {
	case int:
		if vj, ok := j.(int); ok {
			return vi < vj
		}
	case string:
		if vj, ok := j.(string); ok {
			return vi < vj
		}
	case float64:
		if vj, ok := j.(float64); ok {
			return vi < vj
		}
	}
	return false
}
