package main

import (
	"log"

	notificationapp "github.com/gabrielnakaema/project-chat/internal/notification/app"
)

func main() {
	a, err := notificationapp.New()
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
