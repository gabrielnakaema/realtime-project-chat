package main

import (
	"log"

	realtimeapp "github.com/gabrielnakaema/project-chat/internal/realtime/app"
)

func main() {
	app, err := realtimeapp.New()
	if err != nil {
		log.Fatal("error while starting websocket-service", "error", err.Error())
	}
	defer app.Close()

	if err := app.Serve(); err != nil {
		log.Fatal("received error from websocket-service", "error", err.Error())
	}
}
