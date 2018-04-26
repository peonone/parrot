package srv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-redis/redis"
	"github.com/peonone/parrot/chat/proto"
	"github.com/stretchr/testify/mock"
)

type testDataInfo struct {
	fromUID string
	toUID   string
	online  bool
	content string
}

func TestPrivateMessage(t *testing.T) {
	ssMock := new(mockStateStore)
	mqMock := new(mockMQSender)
	baseHandler := &baseHandler{
		stateStore: ssMock,
		mqSender:   mqMock,
	}
	ph := &privateHandler{baseHandler}
	testDatas := []testDataInfo{
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
	}
	res := new(proto.SendPMRes)
	for _, testData := range testDatas {
		req := &proto.SendPMReq{
			FromUID: testData.fromUID,
			ToUID:   testData.toUID,
			Content: testData.content,
		}
		var ocReturnVals []interface{}
		if testData.online {
			ocReturnVals = []interface{}{"111", nil}
		} else {
			ocReturnVals = []interface{}{"", redis.Nil}
		}
		ssMock.On("getOnlineWebNode", testData.toUID).Return(ocReturnVals...).Once()

		if testData.online {
			// TODO verify the message sent to MQ is valid
			mqMock.On("sendMQMsg", mock.Anything, mock.Anything).Return(nil).Once()
		}
		ph.Send(context.Background(), req, res)
		assert.Equal(t, res.Success, testData.online)
		ssMock.AssertExpectations(t)
		mqMock.AssertExpectations(t)
	}
}
