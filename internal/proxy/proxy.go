package proxy

import (
	"github.com/k1v4/load_balancer/internal/balancer"
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
			if lbErr := r.Header.Get("X-LB-Error"); lbErr == "all servers are overloaded (no tokens available)" {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("503 - All backend servers are busy"))

				return
			}

			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}
}
