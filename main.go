package main

import (
	"context"
	"go-backend/db"
	"go-backend/mistral"
	"go-backend/transcribe"
	"go-backend/user"
	"log"
)

func main() {
	database, err := db.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	controller := transcribe.NewController(user.NewUserRepository(database), mistral.NewMistral("https://api.mistral.ai/v1"))
	server := NewServer(controller)
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
