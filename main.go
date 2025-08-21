package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"jaegerGoTest/interceptors"
	jaegerGoTest "jaegerGoTest/proto/gen/proto"
)

type MyServer struct {
	jaegerGoTest.UnimplementedJaegerGoTestServer
	counterContinuous, counterSporadic atomic.Int32
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
		grpc.KeepaliveParams(keepalive.ServerParameters{}),
	}
	grpcServer := grpc.NewServer(opts...)
	jaegerGoTest.RegisterJaegerGoTestServer(grpcServer, service)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}

	go func() {
		fmt.Println(fmt.Sprintf("grpc server listening on 8081"))
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

	//runtime.DefaultContextTimeout = 50 * time.Millisecond
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
		fmt.Println(fmt.Sprintf("HTTP server listening on 8080"))
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}()

	<-ctx.Done()
}

func (s *MyServer) StreamedContinuous(in *jaegerGoTest.StreamedContinuousRequest, stream jaegerGoTest.JaegerGoTest_StreamedContinuousServer) error {
	s.counterContinuous.Store(0)
	for {
		fmt.Println("Sending response", s.counterContinuous.Add(1))
		err := stream.Send(&jaegerGoTest.StreamedContinuousResponse{Value: s.counterContinuous.Load()})
		if err != nil {
			fmt.Println("stream send error:", err)
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *MyServer) StreamedSporadic(in *jaegerGoTest.StreamedSporadicRequest, stream jaegerGoTest.JaegerGoTest_StreamedSporadicServer) error {
	s.counterSporadic.Store(0)
	for {
		fmt.Println("Sending response", s.counterSporadic.Add(1))
		err := stream.Send(&jaegerGoTest.StreamedSporadicResponse{Value: s.counterSporadic.Load()})
		if err != nil {
			fmt.Println("stream send error:", err)
			return err
		}
		time.Sleep(5 * time.Second)
	}
}
