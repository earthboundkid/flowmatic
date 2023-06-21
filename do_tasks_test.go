package workgroup_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/carlmjohnson/workgroup"
)

func TestDoTasks_drainage(t *testing.T) {
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
	manager := func(in, out int, err error) ([]int, bool) {
		m[in] = struct {
			int
			error
		}{out, err}
		if err != nil {
			return nil, false
		}
		return nil, true
	}
	workgroup.DoTasks(5, task, manager, 0, 1)
	if s := fmt.Sprint(m); s != "map[1:text string]" {
		t.Fatal(s)
	}
	if time.Since(start) < sleepTime {
		t.Fatal("didn't sleep enough")
	}
	if !b {
		t.Fatal("didn't finish")
	}
}
