package faceted

import (
	"errors"

	"github.com/k-samuel/faceted/pkg/filter"
	"github.com/k-samuel/faceted/pkg/index"
	"github.com/k-samuel/faceted/pkg/indexer"
	"github.com/k-samuel/faceted/pkg/intersection"
	"github.com/k-samuel/faceted/pkg/query"
	"github.com/k-samuel/faceted/pkg/sort"
	"github.com/k-samuel/faceted/pkg/storage"
	"github.com/k-samuel/faceted/pkg/value"
)

// Storage type constants.
const (
	ArrayStorage = "array"
)

// Factory creates Index instances.
type Search struct {
	converter value.ValueConverter
}

// NewFactory creates a new Factory.
func NewSearch() *Search {
	return &Search{
		converter: value.NewValueConverterDefault(),
	}
}

func (f *Search) WithValueConverter(converter value.ValueConverter) *Search {
	f.converter = converter
	return f
}

// Create creates a new Index with the specified storage type.
func (f *Search) NewIndex(storageType string) (index.IndexInterface, error) {
	var store storage.StorageInterface
	var intersectionImpl intersection.IntersectionInterface
	var querySortImpl sort.QueryResultsInterface
	var scanner *storage.Scanner

	switch storageType {
	case ArrayStorage:
		store = storage.NewArrayStorage(f.converter)
		intersectionImpl = intersection.NewArrayIntersection()
		querySortImpl = sort.NewArrayResults()
		scanner = storage.NewScanner()
	default:
		return nil, errors.New("unknown storage type: " + storageType)
	}

	return index.NewIndex(
		store,
		sort.NewFilters(),
		sort.NewAggregationResults(),
		querySortImpl,
		scanner,
		intersectionImpl,
	), nil
}

func (f *Search) NewRangeIndexer(step int) (*indexer.RangeIndexer, error) {
	return indexer.NewRangeIndexer(step)
}

func (f *Search) NewRangeListIndexer(ranges []int) (*indexer.RangeListIndexer, error) {
	return indexer.NewRangeListIndexer(ranges)
}

func (f *Search) NewAggregationQuery() *query.AggregationQuery {
	return query.NewAggregationQuery()
}

func (f *Search) NewSearchQuery() *query.SearchQuery {
	return query.NewSearchQuery()
}

func (f *Search) NewValueFilter(fieldName string, values interface{}) *filter.ValueFilter {
	return filter.NewValueFilter(fieldName, values, f.converter)
}

func (f *Search) NewValueIntersectionFilter(fieldName string, values interface{}) *filter.ValueIntersectionFilter {
	return filter.NewValueIntersectionFilter(fieldName, values, f.converter)
}

func (f *Search) NewRangeFilter(fieldName string, value *filter.RangeValue) *filter.RangeFilter {
	return filter.NewRangeFilter(fieldName, value)
}

func (f *Search) NewRangeValue(min int, max int) *filter.RangeValue {
	return &filter.RangeValue{Min: min, Max: max}
}

func (f *Search) NewExcludeValueFilter(fieldName string, values interface{}) *filter.ExcludeValueFilter {
	return filter.NewExcludeValueFilter(fieldName, values, f.converter)
}

func (f *Search) NewExcludeRangeFilter(fieldName string, rangeValue *filter.RangeValue) *filter.ExcludeRangeFilter {
	return filter.NewExcludeRangeFilter(fieldName, rangeValue)
}

func (f *Search) NewQuerySort(fieldName string, direction int, sortFlags int) *query.Sort {
	return query.NewSort(fieldName, direction, sortFlags)
}

func (f *Search) NewAggregationSort(direction int, sortFlags int) *query.AggregationSort {
	return query.NewAggregationSort(direction, sortFlags)
}
