package main

import (
	"context"
	"os"
	"time"

	"github.com/chandiniv1/transfers-system/api"
	db "github.com/chandiniv1/transfers-system/db/sqlc"
	"github.com/chandiniv1/transfers-system/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// ✅ Use pgxpool to connect
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to database using pgxpool")
	}

	// ✅ Pass pgx pool to store
	store := db.NewStore(connPool)

	server, err := api.NewServer(store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
