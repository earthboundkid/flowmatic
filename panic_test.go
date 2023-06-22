package workgroup_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/carlmjohnson/workgroup"
)

func try(f func()) (r any) {
	defer func() {
		r = recover()
	}()
	f()
	return
}

func TestDoTasks_panic(t *testing.T) {
	task := func(n int) (int, error) {
		if n == 3 {
			panic("3!!")
		}
		return n * 3, nil
	}
	var triples []int
	manager := func(n, triple int, err error) ([]int, bool) {
		triples = append(triples, triple)
		return nil, true
	}
	r := try(func() {
		workgroup.DoTasks(1, task, manager, 1, 2, 3, 4)
	})
	if r == nil {
		t.Fatal("should have panicked")
	}
	if r != "3!!" {
		t.Fatal(r)
	}
	if fmt.Sprint(triples) != "[3 6]" {
		t.Fatal(triples)
	}
}

func TestDoEach_panic(t *testing.T) {
	var (
		n   atomic.Int64
		err error
	)
	r := try(func() {
		err = workgroup.DoEach(1, []int64{1, 2, 3},
			func(delta int64) error {
				if delta == 2 {
					panic("boom")
				}
				n.Add(delta)
				return nil
			})
	})
	if err != nil {
		t.Fatal("should have panicked")
	}
	if r == nil {
		t.Fatal("should have panicked")
	}
	if r != "boom" {
		t.Fatal(r)
	}
	if n.Load() != 4 {
		t.Fatal(n.Load())
	}
}

func TestDo_panic(t *testing.T) {
	var (
		n   atomic.Int64
		err error
	)
	r := try(func() {
		err = workgroup.Do(
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
	})
	if err != nil {
		t.Fatal("should have panicked")
	}
	if r == nil {
		t.Fatal("should have panicked")
	}
	if r != "boom" {
		t.Fatal(r)
	}
	if n.Load() != 2 {
		t.Fatal(n.Load())
	}
}
