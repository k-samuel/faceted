package query

// Sort direction constants.
const (
	SortAsc  = 0
	SortDesc = 1
)

// Order defines sorting order for query results.
type Sort struct {
	fieldName string
	direction int
	sortFlags int
}

// Sort flags (compatible with PHP sort flags).
const (
	SortRegular      = 0
	SortNumeric      = 1
	SortString       = 2
	SortLocaleString = 3
	SortNatural      = 5
	SortFlagCase     = 8
	SortNaturalCase  = SortNatural | SortFlagCase
)

// NewOrder creates a new Order.
func NewSort(fieldName string, direction int, sortFlags int) *Sort {
	return &Sort{
		fieldName: fieldName,
		direction: direction,
		sortFlags: sortFlags,
	}
}

// GetField returns the field name.
func (o *Sort) GetField() string {
	return o.fieldName
}

// GetDirection returns the sort direction.
func (o *Sort) GetDirection() int {
	return o.direction
}

// GetSortFlags returns the sort flags.
func (o *Sort) GetSortFlags() int {
	return o.sortFlags
}
