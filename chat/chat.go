package chat

import (
	protobuf "github.com/golang/protobuf/proto"
	"github.com/peonone/parrot/chat/proto"
)

const OnlineStateKey = "online-user-nodes"
const PushMsgExchangeName = "push-msg"
const PushPrivateCmd = "private.push"

func EncodePushMsg(cmd string, msg protobuf.Message) ([]byte, error) {
	msgBody, err := protobuf.Marshal(msg)
	if err != nil {
		return nil, err
	}
	pushMsg := &proto.PushMsg{
		Command: cmd,
		Body:    msgBody,
	}
	return protobuf.Marshal(pushMsg)
}

func DecodeFromPushMsg(pushMsg *proto.PushMsg, msg protobuf.Message) error {
	return protobuf.Unmarshal(pushMsg.Body, msg)
}
