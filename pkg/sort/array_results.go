package sort

import (
	"sort"

	"github.com/k-samuel/faceted/pkg/query"
	"github.com/k-samuel/faceted/pkg/storage"
)

// ArrayResults implements QueryResultsInterface for sorting query results.
type ArrayResults struct{}

// NewArrayResults creates a new ArrayResults.
func NewArrayResults() *ArrayResults {
	return &ArrayResults{}
}

// Sort sorts results by field value.
func (a *ArrayResults) Sort(storage storage.StorageInterface, resultsMap map[int]bool, order *query.Sort) []int {
	field := order.GetField()

	if !storage.HasField(field) {
		return []int{}
	}

	fieldData := storage.GetFieldData(field)

	// Get sorted values
	values := make([]string, 0, len(fieldData))
	for v := range fieldData {
		values = append(values, v)
	}

	// Determine sort type and create typed slice for efficient sorting
	var sortedValues []string
	if order.GetDirection() == query.SortAsc {
		if order.GetSortFlags() == query.SortString {
			// Convert to string slice and use optimized sort
			strs := make([]string, len(values))
			copy(strs, values)
			sort.Strings(strs)
			// Convert back to interface{}
			sortedValues = make([]string, len(strs))
			copy(sortedValues, strs)
		} else {
			// Use compareValues for numeric and mixed types
			sortedValues = make([]string, len(values))
			copy(sortedValues, values)
			sort.Slice(sortedValues, func(i, j int) bool {
				return compareValues(sortedValues[i], sortedValues[j], order.GetSortFlags())
			})
		}
	} else {
		// Descending order
		if order.GetSortFlags() == query.SortString {
			// Convert to string slice and use optimized sort (reverse)
			strs := make([]string, len(values))
			copy(strs, values)
			sort.Sort(sort.Reverse(sort.StringSlice(strs)))
			// Convert back to interface{}
			sortedValues = make([]string, len(strs))
			copy(sortedValues, strs)
		} else {
			// Use generic sort.Slice for mixed types (reverse)
			sortedValues = make([]string, len(values))
			copy(sortedValues, values)
			sort.Slice(sortedValues, func(i, j int) bool {
				return compareValues(sortedValues[j], sortedValues[i], order.GetSortFlags())
			})
		}
	}

	// Build sorted result
	sorted := make([]int, 0)
	remaining := make(map[int]bool)
	for k, v := range resultsMap {
		remaining[k] = v
	}

	for _, value := range sortedValues {
		records := fieldData[value]
		if order.GetDirection() == query.SortAsc {
			for _, recId := range records {
				if remaining[recId] {
					sorted = append(sorted, recId)
					delete(remaining, recId)
				}
			}
		} else {
			// Reverse order for descending
			for i := len(records) - 1; i >= 0; i-- {
				recId := records[i]
				if remaining[recId] {
					sorted = append(sorted, recId)
					delete(remaining, recId)
				}
			}
		}
	}

	return sorted
}

// compareValues compares two values based on sort flags.
func compareValues(i, j interface{}, sortFlags int) bool {
	switch sortFlags {
	case query.SortNumeric:
		return toFloat64Value(i) < toFloat64Value(j)
	case query.SortString:
		return toStringValue(i) < toStringValue(j)
	default:
		return compareDefault(i, j)
	}
}
