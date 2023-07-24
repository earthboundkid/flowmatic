package flowsafe

// Enum associates a value with a position in a slice.
type Enum[T any] struct {
	Pos   int
	Value *T
}

// Count is like enumerate in Python.
// It associates the members of a slice with their positions.
func Count[T any, S ~[]T](s S) []Enum[T] {
	r := make([]Enum[T], len(s))
	for i := range s {
		r[i] = Enum[T]{
			Pos:   i,
			Value: &s[i],
		}
	}
	return r
}
