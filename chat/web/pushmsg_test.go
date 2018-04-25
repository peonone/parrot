package web

import (
	"testing"

	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/assert"

	"github.com/streadway/amqp"
)

func TestPushMsg(t *testing.T) {
	handler1 := new(mockCmdHandler)
	handler2 := new(mockCmdHandler)

	handlers := []commandHandler{handler1, handler2}

	deliverCh := make(chan amqp.Delivery, 10)

	req := &proto.SendPMReq{
		FromUID: "peon1",
		ToUID:   "peon2",
		Content: "333",
	}
	cmd := "cmd1"
	body, err := chat.EncodePushMsg(cmd, req)
	assert.Nil(t, err)
	deliverCh <- amqp.Delivery{Body: body}

	handler1.On("canHandlePush", cmd).Return(false).Once()
	handler2.On("canHandlePush", cmd).Return(true).Once()
	expectedPushMsg, err := chat.BuildPushMsg(cmd, req)
	assert.Nil(t, err)
	handler2.On("handlePush", expectedPushMsg).Once()
	close(deliverCh)
	handlePushMsg(deliverCh, handlers)

	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
}
