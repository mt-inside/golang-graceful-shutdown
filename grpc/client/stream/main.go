package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/mt-inside/go-usvc"
	hellopb "github.com/mt-inside/golang-graceful-shutdown/grpc/pb"
	"google.golang.org/grpc"
)

func main() {
	log := usvc.GetLogger(true)

	serverAddr := "localhost:8081"

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	log.V(1).Info("Dialling", "address", serverAddr)
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Error(err, "Can't connect")
		os.Exit(1)
	}
	log.V(1).Info("Connected", "address", serverAddr)
	defer conn.Close()

	go func() {
		for {
			if conn.WaitForStateChange(context.Background(), conn.GetState()) {
				log.Info("Conn state changed", "state", conn.GetState())
			} else {
				panic(nil) // context expired
			}
		}
	}()

	greeter := hellopb.NewGreeterClient(conn)

	streamCtx, streamCancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Info("Got signal", "signal", sig)
		streamCancel()
	}()

	log.V(1).Info("Calling", "method", "ValidateMe")
	friend, err := greeter.ValidateMe(streamCtx)
	if err != nil {
		log.Error(err, "Couldn't call", "method", "ValidateMe")
		os.Exit(1)
	}

	recvCtx, recvCancel := context.WithCancel(context.Background())

	// Get ready to receive. Server is welcome to start replying before we've finished sending
	go func() {
		for {
			greeting, err := friend.Recv()
			if err == io.EOF { // Corresponds to "return nil" server-side. TODO == codes.OK?
				log.Info("Got EOF")
				recvCancel()
				return
			}
			if err != nil {
				log.Error(err, "Failed to get a greeting")
				os.Exit(1)
			}
			log.Info("Received")

			fmt.Println(greeting.GetGreeting())
		}
	}()

	// what would it do if the name resolves to multiple As? Establish several at the start? Reissue the call if one of them GOAWAYS?
	// what does envoy do when it's in the middle? (cause gRPC client only sees one Server)

	// _sending_ a stream kinda just because
	for i := 0; i < 3; i++ {
		err = friend.Send(&hellopb.GreetRequest{Name: "fred"})
		if err != nil {
			log.Error(err, "Couldn't pester my friend")
			os.Exit(1)
		}
		log.Info("Sent")
	}
	friend.CloseSend()
	log.Info("All sent")

	<-recvCtx.Done()
	log.Info("Done")
}
