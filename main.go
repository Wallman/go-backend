package main

import (
	"context"
	"database/sql"
	"go-backend/db"
	"go-backend/mistral"
	"go-backend/otel"
	"go-backend/transcribe"
	"log"
)

func main() {
	shutdown, err := otel.Setup()
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown()

	database, err := db.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer func(database *sql.DB) {
		_ = database.Close()
	}(database)
	controller := transcribe.NewController(db.New(database), mistral.NewMistral("https://api.mistral.ai/v1"))
	server := NewServer(controller)
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
