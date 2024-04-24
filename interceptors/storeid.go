package interceptors

import (
	"context"
	"fmt"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewStoreIDUnaryInterceptor() grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(reportable())
}

func NewStoreIDStreamingInterceptor() grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(reportable())
}

type hasGetStoreID interface {
	GetStoreId() string
}

type reporter struct {
	ctx context.Context
}

func (r *reporter) PostCall(error, time.Duration) {}

func (r *reporter) PostMsgSend(interface{}, error, time.Duration) {}

func (r *reporter) PostMsgReceive(msg interface{}, err error, _ time.Duration) {
	if m, ok := msg.(hasGetStoreID); ok {
		storeID := m.GetStoreId()
		fmt.Println("setting header store-id")
		grpc.SetHeader(r.ctx, metadata.Pairs("store-id-header", storeID))
	}
}

func reportable() interceptors.CommonReportableFunc {
	return func(ctx context.Context, c interceptors.CallMeta) (interceptors.Reporter, context.Context) {
		r := reporter{ctx}
		return &r, r.ctx
	}
}
