package flowmatic_test

import (
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/earthboundkid/flowmatic/v2"
)

func TestManageTasks_drainage(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
		m := map[int]struct {
			int
			error
		}{}
		manager := flowmatic.Manage(2, task)
		manager.Add(0, 1)
		for range manager.Exec() {
			if time.Since(start) > sleepTime {
				t.Fatal("slept too much")
			}
			in, out, err := manager.Result()
			m[in] = struct {
				int
				error
			}{out, err}
			if manager.HasErr() {
				break
			}
		}
		if s := fmt.Sprint(m); s != "map[1:text string]" {
			t.Fatal(s)
		}
		if time.Since(start) < sleepTime {
			t.Fatal("didn't sleep enough")
		}
		if !b {
			t.Fatal("didn't finish")
		}
	})
}

func TestManageTasks_drainage2(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {

		const sleepTime = 10 * time.Millisecond
		b := false
		task := func(n int) (int, error) {
			if n == 0 {
				time.Sleep(sleepTime)
				b = true
				return n, errors.New("text string")
			}
			return n, errors.New("-")
		}
		start := time.Now()
		m := map[int]struct {
			int
			error
		}{}
		manager := flowmatic.Manage(2, task)
		manager.Add(0, 1)
		for range manager.Exec() {
			if time.Since(start) > sleepTime {
				t.Fatal("slept too much")
			}
			in, out, err := manager.Result()
			m[in] = struct {
				int
				error
			}{out, err}
			break
		}
		if s := fmt.Sprint(m); s != "map[1:-]" {
			t.Fatal(s)
		}
		if time.Since(start) < sleepTime {
			t.Fatal("didn't sleep enough")
		}
		if !b {
			t.Fatal("didn't finish")
		}
	})
}
