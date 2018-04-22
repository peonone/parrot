package srv

import (
	micro "github.com/micro/go-micro"
	"github.com/peonone/parrot"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
	"github.com/streadway/amqp"
)

const Name = "go.micro.srv.chat"

type mqSender interface {
	sendMQMsg(key string, body []byte) error
}

type amqpMqSender struct {
	*amqp.Channel
}

func (s *amqpMqSender) sendMQMsg(key string, body []byte) error {
	msg := amqp.Publishing{
		Headers:         amqp.Table{},
		ContentType:     "application/octet-stream",
		ContentEncoding: "",
		Body:            body,
		DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
		Priority:        0,              // 0-9
	}
	return s.Publish(chat.PushMsgExchangeName, key, false, false, msg)
}

type BasicHandler struct {
	stateStore stateStore
	mqSender   mqSender
}

func Init(service micro.Service) {
	rdsClient := parrot.MakeRedisClient()
	amqpClient, err := parrot.MakeAMQPClient()
	if err != nil {
		panic(err)
	}

	err = amqpClient.ExchangeDeclare(chat.PushMsgExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	basicHandler := &BasicHandler{
		stateStore: &redisStateStore{rdsClient},
		mqSender:   &amqpMqSender{amqpClient},
	}
	proto.RegisterPrivateHandler(service.Server(), &PrivateHandler{basicHandler})
	proto.RegisterStateHandler(service.Server(), &stateHandler{basicHandler})
}
