package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	jaegerGoTest "jaegerGoTest/proto/gen/proto"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

var tracer = otel.Tracer("main")

type MyServer struct {
	jaegerGoTest.UnimplementedJaegerGoTestServer
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("otel-collector:4317"),
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

	interceptors := []grpc.UnaryServerInterceptor{
		otelgrpc.UnaryServerInterceptor(),
		RateLimiterUnaryInterceptor(),
	}

	// Create gRPC server
	service := &MyServer{}
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
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
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	}

	mux := runtime.NewServeMux()

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

func RateLimiterUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		err := take(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func take(ctx context.Context) error {
	_, span := tracer.Start(ctx, "rateLimit", trace.WithAttributes(
		attribute.Bool("timed_out", false),
	))
	defer span.End()

	channel := make(chan error, 1)

	// Run the potentially long-running limiter function in its own goroutine and pass back the
	// response into the channel
	go func() {
		channel <- runLimiter(ctx, span.SpanContext())
	}()

	// Listen on the limiter channel and a timeout channel - whichever happens first
	// If the function times out, let the request go through
	select {
	case res := <-channel:
		return res
	case <-time.After(1 * time.Second):
		log.Println(ctx, fmt.Sprintf("Limiter took more than 1 second to execute. Letting request through"))
		span.SetAttributes(attribute.Bool("timed_out", true))
		return nil
	}
}

func runLimiter(ctx context.Context, parentSpanContext trace.SpanContext) error {
	ctx, span := tracer.Start(ctx, "runLimiter", trace.WithLinks(trace.Link{
		SpanContext: parentSpanContext,
	}), trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// actually rate-limit here

	return nil
}

func (s *MyServer) Test(ctx context.Context, in *jaegerGoTest.StringMessage) (*jaegerGoTest.StringMessage, error) {
	_, span := tracer.Start(ctx, "GET /test")
	defer span.End()
	return &jaegerGoTest.StringMessage{Value: "done!"}, nil
}
