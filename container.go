package flowmatic

import "sync"

// Slice is a goroutine safe slice container.
type Slice[T any] struct {
	m    sync.Mutex
	s    []T
	done bool
}

// MakeSlice makes a Slice with the specified capacity.
func MakeSlice[T any](cap int) *Slice[T] {
	return &Slice[T]{s: make([]T, 0, cap)}
}

// Push appends v to s in a goroutine safe way.
// Push panics if the user attempts to append to a slice
// that has already been finalized by [Slice.Slice].
func (s *Slice[T]) Push(v T) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.done {
		panic("must not push to Slice after finalizing")
	}
	s.s = append(s.s, v)
}

// Slice returns the slice underlying s.
// It does not copy.
// After a call to Slice,
// the user must not call Push again.
func (s *Slice[T]) Slice() []T {
	s.m.Lock()
	defer s.m.Unlock()

	s.done = true
	return s.s
}

// Map is a goroutine safe map container.
type Map[K comparable, V any] struct {
	m    sync.Mutex
	mm   map[K]V
	done bool
}

// MakeMap makes a Map with the specified capacity.
func MakeMap[K comparable, V any](cap int) *Map[K, V] {
	return &Map[K, V]{mm: make(map[K]V, cap)}
}

// Add adds its key and value to Map.
// Add panics if the user attempts to add to a map
// that has already been finalized by [Map.Map].
func (m *Map[K, V]) Add(key K, value V) {
	m.m.Lock()
	defer m.m.Unlock()

	if m.done {
		panic("must not add to Map after finalizing")
	}
	if m.mm == nil {
		m.mm = make(map[K]V)
	}
	m.mm[key] = value
}

// Map returns the map underlying m.
// It does not copy.
// After a call to Map,
// the user must not call Add again.
func (m *Map[K, V]) Map() map[K]V {
	m.m.Lock()
	defer m.m.Unlock()
	m.done = true
	return m.mm
}
