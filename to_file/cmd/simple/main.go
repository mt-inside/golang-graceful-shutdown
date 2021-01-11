package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mt-inside/go-deadlines-test/pkg/workers"
)

func main() {
	//timeout, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	//defer cancel() // if it returns before the deadline and we call out of this function, stop the timeout from going off so it doesn't write to a closed channel ?

	ns := workers.WorkerStream()
	for n := range ns {
		workers.Log.Info("got", "val", n)
	}

	time.Sleep(2 * time.Second)
	//cancel()
}

func awk2() {
	timeout, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	ss := make(chan int)

	go func() {
		err := workers.WorkerBlockingStream(timeout, ss)
		if err != nil {
			workers.Log.Error(err, "quitting")
			os.Exit(1)
		}
	}()
	for s := range ss {
		fmt.Println(s)
	}
}
