[![Go](https://github.com/k-samuel/faceted/actions/workflows/go.yml/badge.svg)](https://github.com/k-samuel/faceted/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/k-samuel/faceted?style=flat-square)](https://goreportcard.com/report/github.com/k-samuel/faceted)
[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/k-samuel/faceted/pkg/releases/latest)

# Golang Faceted Search Library v3.2.1
Port of PHP [k-samuel/faceted-search](https://github.com/k-samuel/faceted-search) branched from v3.2.1

Simplified and fast faceted search without using any additional servers such as ElasticSearch, etc.

It can easily process up to 500,000 items with 10 properties. Create individual indices for product groups or categories and you won't need to scale or use more complex tools for a long time.

In addition to faceted filters, it supports exclusive filters.

## Features

- Fast faceted search without using additional servers (ElasticSearch, etc.)
- Support for up to 1,000,000+ records with 10 properties
- Filter aggregation (building available filter values)
- Exclusion filters
- Range filters (RangeFilter)
- Filters with AND conditions (ValueIntersectionFilter)
- Result sorting
- Indexing of numeric ranges (RangeIndexer, RangeListIndexer)

## Supported value types
Input:
```go
bool
int
int64
float32
float64
[]int
[]int64
[]string
[]interface{}
map[string]interface{}
```

*Interfaces must contain the primitives listed in this list*

The results of search.Aggregate() contain a list of available filter values, cast to a string type. Note that this simplifies processing the result structure.

If these types are insufficient, you need to inject your own value.ValueConverterInterface:

```go
import(
    "github.com/k-samuel/faceted"
    "github.com/k-samuel/faceted/pkg/value"
 )
//...
// Create index using Factory
search := faceted.NewSearch()
// Injecting value converter.
// Here you can set your own value.ValueConverter interface realisation
search = search.WithValueConverter(value.NewValueConverterDefault())
//...

```

### Golang version benchmark

Bench Golang (1.25) vs PHP (8.4.4 Opcache JIT, noxdebug) 1M records

|                         | GO         |     PHP   | 
|:------------------------|-----------:|----------:|
| Total Memory, Mb        |  134 Mb    | 417 Mb    |
| Find                    |  0.051789  | 0.022873  |
| Find & Sort             |  0.066998  | 0.030061  |
| Find (unsets)           |  0.067185  | 0.030475  |
| Find (ranges)           |  0.068525  | 0.031425  |
| Filters                 |  0.215391  | 0.065416  |
| Filters & count         |  0.390790  | 0.133543  |
| Filters & count & exc   |  0.423006  | 0.146758  |



# Note

Search index should be created in one thread before using. Currently, Index hash map access not using mutex. 
It can cause problems in concurrent writes and reads.


## Install

```bash
go get github.com/k-samuel/faceted
```

## Project structure

```
pkg/
├── filter/          # Filters (ValueFilter, RangeFilter, ExcludeValueFilter, etc.)
├── index/           # Indexes (Index)
├── indexer/         # Indexers (RangeIndexer, RangeListIndexer)
├── intersection/    # Intersections (ArrayIntersection)
├── query/           # Query (SearchQuery, AggregationQuery, Sort)
├── sort/            # Result sorters (AggregationResults, ArrayResults)
├── storage/         # Storages and scanners (ArrayStorage, Scanner)
cmd/
├── demo/            # Demo application
├── perf/            # Performance test
├── perf-data/       # Performance test data generator
└── tests/           # Unit tests
    └── data/        # Generated test data for performance test
main.go              # Simple examples
go.mod
```


## Quick start

### Creating an index

```go
package main

import (
    "github.com/k-samuel/faceted"
    "github.com/k-samuel/faceted/pkg/filter"
    "github.com/k-samuel/faceted/pkg/query"
)

func main() {
    // Create Index
    search := faceted.NewSearch()
    searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
    storage := searchIndex.GetStorage()

    // Add data
    data := []map[string]interface{}{
        {"id": 7, "color": "black", "price": 100, "sale": true, "size": 36},
        {"id": 9, "color": "green", "price": 100, "sale": true, "size": 40},
    }

    for _, item := range data {
        recordId := int(item["id"].(int))
        delete(item, "id")
        storage.AddRecord(recordId, item)
    }

    // Index optimization
    storage.Optimize()
}
```

### Search with filters

```go
import (
    "github.com/k-samuel/faceted"
    "github.com/k-samuel/faceted/filter"
    "github.com/k-samuel/faceted/query"
)
search := faceted.NewSearch()
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
storage := searchIndex.GetStorage()

// Create filters
filters := []filter.FilterInterface{
    search.NewValueFilter("color", []interface{}{"black", "green"}), // OR условие
    search.NewRangeFilter("size", search.NewRangeValue(36, 40),
}

// Search
searchQuery := search.NewSearchQuery().Filters(filters)
records := searchIndex.Query(searchQuery)
```

### Aggregation (building available filters)

```go
search := faceted.NewSearch()
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
// Aggregation without counting the quantity
aggQuery := search.NewAggregationQuery().Filters(filters)
aggData := searchIndex.Aggregate(aggQuery)

// Aggregation with counting and sorting
aggQuery2 := search.NewAggregationQuery().
    Filters(filters).
    CountItems(true).
    Sort(query.SortAsc, query.SortRegular)
aggData2 := searchIndex.Aggregate(aggQuery2)
```

### Исключающие фильтры

```go
search := faceted.NewSearch()
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
storage := searchIndex.GetStorage()
filters := []filter.FilterInterface{
    search.NewValueFilter("sale", []interface{}{1}),
    search.NewExcludeValueFilter("color", []interface{}{"blue"}),
}
records := searchIndex.Query(search.NewSearchQuery().Filters(filters))
```

### ValueIntersectionFilter (AND condition)

```go
search := faceted.NewSearch()
// For fields with multiple values
// Record: {"purpose": ["hunting", "fishing", "sports"]}
filter := search.NewValueIntersectionFilter("purpose", []interface{}{"hunting", "fishing"})
/ Finds records that contain both hunting and fishing
```

### RangeIndexer for numeric ranges

```go
search := faceted.NewSearch()
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
storage := searchIndex.GetStorage()

// Create an indexer with a step of 100
rangeIndexer, _ := search.NewRangeIndexer(100)
storage.AddIndexer("price", rangeIndexer)

// Add data
storage.AddRecord(1, map[string]interface{}{"price": 90})
storage.AddRecord(2, map[string]interface{}{"price": 150})

// Search by range
filters := []filter.FilterInterface{
    search.NewRangeFilter("price", search.NewRangeValue(100,)),
}
records := searchIndex.Query(search.NewSearchQuery().Filters(filters))
```

### RangeListIndexer for custom ranges

```go
search := faceted.NewSearch()
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
// Создание диапазонов: 0-99, 100-499, 500-999, 1000+
rangeIndexer, _ := search.NewRangeListIndexer([]int{100, 500, 1000})
searchIndex.GetStorage().AddIndexer("price", rangeIndexer)
```

### Sorting results

```go
search := faceted.NewSearch()
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
// Sort by price descending
searchQuery := search.NewSearchQuery().
    Filters(filters).
    Sort("price", query.SortDesc, query.SortNumeric)
records := searchIndex.Query(searchQuery)
```

### Index Export/Import

```go
search := faceted.NewSearch()
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
storage := searchIndex.GetStorage()

// Export
indexData := storage.Export()

// Import
searchIndex, _ := search.NewIndex(faceted.ArrayStorage)
searchIndex.GetStorage().SetData(indexData)
```

## API

### Фильтры

| Filter | Description |
|--------|---------|
| `ValueFilter` | Value filter (OR condition for multiple values) |
| `ValueIntersectionFilter` | Value filter (AND condition) |
| `RangeFilter` | Range filter (min, max) |
| `ExcludeValueFilter` | Value exclusion |
| `ExcludeRangeFilter` | Range exclusion |

### Query

| Query | Description |
|-------|----------|
| `SearchQuery` | Search query with filters and sorting |
| `AggregationQuery` | Aggregation query for building available filters |
| `Sort` | Sorting settings |
| `AggregationSort` | Aggregation sorting settings |

### Storage

| Storage | Description |
|---------|----------|
| `ArrayStorage` | Fast map-based storage |

### Demo application

```bash
cd cmd/demo
go run main.go
```
The local web server will start at http://127.0.0.1:8080/

![](docs/pic.png)


### Test
` go test  ./tests  -coverpkg  ./pkg/... -v -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html `

### Integration performance test (similar to PHP tests/performance/find.php)
Note: Runs from the project root directory.
```bash
# Create a test dataset (only needs to be done once)
go run cmd/perf-data/main.go -size 100000
# Run the test
go run cmd/perf/main.go -size 100000
```

### Simple Examples
```bash
 go run examples/sample/main.go
```

## License

MIT License