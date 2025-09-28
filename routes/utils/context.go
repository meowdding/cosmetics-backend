package utils

import (
	"context"

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

func setupDatabase(ctx *RouteContext) {
	const cosmetics = `
		create table if not exists cosmetics(
			id varchar primary key,
			version int constraint positive_version check ( version > 0 ),
			data json not null
		)
	`

	const player = `
		create table if not exists players(
			player uuid primary key,
			data json not null,
			cosmetics varchar[] default array[]::varchar[]
		)
	`

	_, err := ctx.Pool.Exec(ctx.Context, cosmetics)
	if err != nil {
		panic(err)
	}
	_, err = ctx.Pool.Exec(ctx.Context, player)
	if err != nil {
		panic(err)
	}
}
