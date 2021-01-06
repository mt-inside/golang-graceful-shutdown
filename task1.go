package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Task1(stopCh <-chan struct{}) <-chan error {
	localCh := make(chan error)

	go func() {
		defer close(localCh)

		srv := &http.Server{
			Addr: ":8081",
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/task1/quit", func(w http.ResponseWriter, r *http.Request) {
			// cause run-time error in srv.ListenAndServe()
			srv.Close()
		})
		mux.HandleFunc("/task1", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Task1")
		})
		srv.Handler = mux
		log.Println("Task1 listening...")
		serverCh := channelWrapper(func() error { return srv.ListenAndServe() })

		select {
		case err := <-serverCh:
			/* local server is already shut down because it errored; just tell everyone else to shut down */
			localCh <- err
		case <-stopCh:
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
