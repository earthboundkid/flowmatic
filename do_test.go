package workgroup_test

import (
	"fmt"
	"sync/atomic"
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

func TestDoTasks_panic(t *testing.T) {
	var n atomic.Int64
	err := workgroup.DoTasks(1, []int64{1, 2, 3},
		func(delta int64) error {
			if delta == 2 {
				panic("boom")
			}
			n.Add(delta)
			return nil
		})
	if err == nil {
		t.Fatal("should have panicked")
	}
	if err.Error() != "panic: boom" {
		t.Fatal(err)
	}
	if n.Load() != 1 {
		t.Fatal(n.Load())
	}
}

func TestDoFuncs_panic(t *testing.T) {
	var n atomic.Int64
	err := workgroup.DoFuncs(1,
		func() error {
			n.Add(1)
			return nil
		},
		func() error {
			panic("boom")
		},
		func() error {
			n.Add(1)
			return nil
		})
	if err == nil {
		t.Fatal("should have panicked")
	}
	if err.Error() != "panic: boom" {
		t.Fatal(err)
	}
	if n.Load() != 1 {
		t.Fatal(n.Load())
	}
}
