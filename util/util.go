package util

// https://itnext.io/generic-map-filter-and-reduce-in-go-3845781a591c
type Iterator[T any] interface {
	Next() bool
	Value() T
}
type mapIterator[T any] struct {
	source Iterator[T]
	mapper func(T) T
}

// advance to next element
func (iter *mapIterator[T]) Next() bool {
	return iter.source.Next()
}

func (iter *mapIterator[T]) Value() T {
	value := iter.source.Value()
	return iter.mapper(value)
}
func Map[T any](iter Iterator[T], f func(T) T) Iterator[T] {
	return &mapIterator[T]{
		iter, f,
	}
}

type filterIterator[T any] struct {
	source Iterator[T]
	pred   func(T) bool
}

func (iter *filterIterator[T]) Next() bool {
	for iter.source.Next() {
		if iter.pred(iter.source.Value()) {
			return true
		}
	}
	return false
}

func (iter *filterIterator[T]) Value() T {
	return iter.source.Value()
}

func Filter[T any](iter Iterator[T], pred func(T) bool) Iterator[T] {
	return &filterIterator[T]{
		iter, pred,
	}
}
