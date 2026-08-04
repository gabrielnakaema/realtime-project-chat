package main

import (
	"log"

	notificationapp "github.com/gabrielnakaema/project-chat/internal/notification/app"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
)

func main() {
	if err := apphost.Run("notification-service", notificationapp.New); err != nil {
		log.Fatal(err)
	}
}
