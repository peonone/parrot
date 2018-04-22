package chat

import (
	"testing"
	"time"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodePhshMsg(t *testing.T) {
	req := &proto.SendPMReq{
		FromUID: "peon1",
		ToUID:   "peon2",
		Content: "hihi",
	}
	pbMsg := &proto.SentPrivateMsg{
		Req:           req,
		SentTimestamp: time.Now().Unix(),
	}
	body, err := EncodePushMsg(PushPrivateCmd, pbMsg)
	assert.Nil(t, err)

	pushMsg := new(proto.PushMsg)
	err = protobuf.Unmarshal(body, pushMsg)
	assert.Nil(t, err)

	restoredPbMsg := new(proto.SentPrivateMsg)
	err = DecodeFromPushMsg(pushMsg, restoredPbMsg)
	assert.Nil(t, err)

	assert.Equal(t, pbMsg, restoredPbMsg)
}
