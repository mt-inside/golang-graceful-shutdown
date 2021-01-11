package main

import (
	"fmt"
	"sync"

	"github.com/go-logr/zapr"
	"github.com/mt-inside/go-deadlines-test/pkg/workers"
)

func main() {
	//nolint:errcheck
	defer workers.Log.(zapr.Underlier).GetUnderlying().Sync()

	stopCh := make(chan struct{})
	defer close(stopCh)

	ns := merge(
		stopCh,
		workers.WorkerStream(stopCh),
		workers.WorkerStream(stopCh),
	)
	fmt.Println(<-ns)
	// i := 0
	// for n := range ns {
	// 	workers.Log.Info("got", "val", n)
	// 	i = i + 1
	// 	if i > 2 {
	// 		break
	// 	}
	// }
}

func merge(stopCh <-chan struct{}, cs ...<-chan int) <-chan int {
	var sem sync.WaitGroup // a counting semaphore
	out := make(chan int)

	for _, c := range cs {
		sem.Add(1)
		go func(ch <-chan int) { // not safe to capture the loop variable
			defer sem.Done()
			for n := range ch {
				select {
				case out <- n:
				case <-stopCh:
					workers.Log.Info("Stopped")
					return
				}
			}
		}(c)
	}

	go func() {
		sem.Wait()
		close(out)
	}()

	return out
}
