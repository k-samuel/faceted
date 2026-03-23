package search

import (
	"sort"
)

type FilterResultCache struct {
	Name   string
	Values map[int]struct{}
}

type ResultCache struct {
	Data    []*FilterResultCache
	Filters map[string]struct{}
}

func NewResultCache(size int) *ResultCache {
	return &ResultCache{
		Data:    make([]*FilterResultCache, 0, size),
		Filters: make(map[string]struct{}, size),
	}
}

func (i *ResultCache) Add(cache *FilterResultCache) {
	i.Data = append(i.Data, cache)
	i.Filters[cache.Name] = struct{}{}
}

func (i *ResultCache) SortByCount() {
	sort.Slice(i.Data, func(a, b int) bool {
		return len(i.Data[a].Values) < len(i.Data[b].Values)
	})
}
