package srv

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
)

//PrivateHandler is the private message service handler
type privateHandler struct {
	*baseHandler
}

//Send handles the private message send service call
func (h *privateHandler) Send(ctx context.Context, req *proto.SendPMReq, res *proto.SendPMRes) error {
	onlineNode, err := h.stateStore.getOnlineWebNode(req.ToUID)
	defer func() {
		if err != nil {
			log.Printf("failed to send private message %s", err)
		}
	}()

	if err == redis.Nil {
		res.Success = false
		res.ErrMsg = fmt.Sprintf("user %s is not online", req.ToUID)
		return nil
	} else if err != nil {
		return err
	}
	key := fmt.Sprintf("%s.private.push", onlineNode)
	pbMsg := &proto.SentPrivateMsg{
		Req:           req,
		SentTimestamp: time.Now().Unix(),
	}
	body, err := chat.EncodePushMsg(chat.PushPrivateCmd, pbMsg)
	if err != nil {
		return err
	}
	err = h.mqSender.sendMQMsg(key, body)
	if err != nil {
		return err
	}
	log.Printf("published to RMQ, exchange=%s, routingkey=%s", chat.PushMsgExchangeName, key)
	res.Success = true
	log.Printf("%s sending a private msg to %s(%s), content: %s", req.FromUID, req.ToUID, onlineNode, req.Content)
	return nil
}
