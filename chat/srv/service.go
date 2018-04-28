package srv

import (
	micro "github.com/micro/go-micro"
	"github.com/peonone/parrot"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
	"github.com/streadway/amqp"
)

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

type baseHandler struct {
	stateStore stateStore
	mqSender   mqSender
}

// Init initialize chat service resources and register all handlers
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
	baseHandler := &baseHandler{
		stateStore: &redisStateStore{rdsClient},
		mqSender:   &amqpMqSender{amqpClient},
	}
	proto.RegisterPrivateHandler(service.Server(), &privateHandler{baseHandler})
	proto.RegisterStateHandler(service.Server(), &stateHandler{baseHandler})
	proto.RegisterShoutHandler(service.Server(), &worldShoutHandler{baseHandler})
}
