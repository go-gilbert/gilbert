package containers

import "sync"

// OnceCell is a wrapper around sync.Once which initializes value once
// and returns a stored value.
type OnceCell[T any] struct {
	once  sync.Once
	value T
	ctor  func() T
}

// NewOnceCell constructs a new OnceCell with a provided value constructor.
func NewOnceCell[T any](ctor func() T) *OnceCell[T] {
	return &OnceCell[T]{
		ctor: ctor,
	}
}

// Get returns previously initialized value.
//
// Calls a constructor when value isn't previously initialized.
func (cell *OnceCell[T]) Get() T {
	cell.once.Do(func() {
		cell.value = cell.ctor()
	})

	return cell.value
}

// Reset resets a cell and removes stored value.
func (cell *OnceCell[T]) Reset() {
	cell.value = reset[T]()
	cell.once = sync.Once{}
}

func reset[T any]() T {
	var empty T
	return empty
}
