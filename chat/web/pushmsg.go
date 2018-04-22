package web

import (
	"fmt"
	"log"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/micro/go-web"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
	"github.com/streadway/amqp"
)

func listenPushMsgs(amqpClient *amqp.Channel, cmdHandlers []commandHandler) error {
	err := amqpClient.ExchangeDeclare(chat.PushMsgExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}
	q, err := amqpClient.QueueDeclare("", false, false, true, false, nil)
	key := fmt.Sprintf("%s.#", web.DefaultId)
	err = amqpClient.QueueBind(q.Name, key, chat.PushMsgExchangeName, false, nil)
	log.Printf("listening to exchange=%s, key=%s, queue=%s", chat.PushMsgExchangeName, key, q.Name)
	if err != nil {
		return err
	}
	deliveries, err := amqpClient.Consume(
		q.Name, // name
		"",     // consumerTag,
		true,   // noAck
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		nil,    // arguments
	)
	if err != nil {
		return err
	}
	go handle(deliveries, cmdHandlers)
	return nil
}

func handle(deliveries <-chan amqp.Delivery, handlers []commandHandler) {
	for d := range deliveries {
		pushMsg := &proto.PushMsg{}
		err := protobuf.Unmarshal(d.Body, pushMsg)
		if err != nil {
			log.Printf("failed to decode PushMsg: %s", err)
			continue
		}

		handled := false
		for _, h := range handlers {
			if h.canHandlePush(pushMsg.Command) {
				h.handlePush(pushMsg)
				handled = true
				break
			}
		}
		if !handled {
			log.Printf("unable to find corresponding handler for push msg: %s", pushMsg.Command)
		}
	}
}
