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

## Kubernetes (Local)
1. Install [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
2. Run `./deploy.sh`
3. On another window: `time go run client/main.go`
