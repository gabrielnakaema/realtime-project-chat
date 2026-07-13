package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/notificationapi"
)

func main() {
	a, err := notificationapi.New()
	if err != nil {
		log.Fatal("error while starting notification-service", "error", err.Error())
		return
	}

	defer a.Close()

	err = a.Serve()
	if err != nil {
		log.Fatal("received error from notification-service serve", "error", err.Error())
		return
	}
}
