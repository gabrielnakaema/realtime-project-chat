package main

import (
	"log"

	mcpapp "github.com/gabrielnakaema/project-chat/internal/mcp/app"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
)

func main() {
	if err := apphost.Run("mcp-service", mcpapp.New); err != nil {
		log.Fatal(err)
	}
}
