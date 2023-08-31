package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/ninashvl/avito-backend-test/internal/config"
	server "github.com/ninashvl/avito-backend-test/internal/server"
)

var cfgPath = flag.String("c", "./configs/local_config.toml", "path to config file")

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

	minioClient, err := minio.New(cfg.S3.Host, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
	})
	if err != nil {
		return fmt.Errorf("creating minio client error: %v", err)
	}
	defer db.Close()
	srv := server.NewServer(&cfg, db, minioClient)
	eg.Go(func() error {
		return srv.RunServer(ctx)
	})
	return eg.Wait()
}
