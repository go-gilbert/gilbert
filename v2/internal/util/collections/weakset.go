package collections

import (
	"reflect"
	"unsafe"
)

// WeakSet stores weakly held value references in a collection.
type WeakSet[T comparable] map[unsafe.Pointer]void

// Append appends a new reference to set.
func (set WeakSet[T]) Append(items ...*T) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		set[unsafe.Pointer(item)] = void{}
	}
}

// AppendSlice appends weak references to each slice element to a set.
func (set WeakSet[T]) AppendSlice(slice []T) {
	if len(slice) == 0 {
		return
	}

	elemSize := reflect.TypeOf(slice).Elem().Size()
	header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))

	for i := 0; i < header.Len; i++ {
		offset := elemSize * uintptr(i)
		elemAddr := unsafe.Pointer(header.Data + offset)
		set[elemAddr] = void{}
	}
}

// Has returns if reference exists in a set.
func (set WeakSet[T]) Has(item *T) bool {
	ptr := unsafe.Pointer(item)
	_, ok := set[ptr]
	return ok
}

// Remove removes reference to element from a set.
func (set WeakSet[T]) Remove(item *T) {
	ptr := unsafe.Pointer(item)
	delete(set, ptr)
}

// AsSlice returns slice of references.
//
// Referenced values might be already released by garbage collector.
func (set WeakSet[T]) AsSlice() []*T {
	if len(set) == 0 {
		return nil
	}

	slice := make([]*T, 0, len(set))
	for ptr := range set {
		slice = append(slice, (*T)(ptr))
	}

	return slice
}

// AsValueSlice returns slice of de-referenced values.
//
// Dangerous! Referenced values might be already released by garbage collector.
func (set WeakSet[T]) AsValueSlice() []T {
	if len(set) == 0 {
		return nil
	}

	slice := make([]T, 0, len(set))
	for ptr := range set {
		slice = append(slice, *(*T)(ptr))
	}

	return slice
}
