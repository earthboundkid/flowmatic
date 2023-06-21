package workgroup_test

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

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

func TestDoAll_panic(t *testing.T) {
	var n atomic.Int64
	err := workgroup.DoAll(1, []int64{1, 2, 3},
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

func TestDo_drainage(t *testing.T) {
	const sleepTime = 10 * time.Millisecond
	b := false
	task := func(n int) (int, error) {
		if n == 1 {
			return 0, errors.New("text string")
		}
		time.Sleep(sleepTime)
		b = true
		return 0, nil
	}
	start := time.Now()
	manager := func(in, out int, err error) ([]int, error) {
		t.Log(in, out, err)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	err := workgroup.Do(5, task, manager, 0, 1)
	if err == nil {
		t.Fatal("returned err")
	}
	if err.Error() != "text string" {
		t.Fatal(err)
	}
	if time.Since(start) < sleepTime {
		t.Fatal("didn't sleep enough")
	}
	if !b {
		t.Fatal("didn't finish")
	}
}
