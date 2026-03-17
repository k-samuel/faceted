package index

import (
	"github.com/k-samuel/faceted/pkg/query"
	"github.com/k-samuel/faceted/pkg/storage"
)

// IndexInterface defines the interface for faceted search index.
type IndexInterface interface {
	// Aggregate finds acceptable filter values.
	Aggregate(query *query.AggregationQuery) map[string]map[string]interface{}

	// Query finds records using Query.
	Query(query *query.SearchQuery) []int

	// SetProfiler sets time profiler.
	SetProfiler(profiler *Profile)

	// GetStorage returns index storage.
	GetStorage() storage.StorageInterface

	// GetScanner returns index scanner.
	GetScanner() *storage.Scanner

	// GetCount returns records count.
	GetCount() int

	// SetData loads saved data.
	SetData(data map[string]map[string][]int)

	// Export exports facet index data.
	Export() map[string]map[string][]int

	// Optimize optimizes index structure.
	Optimize()
}
