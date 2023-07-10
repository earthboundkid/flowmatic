package flowsafe_test

import (
	"testing"

	"github.com/carlmjohnson/flowmatic/flowsafe"
)

func TestMap_panic(t *testing.T) {
	var safem flowsafe.Map[string, string]
	safem.Unwrap()

	var r any
	func() {
		defer func() {
			r = recover()
		}()

		safem.Add("a", "b")
	}()
	if r == nil {
		t.Fatal("expected panic for Add after Unwrap")
	}
}

func TestMap_make(t *testing.T) {
	safem := flowsafe.MakeMap[string, int](1)
	safem.Add("foo", 0)
	m := safem.Unwrap()
	if len(m) != 1 {
		t.Fatalf("expected len 1; got %v", len(m))
	}
}
