package storage

import (
	"github.com/k-samuel/faceted/pkg/indexer"
)

// StorageInterface defines the interface for index storage.
type StorageInterface interface {
	// AddRecord adds a record to the index.
	// recordValues: map[fieldName]fieldValue where fieldValue can be a single value or []interface{}
	AddRecord(recordId int, recordValues map[string]interface{}) error

	// GetFieldData returns field data section from index.
	GetFieldData(fieldName string) map[string][]int

	// HasField checks if field exists.
	HasField(fieldName string) bool

	// GetData returns all facet data.
	GetData() map[string]map[string][]int

	// Export exports facet index data.
	Export() map[string]map[string][]int

	// SetData loads saved data.
	SetData(data map[string]map[string][]int)

	// Optimize optimizes index structure.
	Optimize()

	// DeleteRecord deletes a record from the index.
	DeleteRecord(recordId int)

	// ReplaceRecord updates record data.
	ReplaceRecord(recordId int, recordValues map[string]interface{}) error

	// AddIndexer adds a specialized indexer for a field.
	AddIndexer(fieldName string, indexer indexer.IndexerInterface)

	// GetRecordsCount returns the count of records for a field value.
	GetRecordsCount(field string, value interface{}) int

	// Scan returns a channel for iterating over index data.
	Scan() <-chan KeyValue
}

// KeyValue represents a key-value pair for scanning.
type KeyValue struct {
	Key   string
	Value map[string][]int
}
