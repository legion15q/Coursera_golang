package main

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

/*
	Тест, предложенный одним из учащихся курса, Ilya Boltnev
	https://www.coursera.org/learn/golang-webservices-1/discussions/weeks/2/threads/kI2PR_XtEeeWKRIdN7jcig

	В чем его преимущество по сравнению с TestPipeline?
	1. Он проверяет то, что все функции действительно выполнились
	2. Он дает представление о влиянии time.Sleep в одном из звеньев конвейера на время работы

	возможно кому-то будет легче с ним
	при правильной реализации ваш код конечно же должен его проходить
*/

func TestByIlia(t *testing.T) {

	var recieved uint32
	freeFlowJobs := []job{
		job(func(_, out chan interface{}) {
			fmt.Println("out <- uint32(1)")
			out <- uint32(1)
			fmt.Println("out <- uint32(3)")
			out <- uint32(3)
			fmt.Println("out <- uint32(4)")
			out <- uint32(4)
			fmt.Println("func_1 done")
		}),
		job(func(in, out chan interface{}) {
			//range эквивалентен конструкции  val, ok := <-in, то есть происходит считывание из канал in
			fmt.Println("in func2")
			for val := range in {
				fmt.Println("out <- in (previous out)")
				out <- val.(uint32) * 3
				time.Sleep(time.Millisecond * 100)
			}

		}),
		job(func(in, _ chan interface{}) {
			for val := range in {
				fmt.Println("collected", val)
				atomic.AddUint32(&recieved, val.(uint32))
			}
		}),
	}

	start := time.Now()

	ExecutePipeline(freeFlowJobs...)

	end := time.Since(start)

	expectedTime := time.Millisecond * 350

	if end > expectedTime {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}

	if recieved != (1+3+4)*3 {
		t.Errorf("f3 have not collected inputs, recieved = %d", recieved)
	}
}
