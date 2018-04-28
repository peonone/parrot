package main

import (
	"log"

	"github.com/peonone/parrot/chat"

	goweb "github.com/micro/go-web"
	"github.com/peonone/parrot/chat/web"
)

func main() {
	// New web service
	service := goweb.NewService(
		goweb.Name(chat.WebServiceName),
	)

	if err := service.Init(); err != nil {
		log.Fatal("Init", err)
	}
	termFun := web.Init(service)
	if err := service.Run(); err != nil {
		log.Fatal("Run: ", err)
	}
	termFun()
}
