package search

// AggregationQuery defines an aggregation query for building filter aggregates.
type AggregationQuery struct {
	filters       []FilterInterface
	needCount     bool
	records       []int
	sort          *AggregationSort
	selfFiltering bool
}

type AggregationResultField struct {
	Field  string                    `json:"field"`
	Values []*AggregationResultValue `json:"values"`
}

type AggregationResultValue struct {
	Value string `json:"value"`
	Count *int   `json:"count,omitempty"`
}

// NewAggregationQuery creates a new AggregationQuery.
func NewAggregationQuery() *AggregationQuery {
	return &AggregationQuery{
		filters:       make([]FilterInterface, 0),
		records:       make([]int, 0),
		selfFiltering: false,
	}
}

// Filter adds a single filter to the query.
func (q *AggregationQuery) Filter(f FilterInterface) *AggregationQuery {
	q.filters = append(q.filters, f)
	return q
}

// Filters adds multiple filters to the query.
func (q *AggregationQuery) Filters(filters []FilterInterface) *AggregationQuery {
	q.filters = append(q.filters, filters...)
	return q
}

// CountItems enables/disables counting of items per filter value.
func (q *AggregationQuery) CountItems(count bool) *AggregationQuery {
	q.needCount = count
	return q
}

// GetCountItems returns the count items flag.
func (q *AggregationQuery) GetCountItems() bool {
	return q.needCount
}

// InRecords sets the list of record IDs to search in.
func (q *AggregationQuery) InRecords(records []int) *AggregationQuery {
	q.records = records
	return q
}

// GetInRecords returns the list of record IDs.
func (q *AggregationQuery) GetInRecords() []int {
	return q.records
}

// GetFilters returns the list of filters.
func (q *AggregationQuery) GetFilters() []FilterInterface {
	return q.filters
}

// Sort sets the sorting for aggregation results.
func (q *AggregationQuery) Sort(fieldDirection int, valueDirection int) *AggregationQuery {
	q.sort = &AggregationSort{FieldDirection: fieldDirection, ValueDirection: valueDirection}
	return q
}

// GetSort returns the aggregation sort.
func (q *AggregationQuery) GetSort() *AggregationSort {
	return q.sort
}

// SelfFiltering enables/disables self-filtering.
// When enabled, selecting a filter value will filter out other values of the same field.
func (q *AggregationQuery) SelfFiltering(enabled bool) *AggregationQuery {
	q.selfFiltering = enabled
	return q
}

// HasSelfFiltering returns the self-filtering flag.
func (q *AggregationQuery) HasSelfFiltering() bool {
	return q.selfFiltering
}
