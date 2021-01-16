package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Serving for 5s")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		cancel() // mutliple cancels is not an error
	}()

	signalCh := installSignalHandlers()
	t1Ch := Task1(ctx)
	t2Ch := Task2(ctx)

	/* Wait in parallel - any one is sufficient to shutdown; they signal an error (or signal) */

	var err error
	select {
	case err = <-t1Ch:
	case err = <-t2Ch:
	case <-signalCh: // It might seem at first glance that this is redundant, and that we can not wait for signalCh here but pass signalCh direct to the Tasks. Try it, you get stuck.
	}

	log.Println("main thread woken, due to error:", err)

	cancel()

	/* Wait in series - all are needed to quit; they signal that shutdown is complete */

	log.Println("waiting for shutdown")
	for _, ch := range []<-chan error{t1Ch, t2Ch} {
		if err := <-ch; err != nil {
			log.Println("Error during shutdown:", err)
		}
	}
	log.Println("shutdown complete")
}

func channelWrapper(fn func() error) <-chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		ch <- fn()
	}()
	return ch
}

func recoverWrapper(errCh chan<- error, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recover() must be called directly from the unwinding stack frame
			if err := recoverToError(recover()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				errCh <- err
			}
		}()
		h.ServeHTTP(w, r)
	}
}

func recoverToError(r interface{}) error {
	var err error
	if r != nil {
		switch t := r.(type) {
		case string:
			err = errors.New(t)
		case error:
			err = t
		default:
			err = errors.New("Unknown error")
		}
		log.Println("Recovering:", err)
	}
	return err
}

func installSignalHandlers() <-chan struct{} {
	stopCh := make(chan struct{}) // could wrap the signal in an error, but it's not exceptional behaviour
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		close(stopCh)
		<-signalCh
		log.Println("You really mean it huh")
		os.Exit(1)
	}()

	return stopCh
}
