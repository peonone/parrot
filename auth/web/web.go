package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-web"
	"github.com/peonone/parrot/auth/proto"
)

// Name is the service name of auth web
const Name = "go.micro.web.auth"

type authHandler struct {
	Client proto.AuthService
}

type loginResp struct {
	Success bool   `json:"success"`
	ErrMsg  string `json:"errMsg"`
	Token   string `json:"token"`
}

// Init initializes auth web resources and registers all handlers
func Init(service web.Service) {
	auth := &authHandler{Client: proto.AuthServiceClient("go.micro.srv.auth", client.DefaultClient)}

	service.HandleFunc("/login", auth.loginHandler)
}

func (s *authHandler) doLogin(req *http.Request) *loginResp {
	req.ParseForm()
	name, ok := req.PostForm["username"]
	if !ok || len(name) == 0 {
		return &loginResp{
			Success: false,
			ErrMsg:  "username cannot be blank",
		}
	}

	password, ok := req.PostForm["password"]
	if !ok || len(password) == 0 {
		return &loginResp{
			Success: false,
			ErrMsg:  "password cannot be blank",
		}
	}

	response, err := s.Client.Login(context.Background(), &proto.LoginReq{
		Username: name[0],
		Password: password[0],
	})
	if err != nil {
		log.Printf("error occurred during login service call: %s", err)
		return &loginResp{
			Success: false,
			ErrMsg:  "Internal error",
		}
	}
	return &loginResp{
		Success: response.Success,
		ErrMsg:  response.ErrMsg,
		Token:   response.Token,
	}
}

func (s *authHandler) loginHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	res := s.doLogin(req)
	b, _ := json.Marshal(res)
	w.Write(b)
}
