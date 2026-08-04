package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	realtimeapp "github.com/gabrielnakaema/project-chat/internal/realtime/app"
)

func main() {
	if err := apphost.Run("websocket-service", realtimeapp.New); err != nil {
		log.Fatal(err)
	}
}
