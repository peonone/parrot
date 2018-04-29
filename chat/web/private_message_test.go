package web

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/peonone/parrot"

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
	params := parrot.MakeMockServiceParams(ctx, in, opts...)
	returnValues := s.Called(params...)
	return returnValues.Get(0).(*proto.SendPMRes), returnValues.Error(1)
}

func TestPrivateMsgHandler(t *testing.T) {
	oum := newOnlineUsersManager()

	stat := newStat(os.Stderr)
	baseCmdHandler := &baseCmdHandler{oum, stat}
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
	assert.Equal(t, stat.totalInfo.requestReceived, 1)
	assert.Equal(t, stat.totalInfo.requestWithErr, 0)
	assert.Equal(t, stat.minuteInfo.requestReceived, 1)
	errRes := &proto.SendPMRes{Success: false, ErrMsg: "err1"}
	mockService.On("Send", context.Background(), expectedReq).Return(errRes, nil).Once()
	handler.handle(ou, req)
	mockService.AssertExpectations(t)
	assert.Equal(t, len(ou.pushCh), 1)
	assert.Equal(t, errRes, <-ou.pushCh)
	assert.Equal(t, stat.totalInfo.requestReceived, 2)
	assert.Equal(t, stat.totalInfo.requestWithErr, 0)
	assert.Equal(t, stat.minuteInfo.requestReceived, 2)

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
	assert.Equal(t, stat.totalInfo.requestReceived, 3)
	assert.Equal(t, stat.totalInfo.requestWithErr, 1)
	assert.Equal(t, stat.minuteInfo.requestReceived, 3)
}

func TestPrivateMsgPushHandler(t *testing.T) {
	oum := newOnlineUsersManager()
	stat := newStat(os.Stderr)
	baseCmdHandler := &baseCmdHandler{oum, stat}
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
	assert.Equal(t, stat.totalInfo.msgPushed, 1)
	assert.Equal(t, stat.minuteInfo.msgPushed, 1)

	pbMsg.Req.ToUID += "xxx"
	pushMsg, err = chat.BuildPushMsg(chat.PushPrivateCmd, pbMsg)
	assert.Nil(t, err)
	handler.handlePush(pushMsg)
	assert.Equal(t, len(ou.pushCh), 0)
	assert.Equal(t, stat.totalInfo.msgPushed, 1)
	assert.Equal(t, stat.minuteInfo.msgPushed, 1)
}
