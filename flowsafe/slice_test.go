package flowsafe_test

import (
	"testing"

	"github.com/carlmjohnson/flowmatic/flowsafe"
)

func TestSlice_panic(t *testing.T) {
	var safeslice flowsafe.Slice[string]
	safeslice.Unwrap()

	var r any
	func() {
		defer func() {
			r = recover()
		}()

		safeslice.Store("a")
	}()
	if r == nil {
		t.Fatal("expected panic for Add after Unwrap")
	}
}

func TestSlice_make(t *testing.T) {
	safeslice := flowsafe.MakeSlice[string](2)
	safeslice.Store("foo")
	s := safeslice.Unwrap()
	if len(s) != 1 {
		t.Fatalf("expected len 1; got %v", len(s))
	}
	if cap(s) != 2 {
		t.Fatalf("expected cap 1; got %v", cap(s))
	}
}
