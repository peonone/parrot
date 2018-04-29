package web

import (
	"context"
	"errors"
	"log"

	"github.com/micro/go-micro/client"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
)

const worldShoutCmd = "shout.send"

type worldShoutHandler struct {
	*baseCmdHandler
	cli proto.ShoutService
}

type worldShoutPush struct {
	Command       string `json:"command"`
	FromUID       string `json:"fromUID"`
	Content       string `json:"content"`
	SentTimestamp int64  `json:"sentTimestamp"`
}

func newWorldShoutHandler(rpcCli client.Client, bch *baseCmdHandler) commandHandler {
	return &worldShoutHandler{
		baseCmdHandler: bch,
		cli:            proto.ShoutServiceClient(chat.SrvServiceName, rpcCli),
	}
}

func (h *worldShoutHandler) canHandle(cmd string) bool {
	return cmd == worldShoutCmd
}

func (h *worldShoutHandler) validate(req map[string]interface{}) error {
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

func (h *worldShoutHandler) handle(c *onlineUser, req map[string]interface{}) {
	rpcReq := &proto.SendShoutReq{
		FromUID: c.uid,
		Content: req["content"].(string),
	}
	rpcRes, err := h.cli.Send(context.Background(), rpcReq)
	errCnt := 0
	if err != nil {
		c.pushCh <- &genericResp{
			Success: false,
			ErrMsg:  err.Error(),
		}
		errCnt = 1
	} else {
		c.pushCh <- rpcRes
	}
	if h.stat != nil {
		h.stat.addRequestReceived(1, errCnt)
	}
}

func (h *worldShoutHandler) canHandlePush(cmd string) bool {
	return cmd == chat.PushShoutCmd
}

func (h *worldShoutHandler) handlePush(pushMsg *proto.PushMsg) {
	pushShoutMsg := &proto.SentShoutMsg{}
	err := chat.DecodeFromPushMsg(pushMsg, pushShoutMsg)
	if err != nil {
		log.Printf("failed to decode SentPrivateMsg: %s", err)
		return
	}
	resp := &worldShoutPush{
		Command:       chat.PushShoutCmd,
		FromUID:       pushShoutMsg.Req.FromUID,
		Content:       pushShoutMsg.Req.Content,
		SentTimestamp: pushShoutMsg.SentTimestamp,
	}
	pushedCnt := 0
	for _, u := range h.oum.getAllOnlineUsers() {
		u.mu.Lock()
		if !u.closed {
			pushedCnt++
			u.pushCh <- resp
		}
		u.mu.Unlock()
	}
	if h.stat != nil {
		h.stat.addMessagePushed(pushedCnt)
	}
}
