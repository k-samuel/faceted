package sort

import (
	"github.com/k-samuel/faceted/pkg/query"
	"github.com/k-samuel/faceted/pkg/storage"
)

// QueryResultsInterface defines the interface for sorting query results.
type QueryResultsInterface interface {
	// Sort sorts results by field value.
	Sort(storage storage.StorageInterface, resultsMap map[int]bool, order *query.Sort) []int
}
