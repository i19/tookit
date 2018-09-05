package runner

import (
	"fmt"
	"testing"
	"context"
	"time"
)

func worker(name string, score int) (string, error) {
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("%s get %d", name, score), nil
}

func TestConcurrentRunWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	params := [][]interface{}{{"a", 1}, {"b", 2}, {"c", 3}, {"d", 4}, {"e", 5}}
	r := ConcurrentRunWithTimeout(ctx, worker, params)

	go func() {
		for err := range r.Errors() {
			t.Logf("err :%s\n", err.Error())
		}
	}()

	for result := range r.Results() {
		t.Logf("%s\n", result)
	}
}

func TestRunWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond)
	defer cancel()

	if _, err := RunWithTimeout(ctx, worker, "a", 1); err == context.DeadlineExceeded {
		t.Log("OK")
	} else {
		t.Fatal("faile")
	}
}
