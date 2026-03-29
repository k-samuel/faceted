package search

// FilterInterface defines the interface for all filters.
type FilterInterface interface {
	// GetFieldName returns the field name to filter by.
	GetFieldName() string
	// FilterInput filters the faceted data.
	// scan: scanner storage.ScannerInterface  Storage scanner
	// inputIdKeys: map[recordId]bool - input record IDs (modified in place)
	// excludeRecords: map[recordId]bool - records to exclude
	FilterInput(scanner ScannerInterface, inputIdKeys []int, excludeRecords map[int]struct{}) ([]int, error)
	// HasSelfFiltering returns the self-filtering flag.
	HasSelfFiltering() bool
}

// ExcludeFilterInterface extends FilterInterface for exclude filters.
type ExcludeFilterInterface interface {
	FilterInterface
	// AddExcluded adds records to the exclude list.
	AddExcluded(scanner ScannerInterface, excludeRecords map[int]struct{}) error
}
