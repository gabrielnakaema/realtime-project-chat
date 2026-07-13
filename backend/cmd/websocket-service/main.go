package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/websocketapi"
)

func main() {
	app, err := websocketapi.New()
	if err != nil {
		log.Fatal("error while starting websocket-service", "error", err.Error())
	}
	defer app.Close()

	if err := app.Serve(); err != nil {
		log.Fatal("received error from websocket-service", "error", err.Error())
	}
}
