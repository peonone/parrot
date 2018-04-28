package srv

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/peonone/parrot/chat"
	"github.com/stretchr/testify/mock"

	"github.com/micro/go-micro/registry"

	"github.com/peonone/parrot/chat/proto"
)

type shoutTestData struct {
	fromUID       string
	content       string
	hasWebService bool
	nodes         []string
	getServiceErr error
}

func makeMockRegistryServices(name string, nodeNames []string) []*registry.Service {
	nodes := make([]*registry.Node, 0, len(nodeNames))
	for _, nodeName := range nodeNames {
		nodes = append(nodes, &registry.Node{
			Id: nodeName,
		})
	}
	return []*registry.Service{
		&registry.Service{
			Name:  name,
			Nodes: nodes,
		},
	}
}

func TestShoutMessage(t *testing.T) {
	ssMock := new(mockStateStore)
	mqMock := new(mockMQSender)
	regMock := new(mockRegistry)
	baseHandler := &baseHandler{
		stateStore: ssMock,
		mqSender:   mqMock,
	}
	sh := &worldShoutHandler{baseHandler}
	testDatas := []shoutTestData{
		{
			fromUID:       "peon1",
			content:       "hihi",
			hasWebService: true,
			nodes:         []string{"aaa", "bbb"},
		},
		{
			fromUID:       "peon2",
			content:       "hihi2",
			hasWebService: true,
			nodes:         []string{"xxx", "yyy"},
		},
		{
			fromUID:       "peon3",
			content:       "hihi3",
			hasWebService: false,
		},
		{
			fromUID:       "peon1",
			content:       "hihi",
			hasWebService: true,
			nodes:         []string{"aaa", "bbb"},
			getServiceErr: errors.New("get services failed"),
		},
	}
	res := new(proto.SendShoutRes)
	for _, testData := range testDatas {
		registry.DefaultRegistry = regMock

		var mockServices []*registry.Service
		if testData.hasWebService {
			mockServices = makeMockRegistryServices(chat.WebServiceName, testData.nodes)
		} else {
			mockServices = nil
		}
		regMock.On("GetService", chat.WebServiceName).Return(mockServices, testData.getServiceErr).Once()
		req := &proto.SendShoutReq{
			FromUID: testData.fromUID,
			Content: testData.content,
		}
		if testData.getServiceErr == nil {
			for _, nodeID := range testData.nodes {
				key := fmt.Sprintf("%s.private.push", nodeID)
				mqMock.On("sendMQMsg", key, mock.Anything).Return(nil).Once()
			}
		}
		sh.Send(context.Background(), req, res)
		ssMock.AssertExpectations(t)
		mqMock.AssertExpectations(t)
		regMock.AssertExpectations(t)
	}
}
