package srv

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/micro/go-micro/registry"

	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
)

type worldShoutHandler struct {
	*baseHandler
}

//Send handles the world shoult message send service call
func (h *worldShoutHandler) Send(ctx context.Context, req *proto.SendShoutReq, res *proto.SendShoutRes) error {
	webServices, err := registry.DefaultRegistry.GetService(chat.WebServiceName)
	if err != nil {
		return err
	}
	if len(webServices) != 1 {
		return fmt.Errorf("got %d websocket nodes(expected 1)", len(webServices))
	}
	pbMsg := &proto.SentShoutMsg{
		Req:           req,
		SentTimestamp: time.Now().Unix(),
	}
	body, err := chat.EncodePushMsg(chat.PushShoutCmd, pbMsg)
	if err != nil {
		return err
	}
	for _, node := range webServices[0].Nodes {
		key := fmt.Sprintf("%s.private.push", node.Id)
		err = h.mqSender.sendMQMsg(key, body)
		if err != nil {
			log.Printf("failed to send shout msg to RMQ: %s", err)
		} else {
			log.Printf("published to RMQ, exchange=%s, routingkey=%s", chat.PushMsgExchangeName, key)
		}
	}

	res.Success = true
	log.Printf("%s sending a shoult msg to world, content: %s", req.FromUID, req.Content)
	return nil
}
