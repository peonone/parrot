package web

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/micro/go-web"
	authproto "github.com/peonone/parrot/auth/proto"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServe(t *testing.T) {
	maxIdle = time.Second * 5
	mockClient := new(mockUserClient)
	ou := &onlineUser{
		userClient:      mockClient,
		pushCh:          make(chan interface{}, 10),
		lastRequestTime: time.Now(),
	}
	mockAuthService := new(mockAuthService)
	mockStateService := new(mockStateService)
	authHandler := &authHandler{mockAuthService}
	oum := newOnlineUsersManager()

	authReq := &authproto.CheckAuthReq{
		Token: "token1",
	}
	mockClient.On("ReadJSON", mock.Anything).Return(authReq, nil).Once()
	mockClient.On("WriteJSON", mock.Anything).Return(nil).Once()
	authRes := &authproto.CheckAuthRes{Success: true, Uid: "peon1"}
	mockAuthService.On("Check", mock.Anything, mock.Anything).Return(authRes, nil).Once()
	onlineReq := &proto.UserOnlineReq{
		Uid:     authRes.Uid,
		WebNode: web.DefaultId,
	}
	onlineRes := &proto.UserOnlineRes{Success: true}
	mockStateService.On("Online", mock.Anything, onlineReq).Return(onlineRes, nil).Once()
	var readJSONReq map[string]interface{}
	mockClient.On("ReadJSON", &readJSONReq).WaitUntil(time.After(time.Second*2)).Return(&readJSONReq, io.EOF).Once()

	go ou.serve(authHandler, nil, mockStateService, oum)
	time.Sleep(time.Second * 1)
	mockAuthService.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockStateService.AssertExpectations(t)

	offlineReq := &proto.UserOfflineReq{Uid: authRes.Uid}
	offlineRes := &proto.UserOfflineRes{Success: true}
	mockStateService.On("Offline", mock.Anything, offlineReq).Return(offlineRes, nil).Once()

	time.Sleep(time.Second * 2)
	mockClient.AssertExpectations(t)
	mockStateService.AssertExpectations(t)
}

func TestWsCmdProcess(t *testing.T) {
	maxIdle = time.Second * 5
	handler1 := new(mockCmdHandler)
	handler2 := new(mockCmdHandler)

	handlers := []commandHandler{handler1, handler2}

	mockClient := new(mockUserClient)
	ou := &onlineUser{
		userClient:      mockClient,
		pushCh:          make(chan interface{}, 10),
		lastRequestTime: time.Now(),
	}

	cmd := "cmd2"

	// test for passing validate and be processed
	req := map[string]interface{}{
		"command": cmd,
	}
	handler1.On("canHandle", cmd).Return(false).Once()
	handler2.On("canHandle", cmd).Return(true).Once()
	handler2.On("validate", mock.Anything).Return(nil).Once()
	handler2.On("handle", ou, req).Return().Once()
	mockClient.On("ReadJSON", mock.Anything).Return(&req, nil).Once()
	err := processWsRequest(ou, handlers)
	assert.Nil(t, err)
	assert.Equal(t, len(ou.pushCh), 0)
	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	// test of validate error
	handler1.On("canHandle", cmd).Return(false).Once()
	handler2.On("canHandle", cmd).Return(true).Once()
	mockClient.On("ReadJSON", mock.Anything).Return(&req, nil).Once()
	expectErr := errors.New("input error")
	handler2.On("validate", mock.Anything).Return(expectErr).Once()
	err = processWsRequest(ou, handlers)
	assert.Nil(t, err)
	assert.Equal(t, len(ou.pushCh), 1)
	res, ok := (<-ou.pushCh).(*genericResp)
	assert.Equal(t, ok, true)
	assert.Equal(t, res.Success, false)
	assert.Equal(t, res.ErrMsg, expectErr.Error())
	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	// test of invalid cmd
	handler1.On("canHandle", cmd).Return(false).Once()
	handler2.On("canHandle", cmd).Return(false).Once()
	mockClient.On("ReadJSON", mock.Anything).Return(&req, nil).Once()
	err = processWsRequest(ou, handlers)
	assert.Nil(t, err)
	assert.Equal(t, len(ou.pushCh), 1)
	res, ok = (<-ou.pushCh).(*genericResp)
	assert.Equal(t, ok, true)
	assert.Equal(t, res, invalidCmdResp)
}

func TestCheckIdle(t *testing.T) {
	mockClient := new(mockUserClient)
	now := time.Now()
	ou := &onlineUser{
		userClient:      mockClient,
		pushCh:          make(chan interface{}, 10),
		lastRequestTime: now,
	}
	mockAuthService := new(mockAuthService)
	mockStateService := new(mockStateService)
	authHandler := &authHandler{mockAuthService}
	oum := newOnlineUsersManager()

	authReq := &authproto.CheckAuthReq{
		Token: "token1",
	}
	mockClient.On("ReadJSON", mock.Anything).Return(authReq, nil).Once()
	mockClient.On("WriteJSON", mock.Anything).Return(nil).Once()
	authRes := &authproto.CheckAuthRes{Success: true, Uid: "peon1"}
	mockAuthService.On("Check", mock.Anything, mock.Anything).Return(authRes, nil).Once()
	onlineReq := &proto.UserOnlineReq{
		Uid:     authRes.Uid,
		WebNode: web.DefaultId,
	}
	onlineRes := &proto.UserOnlineRes{Success: true}
	mockStateService.On("Online", mock.Anything, onlineReq).Return(onlineRes, nil).Once()
	var readJSONReq map[string]interface{}
	waitUntil := time.After(maxIdle + 2*idleCheckInterval)
	mockClient.On("ReadJSON", &readJSONReq).WaitUntil(waitUntil).Return(&readJSONReq, errors.New("conn closed")).Once()

	go ou.serve(authHandler, nil, mockStateService, oum)
	time.Sleep(maxIdle)
	mockStateService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)

	lastSeenDelta := time.Since(ou.lastRequestTime)
	assert.Equal(t, true, lastSeenDelta < maxIdle+2*time.Second)
	assert.NotEqual(t, now, ou.lastRequestTime)

	offlineReq := &proto.UserOfflineReq{Uid: authRes.Uid}
	offlineRes := &proto.UserOfflineRes{Success: true}
	mockStateService.On("Offline", mock.Anything, offlineReq).Return(offlineRes, nil).Once()
	mockClient.On("WriteJSON", idleToLongResp).Return(nil).Once()
	mockClient.On("Close").Return(nil).Once()
	time.Sleep(3 * idleCheckInterval)
	assert.Equal(t, true, ou.closed)
	mockStateService.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}
