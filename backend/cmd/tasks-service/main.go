package main

import (
	"log"

	tasksapp "github.com/gabrielnakaema/project-chat/internal/tasks/app"
)

func main() {
	a, err := tasksapp.New()
	if err != nil {
		log.Fatal("error while starting tasks-service", "error", err.Error())
		return
	}

	defer a.Close()

	err = a.Serve()
	if err != nil {
		log.Fatal("received error from tasks-service serve", "error", err.Error())
		return
	}
}
