package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"

	hellopb "github.com/mt-inside/golang-graceful-shutdown/grpc/pb"
)

func Task1(ctx context.Context) <-chan error {
	localCh := make(chan error)

	var opts []grpc.ServerOption
	srv := grpc.NewServer(opts...)
	hellopb.RegisterGreeterServer(srv, NewGreeterServer(ctx))

	reflection.Register(srv)
	service.RegisterChannelzServiceToServer(srv)

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
		lis, err := net.Listen("tcp", ":8081")
		if err != nil {
			log.Println(err, "failed to listen")
			localCh <- err
		}
		defer lis.Close()

		log.Println("Task1 serving...")
		serverCh := channelWrapperErr(func() error { return srv.Serve(lis) })

		select {
		case err := <-serverCh:
			/* local server is already shut down because it errored; just tell everyone else to shut down */
			localCh <- err
		case <-ctx.Done():
			log.Println("Gracefully shutting down Task1 server")

			// TODO Lower readiness probe here, so that when the client gets the GOAWAY and tries to reconnect, it doesn't hit you again.

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			select {
			case <-ctx.Done():
				err := ctx.Err()
				log.Println("failed to gracefully shutdown Task1 server", err)
				localCh <- err
			case <-channelWrapperUnit(func() { srv.GracefulStop() }): // takes no context, returns no error
				log.Println("Gracefully shut down Task1 server")
			}
		}
	}()

	return localCh
}
