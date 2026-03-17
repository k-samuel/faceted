package indexer

import (
	"errors"
	"sort"
	"strconv"
)

// RangeListIndexer creates range-based indices with custom range boundaries.
type RangeListIndexer struct {
	ranges      []int
	hasUnsorted bool
	unsortedBuf map[string]map[string][]int
}

// NewRangeListIndexer creates a new RangeListIndexer with the specified ranges.
func NewRangeListIndexer(ranges []int) (*RangeListIndexer, error) {

	if len(ranges) < 2 {
		return nil, errors.New("Invalid step ranges")
	}

	// Sort ranges
	sortedRanges := make([]int, len(ranges))
	copy(sortedRanges, ranges)
	sort.Ints(sortedRanges)

	return &RangeListIndexer{
		ranges:      sortedRanges,
		hasUnsorted: false,
		unsortedBuf: make(map[string]map[string][]int),
	}, nil
}

// Add adds a record to the range index.
func (ri *RangeListIndexer) Add(indexContainer *map[string][]int, recordId int, values []string) (err error) {
	var floatValue float64

	for _, value := range values {

		floatValue, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}

		position := ri.detectRangeKey(floatValue)
		positionKey := strconv.Itoa(position)

		if (*indexContainer)[positionKey] == nil {
			(*indexContainer)[positionKey] = make([]int, 0)
		}
		(*indexContainer)[positionKey] = append((*indexContainer)[positionKey], recordId)

		if ri.unsortedBuf[positionKey] == nil {
			ri.unsortedBuf[positionKey] = make(map[string][]int)
		}
		ri.unsortedBuf[positionKey][value] = append(ri.unsortedBuf[positionKey][value], recordId)
		ri.hasUnsorted = true
	}
	return nil
}

// Optimize optimizes the range index by sorting values within each range.
func (ri *RangeListIndexer) Optimize(indexContainer *map[string][]int) {
	if !ri.hasUnsorted {
		return
	}

	for position, values := range ri.unsortedBuf {
		// Sort values by key
		keys := make([]string, 0, len(values))
		for k := range values {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			left, _ := strconv.ParseFloat(keys[i], 64)
			right, _ := strconv.ParseFloat(keys[j], 64)
			return left < right
		})

		// Rebuild index with sorted values
		sortedRecords := make([]int, 0)
		for _, key := range keys {
			sortedRecords = append(sortedRecords, values[key]...)
		}
		(*indexContainer)[position] = sortedRecords
	}

	ri.unsortedBuf = make(map[string]map[string][]int)
	ri.hasUnsorted = false
}

// detectRangeKey detects the range position key for a value.
func (ri *RangeListIndexer) detectRangeKey(value float64) int {
	lastKey := 0
	for _, key := range ri.ranges {
		if value >= float64(lastKey) && value < float64(key) {
			return lastKey
		}
		lastKey = key
	}
	return lastKey
}
