package limiter

import (
	"fmt"
	"github.com/k1v4/load_balancer/internal/config"
	"net"
	"sync"
	"time"
)

type Bucket struct {
	Client     config.Client
	Tokens     int
	Capacity   int
	RefillRate int
	LastRefill time.Time
	Mu         *sync.Mutex
}

type Limiter struct {
	buckets map[string]*Bucket
	mu      *sync.RWMutex
}

func (l *Limiter) refillBuckets() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.RLock()

		for _, bucket := range l.buckets {
			bucket.Mu.Lock()

			now := time.Now()
			elapsed := now.Sub(bucket.LastRefill).Seconds()

			tokensToAdd := int(elapsed * float64(bucket.RefillRate))

			if tokensToAdd > 0 {
				bucket.Tokens += tokensToAdd

				if bucket.Tokens > bucket.Capacity {
					bucket.Tokens = bucket.Capacity
				}

				bucket.LastRefill = now
			}

			bucket.Mu.Unlock()
		}

		l.mu.RUnlock()
	}
}

func (l *Limiter) Allow(ip net.IP) (bool, error) {
	clientID := ip.String()

	l.mu.RLock()
	bucket, exists := l.buckets[clientID]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()

		bucket = &Bucket{
			Client: config.Client{
				URL:        clientID,
				BucketSize: 10,
				RefillRate: 5,
			},
			Tokens:     10,
			Capacity:   10,
			RefillRate: 5,
			LastRefill: time.Now(),
			Mu:         &sync.Mutex{},
		}

		l.buckets[clientID] = bucket

		l.mu.Unlock()
	}

	bucket.Mu.Lock()
	defer bucket.Mu.Unlock()

	if bucket.Tokens > 0 {
		bucket.Tokens--

		return true, nil
	}

	return false, fmt.Errorf("rate limit exceeded for client %s", clientID)
}
