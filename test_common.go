package parrot

import (
	"context"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/micro/go-micro/client"
)

//MakeMockServiceParams makes params used for the Called method of mock
func MakeMockServiceParams(ctx context.Context, in protobuf.Message, opts ...client.CallOption) []interface{} {
	params := make([]interface{}, 0, len(opts)+2)
	params = append(params, ctx, in)
	for _, opt := range opts {
		params = append(params, opt)
	}
	return params
}
