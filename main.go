package main

import (
	"context"
	"fmt"
	"time"

	"jaegerGoTest/interceptors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"net"
	"net/http"
	"os/signal"
	"syscall"

	jaegerGoTest "jaegerGoTest/proto/gen/proto"
)

type MyServer struct {
	jaegerGoTest.UnimplementedJaegerGoTestServer
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		interceptors.NewStoreIDUnaryInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		interceptors.NewStoreIDStreamingInterceptor(),
	}

	// Create gRPC server
	service := &MyServer{}
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}
	grpcServer := grpc.NewServer(opts...)
	jaegerGoTest.RegisterJaegerGoTestServer(grpcServer, service)

	lis, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		panic(err)
	}

	go func() {
		fmt.Println(fmt.Sprintf("grpc server listening"))
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Errorf("failed to start gRPC server: %w", err)
		}
	}()

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	muxOpts := []runtime.ServeMuxOption{
		runtime.WithErrorHandler(func(ctx context.Context, sr *runtime.ServeMux, mm runtime.Marshaler, w http.ResponseWriter, r *http.Request, e error) {
			fmt.Println("error handler called", e)
			runtime.DefaultHTTPErrorHandler(ctx, sr, mm, w, r, e)
		}),
		runtime.WithStreamErrorHandler(func(ctx context.Context, e error) *status.Status {
			fmt.Println("stream error handler called", e)
			return runtime.DefaultStreamErrorHandler(ctx, e)
		}),
	}

	runtime.DefaultContextTimeout = 50 * time.Millisecond
	mux := runtime.NewServeMux(muxOpts...)

	// Create reverse proxy http -> GRPC
	err = jaegerGoTest.RegisterJaegerGoTestHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:8081"), dialOpts)
	if err != nil {
		return
	}

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	httpServer.RegisterOnShutdown(func() {
		grpcServer.GracefulStop()
	})

	go func() {
		fmt.Println(fmt.Sprintf("HTTP server listening"))
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}()

	<-ctx.Done()
}

func (s *MyServer) GetStoreID(ctx context.Context, in *jaegerGoTest.GetStoreRequest) (*jaegerGoTest.GetStoreResponse, error) {
	// more than the timeout
	time.Sleep(1 * time.Second)
	return &jaegerGoTest.GetStoreResponse{Value: "some data!"}, nil
}

func (s *MyServer) StreamedGetStoreID(in *jaegerGoTest.StreamedGetStoreRequest, stream jaegerGoTest.JaegerGoTest_StreamedGetStoreIDServer) error {
	return nil
}
