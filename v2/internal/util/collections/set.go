package collections

type void = struct{}

// Set stores a sequence of unique items.
//
// Uses plain Go map under the hood.
type Set[T comparable] map[T]void

// NewSet creates a new Set from slice of items.
func NewSet[T comparable](items ...T) Set[T] {
	set := make(Set[T], len(items))
	set.Append(items...)
	return set
}

// Append appends new element to set.
func (s Set[T]) Append(items ...T) {
	if len(items) == 0 {
		return
	}

	for i := range items {
		s[items[i]] = void{}
	}
}

// Remove removes element from set.
//
// Equivalent of:
//
//	delete(set, item)
func (s Set[T]) Remove(item T) {
	delete(s, item)
}

// Has returns if value exists in a set.
func (s Set[T]) Has(item T) bool {
	_, ok := s[item]
	return ok
}

// AsSlice returns slice from elements in Set.
func (s Set[T]) AsSlice() []T {
	if len(s) == 0 {
		return nil
	}

	slice := make([]T, 0, len(s))
	for item := range s {
		slice = append(slice, item)
	}

	return slice
}
