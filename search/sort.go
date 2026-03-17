package search

// Sort direction constants.
const (
	SortAsc         = 0
	SortDesc        = 1
	SortTypeStrings = 0
	SortTypeNumbers = 1
)

// Order defines sorting order for query results.
type Sort struct {
	FieldName string
	Direction int
	SortType  int
}

type AggregationSort struct {
	FieldDirection int
	ValueDirection int
}

// NewOrder creates a new Order.
func NewSort(fieldName string, direction int, sortType int) *Sort {
	return &Sort{
		FieldName: fieldName,
		Direction: direction,
	}
}

// GetField returns the field name.
func (o *Sort) GetField() string {
	return o.FieldName
}

// GetDirection returns the sort direction.
func (o *Sort) GetDirection() int {
	return o.Direction
}

// GetDirection returns the sort direction.
func (o *Sort) GetType() int {
	return o.SortType
}
