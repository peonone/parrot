package web

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWsCmdProcess(t *testing.T) {
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
	mockClient.On("ReadJSON", mock.Anything).Return(req, nil).Once()
	err := processWsRequest(ou, handlers)
	assert.Nil(t, err)
	assert.Equal(t, len(ou.pushCh), 0)
	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	// test of validate error
	handler1.On("canHandle", cmd).Return(false).Once()
	handler2.On("canHandle", cmd).Return(true).Once()
	mockClient.On("ReadJSON", mock.Anything).Return(req, nil).Once()
	expectErr := errors.New("input error")
	handler2.On("validate", mock.Anything).Return(expectErr).Once()
	err = processWsRequest(ou, handlers)
	assert.Nil(t, err)
	assert.Equal(t, len(ou.pushCh), 1)
	res, ok := (<-ou.pushCh).(*genericResp)
	fmt.Printf("%v\n", res)
	assert.Equal(t, ok, true)
	assert.Equal(t, res.Success, false)
	assert.Equal(t, res.ErrMsg, expectErr.Error())
	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	// test of invalid cmd
	handler1.On("canHandle", cmd).Return(false).Once()
	handler2.On("canHandle", cmd).Return(false).Once()
	mockClient.On("ReadJSON", mock.Anything).Return(req, nil).Once()
	err = processWsRequest(ou, handlers)
	assert.Nil(t, err)
	assert.Equal(t, len(ou.pushCh), 1)
	res, ok = (<-ou.pushCh).(*genericResp)
	assert.Equal(t, ok, true)
	assert.Equal(t, res, invalidCmdResp)
}

func TestCheckIdle(t *testing.T) {
	mockClient := new(mockUserClient)
	ou := &onlineUser{
		userClient:      mockClient,
		pushCh:          make(chan interface{}, 10),
		lastRequestTime: time.Now(),
	}

	go ou.checkIdle()
	time.Sleep(maxIdle - time.Second*2)
	mockClient.AssertExpectations(t)

	mockClient.On("Close").Return(nil).Once()
	mockClient.On("WriteJSON", idleToLongResp).Return(nil).Once()
	time.Sleep(time.Second*2 + idleCheckInterval*2)
	mockClient.AssertExpectations(t)
}
