package proxy

import (
	"github.com/k1v4/load_balancer/internal/balancer"
	"log"
	"net/http"
	"net/http/httputil"
)

func NewReverseProxy(b *balancer.Balancer) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			backend, err := b.NextBackend()
			if err != nil {
				log.Printf("Backend selection error: %v", err)
				req.Header.Set("X-LB-Error", err.Error())

				return
			}

			req.URL.Scheme = backend.Scheme
			req.URL.Host = backend.Host
			req.Host = backend.Host
			log.Printf("Routing to %s", backend.String())
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if lbErr := r.Header.Get("X-LB-Error"); lbErr != "" {
				if lbErr == "all servers are overloaded (no tokens available)" {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("503 - All backend servers are busy"))

					return
				}
			}
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}
}
