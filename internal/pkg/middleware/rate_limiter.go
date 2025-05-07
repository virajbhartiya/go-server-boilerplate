package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"go-server-boilerplate/internal/pkg/errors"
)

// IPRateLimiter represents a rate limiter for IP addresses
type IPRateLimiter struct {
	ips   map[string]*rate.Limiter
	mu    *sync.RWMutex
	rate  rate.Limit
	burst int
}

// NewIPRateLimiter creates a new IP rate limiter
func NewIPRateLimiter(requests int, duration time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		mu:    &sync.RWMutex{},
		rate:  rate.Limit(float64(requests) / duration.Seconds()),
		burst: requests,
	}
}

// GetLimiter gets the rate limiter for an IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.ips[ip] = limiter
		i.mu.Unlock()
	}

	return limiter
}

// RateLimiter middleware limits the number of requests per IP
func RateLimiter(requests int, duration time.Duration) gin.HandlerFunc {
	ipLimiter := NewIPRateLimiter(requests, duration)

	return func(c *gin.Context) {
		// Get the IP address from the request
		ip := c.ClientIP()

		// Get the rate limiter for this IP
		limiter := ipLimiter.GetLimiter(ip)

		// Check if this request exceeds the rate limit
		if !limiter.Allow() {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{
					"error":   errors.ErrTimeout.Error(),
					"message": "rate limit exceeded",
				},
			)
			return
		}

		c.Next()
	}
}
