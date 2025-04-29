package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/k1v4/load_balancer/internal/balancer"
	"github.com/k1v4/load_balancer/internal/config"
	"github.com/k1v4/load_balancer/internal/proxy"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("load balancer configuration error: %v", err)
	}

	balance, err := balancer.NewBalancer(cfg.Backends)
	if err != nil {
		log.Fatalf("load balancer error: %v", err)
	}

	reverseProxy := proxy.NewReverseProxy(balance)
	mux := http.NewServeMux()
	mux.Handle("/", reverseProxy)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	// канал для получения сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("load balancer is listening on port %s...", cfg.Port)

		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// блокируем до сигнала завершения
	<-stop
	log.Println("shutting down server...")

	// контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// плавная остановка сервера
	if err = server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
