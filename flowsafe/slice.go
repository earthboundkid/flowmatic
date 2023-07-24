package flowsafe

import "sync"

// Slice is a goroutine safe slice container.
type Slice[T any] struct {
	l    sync.Mutex
	s    []T
	done bool
}

// MakeSlice makes a Slice with the specified capacity.
func MakeSlice[T any](cap int) *Slice[T] {
	return &Slice[T]{s: make([]T, 0, cap)}
}

// Store appends v to s in a goroutine safe way.
// Store panics if the user attempts to append to a slice
// that has already been finalized by [Slice.Unwrap].
func (s *Slice[T]) Store(v T) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.done {
		panic("must not push to Slice after finalizing")
	}
	s.s = append(s.s, v)
}

// Unwrap returns the slice underlying s.
// It does not copy.
// After a call to Unwrap,
// the user must not call Store again.
func (s *Slice[T]) Unwrap() []T {
	s.l.Lock()
	defer s.l.Unlock()

	s.done = true
	return s.s
}
