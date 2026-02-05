package interceptors

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PanicRecoveryHandler() recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) error {
		fmt.Printf("panic recovery handler caught a panic: %v\n", p)
		return status.Error(codes.Unknown, "something bad happened")
	}
}
