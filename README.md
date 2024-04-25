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
1. Run `docker compose up --build`
4. On another window: `curl http://localhost:8080/01HVHZX1S19BBD8J8CVW535KB0 -v`