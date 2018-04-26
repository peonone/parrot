package srv

import (
	"context"

	"github.com/peonone/parrot/chat/proto"
)

type stateHandler struct {
	*baseHandler
}

func (h *stateHandler) Online(ctx context.Context, req *proto.UserOnlineReq, res *proto.UserOnlineRes) error {
	err := h.stateStore.online(req.Uid, req.WebNode)
	if err != nil {
		return err
	}
	res.Success = true
	return nil
}

func (h *stateHandler) Offline(ctx context.Context, req *proto.UserOfflineReq, res *proto.UserOfflineRes) error {
	err := h.stateStore.offline(req.Uid)
	if err != nil {
		return err
	}
	res.Success = true
	return nil
}
