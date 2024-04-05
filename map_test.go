package flowmatic_test

import (
	"context"
	"errors"
	"testing"

	"github.com/earthboundkid/flowmatic"
)

func TestMap(t *testing.T) {
	ctx := context.Background()
	a := errors.New("a")
	b := errors.New("b")
	o, errs := flowmatic.Map(ctx, 1, []int{1, 2, 3}, func(_ context.Context, i int) (int, error) {
		switch i {
		case 1:
			return 1, a
		case 2:
			return 2, b
		default:
			panic("should be canceled by now!")
		}
	})
	if !errors.Is(errs, a) {
		t.Fatal(errs)
	}
	if errors.Is(errs, b) {
		t.Fatal(errs)
	}
	if o != nil {
		t.Fatal(o)
	}
}
