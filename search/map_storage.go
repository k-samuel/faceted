package search

import (
	"slices"
	"sort"

	"github.com/k-samuel/faceted/search/indexer"
	"github.com/k-samuel/faceted/search/value"
)

// ArrayStorage implements StorageInterface using Go maps.
type MapStorage struct {
	data         map[string]map[string][]int
	indexers     map[string]indexer.IndexerInterface
	converter    value.ConverterInterface
	recordCount  int
	needOptimize bool
	sortedKeys   map[string][]string // cache of sorted field value keys
}

// NewStorage creates a new ArrayStorage.
func NewMapStorage(converter value.ConverterInterface) *MapStorage {
	return &MapStorage{
		data:         make(map[string]map[string][]int),
		indexers:     make(map[string]indexer.IndexerInterface),
		converter:    converter,
		needOptimize: false,
		sortedKeys:   make(map[string][]string),
	}
}

func (s *MapStorage) GetTotalCount() int {
	return s.recordCount
}

func (s *MapStorage) GetValueConverter() value.ConverterInterface {
	return s.converter
}

// AddRecord adds a record to the index.
func (s *MapStorage) AddRecord(recordId int, recordValues map[string]interface{}) (err error) {

	var valueSlice []string

	for fieldName, values := range recordValues {

		valueSlice, err = s.converter.ValueToStringSlice(values)

		if err != nil {
			return err
		}

		// If special Indedexer exists
		if indexer, ok := s.indexers[fieldName]; ok {
			if _, ok := s.data[fieldName]; !ok {
				s.data[fieldName] = make(map[string][]int)
			}
			fieldData := s.data[fieldName]

			err = indexer.Add(fieldData, recordId, valueSlice)

			if err != nil {
				return err
			}
			s.data[fieldName] = fieldData

		} else {
			for _, value := range valueSlice {
				if _, ok := s.data[fieldName]; !ok {
					s.data[fieldName] = make(map[string][]int)
				}
				s.data[fieldName][value] = append(s.data[fieldName][value], recordId)
			}
		}
	}
	s.recordCount++
	s.needOptimize = true
	return nil
}

// GetData returns all facet data.
func (s *MapStorage) GetData() map[string]map[string][]int {
	if s.needOptimize {
		s.Optimize()
	}
	return s.data
}

// Export exports facet index data.
func (s *MapStorage) Export() map[string]map[string][]int {
	for fieldName, idx := range s.indexers {
		fieldData := s.data[fieldName]
		idx.Optimize(fieldData)
		s.data[fieldName] = fieldData
	}
	return s.data
}

// SetData loads saved data.
func (s *MapStorage) SetData(data map[string]map[string][]int) {
	s.data = data
	// recalculete total count
	s.RecalculateTotalCount()
	s.Optimize()
}

// GetFieldData returns field data section from index.
func (s *MapStorage) GetFieldData(fieldName string) map[string][]int {

	if s.needOptimize {
		s.Optimize()
	}

	if data, ok := s.data[fieldName]; ok {
		return data
	}
	return make(map[string][]int)
}

// AddIndexer adds a specialized indexer for a field.
func (s *MapStorage) AddIndexer(fieldName string, idx indexer.IndexerInterface) {
	s.indexers[fieldName] = idx
}

// GetRecordsCount returns the count of records for a field value.
func (s *MapStorage) GetRecordsCount(field string, value interface{}) int {
	strValue, _ := s.converter.GetValueString(value)
	if fieldData, ok := s.data[field]; ok {
		if records, ok := fieldData[strValue]; ok {
			return len(records)
		}
	}
	return 0
}

// HasField checks if field exists.
func (s *MapStorage) HasField(fieldName string) bool {
	if data, ok := s.data[fieldName]; ok {
		return len(data) > 0
	}
	return false
}

// Optimize optimizes index structure.
func (s *MapStorage) Optimize() {
	// Optimize indexers
	for fieldName, idx := range s.indexers {
		fieldData := s.data[fieldName]
		idx.Optimize(fieldData)
		s.data[fieldName] = fieldData
	}

	// Sort records by ID and build sorted keys cache
	for fieldName, valueList := range s.data {
		// Count records per value
		for _, list := range valueList {
			// Sort records by ID (except for range indexers)
			if _, hasIndexer := s.indexers[fieldName]; !hasIndexer {
				sort.Ints(list)
			}
		}

		// Build sorted keys cache
		keys := make([]string, 0, len(valueList))
		for k := range valueList {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, func(a, b string) int {
			return s.converter.CompareNumStrings(a, b)
		})
		s.sortedKeys[fieldName] = keys
	}
	s.needOptimize = false
}

// GetSortedFieldValues returns cached sorted field values.
func (s *MapStorage) GetSortedFieldValues(field string) []string {
	if s.needOptimize {
		s.Optimize()
	}
	if keys, ok := s.sortedKeys[field]; ok {
		return keys
	}
	return nil
}

// DeleteRecord deletes a record from the index.
func (s *MapStorage) DeleteRecord(recordId int) {
	var decrement = false
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
				decrement = true
			}
		}
		if len(valueList) == 0 {
			delete(s.data, fieldName)
		}
	}
	if decrement {
		s.recordCount--
	}
}

// ReplaceRecord updates record data.
func (s *MapStorage) ReplaceRecord(recordId int, recordValues map[string]interface{}) error {
	s.DeleteRecord(recordId)
	return s.AddRecord(recordId, recordValues)
}

// GetCount returns records count.
func (s *MapStorage) GetCount() int {
	return s.recordCount
}

// RecalculateTotalCount recalculate records count after data import
func (s *MapStorage) RecalculateTotalCount() {
	resultMap := make(map[int]struct{})
	for _, values := range s.data {
		for _, list := range values {
			for _, id := range list {
				resultMap[id] = struct{}{}
			}
		}
	}
	s.recordCount = len(resultMap)
}
