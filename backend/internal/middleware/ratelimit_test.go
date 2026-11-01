package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func setup(cfg RateLimitConfig) (*gin.Engine, *rateLimiter, *clock) {
	gin.SetMode(gin.TestMode)
	clk := &clock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	rl := newRateLimiter(cfg, clk.now)

	r := gin.New()
	r.Use(rl.handle)
	r.GET("/limited", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r, rl, clk
}

func call(r *gin.Engine, path, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = ip + ":12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimitAllowsBudgetThenRejects(t *testing.T) {
	r, _, _ := setup(RateLimitConfig{Requests: 3, Window: time.Minute})

	for i, wantRemaining := range []string{"2", "1", "0"} {
		w := call(r, "/limited", "192.0.2.1")
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, w.Code)
		}
		if got := w.Header().Get("X-RateLimit-Remaining"); got != wantRemaining {
			t.Errorf("request %d: remaining = %s, want %s", i+1, got, wantRemaining)
		}
		if got := w.Header().Get("X-RateLimit-Limit"); got != "3" {
			t.Errorf("limit header = %s, want 3", got)
		}
	}

	w := call(r, "/limited", "192.0.2.1")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}
	// 3 per minute refills one token every 20 seconds.
	if got := w.Header().Get("Retry-After"); got != "20" {
		t.Errorf("Retry-After = %q, want 20", got)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body["error"] != "too many requests, try again in 20 seconds" {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestRateLimitRefillsOverTime(t *testing.T) {
	r, _, clk := setup(RateLimitConfig{Requests: 2, Window: time.Minute})
	call(r, "/limited", "192.0.2.1")
	call(r, "/limited", "192.0.2.1")
	if w := call(r, "/limited", "192.0.2.1"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}

	// Rejected requests must not consume tokens: one token after 30s.
	clk.advance(30 * time.Second)
	if w := call(r, "/limited", "192.0.2.1"); w.Code != http.StatusOK {
		t.Fatalf("after refill: status = %d, want 200", w.Code)
	}
	if w := call(r, "/limited", "192.0.2.1"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}
}

func TestRateLimitIsPerClient(t *testing.T) {
	r, _, _ := setup(RateLimitConfig{Requests: 1, Window: time.Minute})
	call(r, "/limited", "192.0.2.1")

	if w := call(r, "/limited", "192.0.2.1"); w.Code != http.StatusTooManyRequests {
		t.Errorf("same client: status = %d, want 429", w.Code)
	}
	if w := call(r, "/limited", "198.51.100.7"); w.Code != http.StatusOK {
		t.Errorf("other client: status = %d, want 200", w.Code)
	}
}

func TestRateLimitExemptPaths(t *testing.T) {
	r, _, _ := setup(RateLimitConfig{Requests: 1, Window: time.Minute, Exempt: []string{"/health"}})
	call(r, "/limited", "192.0.2.1")

	for range 5 {
		if w := call(r, "/health", "192.0.2.1"); w.Code != http.StatusOK {
			t.Fatalf("exempt path: status = %d, want 200", w.Code)
		}
	}
}

func TestRateLimitDisabled(t *testing.T) {
	r, _, _ := setup(RateLimitConfig{Requests: 0, Window: time.Minute})
	for range 50 {
		if w := call(r, "/limited", "192.0.2.1"); w.Code != http.StatusOK || w.Header().Get("X-RateLimit-Limit") != "" {
			t.Fatalf("disabled limiter interfered: %d", w.Code)
		}
	}
}

func TestRateLimitSweepsIdleClients(t *testing.T) {
	r, rl, clk := setup(RateLimitConfig{Requests: 5, Window: time.Minute})
	call(r, "/limited", "192.0.2.1")
	call(r, "/limited", "192.0.2.2")

	clk.advance(30 * time.Second)
	call(r, "/limited", "192.0.2.2") // .2 stays active

	clk.advance(40 * time.Second) // .1 idle for 70s, .2 for 40s
	call(r, "/limited", "192.0.2.3")

	if _, ok := rl.clients["192.0.2.1"]; ok {
		t.Error("idle client was not swept")
	}
	if len(rl.clients) != 2 {
		t.Errorf("clients = %d, want 2 (.2 and .3)", len(rl.clients))
	}
}
