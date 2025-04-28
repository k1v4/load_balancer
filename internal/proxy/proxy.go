package proxy

import (
	"github.com/k1v4/load_balancer/internal/balancer"
	"log"
	"net/http"
	"net/http/httputil"
)

func NewReverseProxy(balancer *balancer.Balancer) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			backend := balancer.NextBackend()
			req.URL.Scheme = backend.Scheme
			req.URL.Host = backend.Host
			req.Host = backend.Host // чтобы Host заголовок тоже был правильный
			log.Printf("proxying request to backend: %s", backend.String())
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error: %v", err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		},
	}
}
