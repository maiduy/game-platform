package security

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	// stores IP -> list of request timestamps
	requests     map[string][]time.Time
	mutex        sync.Mutex
	limit        int           // maximum number of requests
	window       time.Duration // time window for rate limiting
	cleanupEvery int           // cleanup counter
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:     make(map[string][]time.Time),
		limit:        limit,
		window:       window,
		cleanupEvery: 100, // cleanup every 100 requests
	}
}

// RateLimitMiddleware returns a middleware that limits request rates
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	var requestCount int

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		now := time.Now()
		windowStart := now.Add(-rl.window)

		// Get existing requests for this IP
		times, exists := rl.requests[clientIP]
		if !exists {
			times = []time.Time{}
		}

		// Filter out old requests
		var recent []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				recent = append(recent, t)
			}
		}

		// Update the requests list
		rl.requests[clientIP] = append(recent, now)

		// Check if rate limit is exceeded
		if len(rl.requests[clientIP]) > rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Periodic cleanup of old entries
		requestCount++
		if requestCount%rl.cleanupEvery == 0 {
			go rl.cleanup(windowStart)
		}

		c.Next()
	}
}

// cleanup removes old entries from the requests map
func (rl *RateLimiter) cleanup(cutoff time.Time) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	for ip, times := range rl.requests {
		var recent []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}

		if len(recent) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = recent
		}
	}
}

// IPWhitelistMiddleware returns a middleware that allows only whitelisted IPs
func IPWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
	// Convert slice to map for O(1) lookup
	whitelistMap := make(map[string]bool)
	for _, ip := range whitelist {
		whitelistMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if IP is whitelisted
		if !whitelistMap[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP not whitelisted",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
