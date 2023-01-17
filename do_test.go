package workgroup_test

import (
	"fmt"
	"testing"

	"github.com/carlmjohnson/workgroup"
)

func TestDo_panic(t *testing.T) {
	task := func(n int) (int, error) {
		if n == 3 {
			panic("3!!")
		}
		return n * 3, nil
	}
	var triples []int
	manager := func(n, triple int, err error) ([]int, error) {
		triples = append(triples, triple)
		return nil, nil
	}
	err := workgroup.Do(1, task, manager, 1, 2, 3, 4)
	if err == nil {
		t.Fatal("should have panicked")
	}
	if err.Error() != "panic: 3!!" {
		t.Fatal(err)
	}
	if fmt.Sprint(triples) != "[3 6]" {
		t.Fatal(triples)
	}
}
