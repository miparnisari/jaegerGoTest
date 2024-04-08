package interceptors

import (
	"context"
	"fmt"

	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PanicRecoveryHandler() recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) error {
		fmt.Println("panic")
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
}
