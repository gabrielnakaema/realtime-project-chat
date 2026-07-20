package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/mcpapi"
)

func main() {
	a, err := mcpapi.New()
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
