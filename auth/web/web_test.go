package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/micro/go-micro/client"
	"github.com/peonone/parrot"
	"github.com/peonone/parrot/auth/proto"
	"github.com/stretchr/testify/mock"
)

type mockSrvClient struct {
	mock.Mock
}

func (m *mockSrvClient) Login(ctx context.Context, in *proto.LoginReq, opts ...client.CallOption) (*proto.LoginRes, error) {
	params := parrot.MakeMockServiceParams(ctx, in, opts...)
	returnVals := m.Called(params...)
	return returnVals.Get(0).(*proto.LoginRes), returnVals.Error(1)
}

func (m *mockSrvClient) Check(ctx context.Context, in *proto.CheckAuthReq, opts ...client.CallOption) (*proto.CheckAuthRes, error) {
	params := parrot.MakeMockServiceParams(ctx, in, opts...)
	returnVals := m.Called(params...)
	return returnVals.Get(0).(*proto.CheckAuthRes), returnVals.Error(1)
}

type testDataInfo struct {
	uname    string
	password string
	success  bool
	errMsg   string
	uid      string
	token    string
}

func TestDoLogin(t *testing.T) {
	mockCli := new(mockSrvClient)
	webService := &authHandler{mockCli}

	testInfos := []testDataInfo{
		{
			uname:    "peon",
			password: "peon!",
			success:  true,
			uid:      "peon",
			token:    "peon1",
		},
		{
			uname:    "blabla",
			password: "peon123",
			success:  false,
			errMsg:   "username and password dose not match",
		},
	}
	for _, testInfo := range testInfos {
		form := url.Values{"username": {testInfo.uname}, "password": {testInfo.password}}
		reader := strings.NewReader(form.Encode())
		req := httptest.NewRequest("POST", "http://localhost/login", reader)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		recorder := httptest.NewRecorder()
		res := &proto.LoginRes{
			Success: testInfo.success,
			ErrMsg:  testInfo.errMsg,
			Uid:     testInfo.uid,
			Token:   testInfo.token,
		}
		mockCli.On("Login", mock.Anything, mock.Anything).Return(res, nil).Once()
		webService.loginHandler(recorder, req)

		webResp := new(loginResp)
		json.Unmarshal(recorder.Body.Bytes(), webResp)
		if webResp.Success != testInfo.success {
			t.Errorf("expected to get success=%v but got %v for %s", testInfo.success, webResp.Success, testInfo.uname)
		}
		mockCli.AssertExpectations(t)
	}

}
