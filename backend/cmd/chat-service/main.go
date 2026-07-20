package main

import (
	"log"

	chatapp "github.com/gabrielnakaema/project-chat/internal/chat/app"
)

func main() {
	a, err := chatapp.New()
	if err != nil {
		log.Fatal("error while starting chat-service", "error", err.Error())
		return
	}

	defer a.Close()

	err = a.Serve()
	if err != nil {
		log.Fatal("received error from chat-service serve", "error", err.Error())
		return
	}
}
