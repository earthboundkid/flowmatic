package flowmatic_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/earthboundkid/flowmatic"
)

func TestManageTasks_drainage(t *testing.T) {
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
	for r := range flowmatic.Tasks(5, task, 0, 1) {
		m[r.In] = struct {
			int
			error
		}{r.Out, r.Err}
		if r.HasErr() {
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
}
