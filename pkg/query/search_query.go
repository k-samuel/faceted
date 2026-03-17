package query

import "github.com/k-samuel/faceted/pkg/filter"

// SearchQuery defines a search query with filters and sorting.
type SearchQuery struct {
	filters []filter.FilterInterface
	sort    *Sort
	records []int
}

// NewSearchQuery creates a new SearchQuery.
func NewSearchQuery() *SearchQuery {
	return &SearchQuery{
		filters: make([]filter.FilterInterface, 0),
		records: make([]int, 0),
	}
}

// Filter adds a single filter to the query.
func (q *SearchQuery) Filter(f filter.FilterInterface) *SearchQuery {
	q.filters = append(q.filters, f)
	return q
}

// Filters adds multiple filters to the query.
func (q *SearchQuery) Filters(filters []filter.FilterInterface) *SearchQuery {
	q.filters = append(q.filters, filters...)
	return q
}

// Order sets the sorting order for results.
func (q *SearchQuery) Sort(fieldName string, direction int, sortFlags int) *SearchQuery {
	q.sort = NewSort(fieldName, direction, sortFlags)
	return q
}

// Order sets the sorting order for results.
func (q *SearchQuery) SortBy(sort *Sort) *SearchQuery {
	q.sort = sort
	return q
}

// InRecords sets the list of record IDs to search in.
func (q *SearchQuery) InRecords(records []int) *SearchQuery {
	q.records = records
	return q
}

// GetInRecords returns the list of record IDs.
func (q *SearchQuery) GetInRecords() []int {
	return q.records
}

// GetOrder returns the sorting order.
func (q *SearchQuery) GetSort() *Sort {
	return q.sort
}

// GetFilters returns the list of filters.
func (q *SearchQuery) GetFilters() []filter.FilterInterface {
	return q.filters
}
