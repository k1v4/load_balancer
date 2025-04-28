package main

import (
	"fmt"
	"github.com/k1v4/load_balancer/internal/balancer"
	"github.com/k1v4/load_balancer/internal/config"
	"github.com/k1v4/load_balancer/internal/proxy"
	"log"
	"net/http"
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

	//навешиваем обработчик
	http.Handle("/", reverseProxy)

	//стартуем HTTP-сервер
	log.Println("load balancer is listening on port 8080...")
	err = http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
