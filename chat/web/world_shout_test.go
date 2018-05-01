package web

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/peonone/parrot"

	"github.com/micro/go-micro/client"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorldShoutService struct {
	mock.Mock
}

func (s *mockWorldShoutService) Send(ctx context.Context, in *proto.SendShoutReq, opts ...client.CallOption) (*proto.SendShoutRes, error) {
	params := parrot.MakeMockServiceParams(ctx, in, opts...)
	returnValues := s.Called(params...)
	return returnValues.Get(0).(*proto.SendShoutRes), returnValues.Error(1)
}

func TestWorldShoutHandler(t *testing.T) {
	oum := newOnlineUsersManager()
	baseCmdHandler := &baseCmdHandler{oum, nil}
	mockService := new(mockWorldShoutService)
	handler := &worldShoutHandler{baseCmdHandler, mockService}

	mockClient := new(mockUserClient)
	ou := &onlineUser{
		uid:             "peon1",
		userClient:      mockClient,
		pushCh:          make(chan interface{}, 10),
		lastRequestTime: time.Now(),
	}

	req := map[string]interface{}{
		"content": "hihi",
	}

	expectedReq := &proto.SendShoutReq{
		FromUID: ou.uid,
		Content: req["content"].(string),
	}
	successRes := &proto.SendShoutRes{
		Success: true,
	}
	mockService.On("Send", context.Background(), expectedReq).Return(successRes, nil).Once()
	handler.handle(ou, req)
	mockService.AssertExpectations(t)
	assert.Equal(t, len(ou.pushCh), 1)
	assert.Equal(t, successRes, <-ou.pushCh)

	errRes := &proto.SendShoutRes{Success: false, ErrMsg: "err1"}
	mockService.On("Send", context.Background(), expectedReq).Return(errRes, nil).Once()
	handler.handle(ou, req)
	mockService.AssertExpectations(t)
	assert.Equal(t, len(ou.pushCh), 1)
	assert.Equal(t, errRes, <-ou.pushCh)

	err := errors.New("internal err")
	var emptyResp *proto.SendShoutRes
	mockService.On("Send", context.Background(), expectedReq).Return(emptyResp, err).Once()
	handler.handle(ou, req)
	mockService.AssertExpectations(t)
	assert.Equal(t, len(ou.pushCh), 1)
	assert.Equal(t, false, (<-ou.pushCh).(*genericResp).Success)
}

func TestWorldShoutPushHandler(t *testing.T) {
	oum := newOnlineUsersManager()
	baseCmdHandler := &baseCmdHandler{oum, nil}
	mockService := new(mockWorldShoutService)
	handler := &worldShoutHandler{baseCmdHandler, mockService}

	mockClient := new(mockUserClient)
	ou1 := &onlineUser{
		uid:        "peon1",
		userClient: mockClient,
		pushCh:     make(chan interface{}, 10),
	}
	oum.online(ou1)

	pbMsg := &proto.SentShoutMsg{
		Req: &proto.SendShoutReq{
			FromUID: "xxxx",
			Content: "hihi",
		},
		SentTimestamp: time.Now().Unix(),
	}
	pushMsg, err := chat.BuildPushMsg(chat.PushPrivateCmd, pbMsg)
	assert.Nil(t, err)
	handler.handlePush(pushMsg)
	assert.Equal(t, len(ou1.pushCh), 1)
	resp := <-ou1.pushCh

	respStruct, ok := resp.(*worldShoutPush)
	assert.Equal(t, ok, true)
	assert.Equal(t, respStruct.FromUID, pbMsg.Req.FromUID)
	assert.Equal(t, respStruct.Content, pbMsg.Req.Content)
}
