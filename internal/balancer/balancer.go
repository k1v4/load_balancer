package balancer

import (
	"fmt"
	"github.com/k1v4/load_balancer/internal/config"
	"net/url"
	"sync"
	"time"
)

type Balancer struct {
	backends []*Backend
	index    int
	mu       sync.Mutex
	stopChan chan struct{}
}

type Backend struct {
	URL        *url.URL
	tokens     int        // текущее количество токенов
	capacity   int        // максимальное количество токенов (BucketSize)
	refillRate int        // токенов в секунду (RefillRate)
	lastRefill time.Time  // время последнего пополнения
	mu         sync.Mutex // защтиа доступа к токенам
}

func NewBalancer(cfgs []config.Client) (*Balancer, error) {
	b := &Balancer{
		stopChan: make(chan struct{}),
	}

	for _, cfg := range cfgs {
		u, err := url.Parse(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid backend URL %s: %v", cfg.URL, err)
		}

		b.backends = append(b.backends, &Backend{
			URL:        u,
			tokens:     cfg.BucketSize,
			capacity:   cfg.BucketSize,
			refillRate: cfg.RefillRate,
			lastRefill: time.Now(),
		})
	}

	go b.refillTokens() // фоновое пополнение токенов
	return b, nil
}

// refillTokens фоновая горутина для пополнения токенов
func (b *Balancer) refillTokens() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			for _, be := range b.backends {
				be.mu.Lock()
				now := time.Now()
				elapsed := now.Sub(be.lastRefill).Seconds()
				tokensToAdd := int(elapsed * float64(be.refillRate))

				if tokensToAdd > 0 {
					be.tokens += tokensToAdd
					if be.tokens > be.capacity {
						be.tokens = be.capacity
					}
					be.lastRefill = now
				}
				be.mu.Unlock()
			}
			b.mu.Unlock()
		case <-b.stopChan:
			return
		}
	}
}

// NextBackend выбор бэкенда с учетом его текущей загрузки
func (b *Balancer) NextBackend() (*url.URL, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.backends) == 0 {
		return nil, fmt.Errorf("no backends available")
	}

	// проверяем все бэкенды по кругу
	for i := 0; i < len(b.backends); i++ {
		be := b.backends[b.index]
		b.index = (b.index + 1) % len(b.backends)

		be.mu.Lock()
		if be.tokens > 0 {
			be.tokens--
			be.mu.Unlock()
			return be.URL, nil
		}
		be.mu.Unlock()
	}

	return nil, fmt.Errorf("all servers are overloaded (no tokens available)")
}

func (b *Balancer) Stop() {
	close(b.stopChan)
}
