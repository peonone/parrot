package main

import (
	"log"

	goweb "github.com/micro/go-web"
	"github.com/peonone/parrot/auth/web"
)

func main() {
	service := goweb.NewService(
		goweb.Name(web.Name),
	)

	// parse command line flags
	service.Init()

	web.Init(service)
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
