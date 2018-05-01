package srv

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	micro "github.com/micro/go-micro"
	proto "github.com/peonone/parrot/auth/proto"
)

const Name = "go.micro.srv.auth"

// TODO use immutable configuration instead of variable
var tokenExpDur = time.Minute * 5

var tokenSignMethod = jwt.SigningMethodHS256

type authService struct {
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var authHmacKey []byte

func genHmacKey() []byte {
	key := make([]byte, 0, 64)
	for i := 0; i < 64; i++ {
		key = append(key, byte(rand.Int31n(63)))
	}
	return key
}

func init() {
	rand.Seed(time.Now().Unix())
	var err error
	authHmacKey, err = ioutil.ReadFile("conf/auth_hmac_key")
	if err != nil {
		log.Printf("can't load HMAC key for auth: %s, generating a random one", err)
		authHmacKey = genHmacKey()
	}
}

//Init initialize the auth service
func Init(service micro.Service) {
	proto.RegisterAuthHandler(service.Server(), new(authService))
}

func (a *authService) generateToken(uid string) (string, error) {
	token := jwt.NewWithClaims(tokenSignMethod, jwt.MapClaims{
		"sub": uid,
		"exp": float64(time.Now().Add(tokenExpDur).Unix()),
	})
	return token.SignedString(authHmacKey)
}

//Login is the login service handler
func (a *authService) Login(ctx context.Context, req *proto.LoginReq, res *proto.LoginRes) error {
	// TODO implement a real auth feature
	if req.Password != req.Username+"!" {
		res.Success = false
		res.ErrMsg = "The username and password are incorrect"
		return nil
	}
	res.Success = true
	token, err := a.generateToken(req.Username)
	if err != nil {
		return err
	}
	res.Token = token
	return nil
}

//Check is the auth check service handler
func (a *authService) Check(ctx context.Context, req *proto.CheckAuthReq, res *proto.CheckAuthRes) error {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method != tokenSignMethod {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return authHmacKey, nil
	})

	if err != nil {
		log.Printf("failed to parse auth token: %s", err)
		res.Success = false
		res.ErrMsg = "Invalid token"
		return nil
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if int64(claims["exp"].(float64)) > time.Now().Unix() {
			res.Success = true
			res.Uid = claims["sub"].(string)
		} else {
			res.Success = false
			res.ErrMsg = "Token expired"
		}
	} else {
		res.Success = false
		res.ErrMsg = "Invalid token"
	}
	return nil
}
