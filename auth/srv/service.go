package srv

import (
	"context"
	"math/rand"
	"time"

	"github.com/peonone/parrot"

	micro "github.com/micro/go-micro"
	proto "github.com/peonone/parrot/auth/proto"
)

const Name = "go.micro.srv.auth"

type AuthService struct {
	tokenStore
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().Unix())
}

//Init initialize the auth service
func Init(service micro.Service) {
	rdsClient := parrot.MakeRedisClient()
	proto.RegisterAuthHandler(service.Server(), &AuthService{
		&redisTokenStore{rdsClient},
	})
}

func (a *AuthService) generateToken() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//Login is the login service handler
func (a *AuthService) Login(ctx context.Context, req *proto.LoginReq, res *proto.LoginRes) error {
	// TODO implement a real auth feature
	if req.Password != req.Username+"!" {
		res.Success = false
		res.ErrMsg = "The username and password are incorrect"
		return nil
	}
	res.Success = true
	res.Uid = req.Username
	res.Token = a.generateToken()
	return a.saveToken(res.Uid, res.Token)
}

//Check is the auth check service handler
func (a *AuthService) Check(ctx context.Context, req *proto.CheckAuthReq, res *proto.CheckAuthRes) error {
	token, err := a.getToken(req.Uid)
	if err != nil {
		res.Success = false
		res.ErrMsg = "Invalid token"
		return nil
	}
	res.Success = token == req.Token
	if !res.Success {
		res.ErrMsg = "Invalid token"
	}
	return nil
}
