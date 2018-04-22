package srv

import (
	"context"
	"errors"
	"testing"

	"github.com/peonone/parrot/auth/proto"
	"github.com/stretchr/testify/mock"
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

type mockTokenStore struct {
	mock.Mock
}

func (ts *mockTokenStore) saveToken(uid string, token string) error {
	returnVals := ts.Called(uid, token)
	return returnVals.Error(0)
}

func (ts *mockTokenStore) getToken(uid string) (string, error) {
	returnVals := ts.Called(uid)
	return returnVals.String(0), returnVals.Error(1)
}
func TestDoAndValidateAuth(t *testing.T) {
	validTokens := make([]string, 0, 2)
	mockTs := &mockTokenStore{}
	a := AuthService{mockTs}
	for _, testData := range testDatas {
		req := &proto.LoginReq{
			Username: testData.username,
			Password: testData.password,
		}
		res := new(proto.LoginRes)
		if testData.expectedSuccess {
			mockTs.On("saveToken", mock.Anything, mock.Anything).Return(nil).Once()
		}
		a.Login(context.Background(), req, res)
		if testData.expectedSuccess != res.Success {
			t.Errorf("DoAuth: expect get success=%v but got %v for %s",
				testData.expectedSuccess, res.Success, testData.username)
			return
		}
		mockTs.AssertExpectations(t)
		if res.Success {
			if len(res.Token) == 0 {
				t.Errorf("DoAuth: got empty token for a succeeded auth request from %s",
					testData.username)
				return
			}
			validTokens = append(validTokens, res.Token)
		}
		validateReq := &proto.CheckAuthReq{
			Uid:   res.Uid,
			Token: res.Token,
		}
		validateRes := new(proto.CheckAuthRes)

		var returnVals []interface{}
		if testData.expectedSuccess {
			returnVals = []interface{}{res.Token, nil}
		} else {
			returnVals = []interface{}{"", errors.New("not found")}
		}
		mockTs.On("getToken", res.Uid).Return(returnVals...).Once()
		a.Check(context.Background(), validateReq, validateRes)
		mockTs.AssertExpectations(t)
		if validateRes.Success != testData.expectedSuccess {
			t.Errorf("ValidateAuth: expect get success=%v but got %v for %s",
				testData.expectedSuccess, validateRes.Success, testData.username)
			return
		}
		if testData.expectedSuccess {
			mockTs.On("getToken", res.Uid).Return("", errors.New("not found")).Once()
			validateReq.Token += "abc"
			a.Check(context.Background(), validateReq, validateRes)
			mockTs.AssertExpectations(t)
			if validateRes.Success {
				t.Errorf("ValidateAuth: got success=true for an invalid(by add a postfix) token for %s",
					testData.username)
				return
			}
		}

	}
	if len(validTokens) >= 2 {
		if validTokens[0] == validTokens[1] {
			t.Errorf("DoAuth: got same token for two auth requests")
			return
		}
	}
}
