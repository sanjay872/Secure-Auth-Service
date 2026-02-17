package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	requests int
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		window:  window,
	}

	// Cleanup old clients periodically
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > rl.window {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Unable to parse IP", http.StatusInternalServerError)
			return
		}

		rl.mu.Lock()
		c, exists := rl.clients[ip]
		if !exists {
			rl.clients[ip] = &client{
				requests: 1,
				lastSeen: time.Now(),
			}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		// Reset window
		if time.Since(c.lastSeen) > rl.window {
			c.requests = 0
		}

		if c.requests >= rl.limit {
			rl.mu.Unlock()
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		c.requests++
		c.lastSeen = time.Now()
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
