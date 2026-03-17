package value

type ConverterInterface interface {
	// GetValueString  convert value interface{} into string representation
	GetValueString(val interface{}) (value string, err error)
	// ValueToStringSlice convert input value into []string slice
	ValueToStringSlice(val interface{}) (valuesSlice []string, err error)
	// Compare numeric strings -1,0,1
	CompareNumStrings(a, b string) int
}
