package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/tasksapi"
)

func main() {
	a, err := tasksapi.New()
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
