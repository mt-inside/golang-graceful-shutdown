package main

import (
	"context"
	"io"
	"log"
	"time"

	hellopb "github.com/mt-inside/golang-graceful-shutdown/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type greeterServer struct {
	hellopb.UnimplementedGreeterServer

	ctx context.Context
}

func NewGreeterServer(ctx context.Context) *greeterServer {
	return &greeterServer{ctx: ctx}
}

func (g *greeterServer) Greet(ctx context.Context, req *hellopb.GreetRequest) (*hellopb.GreetReply, error) {
	name := req.GetName()
	log.Println("Handling request", "name", name)
	return &hellopb.GreetReply{Greeting: "Hello " + name}, nil
}

func (g *greeterServer) PenFriend(reqs hellopb.Greeter_PenFriendServer) error {
	for {
		req, err := reqs.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		select {
		case <-g.ctx.Done():
			return status.Error(codes.Canceled, "Graceful shutdown")
		default:
		}

		name := req.GetName()
		log.Println("Handling request", "name", name)
		err = reqs.Send(&hellopb.GreetReply{Greeting: "Hello " + name})
		if err != nil {
			return err
		}
	}
}

func (g *greeterServer) ValidateMe(reqs hellopb.Greeter_ValidateMeServer) error {
	for {
		_, err := reqs.Recv()
		log.Println("Received")
		if err == io.EOF {
			log.Println("Got em all")
			break
		}
		if err != nil {
			log.Println("Receive error", err)
			return err
		}
	}

	for {
		select {
		case <-g.ctx.Done():
			return status.Error(codes.Canceled, "Graceful shutdown")
		default:
		}

		log.Println("Responding")
		err := reqs.Send(&hellopb.GreetReply{Greeting: "Hello you"})
		if err != nil {
			log.Println("Send error", err)
			return err
		}

		time.Sleep(1 * time.Second)
	}
}
