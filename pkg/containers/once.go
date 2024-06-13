package containers

import "sync"

// SyncOnceCell is a wrapper around sync.Once which initializes value once
// and returns a stored value.
//
// Unlike `OnceCell`, this is a thread-safe implementation.
type SyncOnceCell[T any] struct {
	once  sync.Once
	value T
	ctor  func() T
}

// NewSyncOnceCell constructs a new SyncOnceCell with a provided value constructor.
func NewSyncOnceCell[T any](ctor func() T) *SyncOnceCell[T] {
	return &SyncOnceCell[T]{
		ctor: ctor,
	}
}

// Get returns previously initialized value.
//
// Calls a constructor when value isn't previously initialized.
func (cell *SyncOnceCell[T]) Get() T {
	cell.once.Do(func() {
		cell.value = cell.ctor()
	})

	return cell.value
}

// Reset resets a cell and removes stored value.
func (cell *SyncOnceCell[T]) Reset() {
	cell.value = reset[T]()
	cell.once = sync.Once{}
}

// OnceCell is a wrapper that provides lazy initialization of a singleton value.
//
// This is not a thread-safe implementation. Use `SyncOnceCell` for safe concurrent use.
type OnceCell[T any] struct {
	initialized bool
	value       T
	ctor        func() T
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
	if !cell.initialized {
		cell.value = cell.ctor()
	}

	return cell.value
}

// Reset resets a cell and removes stored value.
func (cell *OnceCell[T]) Reset() {
	cell.initialized = false
	cell.value = reset[T]()
}

func reset[T any]() T {
	var empty T
	return empty
}
