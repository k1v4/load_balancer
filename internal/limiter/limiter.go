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

// refillBuckets регулярно (раз в секунду) пополняет токены в каждом бакете в соответствии с RefillRate.
func (l *Limiter) refillBuckets() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.RLock()

		for _, bucket := range l.buckets {
			bucket.Mu.Lock()

			now := time.Now()
			elapsed := now.Sub(bucket.LastRefill).Seconds()

			// считаем сколько токенов нужно добавить
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

// Allow проверяет, доступен ли клиенту доступ на основе токенов.
// Если токен есть — он тратится и доступ разрешается. Иначе — отказ.
func (l *Limiter) Allow(ip net.IP) (bool, error) {
	clientID := ip.String()

	l.mu.RLock()
	bucket, exists := l.buckets[clientID]
	l.mu.RUnlock()

	// неизвестный адрес
	if !exists {
		return false, fmt.Errorf("clientID %s does not exist", clientID)
	}

	bucket.Mu.Lock()
	defer bucket.Mu.Unlock()

	if bucket.Tokens > 0 {
		bucket.Tokens--

		return true, nil
	}

	return false, fmt.Errorf("rate limit exceeded for client %s", clientID)
}
