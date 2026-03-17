package indexer

// IndexerInterface defines the interface for field indexers.
// Indexers are used to create specialized index structures for fields
// (e.g., range indexing for numeric fields).
type IndexerInterface interface {
	// Add adds a record to the index.
	// indexContainer: the field's index data structure
	// recordId: the ID of the record to add
	// values: the field values for this record
	Add(indexContainer *map[string][]int, recordId int, values []string) error

	// Optimize optimizes the index data structures after all records are added.
	Optimize(indexContainer *map[string][]int)
}
