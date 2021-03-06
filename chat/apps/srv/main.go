package main

import (
	"log"
	"time"

	"github.com/peonone/parrot/chat"

	"github.com/micro/go-micro"
	"github.com/peonone/parrot/chat/srv"
)

func main() {
	service := micro.NewService(
		micro.Name(chat.SrvServiceName),
		micro.RegisterTTL(time.Second*10),
		micro.RegisterInterval(time.Second*5),
	)
	service.Init()
	termFun := srv.Init(service)

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
	termFun()
}
