// Package tests exercises the full HTTP stack (router → handler → use case →
// domain) with net/http/httptest.
package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"calculator-api/internal/config"
	"calculator-api/internal/router"
)

// newRouter builds the real router. Rate limiting is off unless a test enables
// it, so the functional tests stay independent of each other.
func newRouter(t *testing.T, cfg config.Config) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	if cfg.TrustedProxies == nil {
		cfg.TrustedProxies = config.DefaultTrustedProxies
	}
	r, err := router.New(cfg)
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}
	return r
}

func do(t *testing.T, method, path, body string) (int, map[string]any) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(t, config.Config{}).ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var decoded map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("response is not JSON: %q", w.Body.String())
	}
	return w.Code, decoded
}

func TestCalculateSuccess(t *testing.T) {
	tests := []struct {
		body string
		want float64
	}{
		{`{"operation":"add","a":10,"b":5}`, 15},
		{`{"operation":"subtract","a":10,"b":5}`, 5},
		{`{"operation":"multiply","a":10,"b":5}`, 50},
		{`{"operation":"divide","a":10,"b":4}`, 2.5},
		{`{"operation":"power","a":2,"b":8}`, 256},
		{`{"operation":"sqrt","a":81}`, 9},
		{`{"operation":"percentage","a":15,"b":200}`, 30},
	}

	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			status, body := do(t, http.MethodPost, "/api/calculate", tt.body)
			if status != http.StatusOK {
				t.Fatalf("status = %d, body = %v", status, body)
			}
			if body["result"] != tt.want {
				t.Errorf("result = %v, want %v", body["result"], tt.want)
			}
		})
	}
}

func TestCalculateErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{"division by zero", `{"operation":"divide","a":10,"b":0}`, "division by zero"},
		{"negative square root", `{"operation":"sqrt","a":-1}`, "square root of a negative number"},
		{"unsupported operation", `{"operation":"modulo","a":10,"b":3}`, `unsupported operation "modulo"`},
		{"missing operation", `{"a":10,"b":3}`, `unsupported operation ""`},
		{"missing a", `{"operation":"add","b":3}`, `missing operand: field "a" is required`},
		{"missing b", `{"operation":"add","a":3}`, `missing operand: field "b" is required for add`},
		{"null operand", `{"operation":"add","a":null,"b":3}`, `missing operand: field "a" is required`},
		{"invalid operand type", `{"operation":"add","a":"1","b":3}`, `field "a" must be a finite number`},
		{"malformed JSON", `not json`, "request body must be valid JSON"},
		{"overflow", `{"operation":"multiply","a":1e308,"b":10}`, "result is not a finite number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := do(t, http.MethodPost, "/api/calculate", tt.body)
			if status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", status)
			}
			if body["error"] != tt.wantError {
				t.Errorf("error = %v, want %q", body["error"], tt.wantError)
			}
		})
	}
}

func TestHealth(t *testing.T) {
	status, body := do(t, http.MethodGet, "/api/health", "")
	if status != http.StatusOK || body["status"] != "ok" {
		t.Errorf("got %d %v", status, body)
	}
}

func TestAPIDocs(t *testing.T) {
	r := newRouter(t, config.Config{})
	for path, wantType := range map[string]string{
		"/api/openapi.yaml": "application/yaml",
		"/api/docs":         "text/html",
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

		if w.Code != http.StatusOK {
			t.Errorf("%s: status = %d", path, w.Code)
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, wantType) {
			t.Errorf("%s: Content-Type = %q, want %s", path, ct, wantType)
		}
	}
}

func TestUnknownRoutesReturnJSON(t *testing.T) {
	if status, body := do(t, http.MethodGet, "/api/nope", ""); status != http.StatusNotFound || body["error"] != "not found" {
		t.Errorf("unknown route: got %d %v", status, body)
	}
	if status, body := do(t, http.MethodGet, "/api/calculate", ""); status != http.StatusMethodNotAllowed || body["error"] != "method not allowed" {
		t.Errorf("wrong method: got %d %v", status, body)
	}
}

func post(r *gin.Engine, remoteAddr, forwardedFor string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/calculate", strings.NewReader(`{"operation":"add","a":1,"b":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimitThroughTheRouter(t *testing.T) {
	r := newRouter(t, config.Config{RateLimitRequests: 2, RateLimitWindow: time.Minute})

	for i := range 2 {
		if w := post(r, "203.0.113.10:1000", ""); w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d", i+1, w.Code)
		}
	}
	w := post(r, "203.0.113.10:1000", "")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" || !strings.Contains(w.Body.String(), "too many requests") {
		t.Errorf("429 response missing details: %v %s", w.Header(), w.Body.String())
	}

	// The health check is exempt, so the UI never reports "API offline" because of the limit.
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.RemoteAddr = "203.0.113.10:1000"
	hw := httptest.NewRecorder()
	r.ServeHTTP(hw, req)
	if hw.Code != http.StatusOK {
		t.Errorf("health: status = %d, want 200", hw.Code)
	}
}

func TestRateLimitUsesRealClientIPBehindTrustedProxy(t *testing.T) {
	r := newRouter(t, config.Config{RateLimitRequests: 1, RateLimitWindow: time.Minute})
	const proxy = "172.18.0.3:5000" // e.g. nginx on the Docker network

	// Two users behind the same proxy get separate budgets.
	if w := post(r, proxy, "198.51.100.1"); w.Code != http.StatusOK {
		t.Fatalf("user 1: status = %d", w.Code)
	}
	if w := post(r, proxy, "198.51.100.2"); w.Code != http.StatusOK {
		t.Fatalf("user 2: status = %d", w.Code)
	}
	if w := post(r, proxy, "198.51.100.1"); w.Code != http.StatusTooManyRequests {
		t.Errorf("user 1 again: status = %d, want 429", w.Code)
	}
}

func TestRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	r := newRouter(t, config.Config{RateLimitRequests: 1, RateLimitWindow: time.Minute})
	const attacker = "203.0.113.66:4000" // public IP, not a trusted proxy

	post(r, attacker, "198.51.100.1")
	// Rotating a fake X-Forwarded-For must not grant a fresh budget.
	if w := post(r, attacker, "198.51.100.99"); w.Code != http.StatusTooManyRequests {
		t.Errorf("spoofed header bypassed the limit: status = %d", w.Code)
	}
}

func TestRouterRejectsInvalidTrustedProxies(t *testing.T) {
	if _, err := router.New(config.Config{TrustedProxies: []string{"not-an-ip"}}); err == nil {
		t.Error("expected an error for an invalid proxy")
	}
}
