// Package main is the entrypoint for the spendly backend server.
// It wires configuration, database, repositories, services, handlers,
// middleware, and the router, then starts an HTTP server with graceful shutdown.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mariano145/spendly/backend/internal/config"
	"github.com/Mariano145/spendly/backend/internal/database"
	"github.com/Mariano145/spendly/backend/internal/handler"
	"github.com/Mariano145/spendly/backend/internal/middleware"
	"github.com/Mariano145/spendly/backend/internal/repository"
	"github.com/Mariano145/spendly/backend/internal/router"
	"github.com/Mariano145/spendly/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	pool, err := database.NewPostgresDB(ctx, database.DefaultPoolConfig(cfg.DatabaseURL))
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewRefreshTokenRepository(pool)

	authSvc := service.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWTSecret,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
		cfg.BCryptCost,
	)

	authHandler := handler.NewAuthHandler(authSvc)

	authMW := middleware.Auth(authSvc)
	corsMW := middleware.CORS(cfg.CORSOrigin)

	h := router.New(router.Deps{
		AuthHandler: authHandler,
		AuthMW:      authMW,
		CORSMW:      corsMW,
	})

	addr := fmt.Sprintf(":%s", cfg.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("shutdown signal received, draining connections...")
	case err := <-errCh:
		log.Printf("server error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	pool.Close()
	log.Println("server stopped")
}
