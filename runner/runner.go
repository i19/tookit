package runner

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

var (
	illegalInputError  = errors.New("illegal input")
	illegalOutputError = errors.New("illegal output")
)

type r struct {
	results chan interface{}
	errors  chan error
}

func (r *r) Results() <-chan interface{} {
	return r.results
}

func (r *r) Errors() <-chan error {
	return r.errors
}

func (r *r) run(ctx context.Context, fn interface{}, params [][]interface{}) {
	var wgForRequest sync.WaitGroup

	for _, param := range params {
		wgForRequest.Add(1)
		go func(p []interface{}) {
			defer wgForRequest.Done()
			if result, err := RunWithTimeout(ctx, fn, p...); err != nil {
				r.errors <- err
			} else {
				r.results <- result
			}
		}(param)
	}

	wgForRequest.Wait()
	close(r.results)
	close(r.errors)
}

func RunWithTimeout(ctx context.Context, fn interface{}, params ...interface{}, ) (interface{}, error) {
	// 检测输入的是否是函数，输入的参数和返回的结果数量是否正确
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return nil, illegalInputError
	}

	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != len(params) {
		return nil, illegalInputError
	}
	if f.Type().NumOut() != 2 {
		return nil, illegalOutputError
	}

	// 构建输入
	inputs := make([]reflect.Value, len(params))
	for k, in := range params {
		inputs[k] = reflect.ValueOf(in)
	}

	// 超时和完成处理
	done := make(chan struct{})
	resultError := make(chan error)

	var v interface{}
	go func(ins []reflect.Value) {
		results := f.Call(ins)
		if err := results[1].Interface(); err != nil {
			resultError <- err.(error)
		} else {
			v = results[0].Interface()
			done <- struct{}{}
		}
	}(inputs)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-resultError:
		return nil, err
	case <-done:
		return v, nil
	}
}

func ConcurrentRunWithTimeout(ctx context.Context, fn interface{}, params [][]interface{}) *r {
	tr := r{
		results: make(chan interface{}, 100),
		errors:  make(chan error, 100),
	}

	go tr.run(ctx, fn, params)

	return &tr
}
