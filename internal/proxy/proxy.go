package proxy

import (
	"encoding/json"
	"github.com/k1v4/load_balancer/internal/balancer"
	"github.com/k1v4/load_balancer/internal/entity"
	"log"
	"net/http"
	"net/http/httputil"
)

// NewReverseProxy создаёт и возвращает настроенный reverse proxy,
// который будет перенаправлять входящие HTTP-запросы на один из бэкенд-серверов
func NewReverseProxy(b *balancer.Balancer) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{

		// перенастройка запроса перед его отправкой на бэкенд.
		Director: func(req *http.Request) {
			backend, err := b.NextBackend()
			if err != nil {
				log.Printf("Backend selection error: %v", err)
				req.Header.Set("X-LB-Error", err.Error())

				return
			}

			// смена параметров запроса
			req.URL.Scheme = backend.Scheme
			req.URL.Host = backend.Host
			req.Host = backend.Host

			log.Printf("Routing to %s", backend.String())
		},
		// отлавливает ошибки при перенаправлении запроса
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			switch r.Header.Get("X-LB-Error") {
			case "all servers are overloaded (no tokens available)":
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("503 - All backend servers are busy"))
				json.NewEncoder(w).Encode(entity.ErrorResponse{
					Code:    http.StatusServiceUnavailable,
					Message: "All backend servers are busy\"",
				})
				return
			case "rate limit exceeded":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(entity.ErrorResponse{
					Code:    http.StatusTooManyRequests,
					Message: "Rate limit exceeded",
				})
				return
			default:
				log.Printf("Proxy error: %v", err)
				w.WriteHeader(http.StatusBadGateway)
				_ = json.NewEncoder(w).Encode(entity.ErrorResponse{
					Code:    http.StatusBadGateway,
					Message: "Bad Gateway: backend service is unavailable",
				})
			}
		},
	}
}
