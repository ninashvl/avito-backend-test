package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/errgroup"

	"github.com/ninashvl/avito-backend-test/internal/config"
	"github.com/ninashvl/avito-backend-test/internal/repository/reports"
	segment2 "github.com/ninashvl/avito-backend-test/internal/repository/segment"
	"github.com/ninashvl/avito-backend-test/internal/repository/user"
	v1 "github.com/ninashvl/avito-backend-test/internal/server/v1"
	"github.com/ninashvl/avito-backend-test/internal/store"
	"github.com/ninashvl/avito-backend-test/internal/usecase/segment"
	user2 "github.com/ninashvl/avito-backend-test/internal/usecase/user"
)

type Server struct {
	srv *http.Server
}

func NewServer(cfg *config.Config, db *sqlx.DB, client *minio.Client) *Server {
	e := echo.New()
	e.Use(middleware.Recover(), middleware.Logger(), middleware.CORS())

	txtr := store.NewTransactor(db)

	segmentRepo := segment2.NewRepo(db)
	userRepo := user.NewRepo(db)
	reportRepo := reports.NewRepo(cfg.S3.Bucket, client)

	segmentUseCase := segment.NewUseCase(segmentRepo)
	userUseCase := user2.NewUseCase(userRepo, reportRepo, txtr)

	handlers := v1.NewHandler(segmentUseCase, userUseCase)
	v1.RegisterHandlers(e, handlers)

	return &Server{
		&http.Server{
			Addr:              cfg.ServerConf.Addr,
			Handler:           e,
			ReadHeaderTimeout: time.Second * 5,
		},
	}
}

func (s *Server) RunServer(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		return s.srv.Shutdown(ctx)
	})
	eg.Go(func() error {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}
