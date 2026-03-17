package search

import (
	"github.com/k-samuel/faceted/search/indexer"
	"github.com/k-samuel/faceted/search/value"
)

// StorageInterface defines the interface for index storage.
type StorageInterface interface {
	// AddRecord adds a record to the index.
	// recordValues: map[fieldName]fieldValue where fieldValue can be a single value or []interface{}
	AddRecord(recordId int, recordValues map[string]interface{}) error

	// HasField checks if field exists.
	HasField(fieldName string) bool

	// DeleteRecord deletes a record from the index.
	DeleteRecord(recordId int)

	// ReplaceRecord updates record data.
	ReplaceRecord(recordId int, recordValues map[string]interface{}) error

	// AddIndexer adds a specialized indexer for a field.
	AddIndexer(fieldName string, indexer indexer.IndexerInterface)

	// GetRecordsCount returns the count of records for a field value.
	GetRecordsCount(field string, value interface{}) int

	// GetCount returns records count.
	GetCount() int

	// Optimize storage structure adding new records
	Optimize()

	// Export exports facet index data.
	Export() map[string]map[string][]int

	// SetData loads previosly exported data.
	SetData(data map[string]map[string][]int)

	// GetValueConverter
	GetValueConverter() value.ConverterInterface
}
