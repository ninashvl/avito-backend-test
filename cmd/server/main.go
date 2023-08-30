package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ninashvl/avito-backend-test/internal/config"
	server "github.com/ninashvl/avito-backend-test/internal/server"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os/signal"
	"syscall"
)

var cfgPath = flag.String("c", "./configs/config.toml", "path to config file")

func main() {
	flag.Parse()
	err := Run()
	if err != nil {
		log.Fatal(err)
	}
}

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	cfg, err := config.ParseAndValidate(*cfgPath)
	if err != nil {
		return fmt.Errorf("config parse error: %w", err)
	}
	db, err := sqlx.Connect("postgres", cfg.PgConfig.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	srv := server.NewServer(&cfg, db)
	eg.Go(func() error {
		return srv.RunServer(ctx)
	})
	return eg.Wait()
}
