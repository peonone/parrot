package main

import (
	"log"

	goweb "github.com/micro/go-web"
	"github.com/peonone/parrot/chat/web"
)

func main() {
	// New web service
	service := goweb.NewService(
		goweb.Name(web.Name),
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
