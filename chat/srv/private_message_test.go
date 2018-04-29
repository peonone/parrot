package srv

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-redis/redis"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/mock"
)

type pmTestData struct {
	fromUID string
	toUID   string
	online  bool
	content string
	rdsErr  error
}

func TestPrivateMessage(t *testing.T) {
	ssMock := new(mockStateStore)
	mqMock := new(mockMQSender)
	baseHandler := &baseHandler{
		stateStore: ssMock,
		mqSender:   mqMock,
	}
	ph := &privateHandler{baseHandler}
	testDatas := []pmTestData{
		{
			fromUID: "3x",
			toUID:   "peon",
			online:  true,
		},
		{
			fromUID: "3x",
			toUID:   "bla",
			online:  false,
		},
		{
			fromUID: "3x",
			toUID:   "bla",
			online:  false,
			rdsErr:  errors.New("err1"),
		},
	}
	res := new(proto.SendPMRes)
	for _, testData := range testDatas {
		req := &proto.SendPMReq{
			FromUID: testData.fromUID,
			ToUID:   testData.toUID,
			Content: testData.content,
		}
		var ocReturnVals []interface{}
		if testData.rdsErr != nil {
			ocReturnVals = []interface{}{"", testData.rdsErr}
		} else if testData.online {
			ocReturnVals = []interface{}{"111", nil}
		} else {
			ocReturnVals = []interface{}{"", redis.Nil}
		}
		ssMock.On("getOnlineWebNode", testData.toUID).Return(ocReturnVals...).Once()

		if testData.online {
			// TODO verify the message sent to MQ is valid
			mqMock.On("sendMQMsg", mock.Anything, mock.Anything).Return(nil).Once()
		}
		err := ph.Send(context.Background(), req, res)

		if testData.rdsErr != nil {
			assert.Equal(t, err, testData.rdsErr)
		} else {
			assert.Equal(t, res.Success, testData.online)
		}
		ssMock.AssertExpectations(t)
		mqMock.AssertExpectations(t)
	}
}
