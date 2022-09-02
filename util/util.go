package util

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
