package interceptors

import (
	"context"
	"fmt"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/oklog/ulid/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewStoreIDUnaryInterceptor() grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(reportable())
}

func NewStoreIDStreamingInterceptor() grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(reportable())
}

type reporter struct {
	ctx context.Context
}

func (r *reporter) PostCall(error, time.Duration) {}

func (r *reporter) PostMsgSend(interface{}, error, time.Duration) {}

func (r *reporter) PostMsgReceive(msg interface{}, err error, _ time.Duration) {
	requestID := ulid.Make().String()
	err = grpc.SetHeader(r.ctx, metadata.Pairs("request-id", requestID))
	if err != nil {
		fmt.Println("error setting header", err)
	} else {
		fmt.Println("set header request-id:", requestID)
	}
}

func reportable() interceptors.CommonReportableFunc {
	return func(ctx context.Context, c interceptors.CallMeta) (interceptors.Reporter, context.Context) {
		r := reporter{ctx}
		return &r, r.ctx
	}
}
