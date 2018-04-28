package chat

import (
	protobuf "github.com/golang/protobuf/proto"
	"github.com/peonone/parrot/chat/proto"
)

const (
	WebServiceName = "go.micro.web.chat"
	SrvServiceName = "go.micro.srv.chat"
)

const (
	OnlineStateKey      = "online-user-nodes"
	PushMsgExchangeName = "push-msg"
	PushPrivateCmd      = "private.push"
	PushShoutCmd        = "shout.push"
)

func BuildPushMsg(cmd string, msg protobuf.Message) (*proto.PushMsg, error) {
	msgBody, err := protobuf.Marshal(msg)
	if err != nil {
		return nil, err
	}
	pushMsg := &proto.PushMsg{
		Command: cmd,
		Body:    msgBody,
	}
	return pushMsg, nil
}

func EncodePushMsg(cmd string, msg protobuf.Message) ([]byte, error) {
	msg, err := BuildPushMsg(cmd, msg)
	if err != nil {
		return nil, err
	}
	return protobuf.Marshal(msg)
}

func DecodeFromPushMsg(pushMsg *proto.PushMsg, msg protobuf.Message) error {
	return protobuf.Unmarshal(pushMsg.Body, msg)
}
