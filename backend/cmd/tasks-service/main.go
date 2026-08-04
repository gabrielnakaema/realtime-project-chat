package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	tasksapp "github.com/gabrielnakaema/project-chat/internal/tasks/app"
)

func main() {
	if err := apphost.Run("tasks-service", tasksapp.New); err != nil {
		log.Fatal(err)
	}
}
