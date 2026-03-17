package value

type ValueConverter interface {
	// GetValueString  convert value interface{} into string representation
	GetValueString(val interface{}) (value string, err error)
	// ValueToStringSlice convert input value into []string slice
	ValueToStringSlice(val interface{}) (valuesSlice []string, err error)
}
