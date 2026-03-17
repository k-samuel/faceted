package storage

import (
	"sort"

	"github.com/k-samuel/faceted/pkg/indexer"
	"github.com/k-samuel/faceted/pkg/value"
)

// ArrayStorage implements StorageInterface using Go maps.
type ArrayStorage struct {
	data      map[string]map[string][]int
	indexers  map[string]indexer.IndexerInterface
	converter value.ValueConverter
}

// NewArrayStorage creates a new ArrayStorage.
func NewArrayStorage(converter value.ValueConverter) *ArrayStorage {
	return &ArrayStorage{
		data:      make(map[string]map[string][]int),
		indexers:  make(map[string]indexer.IndexerInterface),
		converter: converter,
	}
}

// AddRecord adds a record to the index.
func (s *ArrayStorage) AddRecord(recordId int, recordValues map[string]interface{}) (err error) {

	var valueSlice []string

	for fieldName, values := range recordValues {

		valueSlice, err = s.converter.ValueToStringSlice(values)

		if err != nil {
			return err
		}

		// If special Indedexer exists
		if indexer, ok := s.indexers[fieldName]; ok {
			if s.data[fieldName] == nil {
				s.data[fieldName] = make(map[string][]int)
			}
			fieldData := s.data[fieldName]

			err = indexer.Add(&fieldData, recordId, valueSlice)

			if err != nil {
				return err
			}
			s.data[fieldName] = fieldData

		} else {
			for _, value := range valueSlice {
				if s.data[fieldName] == nil {
					s.data[fieldName] = make(map[string][]int)
				}
				s.data[fieldName][value] = append(s.data[fieldName][value], recordId)
			}
		}
	}
	return nil
}

// GetData returns all facet data.
func (s *ArrayStorage) GetData() map[string]map[string][]int {
	return s.data
}

// Export exports facet index data.
func (s *ArrayStorage) Export() map[string]map[string][]int {
	for fieldName, idx := range s.indexers {
		fieldData := s.data[fieldName]
		idx.Optimize(&fieldData)
		s.data[fieldName] = fieldData
	}
	return s.data
}

// SetData loads saved data.
func (s *ArrayStorage) SetData(data map[string]map[string][]int) {
	s.data = data
}

// GetFieldData returns field data section from index.
func (s *ArrayStorage) GetFieldData(fieldName string) map[string][]int {
	if data, ok := s.data[fieldName]; ok {
		return data
	}
	return make(map[string][]int)
}

// AddIndexer adds a specialized indexer for a field.
func (s *ArrayStorage) AddIndexer(fieldName string, idx indexer.IndexerInterface) {
	s.indexers[fieldName] = idx
}

// GetRecordsCount returns the count of records for a field value.
func (s *ArrayStorage) GetRecordsCount(field string, value interface{}) int {
	strValue, _ := s.converter.GetValueString(value)
	if fieldData, ok := s.data[field]; ok {
		if records, ok := fieldData[strValue]; ok {
			return len(records)
		}
	}
	return 0
}

// HasField checks if field exists.
func (s *ArrayStorage) HasField(fieldName string) bool {
	if data, ok := s.data[fieldName]; ok {
		return len(data) > 0
	}
	return false
}

// Optimize optimizes index structure.
func (s *ArrayStorage) Optimize() {
	// Optimize indexers
	for fieldName, idx := range s.indexers {
		fieldData := s.data[fieldName]
		idx.Optimize(&fieldData)
		s.data[fieldName] = fieldData
	}

	// Sort records by ID and values by record count
	for fieldName, valueList := range s.data {
		// Count records per value
		valueCounts := make(map[string]int)
		for value, list := range valueList {
			valueCounts[value] = len(list)

			// Sort records by ID (except for range indexers)
			if _, hasIndexer := s.indexers[fieldName]; !hasIndexer {
				sort.Ints(list)
			}
		}

		// Sort values by record count
		sortedValues := make([]string, len(valueCounts))
		i := 0
		for value := range valueCounts {
			sortedValues[i] = value
			i++
		}
		sort.Slice(sortedValues, func(i, j int) bool {
			return valueCounts[sortedValues[i]] < valueCounts[sortedValues[j]]
		})

		// Rebuild valueList in sorted order
		newList := make(map[string][]int)
		for _, value := range sortedValues {
			newList[value] = valueList[value]
		}
		s.data[fieldName] = newList
	}
}

// DeleteRecord deletes a record from the index.
func (s *ArrayStorage) DeleteRecord(recordId int) {
	for fieldName, valueList := range s.data {
		for fieldValue, list := range valueList {
			hasDeletion := false
			newList := make([]int, 0, len(list))
			for _, id := range list {
				if id != recordId {
					newList = append(newList, id)
				} else {
					hasDeletion = true
				}
			}
			if hasDeletion {
				if len(newList) == 0 {
					delete(valueList, fieldValue)
				} else {
					valueList[fieldValue] = newList
				}
			}
		}
		if len(valueList) == 0 {
			delete(s.data, fieldName)
		}
	}
}

// ReplaceRecord updates record data.
func (s *ArrayStorage) ReplaceRecord(recordId int, recordValues map[string]interface{}) error {
	s.DeleteRecord(recordId)
	return s.AddRecord(recordId, recordValues)
}

// Scan returns a channel for iterating over index data.
func (s *ArrayStorage) Scan() <-chan KeyValue {
	ch := make(chan KeyValue)
	go func() {
		defer close(ch)
		for k, v := range s.data {
			ch <- KeyValue{Key: k, Value: v}
		}
	}()
	return ch
}
