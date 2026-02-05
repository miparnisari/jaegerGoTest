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

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"jaegerGoTest/interceptors"
	jaegerGoTest "jaegerGoTest/proto/gen/proto"
)

type MyPanicServer struct {
	jaegerGoTest.UnimplementedPanicServiceServer
}

type MyStreamingService struct {
	jaegerGoTest.UnimplementedStreamingServiceServer
	counterContinuous, counterSporadic atomic.Int32
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	panicRecovery := interceptors.PanicRecoveryHandler()

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		recovery.UnaryServerInterceptor(
			recovery.WithRecoveryHandlerContext(panicRecovery),
		),
		interceptors.NewRequestIDUnaryInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		recovery.StreamServerInterceptor(
			recovery.WithRecoveryHandlerContext(panicRecovery),
		),
		interceptors.NewRequestIDStreamingInterceptor(),
	}

	// Create gRPC server
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
		grpc.KeepaliveParams(keepalive.ServerParameters{}),
	}
	grpcServer := grpc.NewServer(opts...)
	jaegerGoTest.RegisterPanicServiceServer(grpcServer, &MyPanicServer{})
	jaegerGoTest.RegisterStreamingServiceServer(grpcServer, &MyStreamingService{})
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}

	grpcDone := make(chan bool)
	go func() {
		defer close(grpcDone)
		fmt.Println(fmt.Sprintf("grpc server listening on 8081"))
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Errorf("failed to start gRPC server: %w", err)
		}
		fmt.Println("gRPC server stopped serving")
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

	//runtime.DefaultContextTimeout = 10 * time.Second
	mux := runtime.NewServeMux(muxOpts...)

	// Create reverse proxy http -> GRPC
	err = jaegerGoTest.RegisterPanicServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:8081"), dialOpts)
	if err != nil {
		panic(err)
	}
	err = jaegerGoTest.RegisterStreamingServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:8081"), dialOpts)
	if err != nil {
		panic(err)
	}

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	httpDone := make(chan bool)
	go func() {
		defer close(httpDone)
		fmt.Println(fmt.Sprintf("HTTP server listening on 8080"))
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Errorf("failed to start HTTP server: %w", err)
		}
		fmt.Println("HTTP server stopped serving")
	}()

	<-ctx.Done()
	fmt.Println("initiating shutdown...")

	shutdownGraceful := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		fmt.Println("gRPC server shutdown gracefully")
		close(shutdownGraceful)
	}()

	select {
	case <-shutdownGraceful:
	case <-time.After(10 * time.Second):
		fmt.Println("timeout waiting for server to shutdown gracefully; forcing shutdown")
		grpcServer.Stop()
	}

	httpServer.Shutdown(context.Background())

	<-httpDone
	<-grpcDone

	fmt.Printf("good bye!")
}

func (s *MyPanicServer) CausePanic(in *jaegerGoTest.PanicCausingReq, out grpc.ServerStreamingServer[jaegerGoTest.PanicCausingRes]) error {
	causePanic()
	return nil
}

func (s *MyPanicServer) CausePanicInGoroutine(in *jaegerGoTest.PanicCausingReq, out grpc.ServerStreamingServer[jaegerGoTest.PanicCausingRes]) error {
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()
		causePanic()
	}()
	<-done
	return nil
}

func (s *MyStreamingService) StreamedContinuous(in *jaegerGoTest.StreamedContinuousRequest, stream jaegerGoTest.StreamingService_StreamedContinuousServer) error {
	s.counterContinuous.Store(0)
	for {
		select {
		case <-stream.Context().Done():
			fmt.Println("stream context had an error:", stream.Context().Err())
			return status.Error(codes.Unknown, "stream context had an error")
		case <-time.After(5 * time.Millisecond):
			fmt.Println("Sending response", s.counterContinuous.Add(1))
			err := stream.Send(&jaegerGoTest.StreamedContinuousResponse{Value: s.counterContinuous.Load()})
			if err != nil {
				fmt.Println("stream send error:", err)
				return err
			}
		}
	}
}

func (s *MyStreamingService) StreamedSporadic(in *jaegerGoTest.StreamedSporadicRequest, stream jaegerGoTest.StreamingService_StreamedSporadicServer) error {
	s.counterSporadic.Store(0)
	for {
		select {
		case <-stream.Context().Done():
			fmt.Println("stream context had an error:", stream.Context().Err())
			return status.Error(codes.Unknown, "stream context had an error")
		case <-time.After(5 * time.Second):
			fmt.Println("Sending response", s.counterSporadic.Add(1))
			err := stream.Send(&jaegerGoTest.StreamedSporadicResponse{Value: s.counterSporadic.Load()})
			if err != nil {
				fmt.Println("stream send error:", err)
				return err
			}
		}
	}
}

func causePanic() {
	arr := []string{"1", "2", "3"}
	fmt.Println(arr[3])
}
