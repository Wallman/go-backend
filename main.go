package main

import (
	"context"
	"database/sql"
	"go-backend/mistral"
	"go-backend/transcribe"
	"go-backend/user"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = db.PingContext(context.Background()); err != nil {
		log.Fatal(err)
	}
	if err = runMigrations(dbURL); err != nil {
		log.Fatal(err)
	}

	controller := transcribe.NewController(user.NewUserRepository(db), mistral.NewMistral())
	server := NewServer(controller)
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func runMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
