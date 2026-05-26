package db

import (
	"context"
	"database/sql"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(ctx context.Context) (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		db.Close()
		return nil, err
	}
	defer m.Close()
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		db.Close()
		return nil, err
	}
	return db, nil
}
