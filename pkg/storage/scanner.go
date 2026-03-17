package storage

import (
	"github.com/k-samuel/faceted/pkg/filter"
)

// Scanner provides methods for scanning and filtering index data.
type Scanner struct{}

// NewScanner creates a new Scanner.
func NewScanner() *Scanner {
	return &Scanner{}
}

// FindRecordsMap finds records by filters using optimized batch processing.
// All filters are processed in a single pass through the data.
func (sc *Scanner) FindRecordsMap(storage StorageInterface, filters []filter.FilterInterface, inputRecords map[int]bool, excludeRecords map[int]bool) map[int]bool {
	// Initialize inputRecords if nil
	if inputRecords == nil {
		inputRecords = make(map[int]bool)
	}

	// If no filters, find by input records only
	if len(filters) == 0 {
		return sc.FindInput(storage, inputRecords, excludeRecords)
	}

	data := storage.GetData()

	// Process each filter
	for _, f := range filters {
		fieldName := f.GetFieldName()
		fieldData, ok := data[fieldName]
		if !ok {
			return make(map[int]bool)
		}

		// Apply filter
		f.FilterInput(fieldData, inputRecords, excludeRecords)

		// Early exit if no matches
		if len(inputRecords) == 0 {
			return inputRecords
		}
	}

	return inputRecords
}

// FindExcludeRecordsMap finds records by exclude filters.
func (sc *Scanner) FindExcludeRecordsMap(storage StorageInterface, filters []filter.ExcludeFilterInterface, excludeRecords *map[int]bool) {
	if len(filters) == 0 {
		return
	}

	data := storage.GetData()

	for _, f := range filters {
		fieldName := f.GetFieldName()
		if fieldData, ok := data[fieldName]; ok {
			f.AddExcluded(fieldData, excludeRecords)
		}
	}
}

// FindInput finds records without filters.
func (sc *Scanner) FindInput(storage StorageInterface, inputRecords map[int]bool, excludeRecords map[int]bool) map[int]bool {
	total := sc.GetAllRecordIdMap(storage)

	if len(inputRecords) == 0 && len(excludeRecords) == 0 {
		return total
	}

	// Intersect with input records
	if len(inputRecords) > 0 {
		intersected := make(map[int]bool, minInt(len(total), len(inputRecords)))
		if len(total) > len(inputRecords) {
			for recId := range inputRecords {
				if total[recId] {
					intersected[recId] = true
				}
			}
		} else {
			for recId := range total {
				if inputRecords[recId] {
					intersected[recId] = true
				}
			}
		}
		total = intersected
	}

	// Remove excluded records
	if len(excludeRecords) > 0 {
		if len(total) > len(excludeRecords) {
			for recId := range excludeRecords {
				delete(total, recId)
			}
		} else {
			for recId := range total {
				if excludeRecords[recId] {
					delete(total, recId)
				}
			}
		}
	}

	return total
}

// GetAllRecordIdMap returns all record IDs from the index.
func (sc *Scanner) GetAllRecordIdMap(storage StorageInterface) map[int]bool {
	data := storage.GetData()

	// Estimate capacity
	totalRecords := 0
	for _, values := range data {
		for _, list := range values {
			totalRecords += len(list)
			break
		}
		break
	}

	// Count unique fields for better estimation
	fieldCount := len(data)
	if fieldCount > 0 {
		totalRecords /= fieldCount
	}

	result := make(map[int]bool, totalRecords)

	for kv := range storage.Scan() {
		for _, list := range kv.Value {
			for _, recId := range list {
				result[recId] = true
			}
		}
	}

	return result
}

// Scan returns a channel for iterating over index data.
func (sc *Scanner) Scan(storage StorageInterface) <-chan KeyValue {
	return storage.Scan()
}
