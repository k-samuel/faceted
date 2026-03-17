package search

import (
	"slices"
	"strings"
)

// ArrayResults implements QueryResultsInterface for sorting query results.
type QueryResults struct{}

// NewArrayResults creates a new ArrayResults.
func NewQueryResults() *QueryResults {
	return &QueryResults{}
}

// Sort sorts results by field value.
func (a *QueryResults) Sort(values map[string][]int, resultsMap map[int]struct{}, order *Sort) []int {

	// Determine sort type and create typed slice for efficient sorting

	sortedValues := make([]string, len(values))
	for k := range values {
		sortedValues = append(sortedValues, k)
	}

	if order.GetDirection() == SortAsc {
		slices.SortStableFunc(sortedValues, func(i, j string) int {
			return strings.Compare(i, j)
		})

	} else {
		// Descending order
		slices.SortStableFunc(sortedValues, func(i, j string) int {
			return strings.Compare(j, i)
		})
	}

	// Build sorted result
	sorted := make([]int, 0)

	for _, value := range sortedValues {
		records := values[value]
		if order.GetDirection() == SortAsc {
			for _, recId := range records {
				if _, ok := resultsMap[recId]; ok {
					sorted = append(sorted, recId)
					delete(resultsMap, recId)
				}
			}
		} else {
			// Reverse order for descending
			for i := len(records) - 1; i >= 0; i-- {
				recId := records[i]
				if _, ok := resultsMap[recId]; ok {
					sorted = append(sorted, recId)
					delete(resultsMap, recId)
				}
			}
		}
	}

	return sorted
}
