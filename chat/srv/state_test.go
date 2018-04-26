package srv

import (
	"context"
	"testing"

	"github.com/peonone/parrot/chat/proto"
)

func TestOnline(t *testing.T) {
	ssMock := new(mockStateStore)
	mqMock := new(mockMQSender)
	baseHandler := &baseHandler{ssMock, mqMock}
	h := &stateHandler{baseHandler}

	uid := "peon1"
	node := "web-node1"
	ssMock.On("online", uid, node).Return(nil).Once()
	onReq := &proto.UserOnlineReq{Uid: uid, WebNode: node}
	onRes := &proto.UserOnlineRes{}
	h.Online(context.Background(), onReq, onRes)
	ssMock.AssertExpectations(t)

	ssMock.On("offline", uid).Return(nil).Once()
	offReq := &proto.UserOfflineReq{Uid: uid}
	offRes := &proto.UserOfflineRes{}
	h.Offline(context.Background(), offReq, offRes)
	ssMock.AssertExpectations(t)
}
