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

	var err error
	select {
	case err = <-t1Ch:
	case err = <-t2Ch:
	case <-signalCh:
	}

	log.Println("main thread woken, due to error:", err)

	shutdown = true
	close(stopCh)

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
