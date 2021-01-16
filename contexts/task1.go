package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Task1(ctx context.Context) <-chan error {
	localCh := make(chan error)

	srv := &http.Server{
		Addr: ":8081",
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Task1 /quit")
		// cause run-time error in srv.ListenAndServe()
		srv.Close()
	})

	// Handlers are run on separate goroutines, so send an error sideways to cause the "main" goroutine to panic
	panicCh := make(chan error)
	mux.HandleFunc("/panic-serve", func(w http.ResponseWriter, r *http.Request) {
		// make the "main" thread longjmp
		panicCh <- errors.New("Task1 /panic-serve")
	})

	// Handlers are run on separate goroutines, so a panic on one needs catching by a "middleware" introduced into its callstack, which then needs to send the error sideways to the "main" goroutine. Note that here, BOTH tasks gracefully shutdown, as neither had an error; panicing handlers just kill their own goroutine and you can just carry on if you want.
	mux.HandleFunc("/panic-handler", recoverWrapper(localCh, func(w http.ResponseWriter, r *http.Request) {
		// try to longjmp out
		panic("Task1 /panic-handler")
	}))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Task1")
	})

	srv.Handler = mux

	// Only this part of the function needs to be off the main thread for avoid deadlocks. There may be non-functional performance advantages to running the above setup in parallel, but _concurrency is not parallelism_
	go func() {
		defer close(localCh)

		// Catch any panics in this func or its children, and send back to main().
		// Note: handlers do NOT run on this thread; they are not children of ListenAndServe() */
		// Ordering with closing localCh is important; defers are run LIFO
		defer func() {
			// recover() must be called directly from the unwinding stack frame
			err := recoverToError(recover())
			localCh <- err
		}()

		log.Println("Task1 listening...")
		serverCh := channelWrapper(func() error { return srv.ListenAndServe() })

		select {
		case err := <-panicCh:
			// Bit of gymnastics to cause a panic on this thread as a result of a handler call
			panic(err)
		case err := <-serverCh:
			/* local server is already shut down because it errored; just tell everyone else to shut down */
			localCh <- err
		case <-ctx.Done():
			log.Println("Gracefully shutting down Task1 server")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

			if false {
				// sabotage shutdown
				srv.RegisterOnShutdown(func() { cancel() })
			} else {
				// normal behaviour
				defer cancel()
			}

			if err := srv.Shutdown(ctx); err != nil {
				log.Println("failed to gracefully shutdown Task1 server", err)
				localCh <- err
			}
			log.Println("Gracefully shut down Task1 server")
		}
	}()

	return localCh
}
