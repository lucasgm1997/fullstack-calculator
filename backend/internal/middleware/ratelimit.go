// Package middleware contains cross-cutting Gin middleware.
package middleware

import (
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"calculator-api/internal/dto"
)

// RateLimitConfig configures per-client rate limiting.
//
// Each client (by IP) gets a token bucket holding Requests tokens that refills
// at Requests per Window, so a client can burst up to Requests calls and then
// sustain Requests per Window.
type RateLimitConfig struct {
	// Requests allowed per Window for each client. Zero or less disables limiting.
	Requests int
	Window   time.Duration
	// Exempt lists exact request paths that are never limited (e.g. health checks).
	Exempt []string
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	cfg       RateLimitConfig
	refill    rate.Limit
	now       func() time.Time
	mu        sync.Mutex
	clients   map[string]*client
	lastSweep time.Time
}

// RateLimit returns middleware that answers 429 Too Many Requests once a client
// exceeds its budget. The client is identified by gin's ClientIP, so configure
// the engine's trusted proxies to get real IPs behind a reverse proxy.
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	return newRateLimiter(cfg, time.Now).handle
}

func newRateLimiter(cfg RateLimitConfig, now func() time.Time) *rateLimiter {
	rl := &rateLimiter{cfg: cfg, now: now, clients: map[string]*client{}, lastSweep: now()}
	if cfg.Requests > 0 && cfg.Window > 0 {
		rl.refill = rate.Limit(float64(cfg.Requests) / cfg.Window.Seconds())
	}
	return rl
}

func (rl *rateLimiter) handle(c *gin.Context) {
	if rl.refill == 0 || slices.Contains(rl.cfg.Exempt, c.Request.URL.Path) {
		c.Next()
		return
	}

	now := rl.now()
	limiter := rl.limiterFor(c.ClientIP(), now)
	c.Header("X-RateLimit-Limit", strconv.Itoa(rl.cfg.Requests))

	reservation := limiter.ReserveN(now, 1)
	if delay := reservation.DelayFrom(now); delay > 0 {
		reservation.CancelAt(now) // the request is rejected, so give the token back
		seconds := int(math.Ceil(delay.Seconds()))
		c.Header("Retry-After", strconv.Itoa(seconds))
		c.Header("X-RateLimit-Remaining", "0")
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
			Error: fmt.Sprintf("too many requests, try again in %d seconds", seconds),
		})
		return
	}

	c.Header("X-RateLimit-Remaining", strconv.Itoa(int(limiter.TokensAt(now))))
	c.Next()
}

// limiterFor returns the client's bucket, creating it on first use. Clients
// idle for a full window are dropped: their bucket would be full again anyway,
// and this keeps memory bounded when many distinct IPs show up.
func (rl *rateLimiter) limiterFor(ip string, now time.Time) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if now.Sub(rl.lastSweep) >= rl.cfg.Window {
		for key, cl := range rl.clients {
			if now.Sub(cl.lastSeen) >= rl.cfg.Window {
				delete(rl.clients, key)
			}
		}
		rl.lastSweep = now
	}

	cl, ok := rl.clients[ip]
	if !ok {
		cl = &client{limiter: rate.NewLimiter(rl.refill, rl.cfg.Requests)}
		rl.clients[ip] = cl
	}
	cl.lastSeen = now
	return cl.limiter
}
