package web

import (
	"context"
	"testing"

	"github.com/micro/go-micro/client"
	authproto "github.com/peonone/parrot/auth/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (s *mockAuthService) Login(ctx context.Context, req *authproto.LoginReq, opts ...client.CallOption) (*authproto.LoginRes, error) {
	args := []interface{}{ctx, req}
	for _, opt := range opts {
		args = append(args, opt)
	}
	returnValues := s.Called(args...)
	return returnValues.Get(0).(*authproto.LoginRes), returnValues.Error(1)
}

func (s *mockAuthService) Check(ctx context.Context, req *authproto.CheckAuthReq, opts ...client.CallOption) (*authproto.CheckAuthRes, error) {
	args := []interface{}{ctx, req}
	for _, opt := range opts {
		args = append(args, opt)
	}
	returnValues := s.Called(args...)
	return returnValues.Get(0).(*authproto.CheckAuthRes), returnValues.Error(1)
}

func TestAuth(t *testing.T) {
	mockClient := new(mockUserClient)
	mockService := new(mockAuthService)
	handler := &authHandler{mockService}
	ou := &onlineUser{
		userClient: mockClient,
		pushCh:     make(chan interface{}, 10),
	}
	req := &authproto.CheckAuthReq{
		Uid:   "user1",
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
	}
	mockService.On("Check", context.Background(), req).Return(correctRes, nil).Once()
	assert.Nil(t, handler.doAuth(ou))
	mockClient.AssertExpectations(t)
	mockService.AssertExpectations(t)
}
