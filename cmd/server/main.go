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

	"pcclub-server/internal/app"
)

func main() {
	cfg := app.ConfigFromEnv()

	serverApp, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer serverApp.Close()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           serverApp.Routes(),
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("pcclub server listening on :%s", cfg.Port)
		log.Printf("server config: read_timeout=%v write_timeout=%v idle_timeout=%v",
			cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}
}
