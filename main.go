package main

import (
	"context"
	"fmt"
	"time"

	"jaegerGoTest/interceptors"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"net"
	"net/http"
	"os/signal"
	"syscall"

	jaegerGoTest "jaegerGoTest/proto/gen/proto"
)

var tracer = otel.Tracer("main")

type MyServer struct {
	jaegerGoTest.UnimplementedJaegerGoTestServer
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("otel-collector-demo:4317"),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		panic(err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String("jaegerGoTest"),
		))
	if err != nil {
		panic(err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1))),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(traceExporter)),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)

	panicHandler := interceptors.PanicRecoveryHandler()

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(
			grpc_recovery.WithRecoveryHandlerContext(panicHandler),
		),
		grpc_validator.UnaryServerInterceptor(),
		interceptors.NewStoreIDUnaryInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpc_recovery.StreamServerInterceptor(
			grpc_recovery.WithRecoveryHandlerContext(panicHandler),
		),
		grpc_validator.StreamServerInterceptor(),
		interceptors.NewStoreIDStreamingInterceptor(),
	}

	// Create gRPC server
	service := &MyServer{}
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
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
		tracerProvider.ForceFlush(ctx)
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
	ctx, span := tracer.Start(ctx, "GET /store-id")
	defer span.End()

	// more than the timeout
	time.Sleep(1 * time.Second)
	return &jaegerGoTest.GetStoreResponse{Value: "some data!"}, nil
}

func (s *MyServer) StreamedGetStoreID(in *jaegerGoTest.StreamedGetStoreRequest, stream jaegerGoTest.JaegerGoTest_StreamedGetStoreIDServer) error {
	ctx := stream.Context()
	_, span := tracer.Start(ctx, "GET /stream/store-id")
	defer span.End()

	return nil
}
