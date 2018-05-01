package web

import (
	"context"
	"errors"
	"log"

	"github.com/micro/go-micro/client"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
)

const sendPMCmd = "private.send"

type privateMessageHandler struct {
	*baseCmdHandler
	cli proto.PrivateService
}

type privateMessagePush struct {
	Command       string `json:"command"`
	FromUID       string `json:"fromUID"`
	Content       string `json:"content"`
	SentTimestamp int64  `json:"sentTimestamp"`
}

func newPrivateMessageHandler(rpcCli client.Client, bch *baseCmdHandler) commandHandler {
	return &privateMessageHandler{
		baseCmdHandler: bch,
		cli:            proto.PrivateServiceClient(chat.SrvServiceName, rpcCli),
	}
}

func (h *privateMessageHandler) canHandle(cmd string) bool {
	return cmd == sendPMCmd
}

func (h *privateMessageHandler) validate(req map[string]interface{}) error {
	toUID, ok := req["toUID"]
	if !ok {
		return errors.New("mush specify a toUID(str)")
	}
	_, ok = toUID.(string)
	if !ok {
		return errors.New("mush specify a toUID(str)")
	}

	content, ok := req["content"]
	if !ok {
		return errors.New("mush specify a content(str)")
	}
	_, ok = content.(string)
	if !ok {
		return errors.New("mush specify a content(str)")
	}
	return nil
}

func (h *privateMessageHandler) handle(c *onlineUser, req map[string]interface{}) {
	rpcReq := &proto.SendPMReq{
		FromUID: c.uid,
		ToUID:   req["toUID"].(string),
		Content: req["content"].(string),
	}
	rpcRes, err := h.cli.Send(context.Background(), rpcReq)
	errCnt := 0
	if err != nil {
		log.Printf("error occurred during private message send service call: %s", err)
		c.pushCh <- &genericResp{
			Success: false,
			ErrMsg:  "Internal error",
		}
		errCnt = 1
	} else {
		c.pushCh <- rpcRes
	}
	if h.stat != nil {
		h.stat.addRequestReceived(1, errCnt)
	}
}

func (h *privateMessageHandler) canHandlePush(cmd string) bool {
	return cmd == chat.PushPrivateCmd
}

func (h *privateMessageHandler) handlePush(pushMsg *proto.PushMsg) {
	pushPrivMsg := &proto.SentPrivateMsg{}
	err := chat.DecodeFromPushMsg(pushMsg, pushPrivMsg)
	if err != nil {
		log.Printf("failed to decode SentPrivateMsg: %s", err)
		return
	}
	resp := &privateMessagePush{
		Command:       chat.PushPrivateCmd,
		FromUID:       pushPrivMsg.Req.FromUID,
		Content:       pushPrivMsg.Req.Content,
		SentTimestamp: pushPrivMsg.SentTimestamp,
	}

	c := h.oum.getOnlineClient(pushPrivMsg.Req.ToUID)
	if c == nil {
		log.Printf("user[%s] is not online", pushPrivMsg.Req.ToUID)
	} else {
		c.pushCh <- resp
		if h.stat != nil {
			h.stat.addMessagePushed(1)
		}
	}
}
