package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stopCh := make(chan struct{})
	shutdown := false
	defer func() {
		if !shutdown {
			close(stopCh)
		}
	}()

	signalCh := installSignalHandlers()
	t1Ch := Task1(stopCh)
	t2Ch := Task2(stopCh)

	/* Wait in parallel - any one is sufficient to shutdown; they signal an error (or signal) */

	var err error
	select {
	case err = <-t1Ch:
	case err = <-t2Ch:
	case <-signalCh: // It might seem at first glance that this is redundant, and that we can not wait for signalCh here but pass signalCh direct to the Tasks. Try it, you get stuck.
	}

	log.Println("main thread woken, due to:", err)

	shutdown = true
	close(stopCh)

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

func installSignalHandlers() <-chan struct{} { // tempting to type this as an error, but it isn't - it's non-failure, non-exceptional behavour to get a SIGTERM from k8s
	stopCh := make(chan struct{}) // could wrap the signal in an error, but it's not exceptional behaviour
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalCh
		close(stopCh)
		log.Printf("Got signal: %v", sig)
		sig = <-signalCh
		log.Printf("Got signal: %v", sig)
		log.Println("You really mean it huh")
		os.Exit(1)
	}()

	return stopCh
}
