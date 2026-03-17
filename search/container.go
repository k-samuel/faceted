package search

import (
	"github.com/k-samuel/faceted/search/value"
)

// Factory creates Index instances.
type Container struct {
	Converter value.ConverterInterface
}

type searchOption func(*Container)

func WithValueConverter(converter value.ConverterInterface) searchOption {
	return func(s *Container) {
		s.Converter = converter
	}
}

// NewFactory creates a new Search Factory.
func NewContainer(opts ...searchOption) *Container {
	s := &Container{
		Converter: value.NewConverter(),
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

// NewIndex creates a new Index with MapStorage
func (f *Container) NewDb() *Db {
	st := NewMapStorage(f.Converter)
	return NewDb(
		st,
		NewMapScanner(st),
	)
}
