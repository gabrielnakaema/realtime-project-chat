package main

import (
	"log"

	chatapp "github.com/gabrielnakaema/project-chat/internal/chat/app"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
)

func main() {
	if err := apphost.Run("chat-service", chatapp.New); err != nil {
		log.Fatal(err)
	}
}
