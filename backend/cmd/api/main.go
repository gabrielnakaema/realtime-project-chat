package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/api"
)

func main() {
	a, err := api.NewApi()
	if err != nil {
		log.Fatal("error while starting api", "error", err.Error())
		return
	}

	defer a.Publisher.Close()

	err = a.Serve()
	if err != nil {
		log.Fatal("received error from api serve", "error", err.Error())
		return
	}
}
