package srv

import (
	"context"
	"testing"

	"github.com/peonone/parrot/chat/proto"
)

func TestOnline(t *testing.T) {
	ssMock := new(mockStateStore)
	mqMock := new(mockMQSender)
	basicHandler := &BasicHandler{ssMock, mqMock}
	h := &stateHandler{basicHandler}

	uid := "peon1"
	node := "web-node1"
	ssMock.On("online", uid, node).Return(nil).Once()
	req := &proto.UserOnlineReq{Uid: uid, WebNode: node}
	res := &proto.UserOnlineRes{}
	h.Online(context.Background(), req, res)
	ssMock.AssertExpectations(t)
}
