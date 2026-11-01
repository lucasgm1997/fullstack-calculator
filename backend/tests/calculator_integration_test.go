// Package tests exercises the full HTTP stack (router → handler → use case →
// domain) with net/http/httptest.
package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"calculator-api/internal/router"
)

func do(t *testing.T, method, path, body string) (int, map[string]any) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.New().ServeHTTP(w, req)

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

func TestUnknownRoutesReturnJSON(t *testing.T) {
	if status, body := do(t, http.MethodGet, "/api/nope", ""); status != http.StatusNotFound || body["error"] != "not found" {
		t.Errorf("unknown route: got %d %v", status, body)
	}
	if status, body := do(t, http.MethodGet, "/api/calculate", ""); status != http.StatusMethodNotAllowed || body["error"] != "method not allowed" {
		t.Errorf("wrong method: got %d %v", status, body)
	}
}
