package flowsafe

import "sync"

// Map is a goroutine safe map container.
type Map[K comparable, V any] struct {
	l    sync.Mutex
	m    map[K]V
	done bool
}

// MakeMap makes a Map with the specified capacity.
func MakeMap[K comparable, V any](cap int) *Map[K, V] {
	return &Map[K, V]{m: make(map[K]V, cap)}
}

// Store adds its key and value to Map.
// Store panics if the user attempts to add to a map
// that has already been finalized by [Map.Unwrap].
func (m *Map[K, V]) Store(key K, value V) {
	m.l.Lock()
	defer m.l.Unlock()

	if m.done {
		panic("must not add to Map after finalizing")
	}
	if m.m == nil {
		m.m = make(map[K]V)
	}
	m.m[key] = value
}

// Unwrap returns the map underlying m.
// It does not copy.
// After a call to Unwrap,
// the user must not call Store again.
func (m *Map[K, V]) Unwrap() map[K]V {
	m.l.Lock()
	defer m.l.Unlock()
	m.done = true
	return m.m
}
