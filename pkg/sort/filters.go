package sort

import (
	"sort"

	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/storage"
)

// Filters provides sorting functionality for filters.
type Filters struct{}

// NewFilters creates a new Filters.
func NewFilters() *Filters {
	return &Filters{}
}

// ByCount sorts filters by minimum values count.
// This is an optimization for aggregation - filters with fewer values are processed first.
func (f *Filters) ByCount(storage storage.StorageInterface, filters []filter.FilterInterface) []filter.FilterInterface {
	type filterWithIndex struct {
		index int
		count int
	}

	counts := make([]filterWithIndex, len(filters))

	for i, flt := range filters {
		// Non-ValueFilter types get max priority
		vf, ok := flt.(*filter.ValueFilter)
		if !ok {
			counts[i] = filterWithIndex{index: i, count: int(^uint(0) >> 1)} // Max int
			continue
		}

		fieldName := vf.GetFieldName()

		if !storage.HasField(fieldName) {
			counts[i] = filterWithIndex{index: i, count: 0}
			continue
		}

		filterValues := vf.GetValue()
		filterValuesCount := make(map[interface{}]int)
		valuesInFilter := len(filterValues)
		minCount := int(^uint(0) >> 1)

		for _, value := range filterValues {
			cnt := storage.GetRecordsCount(fieldName, value)
			if valuesInFilter > 1 {
				filterValuesCount[value] = cnt
			}

			if cnt < minCount {
				minCount = cnt
			}
		}

		counts[i] = filterWithIndex{index: i, count: minCount}

		// Sort filter values by records count
		if valuesInFilter > 1 {
			sortedValues := make([]interface{}, 0, len(filterValuesCount))
			for value := range filterValuesCount {
				sortedValues = append(sortedValues, value)
			}
			sort.Slice(sortedValues, func(i, j int) bool {
				return filterValuesCount[sortedValues[i]] < filterValuesCount[sortedValues[j]]
			})
			vf.SetValue(sortedValues)
		}
	}

	// Sort filters by count
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count < counts[j].count
	})

	// Build result
	result := make([]filter.FilterInterface, len(counts))
	for i, fi := range counts {
		result[i] = filters[fi.index]
	}

	return result
}
