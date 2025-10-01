package internal

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouteContext struct {
	Config  *Config
	Pool    *pgxpool.Pool
	Context context.Context
}

func NewRouteContext() RouteContext {
	config := NewConfig()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, config.PostgresUri)
	if err != nil {
		panic(err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		panic(err)
	}

	routeContext := RouteContext{&config, pool, ctx}

	setupDatabase(&routeContext)

	return routeContext
}

//go:embed migrations/*.sql
var migrationFS embed.FS

func setupDatabase(ctx *RouteContext) {
	// Create a dedicated connection for migrations because migrate wont take a pgx conn (needs database/sql conn)
	migrateConn, err := sql.Open("pgx", ctx.Config.PostgresUri)
	if err != nil {
		panic(fmt.Sprintf("failed to acquire connection for migrations: %w", err))
	}
	defer migrateConn.Close()
	migrateDriver, err := migratepgx.WithInstance(migrateConn, &migratepgx.Config{
		MigrationsTable: "cosmetics_migrations",
	})

	if err != nil {
		panic(fmt.Sprintf("failed to create migrate driver: %w", err))
	}
	migrateSource, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		panic(fmt.Sprintf("failed to create migrate source: %w", err))
	}
	m, err := migrate.NewWithInstance("migration-fs", migrateSource, "migration-db", migrateDriver)
	if err != nil {
		panic(fmt.Sprintf("failed to create migrate instance: %w", err))
	}

	// Apply all migrations up to the latest
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Sprintf("failed to apply migrations: %w", err))
	}
}
