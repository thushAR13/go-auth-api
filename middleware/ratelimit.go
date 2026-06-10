package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type visitorStore struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

func newVisitorStore() *visitorStore {
	store := &visitorStore{
		visitors: make(map[string]*visitor),
	}

	go store.cleanupLoop()

	return store
}

func (s *visitorStore) cleanupLoop() {
	for {
		time.Sleep(1 * time.Minute)
		s.mu.Lock()

		for ip, v := range s.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(s.visitors, ip)
			}
		}
		s.mu.Unlock()
	}
}

func (s *visitorStore) getVisitor(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.visitors[ip]

	if !exists {
		limiter := rate.NewLimiter(rate.Every(2*time.Second), 5)
		s.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

var store = newVisitorStore()

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := store.getVisitor(ip)

		if !limiter.Allow() {
			slog.Warn("rate limit exceeded",
				"ip", ip,
				"path", r.URL.Path)
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
