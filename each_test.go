package flowmatic_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/earthboundkid/flowmatic/v2"
)

func TestEach_err(t *testing.T) {
	a := errors.New("a")
	b := errors.New("b")
	errs := flowmatic.Each(1, slices.Values([]int{1, 2, 3}), func(i int) error {
		switch i {
		case 1:
			return a
		case 2:
			return b
		default:
			return nil
		}
	})
	if !errors.Is(errs, a) {
		t.Fatal(errs)
	}
	if !errors.Is(errs, b) {
		t.Fatal(errs)
	}
}
