package balancer

import (
	"net/url"
	"sync"
)

type Balancer struct {
	Backends []*url.URL
	index    int
	mu       *sync.Mutex
}

func NewBalancer(backendUrls []string) (*Balancer, error) {
	var backends []*url.URL
	for _, addr := range backendUrls {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, err
		}

		backends = append(backends, u)
	}

	return &Balancer{
		Backends: backends,
		mu:       &sync.Mutex{},
		index:    0,
	}, nil
}

func (b *Balancer) NextBackend() *url.URL {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.Backends) == 0 {
		return nil
	}

	backend := b.Backends[b.index]
	b.index = (b.index + 1) % len(b.Backends)

	return backend
}
