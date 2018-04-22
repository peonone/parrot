package web

import (
	"context"

	"github.com/micro/go-micro/client"
	authproto "github.com/peonone/parrot/auth/proto"
	authsrv "github.com/peonone/parrot/auth/srv"
)

type authHandler struct {
	cli authproto.AuthService
}

func newAuthHandler(rpcClient client.Client) *authHandler {
	return &authHandler{
		cli: authproto.AuthServiceClient(authsrv.Name, rpcClient),
	}
}

func (h *authHandler) doAuth(c *onlineUser) error {
	req := new(authproto.CheckAuthReq)
	for {
		err := c.userClient.ReadJSON(req)
		if err != nil {
			return err
		}

		res, err := h.cli.Check(context.Background(), req)
		if err != nil {
			return err
		}
		c.pushCh <- res
		if res.Success {
			c.uid = req.Uid
			break
		}
	}
	return nil
}
