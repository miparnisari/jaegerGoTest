package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	jaegerGoTest "jaegerGoTest/proto/gen/proto"
)

func main() {
	var (
		serverAddr = flag.String("server", "grpc.localhost:8081", "gRPC server address")
	)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time: 2 * time.Second,
		}))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := jaegerGoTest.NewJaegerGoTestClient(conn)

	fmt.Printf("Connecting to %s\n", *serverAddr)
	fmt.Println("Receiving streaming data... (Press Ctrl+C to stop)")

	stream, err := client.StreamedContinuous(context.Background(), &jaegerGoTest.StreamedContinuousRequest{})
	if err != nil {
		log.Fatalf("Failed to call StreamedGetStoreID: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nShutting down...")
			return
		default:
			resp, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("Stream ended")
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive from stream: %v", err)
			}

			fmt.Printf("Received value: %d\n", resp.Value)
		}
	}
}
