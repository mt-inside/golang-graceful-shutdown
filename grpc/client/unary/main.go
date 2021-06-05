package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	greeter := hellopb.NewGreeterClient(conn)

	log.V(1).Info("Calling", "method", "Greet")
	greeting, err := greeter.Greet(context.TODO(), &hellopb.GreetRequest{Name: "fred"})
	if err != nil {
		log.Error(err, "Couldn't request greeting")
		os.Exit(1)
	}
	log.Info("Received")

	fmt.Println(greeting.GetGreeting())

	time.Sleep(10 * time.Second) // Outlive the server

	log.Info("Done")
}
