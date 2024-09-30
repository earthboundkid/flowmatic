package flowmatic_test

import (
	"context"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/earthboundkid/flowmatic/v2"
)

func try(f func()) (r any) {
	defer func() {
		r = recover()
	}()
	f()
	return
}

func TestManage_panic(t *testing.T) {
	task := func(n int) (int, error) {
		if n == 3 {
			panic("3!!")
		}
		return n * 3, nil
	}
	var triples []int
	r := try(func() {
		m := flowmatic.Manage(1, task)
		m.Queue(1, 2, 3, 4, 5, 6, 7)
		for range m.Exec() {
			triples = append(triples, m.Output())
		}
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
	task2 := func(n int) (int, error) {
		return n * 3, nil
	}
	triples = nil
	r2 := try(func() {
		m := flowmatic.Manage(flowmatic.MaxProcs, task2)
		m.Queue(1, 2, 3, 4, 5, 6, 7)
		for range m.Exec() {
			for range m.Exec() {
				triples = append(triples, m.Output())
			}
		}
	})
	if r2 != "already executing" {
		t.Fatal("should have panicked")
	}
	var triples2 []int
	r3 := try(func() {
		m := flowmatic.Manage(flowmatic.MaxProcs, task2)
		m.Queue(1, 2, 3, 4)
		for range m.Exec() {
			triples = append(triples, m.Output())
		}
		slices.Sort(triples)
		m.Queue(5, 6, 7, 8)
		for range m.Exec() {
			triples2 = append(triples2, m.Output())
		}
		slices.Sort(triples2)
	})
	if r3 != nil {
		t.Fatal("should not have panicked")
	}
	if fmt.Sprint(triples) != "[3 6 9 12]" {
		t.Fatal(triples)
	}
	if fmt.Sprint(triples2) != "[15 18 21 24]" {
		t.Fatal(triples2)
	}
}

func TestEach_panic(t *testing.T) {
	var (
		n   atomic.Int64
		err error
	)
	r := try(func() {
		err = flowmatic.Each(1, slices.Values([]int64{1, 2, 4, 8, 16}),
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
	if n.Load() != 29 {
		t.Fatal(n.Load())
	}
}

func TestDo_panic(t *testing.T) {
	var (
		n   atomic.Int64
		err error
	)
	r := try(func() {
		err = flowmatic.Do(
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

func TestRace_panic(t *testing.T) {
	var (
		n   atomic.Int64
		err error
	)
	r := try(func() {
		err = flowmatic.Race(context.Background(),
			func(context.Context) error {
				n.Add(1)
				return nil
			},
			func(context.Context) error {
				panic("boom")
			},
			func(context.Context) error {
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

func TestAll_panic(t *testing.T) {
	var (
		n   atomic.Int64
		err error
	)
	r := try(func() {
		err = flowmatic.All(context.Background(),
			func(context.Context) error {
				n.Add(1)
				return nil
			},
			func(context.Context) error {
				panic("boom")
			},
			func(context.Context) error {
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

func TestMap_panic(t *testing.T) {
	var (
		err error
		o   []int64
	)
	ctx := context.Background()
	r := try(func() {
		o, err = flowmatic.Map(ctx, 1, []int64{1, 2, 3},
			func(_ context.Context, delta int64) (int64, error) {
				if delta == 2 {
					panic("boom")
				}
				return 2 * delta, nil
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
	if o != nil {
		t.Fatal(o)
	}
}
