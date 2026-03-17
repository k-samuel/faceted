package value

import (
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type Converter struct {
}

// NewValueFilter creates a new ValueFilter.
func NewConverter() *Converter {
	f := &Converter{}
	return f
}

func (conv *Converter) GetValueString(val interface{}) (value string, err error) {
	switch v := val.(type) {
	case bool:
		if v {
			return "1", nil
		}
		return "0", nil
	case int:
		return strconv.Itoa(v), nil

	case int64:
		return strconv.FormatInt(v, 10), nil
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case float32:
		float64Val := float64(v)
		return strconv.FormatFloat(float64Val, 'f', -1, 32), nil
	}

	err = errors.New("Undefined value type")
	return value, err
}
func (conv *Converter) ValueToStringSlice(val interface{}) (valuesSlice []string, err error) {

	var valueString string
	valuesSlice = make([]string, 0)

	valKind := reflect.TypeOf(val).Kind()

	if valKind == reflect.Slice || valKind == reflect.Array || valKind == reflect.Map {

		// map
		if s, ok := val.(map[string]interface{}); ok {
			for _, v := range s {
				valueString, err = conv.GetValueString(v)
				if err != nil {
					return
				}
				valuesSlice = append(valuesSlice, valueString)
			}

			if len(valuesSlice) > 2 {
				slices.Sort(valuesSlice)
				valuesSlice = slices.Compact(valuesSlice)
			}

			return
		}

		// slice of interface{}
		if s, ok := val.([]interface{}); ok {
			for _, v := range s {
				valueString, err = conv.GetValueString(v)
				if err != nil {
					return
				}
				valuesSlice = append(valuesSlice, valueString)
			}
			if len(valuesSlice) > 2 {
				slices.Sort(valuesSlice)
				valuesSlice = slices.Compact(valuesSlice)
			}
			return
		}

		// slice of int
		if s, ok := val.([]int); ok {
			for _, v := range s {
				valueString, err = conv.GetValueString(v)
				if err != nil {
					return
				}
				valuesSlice = append(valuesSlice, valueString)
			}
			if len(valuesSlice) > 2 {
				slices.Sort(valuesSlice)
				valuesSlice = slices.Compact(valuesSlice)
			}
			return
		}

		// slice of int64
		if s, ok := val.([]int64); ok {
			for _, v := range s {
				valueString, err = conv.GetValueString(v)
				if err != nil {
					return
				}
				valuesSlice = append(valuesSlice, valueString)
			}
			if len(valuesSlice) > 2 {
				slices.Sort(valuesSlice)
				valuesSlice = slices.Compact(valuesSlice)
			}
			return
		}

		// slice of string
		if s, ok := val.([]string); ok {
			valuesSlice = make([]string, len(s))
			copy(valuesSlice, s)
			if len(valuesSlice) > 2 {
				slices.Sort(valuesSlice)
				valuesSlice = slices.Compact(valuesSlice)
			}
			return
		}

		err = errors.New("Type cannot be converted into value " + reflect.TypeOf(val).Name() + reflect.TypeOf(val).Elem().Name())
		return nil, err
	}

	// other
	valueString, err = conv.GetValueString(val)
	if err != nil {
		return
	}

	valuesSlice = append(valuesSlice, valueString)
	return
}

func (conv *Converter) CompareNumStrings(a, b string) int {

	// find decimal separator
	idxA := strings.Index(a, ".")
	idxB := strings.Index(b, ".")

	// compare  cail parg
	if idxA == -1 {
		idxA = len(a)
	}
	if idxB == -1 {
		idxB = len(b)
	}

	// 2. compare length
	if idxA > idxB {
		return 1
	}
	if idxA < idxB {
		return -1
	}

	// 3. compare decimal part
	return strings.Compare(a, b)
}
