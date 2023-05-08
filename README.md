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
1. Go to https://app.lightstep.com/ACCOUNT-HERE/developer-mode and create a satellite token
2. Update `docker-compose.yaml`
3. Run `docker compose up --build`
4. On another window: `curl localhost:8080/test`
5. Open Jaeger (`http://localhost:16686`) and Lightstep (`https://app.lightstep.com/ACCOUNT/developer-mode`) to view traces

![jaeger 1](img/1.png)

![jaeger 2](img/2.png)