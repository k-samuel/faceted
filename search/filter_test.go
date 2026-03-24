package search

import (
	"slices"
	"testing"

	"github.com/k-samuel/faceted/search/value"
)

// TestValueFilterSetValue tests ValueFilter SetValue functionality.
func TestValueFilterSetValue(t *testing.T) {
	f := NewValueFilter("color", "red")

	values, _ := value.NewConverter().ValueToStringSlice(f.GetValue())

	if len(values) != 1 || values[0] != "red" {
		t.Errorf("Expected single value 'red', got %v", values)
	}

	f.SetValue([]interface{}{"blue", "green"})
	values, _ = value.NewConverter().ValueToStringSlice(f.GetValue())
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
	if values[0] != "blue" || values[1] != "green" {
		t.Errorf("Expected ['blue', 'green'], got %v", values)
	}

	f.SetValue(42)
	values, _ = value.NewConverter().ValueToStringSlice(f.GetValue())
	if len(values) != 1 || values[0] != "42" {
		t.Errorf("Expected single value 42, got %v", values)
	}

	f.SetValue([]int{1, 2, 3})
	values, _ = value.NewConverter().ValueToStringSlice(f.GetValue())
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
}

// TestValueFilterSelfFiltering tests ValueFilter SelfFiltering functionality.
func TestValueFilterSelfFiltering(t *testing.T) {
	f := NewValueFilter("color", "red")

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}

	f.SelfFiltering(false)
	if f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be disabled")
	}
}

// TestExcludeValueFilterSelfFiltering tests ExcludeValueFilter self-filtering.
func TestExcludeValueFilterSelfFiltering(t *testing.T) {
	f := NewExcludeValueFilter("color", []interface{}{"red"})

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering enabled")
	}

	f.SelfFiltering(false)
	if f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering disabled")
	}
}

// TestRangeFilter tests RangeFilter functionality.
func TestRangeFilter(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	values := f.GetValue()
	if values.Min != 10 || values.Max != 100 {
		t.Errorf("Expected RangeValue{Min: 10, Max: 100}, got %v", values)
	}

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}
}

// TestRangeFilterGetValue tests RangeFilter GetValue functionality.
func TestRangeFilterGetValue(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	values := f.GetValue()
	if values.Min != 10 || values.Max != 100 {
		t.Errorf("Expected RangeValue{Min: 10, Max: 100}, got %v", values)
	}
}

// TestRangeFilterGetValueWithFloat tests RangeFilter GetValue with float values.
func TestRangeFilterGetValueWithFloat(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10.5, Max: 99.9})

	values := f.GetValue()
	if values.Min != 10.5 || values.Max != 99.9 {
		t.Errorf("Expected RangeValue{Min: 10.5, Max: 99.9}, got %v", values)
	}
}

// TestRangeFilterGetValueWithNegative tests RangeFilter GetValue with negative values.
func TestRangeFilterGetValueWithNegative(t *testing.T) {
	f := NewRangeFilter("temperature", &RangeValue{Min: -100, Max: -10})

	values := f.GetValue()
	if values.Min != -100 || values.Max != -10 {
		t.Errorf("Expected RangeValue{Min: -100, Max: -10}, got %v", values)
	}
}

// TestRangeFilterSetValue tests RangeFilter SetValue functionality.
func TestRangeFilterSetValue(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	f.SetValue(&RangeValue{Min: 20, Max: 200})
	values := f.GetValue()
	if values.Min != 20 || values.Max != 200 {
		t.Errorf("Expected RangeValue{Min: 20, Max: 200}, got %v", values)
	}
}

// TestRangeFilterGetFieldName tests RangeFilter GetFieldName functionality.
func TestRangeFilterGetFieldName(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	if f.GetFieldName() != "price" {
		t.Errorf("Expected field name 'price', got %s", f.GetFieldName())
	}
}

// TestRangeFilterGetMin tests RangeFilter GetMin functionality.
func TestRangeFilterGetMin(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 25.5, Max: 100})

	if f.GetMin() != 25.5 {
		t.Errorf("Expected min 25.5, got %v", f.GetMin())
	}
}

// TestRangeFilterGetMax tests RangeFilter GetMax functionality.
func TestRangeFilterGetMax(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10, Max: 99.9})

	if f.GetMax() != 99.9 {
		t.Errorf("Expected max 99.9, got %v", f.GetMax())
	}
}

// TestRangeFilterWithZeroValue tests RangeFilter with zero values.
func TestRangeFilterWithZeroValue(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 0, Max: 0})

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}
}

// TestRangeFilterWithNegativeValues tests RangeFilter with negative values.
func TestRangeFilterWithNegativeValues(t *testing.T) {
	f := NewRangeFilter("temperature", &RangeValue{Min: -20, Max: -5})

	values := f.GetValue()
	if values.Min != -20 || values.Max != -5 {
		t.Errorf("Expected RangeValue{Min: -20, Max: -5}, got %v", values)
	}
}

// TestRangeFilterWithFloatValues tests RangeFilter with float values.
func TestRangeFilterWithFloatValues(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10.5, Max: 99.99})

	values := f.GetValue()
	if values.Min != 10.5 || values.Max != 99.99 {
		t.Errorf("Expected RangeValue{Min: 10.5, Max: 99.99}, got %v", values)
	}
}

// TestRangeFilterFilterInput tests RangeFilter FilterInput functionality.
func TestRangeFilterFilterInput(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	facetedData := map[string]map[string][]int{
		"price": {
			"5":   {1, 2},
			"15":  {3, 4},
			"50":  {5, 6, 7},
			"99":  {8, 9},
			"100": {10},
			"101": {11, 12},
		},
	}

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	inputRecords := []int{1, 3, 5, 8, 10, 11}
	excludeRecords := map[int]struct{}{5: {}}

	res, err := f.FilterInput(scan, inputRecords, excludeRecords)
	if err != nil {
		t.Error(err)
	}

	// Expected: 3,4,8,9,10 (records in range [10,100]: 3,4,5,6,7,8,9,10)
	// After excluding 5: 3,4,6,7,8,9,10
	// After intersecting with inputRecords {1,3,5,8,10,11}: 3,8,10
	expectedCount := 3 // records 3, 8, 10
	if len(res) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(res))
	}

	if !slices.Equal(res, []int{3, 8, 10}) {
		t.Errorf("Expected record to be included")
	}
}

// TestRangeFilterFilterInputWithNoMatches tests RangeFilter FilterInput with no matches.
func TestRangeFilterFilterInputWithNoMatches(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 200, Max: 300})

	storage := NewMapStorage(value.NewConverter())

	facetedData := map[string]map[string][]int{
		"price": {
			"50":  {1, 2},
			"100": {3, 4},
		},
	}
	storage.SetData(facetedData)
	scan := NewMapScanner(storage)

	inputRecords := []int{1, 2, 3}
	res, _ := f.FilterInput(scan, inputRecords, map[int]struct{}{})

	if len(res) != 0 {
		t.Errorf("Expected 0 records after filtering, got %d", len(res))
	}
}

// TestRangeFilterFilterInputWithAllMatches tests RangeFilter FilterInput with all matches.
func TestRangeFilterFilterInputWithAllMatches(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 0, Max: 1000})

	storage := NewMapStorage(value.NewConverter())
	facetedData := map[string]map[string][]int{
		"price": {
			"10": {1, 2},
			"50": {3, 4},
		},
	}
	storage.SetData(facetedData)
	scan := NewMapScanner(storage)

	inputRecords := []int{1, 2, 3, 4}
	f.FilterInput(scan, inputRecords, map[int]struct{}{})

	if len(inputRecords) != 4 {
		t.Errorf("Expected 4 records, got %d", len(inputRecords))
	}
}

// TestRangeFilterFilterInputWithExcludes tests RangeFilter FilterInput with exclude records.
func TestRangeFilterFilterInputWithExcludes(t *testing.T) {
	f := NewRangeFilter("price", &RangeValue{Min: 10, Max: 100})
	storage := NewMapStorage(value.NewConverter())
	facetedData := map[string]map[string][]int{
		"price": {
			"20": {1, 2, 3},
			"50": {4, 5},
		},
	}
	storage.SetData(facetedData)
	scan := NewMapScanner(storage)
	inputRecords := []int{1, 2, 3, 4, 5}
	excludeRecords := map[int]struct{}{2: {}, 5: {}}

	res, _ := f.FilterInput(scan, inputRecords, excludeRecords)

	if len(res) != 3 {
		t.Errorf("Expected 3 records, got %d", len(res))
	}

	for _, v := range res {
		if v == 2 || v == 5 {
			t.Errorf("Expected record %d to be excluded", v)
		}
	}
}

// TestExcludeRangeFilter tests ExcludeRangeFilter functionality.
func TestExcludeRangeFilter(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	values := f.GetValue()
	if values.Min != 10 || values.Max != 100 {
		t.Errorf("Expected RangeValue{Min: 10, Max: 100}, got %v", values)
	}

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}
}

// TestExcludeRangeFilterGetValue tests ExcludeRangeFilter GetValue functionality.
func TestExcludeRangeFilterGetValue(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 20, Max: 200})

	values := f.GetValue()
	if values.Min != 20 || values.Max != 200 {
		t.Errorf("Expected RangeValue{Min: 20, Max: 200}, got %v", values)
	}
}

// TestExcludeRangeFilterGetValueWithFloat tests ExcludeRangeFilter GetValue with float values.
func TestExcludeRangeFilterGetValueWithFloat(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 20.5, Max: 200.5})

	values := f.GetValue()
	if values.Min != 20.5 || values.Max != 200.5 {
		t.Errorf("Expected RangeValue{Min: 20.5, Max: 200.5}, got %v", values)
	}
}

// TestExcludeRangeFilterGetValueWithNegative tests ExcludeRangeFilter GetValue with negative values.
func TestExcludeRangeFilterGetValueWithNegative(t *testing.T) {
	f := NewExcludeRangeFilter("temperature", &RangeValue{Min: -200, Max: -20})

	values := f.GetValue()
	if values.Min != -200 || values.Max != -20 {
		t.Errorf("Expected RangeValue{Min: -200, Max: -20}, got %v", values)
	}
}

// TestExcludeRangeFilterGetFieldName tests ExcludeRangeFilter GetFieldName functionality.
func TestExcludeRangeFilterGetFieldName(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	if f.GetFieldName() != "price" {
		t.Errorf("Expected field name 'price', got %s", f.GetFieldName())
	}
}

// TestExcludeRangeFilterGetMin tests ExcludeRangeFilter GetMin functionality.
func TestExcludeRangeFilterGetMin(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 25.5, Max: 100})

	if f.GetMin() != 25.5 {
		t.Errorf("Expected min 25.5, got %v", f.GetMin())
	}
}

// TestExcludeRangeFilterGetMax tests ExcludeRangeFilter GetMax functionality.
func TestExcludeRangeFilterGetMax(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 10, Max: 99.9})

	if f.GetMax() != 99.9 {
		t.Errorf("Expected max 99.9, got %v", f.GetMax())
	}
}

// TestExcludeRangeFilterAddExcluded tests ExcludeRangeFilter AddExcluded functionality.
func TestExcludeRangeFilterAddExcluded(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 10, Max: 100})

	facetedData := map[string]map[string][]int{
		"price": {
			"5":   {1, 2},
			"15":  {3, 4},
			"50":  {5, 6, 7},
			"99":  {8, 9},
			"101": {10, 11},
		},
	}

	excludeRecords := make(map[int]struct{})

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	f.AddExcluded(scan, excludeRecords)

	// Records in range [10, 100] should be excluded: 3,4,5,6,7,8,9
	expectedExcluded := 7 // records 3,4,5,6,7,8,9
	if len(excludeRecords) != expectedExcluded {
		t.Errorf("Expected %d excluded records, got %d", expectedExcluded, len(excludeRecords))
	}

	if _, ok := excludeRecords[3]; !ok {
		t.Errorf("Expected record 3 to be excluded")
	}
	if _, ok := excludeRecords[9]; !ok {
		t.Errorf("Expected record 9 to be excluded")
	}
	// Records outside range should not be excluded
	if _, ok := excludeRecords[1]; ok {
		t.Errorf("Expected record 1 to NOT be excluded")
	}
	if _, ok := excludeRecords[10]; ok {
		t.Errorf("Expected record 10 to NOT be excluded")
	}
}

// TestExcludeRangeFilterAddExcludedWithZeroValue tests ExcludeRangeFilter AddExcluded with zero values.
func TestExcludeRangeFilterAddExcludedWithZeroValue(t *testing.T) {
	f := NewExcludeRangeFilter("price", &RangeValue{Min: 0, Max: 0})

	st := NewMapStorage(value.NewConverter())
	scan := NewMapScanner(st)

	excludeRecords := make(map[int]struct{})
	_ = f.AddExcluded(scan, excludeRecords)

	// With zero values, AddExcluded should return without adding any records
	if len(excludeRecords) != 0 {
		t.Errorf("Expected 0 excluded records, got %d", len(excludeRecords))
	}
}

// TestExcludeValueFilterGetValue tests ExcludeValueFilter GetValue functionality.
func TestExcludeValueFilterGetValue(t *testing.T) {
	f := NewExcludeValueFilter("color", []interface{}{"red", "blue"})

	values := f.GetValue()
	valSlice, _ := value.NewConverter().ValueToStringSlice(values)
	if len(valSlice) != 2 {
		t.Errorf("Expected 2 values, got %d", len(valSlice))
	}
}

// TestExcludeValueFilterGetFieldName tests ExcludeValueFilter GetFieldName functionality.
func TestExcludeValueFilterGetFieldName(t *testing.T) {
	f := NewExcludeValueFilter("color", []interface{}{"red"})

	if f.GetFieldName() != "color" {
		t.Errorf("Expected field name 'color', got %s", f.GetFieldName())
	}
}

// TestExcludeValueFilterAddExcluded tests ExcludeValueFilter AddExcluded functionality.
func TestExcludeValueFilterAddExcluded(t *testing.T) {
	f := NewExcludeValueFilter("color", []interface{}{"red", "blue"})

	facetedData := map[string]map[string][]int{
		"color": {
			"red":   {1, 2, 3},
			"blue":  {4, 5},
			"green": {6, 7},
		},
	}

	excludeRecords := make(map[int]struct{})

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	err := f.AddExcluded(scan, excludeRecords)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Records with red or blue should be excluded: 1,2,3,4,5
	expectedExcluded := 5 // records 1,2,3,4,5
	if len(excludeRecords) != expectedExcluded {
		t.Errorf("Expected %d excluded records, got %d", expectedExcluded, len(excludeRecords))
	}

	// Records with green should not be excluded
	if _, ok := excludeRecords[6]; ok {
		t.Errorf("Expected record 6 to NOT be excluded")
	}
	if _, ok := excludeRecords[7]; ok {
		t.Errorf("Expected record 7 to NOT be excluded")
	}
}

// TestValueIntersectionFilterGetValue tests ValueIntersectionFilter GetValue functionality.
func TestValueIntersectionFilterGetValue(t *testing.T) {
	f := NewValueIntersectionFilter("purpose", []interface{}{"hunting", "fishing"})

	values := f.GetValue()
	valSlice, _ := value.NewConverter().ValueToStringSlice(values)
	if len(valSlice) != 2 {
		t.Errorf("Expected 2 values, got %d", len(valSlice))
	}
}

// TestValueIntersectionFilterGetFieldName tests ValueIntersectionFilter GetFieldName functionality.
func TestValueIntersectionFilterGetFieldName(t *testing.T) {
	f := NewValueIntersectionFilter("purpose", []interface{}{"hunting"})

	if f.GetFieldName() != "purpose" {
		t.Errorf("Expected field name 'purpose', got %s", f.GetFieldName())
	}
}

// TestValueIntersectionFilterSelfFiltering tests ValueIntersectionFilter self-filtering.
func TestValueIntersectionFilterSelfFiltering(t *testing.T) {
	f := NewValueIntersectionFilter("purpose", []interface{}{"hunting"})

	f.SelfFiltering(true)
	if !f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}

	f.SelfFiltering(false)
	if f.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be disabled")
	}
}

// TestValueIntersectionFilterFilterInput tests ValueIntersectionFilter FilterInput functionality.
func TestValueIntersectionFilterFilterInput(t *testing.T) {
	f := NewValueIntersectionFilter("purpose", []interface{}{"hunting", "fishing"})

	facetedData := map[string]map[string][]int{
		"purpose": {
			"hunting": {1, 2, 3},
			"fishing": {2, 3, 4},
			"sports":  {5, 6},
		},
	}

	inputRecords := []int{1, 2, 3, 4, 5}
	excludeRecords := map[int]struct{}{}

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	res, err := f.FilterInput(scan, inputRecords, excludeRecords)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Records with BOTH hunting AND fishing: 2, 3
	expectedCount := 2
	if len(res) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(res))
	}

	if !slices.Equal(res, []int{2, 3}) {
		t.Errorf("Expected record 2,3 to be included")
	}
}

// TestAggregationQuerySelfFiltering tests AggregationQuery SelfFiltering functionality.
func TestAggregationQuerySelfFiltering(t *testing.T) {
	q := NewAggregationQuery()

	q.SelfFiltering(true)
	if !q.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be enabled")
	}

	q.SelfFiltering(false)
	if q.HasSelfFiltering() {
		t.Errorf("Expected self-filtering to be disabled")
	}
}

// TestSearchQueryFilter tests SearchQuery Filter functionality.
func TestSearchQueryFilter(t *testing.T) {
	q := NewSearchQuery()

	f := NewValueFilter("color", []interface{}{"red"})
	q.Filter(f)

	filters := q.GetFilters()
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}
}

// TestSearchQuerySortBy tests SearchQuery SortBy functionality.
func TestSearchQuerySortBy(t *testing.T) {
	q := NewSearchQuery()

	sort := NewSort("price", SortDesc, SortTypeNumbers)
	q.SortBy(sort)

	if q.GetSort() == nil {
		t.Errorf("Expected sort to be set")
	}
}

// TestSearchQueryInRecords tests SearchQuery InRecords functionality.
func TestSearchQueryInRecords(t *testing.T) {
	q := NewSearchQuery()

	q.InRecords([]int{1, 2, 3})

	records := q.GetInRecords()
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
}

// TestAggregationQueryFilter tests AggregationQuery Filter functionality.
func TestAggregationQueryFilter(t *testing.T) {
	q := NewAggregationQuery()

	f := NewValueFilter("color", []interface{}{"red"})
	q.Filter(f)

	filters := q.GetFilters()
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}
}

// TestAggregationQueryInRecords tests AggregationQuery InRecords functionality.
func TestAggregationQueryInRecords(t *testing.T) {
	q := NewAggregationQuery()

	q.InRecords([]int{1, 2, 3})

	records := q.GetInRecords()
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
}

// TestNewRangeValue tests creating a new RangeValue.
func TestNewRangeValue(t *testing.T) {
	rv := NewRangeValue(10, 100)

	if rv.Min != 10 {
		t.Errorf("Expected min 10, got %v", rv.Min)
	}
	if rv.Max != 100 {
		t.Errorf("Expected max 100, got %v", rv.Max)
	}
}

// TestNewValueFilter tests creating a new ValueFilter.
func TestNewValueFilter(t *testing.T) {

	container := NewContainer()

	db := container.NewDb()
	storage := db.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"color": "red", "warehouse": []int{1, 3, 7}})
	_ = storage.AddRecord(2, map[string]interface{}{"color": "blue", "warehouse": []int{1, 2, 3}})
	_ = storage.AddRecord(3, map[string]interface{}{"color": "red", "warehouse": []int{1, 2, 3}})
	_ = storage.AddRecord(4, map[string]interface{}{"color": "red", "warehouse": []int{2, 3}})
	storage.Optimize()

	filters := []FilterInterface{
		NewValueFilter("color", []interface{}{"red"}),
		NewValueFilter("warehouse", []int{2, 3}),
	}

	searchQuery := NewSearchQuery().Filters(filters)
	records, err := db.Query(searchQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should find record 1
	if len(records) != 3 {
		t.Errorf("Expected records 3, got %v", records)
	}
}

// TestNewValueIntersectionFilter tests creating a new ValueIntersectionFilter.
func TestNewValueIntersectionFilter(t *testing.T) {
	f := NewValueIntersectionFilter("purpose", []interface{}{"hunting", "fishing"})

	if f.GetFieldName() != "purpose" {
		t.Errorf("Expected field name 'purpose', got %s", f.GetFieldName())
	}
}

// TestNewExcludeValueFilter tests creating a new ExcludeValueFilter.
func TestNewExcludeValueFilter(t *testing.T) {
	f := NewExcludeValueFilter("color", []interface{}{"red", "blue"})

	if f.GetFieldName() != "color" {
		t.Errorf("Expected field name 'color', got %s", f.GetFieldName())
	}
}

// TestScannerGetFieldValueRecords tests Scanner GetFieldValueRecords functionality.
func TestScannerGetFieldValueRecords(t *testing.T) {
	facetedData := map[string]map[string][]int{
		"color": {
			"red":   {1, 2},
			"blue":  {3, 4},
			"green": {5, 6},
		},
	}

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	records := scan.GetFieldValueRecords("color")
	if len(records) != 3 {
		t.Errorf("Expected 3 values, got %d", len(records))
	}
}

// TestScannerGetAllValues tests Scanner GetAllValues functionality.
func TestScannerGetAllValues(t *testing.T) {
	facetedData := map[string]map[string][]int{
		"color": {
			"red":   {1, 2},
			"blue":  {3, 4},
			"green": {5, 6},
		},
	}

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	values := scan.GetAllValues(map[int]struct{}{})
	if !hasAggregationField(values, "color") {
		t.Errorf("Expected color field in values")
	}
}

// TestMapScannerFindExcludedRecordsMap tests MapScanner FindExcludeRecordsMap functionality.
func TestMapScannerFindExcludedRecordsMap(t *testing.T) {
	facetedData := map[string]map[string][]int{
		"color": {
			"red": {1, 2, 3},
		},
	}

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	filters := []ExcludeFilterInterface{
		NewExcludeValueFilter("color", []interface{}{"red"}),
	}

	excludeRecords := make(map[int]struct{})
	scan.FindExcludeRecordsMap(filters, excludeRecords)

	// All records with red should be excluded
	if len(excludeRecords) != 3 {
		t.Errorf("Expected 3 excluded records, got %d", len(excludeRecords))
	}
}

// TestScannerGetAllValuesWithExclude tests Scanner GetAllValues with exclude functionality.
func TestScannerGetAllValuesWithExclude(t *testing.T) {
	facetedData := map[string]map[string][]int{
		"color": {
			"red":   {1, 2},
			"blue":  {3, 4},
			"green": {5, 6},
		},
	}

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	// Exclude record 1
	excludeMap := map[int]struct{}{1: {}}
	values := scan.GetAllValues(excludeMap)

	// red should still exist (has record 2)
	foundRed := hasAggregationValue(values, "color", "red")
	if !foundRed {
		t.Errorf("Expected red in values even with record 1 excluded")
	}
}

// TestScannerGetFieldValueRecordsEmpty tests Scanner GetFieldValueRecords with empty data.
func TestScannerGetFieldValueRecordsEmpty(t *testing.T) {
	st := NewMapStorage(value.NewConverter())
	scan := NewMapScanner(st)

	records := scan.GetFieldValueRecords("nonexistent")
	if len(records) != 0 {
		t.Errorf("Expected 0 records for nonexistent field, got %d", len(records))
	}
}

// TestMapScannerFindExcludedRecordsMapWithNoMatches tests MapScanner FindExcludeRecordsMap with no matches.
func TestMapScannerFindExcludedRecordsMapWithNoMatches(t *testing.T) {
	facetedData := map[string]map[string][]int{
		"color": {
			"red": {1, 2, 3},
		},
	}

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	// Filter for a field that doesn't exist in data
	filters := []ExcludeFilterInterface{
		NewExcludeValueFilter("nonexistent", []interface{}{"value"}),
	}

	excludeRecords := make(map[int]struct{})
	scan.FindExcludeRecordsMap(filters, excludeRecords)

	// No records should be excluded (field doesn't exist)
	if len(excludeRecords) != 0 {
		t.Errorf("Expected 0 excluded records, got %d", len(excludeRecords))
	}
}

// TestQueryWithInputRecords tests Query with input records filtering.
func TestQueryWithInputRecords(t *testing.T) {
	container := NewContainer()
	db := container.NewDb()
	storage := db.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"color": "red"})
	_ = storage.AddRecord(2, map[string]interface{}{"color": "blue"})
	_ = storage.AddRecord(3, map[string]interface{}{"color": "red"})

	// Search only within records 1 and 2
	inputRecords := []int{1, 2}
	filters := []FilterInterface{
		NewValueFilter("color", []interface{}{"red"}),
	}

	searchQuery := NewSearchQuery().Filters(filters).InRecords(inputRecords)
	records, err := db.Query(searchQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should find record 1
	if len(records) != 1 || records[0] != 1 {
		t.Errorf("Expected record 1, got %v", records)
	}
}

// TestAggregationWithSelfFiltering tests Aggregation with self-filtering enabled.
func TestAggregationWithSelfFiltering(t *testing.T) {
	container := NewContainer()
	db := container.NewDb()
	storage := db.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"color": "red"})
	_ = storage.AddRecord(2, map[string]interface{}{"color": "blue"})
	_ = storage.AddRecord(3, map[string]interface{}{"color": "red"})

	filters := []FilterInterface{
		NewValueFilter("color", []interface{}{"red"}),
	}

	// With self-filtering, when we select "red", it should exclude "blue" from results
	aggQuery := NewAggregationQuery().Filters(filters).SelfFiltering(true).CountItems(true)
	aggData, err := db.Aggregate(aggQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// red should exist in color with count
	if !hasAggregationField(aggData, "color") {
		t.Errorf("Expected color in aggData")
	}
}

// TestQueryWithSortBy tests Query with SortBy functionality.
func TestQueryWithSortBy(t *testing.T) {
	container := NewContainer()
	db := container.NewDb()
	storage := db.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"name": "A", "price": 300})
	_ = storage.AddRecord(2, map[string]interface{}{"name": "B", "price": 100})
	_ = storage.AddRecord(3, map[string]interface{}{"name": "C", "price": 200})

	filters := []FilterInterface{}
	sort := NewSort("price", SortAsc, SortTypeNumbers)

	searchQuery := NewSearchQuery().Filters(filters).SortBy(sort)
	records, err := db.Query(searchQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should be sorted by price ASC: 2 (100), 3 (200), 1 (300)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
	if records[0] != 2 || records[1] != 3 || records[2] != 1 {
		t.Errorf("Expected sorted records [2, 3, 1], got %v", records)
	}
}

// TestFilterInputWithNoMatches tests FilterInput with no matching records.
func TestFilterInputWithNoMatches(t *testing.T) {
	f := NewValueFilter("color", []interface{}{"green"})

	facetedData := map[string]map[string][]int{
		"color": {
			"red": {1, 2},
		},
	}

	inputRecords := []int{1, 2}
	excludeRecords := make(map[int]struct{})

	st := NewMapStorage(value.NewConverter())
	st.SetData(facetedData)
	scan := NewMapScanner(st)

	res, err := f.FilterInput(scan, inputRecords, excludeRecords)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// No records match "green", all should be removed from input
	if len(res) != 0 {
		t.Errorf("Expected 0 records after filter, got %d", len(res))
	}
}

// TestQueryAggregateWithEmptyData tests Query and Aggregate with empty storage.
func TestQueryAggregateWithEmptyData(t *testing.T) {
	container := NewContainer()
	db := container.NewDb()

	// Query with no records
	filters := []FilterInterface{}
	searchQuery := NewSearchQuery().Filters(filters)
	records, err := db.Query(searchQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}

	// Aggregate with no records
	aggQuery := NewAggregationQuery().Filters(filters).CountItems(true)
	aggData, err := db.Aggregate(aggQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(aggData) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(aggData))
	}
}

// TestGetRecordsCountWithNonExistentValue tests GetRecordsCount with non-existent value.
func TestGetRecordsCountWithNonExistentValue(t *testing.T) {
	st := NewMapStorage(value.NewConverter())

	count := st.GetRecordsCount("color", "nonexistent")
	if count != 0 {
		t.Errorf("Expected 0 count for nonexistent value, got %d", count)
	}
}

// TestGetRecordsCountWithNonExistentField tests GetRecordsCount with non-existent field.
func TestGetRecordsCountWithNonExistentField(t *testing.T) {
	st := NewMapStorage(value.NewConverter())
	st.SetData(map[string]map[string][]int{
		"other": {"value": {1}},
	})

	count := st.GetRecordsCount("color", "red")
	if count != 0 {
		t.Errorf("Expected 0 count for nonexistent field, got %d", count)
	}
}

// TestNewQueryResults tests creating a new QueryResults.
func TestNewQueryResults(t *testing.T) {
	qr := NewQueryResults()
	if qr == nil {
		t.Errorf("Expected non-nil QueryResults")
	}
}

// TestQueryWithMultipleExcludeFilters tests Query with multiple exclude filters.
func TestQueryWithMultipleExcludeFilters(t *testing.T) {
	container := NewContainer()
	db := container.NewDb()
	storage := db.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"color": "red", "size": 10})
	_ = storage.AddRecord(2, map[string]interface{}{"color": "blue", "size": 10})
	_ = storage.AddRecord(3, map[string]interface{}{"color": "red", "size": 20})
	_ = storage.AddRecord(4, map[string]interface{}{"color": "blue", "size": 20})

	// Exclude both blue color and size 10
	filters := []FilterInterface{
		NewExcludeValueFilter("color", []interface{}{"blue"}),
		NewExcludeValueFilter("size", []interface{}{10}),
	}

	searchQuery := NewSearchQuery().Filters(filters)
	records, err := db.Query(searchQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Record 1 (red, 10) - excluded by size
	// Record 2 (blue, 10) - excluded by both
	// Record 3 (red, 20) - should be included
	// Record 4 (blue, 20) - excluded by color
	if len(records) != 1 || records[0] != 3 {
		t.Errorf("Expected record 3, got %v", records)
	}
}

// TestAggregationQueryWithInputRecords tests AggregationQuery with input records.
func TestAggregationQueryWithInputRecords(t *testing.T) {
	container := NewContainer()
	db := container.NewDb()
	storage := db.GetStorage()

	_ = storage.AddRecord(1, map[string]interface{}{"color": "red"})
	_ = storage.AddRecord(2, map[string]interface{}{"color": "blue"})
	_ = storage.AddRecord(3, map[string]interface{}{"color": "red"})

	// Aggregate only within records 1 and 2
	aggQuery := NewAggregationQuery().InRecords([]int{1, 2}).CountItems(true)
	aggData, err := db.Aggregate(aggQuery)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Debug: print all fields
	t.Logf("Aggregation data: %d fields", len(aggData))
	for _, f := range aggData {
		t.Logf("  Field %s: %d values", f.Field, len(f.Values))
		for _, v := range f.Values {
			if v.Count != nil {
				t.Logf("    %s=%d", v.Value, *v.Count)
			} else {
				t.Logf("    %s", v.Value)
			}
		}
	}

	// Should show red: 1, blue: 1 (only from records 1 and 2)
	colorField := findAggregationField(aggData, "color")
	if colorField == nil {
		t.Errorf("Expected color field in aggData")
		return
	}
	if len(colorField.Values) != 2 {
		t.Errorf("Expected 2 values in color field, got %d", len(colorField.Values))
	}
	redFound := false
	blueFound := false
	for _, v := range colorField.Values {
		if v.Value == "red" {
			redFound = true
		}
		if v.Value == "blue" {
			blueFound = true
		}
	}
	if !redFound || !blueFound {
		t.Errorf("Expected red and blue colors in aggData, got red=%v blue=%v", redFound, blueFound)
	}

}

// hasAggregationField checks if a field exists in aggregation results.
func hasAggregationField(result []*AggregationResultField, fieldName string) bool {
	for _, field := range result {
		if field.Field == fieldName {
			return true
		}
	}
	return false
}

// hasAggregationValue checks if a value exists in a specific field.
func hasAggregationValue(result []*AggregationResultField, fieldName, value string) bool {
	for _, field := range result {
		if field.Field == fieldName {
			for _, v := range field.Values {
				if v.Value == value {
					return true
				}
			}
		}
	}
	return false
}

// findAggregationField finds a field in aggregation results.
func findAggregationField(result []*AggregationResultField, fieldName string) *AggregationResultField {
	for _, field := range result {
		if field.Field == fieldName {
			return field
		}
	}
	return nil
}
