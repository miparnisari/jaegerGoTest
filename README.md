# Setup

1. Install Go and Docker.
2. Install `buf`: https://buf.build/docs/installation/
3. Run

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

# Demo

## Docker Compose
1. Run `docker compose up --build`
2. On another window: `curl http://localhost:8080/streamed -v`

## Kubernetes (Local)
1. Install [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
2. Run `./deploy.sh`
3. On another window: `curl http://localhost:8080/streamed -v`
4. `grpcurl -plaintext grpc.localhost:8081 list` and `grpcurl -plaintext grpc.localhost:8081 jaegerGoTest.JaegerGoTest/StreamedGetStoreID`

### Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                          Kind Cluster                                │
│                                                                      │
│  ┌─────────────────┐                                                 │
│  │ projectcontour  │                                                 │
│  │   namespace     │                                                 │
│  │                 │                                                 │
│  │ ┌─────────────┐ │     ┌──────────────────────────────────────┐    │
│  │ │   Contour   │ │────▶│              Envoy                   │    │
│  │ │ Deployment  │ │     │           DaemonSet                  │    │
│  │ │             │ │     │                                      │    │
│  │ │ - HTTPProxy │ │     │ - Load Balancer                      │    │
│  │ │   config    │ │     │ - HTTP/gRPC Proxy                    │    │
│  │ │ - Timeouts  │ │     │ - Access Logs                        │    │
│  │ │ - Routing   │ │     │ - Stream Timeouts (10s)              │    │
│  │ └─────────────┘ │     └──────────────────┬───────────────────┘    │
│  └─────────────────┘                        │                        │
│                                             │ :80                    │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │                    default namespace                            │ │
│  │                                                                 │ │
│  │  ┌───────────────────┐           ┌─────────────────────────────┐│ │
│  │  │ jaeger-go-test    │◀──────────│    jaeger-go-test           ││ │
│  │  │    Service        │           │      Deployment             ││ │
│  │  │                   │           │                             ││ │
│  │  │ - ClusterIP       │           │  ┌─────────────────────────┐││ │
│  │  │ - Port: 8080      │           │  │        Pod              │││ │
│  │  └───────────────────┘           │  │                         │││ │
│  │                                  │  │  ┌─────────────────────┐│││ │
│  │                                  │  │  │  Go Server          ││││ │
│  │                                  │  │  │                     ││││ │
│  │                                  │  │  │ - gRPC :8081        ││││ │
│  │                                  │  │  │ - HTTP Gateway :8080││││ │
│  │                                  │  │  │ - Streaming API     ││││ │
│  │                                  │  │  └─────────────────────┘│││ │
│  │                                  │  └─────────────────────────┘││ │
│  │                                  └─────────────────────────────┘│ │
│  └─────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
           ▲
           │ kubectl port-forward 8080:80
           │
    ┌─────────────┐
    │   localhost │
    │     :8080   │
    └─────────────┘

Request Flow:
1. curl http://localhost:8080/streamed
2. → localhost:8080 (port-forward)
3. → Envoy DaemonSet :80
4. → jaeger-go-test Service :8080
5. → jaeger-go-test Pod :8080
6. → Go HTTP Gateway → gRPC Server :8081
7. ← Streaming response
```