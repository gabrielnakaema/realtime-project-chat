package main

import (
	"log"

	mcpapp "github.com/gabrielnakaema/project-chat/internal/mcp/app"
)

func main() {
	a, err := mcpapp.New()
	if err != nil {
		log.Fatal("error while starting mcp-service", "error", err.Error())
		return
	}

	defer a.Close()

	err = a.Serve()
	if err != nil {
		log.Fatal("received error from mcp-service serve", "error", err.Error())
		return
	}
}
