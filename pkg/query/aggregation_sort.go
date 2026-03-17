package query

// AggregationSort defines sorting for aggregation results.
type AggregationSort struct {
	direction int
	sortFlags int
}

// NewAggregationSort creates a new AggregationSort.
func NewAggregationSort(direction int, sortFlags int) *AggregationSort {
	return &AggregationSort{
		direction: direction,
		sortFlags: sortFlags,
	}
}

// GetDirection returns the sort direction.
func (s *AggregationSort) GetDirection() int {
	return s.direction
}

// GetSortFlags returns the sort flags.
func (s *AggregationSort) GetSortFlags() int {
	return s.sortFlags
}
