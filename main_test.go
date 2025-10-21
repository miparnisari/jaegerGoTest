package main_test

import (
	jaegerGoTest "jaegerGoTest/proto/gen/proto"
	"testing"

	"github.com/planetscale/vtprotobuf/codec/grpc"
	_ "google.golang.org/grpc/encoding/proto"
)

func TestMarshal(t *testing.T) {
	test := &jaegerGoTest.TestMsg{Test: &jaegerGoTest.TestMsg_A{}}
	_, err := grpc.Codec{}.Marshal(test)
	if err != nil {
		panic(err)
	}
}
