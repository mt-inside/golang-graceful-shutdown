package main

import (
	"context"
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
		// cause run-time error in srv.ListenAndServe()
		srv.Close()
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Task1")
	})
	srv.Handler = mux

	// Only this part of the function needs to be off the main thread for avoid deadlocks. There may be non-functional performance advantages to running the above setup in parallel, but _concurrency is not parallelism_
	go func() {
		defer close(localCh)
		log.Println("Task1 listening...")
		serverCh := channelWrapper(func() error { return srv.ListenAndServe() })

		select {
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
