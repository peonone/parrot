package web

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPrivateMessageService struct {
	mock.Mock
}

func (s *mockPrivateMessageService) Send(ctx context.Context, in *proto.SendPMReq, opts ...client.CallOption) (*proto.SendPMRes, error) {
	params := make([]interface{}, 0, 2+len(opts))
	params = append(params, ctx, in)
	for _, opt := range opts {
		params = append(params, opt)
	}
	returnValues := s.Called(params...)
	return returnValues.Get(0).(*proto.SendPMRes), returnValues.Error(1)
}

func TestPrivateMsgHandler(t *testing.T) {
	oum := newOnlineUsersManager()
	baseCmdHandler := &baseCmdHandler{oum}
	mockService := new(mockPrivateMessageService)
	handler := &privateMessageHandler{baseCmdHandler, mockService}

	mockClient := new(mockUserClient)
	ou := &onlineUser{
		uid:             "peon1",
		userClient:      mockClient,
		pushCh:          make(chan interface{}, 10),
		lastRequestTime: time.Now(),
	}

	req := map[string]interface{}{
		"toUID":   "333",
		"content": "hihi",
	}

	expectedReq := &proto.SendPMReq{
		FromUID: ou.uid,
		ToUID:   req["toUID"].(string),
		Content: req["content"].(string),
	}
	successRes := &proto.SendPMRes{
		Success: true,
	}
	mockService.On("Send", context.Background(), expectedReq).Return(successRes, nil).Once()
	handler.handle(ou, req)
	mockService.AssertExpectations(t)
	assert.Equal(t, len(ou.pushCh), 1)
	assert.Equal(t, successRes, <-ou.pushCh)

	errRes := &proto.SendPMRes{Success: false, ErrMsg: "err1"}
	mockService.On("Send", context.Background(), expectedReq).Return(errRes, nil).Once()
	handler.handle(ou, req)
	mockService.AssertExpectations(t)
	assert.Equal(t, len(ou.pushCh), 1)
	assert.Equal(t, errRes, <-ou.pushCh)

	err := errors.New("internal err")
	var emptyResp *proto.SendPMRes
	mockService.On("Send", context.Background(), expectedReq).Return(emptyResp, err).Once()
	handler.handle(ou, req)
	mockService.AssertExpectations(t)
	assert.Equal(t, len(ou.pushCh), 1)
	expectedRes := &genericResp{
		Success: false,
		ErrMsg:  err.Error(),
	}
	assert.Equal(t, expectedRes, <-ou.pushCh)
}

func TestPrivateMsgPushHandler(t *testing.T) {
	oum := newOnlineUsersManager()
	baseCmdHandler := &baseCmdHandler{oum}
	mockService := new(mockPrivateMessageService)
	handler := &privateMessageHandler{baseCmdHandler, mockService}

	mockClient := new(mockUserClient)
	ou := &onlineUser{
		uid:        "peon1",
		userClient: mockClient,
		pushCh:     make(chan interface{}, 10),
	}
	oum.online(ou)

	pbMsg := &proto.SentPrivateMsg{
		Req: &proto.SendPMReq{
			FromUID: "peon2",
			ToUID:   ou.uid,
			Content: "hihi",
		},
		SentTimestamp: time.Now().Unix(),
	}
	pushMsg, err := chat.BuildPushMsg(chat.PushPrivateCmd, pbMsg)
	assert.Nil(t, err)
	handler.handlePush(pushMsg)
	assert.Equal(t, len(ou.pushCh), 1)
	resp := <-ou.pushCh

	respStruct, ok := resp.(*privateMessagePush)
	assert.Equal(t, ok, true)
	assert.Equal(t, respStruct.FromUID, pbMsg.Req.FromUID)
	assert.Equal(t, respStruct.Content, pbMsg.Req.Content)

	pbMsg.Req.ToUID += "xxx"
	pushMsg, err = chat.BuildPushMsg(chat.PushPrivateCmd, pbMsg)
	assert.Nil(t, err)
	handler.handlePush(pushMsg)
	assert.Equal(t, len(ou.pushCh), 0)
}
