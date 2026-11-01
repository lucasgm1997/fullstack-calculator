package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"calculator-api/internal/domain"
	"calculator-api/internal/usecase"
)

type fakeCalculator struct {
	result float64
	err    error
	got    usecase.CalculateInput
}

func (f *fakeCalculator) Execute(in usecase.CalculateInput) (float64, error) {
	f.got = in
	return f.result, f.err
}

func serve(t *testing.T, calc Calculator, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/calculate", NewCalculatorHandler(calc).Calculate)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %q", w.Body.String())
	}
	return body
}

func TestCalculatePassesInputToUseCase(t *testing.T) {
	calc := &fakeCalculator{result: 42}
	w := serve(t, calc, `{"operation":"add","a":40,"b":2}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := decode(t, w)["result"]; got != 42.0 {
		t.Errorf("result = %v, want 42", got)
	}
	if calc.got.Operation != "add" || *calc.got.A != 40 || *calc.got.B != 2 {
		t.Errorf("use case received %+v", calc.got)
	}
}

func TestCalculateMapsErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{"domain error", domain.ErrDivisionByZero, http.StatusBadRequest, "division by zero"},
		{"validation error", usecase.ErrMissingOperand, http.StatusBadRequest, "missing operand"},
		{"unexpected error is hidden", errors.New("boom"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(t, &fakeCalculator{err: tt.err}, `{"operation":"divide","a":1,"b":0}`)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if got := decode(t, w)["error"]; got != tt.wantError {
				t.Errorf("error = %v, want %q", got, tt.wantError)
			}
		})
	}
}

func TestCalculateRejectsBadBodies(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{"malformed JSON", `{"operation":`, "request body must be valid JSON"},
		{"empty body", ``, "request body must be valid JSON"},
		{"string operand", `{"operation":"add","a":"ten","b":5}`, `field "a" must be a finite number`},
		{"out of range operand", `{"operation":"add","a":1,"b":1e400}`, `field "b" must be a finite number`},
		{"numeric operation", `{"operation":1,"a":1,"b":5}`, `field "operation" must be a string`},
		{"oversized body", `{"operation":"` + strings.Repeat("x", maxBodyBytes) + `"}`, "request body too large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := &fakeCalculator{}
			w := serve(t, calc, tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", w.Code)
			}
			if got := decode(t, w)["error"]; got != tt.wantError {
				t.Errorf("error = %v, want %q", got, tt.wantError)
			}
		})
	}
}
