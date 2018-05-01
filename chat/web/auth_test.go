package web

import (
	"context"
	"testing"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/peonone/parrot"
	authproto "github.com/peonone/parrot/auth/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (s *mockAuthService) Login(ctx context.Context, req *authproto.LoginReq, opts ...client.CallOption) (*authproto.LoginRes, error) {
	params := parrot.MakeMockServiceParams(ctx, req, opts...)
	returnValues := s.Called(params...)
	return returnValues.Get(0).(*authproto.LoginRes), returnValues.Error(1)
}

func (s *mockAuthService) Check(ctx context.Context, req *authproto.CheckAuthReq, opts ...client.CallOption) (*authproto.CheckAuthRes, error) {
	params := parrot.MakeMockServiceParams(ctx, req, opts...)
	returnValues := s.Called(params...)
	return returnValues.Get(0).(*authproto.CheckAuthRes), returnValues.Error(1)
}

func TestAuth(t *testing.T) {
	mockClient := new(mockUserClient)
	mockService := new(mockAuthService)
	handler := &authHandler{mockService}
	ou := &onlineUser{
		userClient:      mockClient,
		pushCh:          make(chan interface{}, 10),
		lastRequestTime: time.Now(),
	}
	req := &authproto.CheckAuthReq{
		Token: "token",
	}
	mockClient.On("ReadJSON", mock.Anything).Return(req, nil).Times(2)
	errRes := &authproto.CheckAuthRes{
		Success: false,
		ErrMsg:  "err1",
	}
	mockService.On("Check", context.Background(), req).Return(errRes, nil).Once()

	correctRes := &authproto.CheckAuthRes{
		Success: true,
		Uid:     "user1",
	}
	mockService.On("Check", context.Background(), req).Return(correctRes, nil).Once()
	assert.Nil(t, handler.doAuth(ou))
	assert.Equal(t, correctRes.Uid, ou.uid)
	mockClient.AssertExpectations(t)
	mockService.AssertExpectations(t)
}
