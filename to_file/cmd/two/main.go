package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mt-inside/go-deadlines-test/pkg/workers"
)

func main() {
	timeout, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel() // if it returns before the deadline and we call out of this function, stop the timeout from going off so it doesn't write to a closed channel ?

	// read the blod
	// stream should close its channel when it's done
	// lump should return; deal with both styles
	// comment: no need to wait for termination of the work - either we're quitting, in which case the timeouts weren't necessary, or we're gonna go do something else, in which case why block?
	data, err := workers.WorkerLump(timeout)
	if err != nil {
		log.Fatal("failed to calcualte data:", err)
	}
	fmt.Println(data)
}
