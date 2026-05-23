package main

import (
	"context"
	"database/sql"
	"go-backend/mistral"
	"go-backend/transcribe"
	"go-backend/user"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/postgres")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = db.PingContext(context.Background()); err != nil {
		log.Fatal(err)
	}
	controller := transcribe.NewController(user.NewUserRepository(db), mistral.NewMistral())
	server := NewServer(controller)
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
