package srv

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/peonone/parrot/auth/proto"
)

type authTestInfo struct {
	username        string
	password        string
	expectedSuccess bool
}

var testDatas = []authTestInfo{
	authTestInfo{"peonone", "peonone!", true},
	authTestInfo{"invalid", "blabla", false},
	authTestInfo{"valid", "valid!", true},
}

func TestDoAndValidateAuth(t *testing.T) {
	tokenExpDur = time.Second * 5
	validTokens := make([]string, 0, 2)
	service := authService{}
	for _, testData := range testDatas {
		loginReq := &proto.LoginReq{
			Username: testData.username,
			Password: testData.password,
		}
		loginRes := new(proto.LoginRes)
		service.Login(context.Background(), loginReq, loginRes)
		if testData.expectedSuccess != loginRes.Success {
			t.Errorf("DoAuth: expect get success=%v but got %v for %s",
				testData.expectedSuccess, loginRes.Success, testData.username)
			return
		}
		if loginRes.Success {
			if len(loginRes.Token) == 0 {
				t.Errorf("DoAuth: got empty token for a succeeded auth request from %s",
					testData.username)
				return
			}
			validTokens = append(validTokens, loginRes.Token)
		}
		checkReq := &proto.CheckAuthReq{
			Token: loginRes.Token,
		}
		checkRes := new(proto.CheckAuthRes)

		err := service.Check(context.Background(), checkReq, checkRes)
		if testData.expectedSuccess {
			assert.Nil(t, err)
			assert.Equal(t, loginReq.Username, checkRes.Uid)
		} else {
			assert.Equal(t, "", checkRes.Uid)
		}
		assert.Equal(t, testData.expectedSuccess, checkRes.Success)
		if testData.expectedSuccess {
			checkReq.Token += "abc"
			err = service.Check(context.Background(), checkReq, checkRes)
			assert.Nil(t, err)
			assert.False(t, checkRes.Success)

			time.Sleep(tokenExpDur)
			checkReq.Token = loginRes.Token
			err = service.Check(context.Background(), checkReq, checkRes)
			assert.Nil(t, err)
			assert.False(t, checkRes.Success)
		}
	}
	if len(validTokens) >= 2 {
		assert.NotEqual(t, validTokens[0], validTokens[1])
	}
}
