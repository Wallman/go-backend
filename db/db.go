package db

import (
	"context"
	"database/sql"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/semconv/v1.26.0"
)

func Connect(ctx context.Context) (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := otelsql.Open("pgx", dbURL, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL)); err != nil {
		_ = db.Close()
		return nil, err
	}
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	defer func(m *migrate.Migrate) {
		_, _ = m.Close()
	}(m)
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
