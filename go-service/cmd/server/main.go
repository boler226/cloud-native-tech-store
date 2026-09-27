package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tech-store-go/internal/config"
	"tech-store-go/internal/handler"
	"tech-store-go/internal/repository"
	"tech-store-go/internal/router"
	"tech-store-go/internal/seed"
	"tech-store-go/internal/service"
)

func main() {
	cfg := config.Load()

	// Збираємо застосунок: repository -> service -> handler -> router.
	repo := repository.NewMemoryProductRepository(seed.Products())
	productService := service.NewProductService(repo)
	concurrencyService := service.NewConcurrencyService()

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: router.New(
			handler.NewHealthHandler(),
			handler.NewProductHandler(productService),
			handler.NewConcurrencyHandler(concurrencyService),
		),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("%s listening on http://localhost:%s", config.ServiceName, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Коректне завершення по Ctrl+C / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
