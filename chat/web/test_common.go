package web

import (
	"context"
	"reflect"

	"github.com/peonone/parrot"

	"github.com/micro/go-micro/client"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/mock"
)

type mockUserClient struct {
	mock.Mock
}

func (c *mockUserClient) ReadJSON(v interface{}) error {
	returnVals := c.Called(v)
	// TODO use the first of return values as the value need to update the arg pointer
	// need find the best practice to do it

	vv := reflect.ValueOf(returnVals.Get(0))
	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr:
		ptrVal := reflect.ValueOf(v).Elem()
		for vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}
		ptrVal.Set(vv)
	case reflect.Map:
		mapV := v.(map[string]interface{})
		mapVv := returnVals.Get(0).(map[string]interface{})
		for k, v := range mapVv {
			mapV[k] = v
		}
	}
	return returnVals.Error(1)
}

func (c *mockUserClient) WriteJSON(v interface{}) error {
	returnVals := c.Called(v)
	return returnVals.Error(0)
}

func (c *mockUserClient) Close() error {
	return c.Called().Error(0)
}

type mockCmdHandler struct {
	mock.Mock
}

func (h *mockCmdHandler) canHandle(cmd string) bool {
	return h.Called(cmd).Bool(0)
}

func (h *mockCmdHandler) validate(req map[string]interface{}) error {
	return h.Called(req).Error(0)
}

func (h *mockCmdHandler) handle(client *onlineUser, req map[string]interface{}) {
	h.Called(client, req)
}

func (h *mockCmdHandler) canHandlePush(cmd string) bool {
	return h.Called(cmd).Bool(0)
}

func (h *mockCmdHandler) handlePush(msg *proto.PushMsg) {
	h.Called(msg)
}

type mockStateService struct {
	mock.Mock
}

func (s *mockStateService) Online(ctx context.Context, in *proto.UserOnlineReq, opts ...client.CallOption) (*proto.UserOnlineRes, error) {
	params := parrot.MakeMockServiceParams(ctx, in, opts...)
	returnValues := s.Called(params...)
	return returnValues.Get(0).(*proto.UserOnlineRes), returnValues.Error(1)
}

func (s *mockStateService) Offline(ctx context.Context, in *proto.UserOfflineReq, opts ...client.CallOption) (*proto.UserOfflineRes, error) {
	params := parrot.MakeMockServiceParams(ctx, in, opts...)
	returnValues := s.Called(params...)
	return returnValues.Get(0).(*proto.UserOfflineRes), returnValues.Error(1)
}
