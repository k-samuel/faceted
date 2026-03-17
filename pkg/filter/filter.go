package filter

// FilterInterface defines the interface for all filters.
type FilterInterface interface {
	// GetFieldName returns the field name to filter by.
	GetFieldName() string
	// FilterInput filters the faceted data.
	// facetedData: map[fieldValue][]recordId
	// inputIdKeys: map[recordId]bool - input record IDs (modified in place)
	// excludeRecords: map[recordId]bool - records to exclude
	FilterInput(facetedData map[string][]int, inputIdKeys map[int]bool, excludeRecords map[int]bool)
	// HasSelfFiltering returns the self-filtering flag.
	HasSelfFiltering() bool
}

// ExcludeFilterInterface extends FilterInterface for exclude filters.
type ExcludeFilterInterface interface {
	FilterInterface
	// AddExcluded adds records to the exclude list.
	AddExcluded(facetedData map[string][]int, excludeRecords *map[int]bool)
}
